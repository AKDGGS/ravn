package main

import (
	"flag"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/blevesearch/bleve/v2"
	"gopkg.in/yaml.v2"
)

var Year_rx *regexp.Regexp = regexp.MustCompile(`\d{4}`)
var YearAB_rx *regexp.Regexp = regexp.MustCompile(`(?:(\d{4})([a-z]{0,1})[;, ]*)+`)
var Flags_rx *regexp.Regexp = regexp.MustCompile(` {0,1}\(\*{0,1}[TtRrNn]\)`)
var Nupper_rx *regexp.Regexp = regexp.MustCompile(`([A-Z]+)`)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "error: action missing\n")
		PrintHelp()
		os.Exit(1)
	}

	switch strings.ToLower(os.Args[1]) {
	case "start":
		cmd := flag.NewFlagSet("start", flag.ExitOnError)
		cmd.SetOutput(os.Stdout)
		cmd.Usage = func() {
			fmt.Fprintf(cmd.Output(),
				"usage: %s %s [options] [filename ...]\n", os.Args[0], cmd.Name())
			fmt.Fprintf(cmd.Output(), "options:\n")
			cmd.PrintDefaults()
		}
		laddr := cmd.String("listen", "127.0.0.1:0", "start listening on address")
		cmd.Parse(os.Args[2:])

		sidx, err := bleve.Open("species.bleve")
		if err != nil {
			fmt.Fprintf(os.Stderr, "bleve.Open(): %s\n", err.Error())
			os.Exit(1)
		}
		tidx, err := bleve.Open("taxon.bleve")
		if err != nil {
			fmt.Fprintf(os.Stderr, "bleve.Open(): %s\n", err.Error())
			os.Exit(1)
		}
		gidx, err := bleve.Open("genera.bleve")
		if err != nil {
			fmt.Fprintf(os.Stderr, "bleve.Open(): %s\n", err.Error())
			os.Exit(1)
		}

		srv := WebServer{
			ListenAddress: *laddr, SpeciesIndex: sidx,
			TaxonIndex: tidx, GeneraIndex: gidx,
		}
		if err = srv.Start(); err != nil {
			fmt.Fprintf(os.Stderr, "server.Start(): %s\n", err.Error())
			os.Exit(1)
		}

	case "taxon":
		cmd := flag.NewFlagSet("taxon", flag.ExitOnError)
		cmd.SetOutput(os.Stdout)
		cmd.Usage = func() {
			fmt.Fprintf(cmd.Output(),
				"usage: %s %s [options] [filename ...]\n", os.Args[0], cmd.Name())
			fmt.Fprintf(cmd.Output(), "options:\n")
			cmd.PrintDefaults()
		}
		output_yaml := cmd.Bool("yaml", false, "output yaml")

		cmd.Parse(os.Args[2:])

		if cmd.NArg() < 1 {
			fmt.Fprintf(os.Stderr, "error: no files specified\n")
			cmd.Usage()
			os.Exit(1)
		}

		var taxons []*TaxonReference
		for _, fn := range cmd.Args() {
			err := ParseTaxonReference(fn, &taxons)
			if err != nil {
				fmt.Fprintf(os.Stderr, "ParseTaxonReference(): %s\n", err.Error())
				os.Exit(1)
			}
		}

		if *output_yaml {
			b, err := yaml.Marshal(taxons)
			if err != nil {
				fmt.Fprintf(os.Stderr, "yaml.Marshal(): %s\n", err.Error())
				os.Exit(1)
			}
			fmt.Printf("%s", string(b))
		} else {
			index, err := bleve.Open("taxon.bleve")
			if err == bleve.ErrorIndexPathDoesNotExist {
				imap := bleve.NewIndexMapping()
				line_f := bleve.NewNumericFieldMapping()
				line_f.Store = false
				line_f.Index = false
				file_f := bleve.NewTextFieldMapping()
				file_f.Store = false
				file_f.Index = false
				imap.DefaultMapping.AddFieldMappingsAt("Line", line_f)
				imap.DefaultMapping.AddFieldMappingsAt("File", file_f)
				index, err = bleve.New("taxon.bleve", imap)
			}
			if err != nil {
				fmt.Fprintf(os.Stderr, "bleve.New(): %s\n", err.Error())
				os.Exit(1)
			}
			defer index.Close()

			batch := index.NewBatch()
			for _, tx := range taxons {
				batch.Index(fmt.Sprintf("%s-%d", tx.File, tx.Line), tx)
				if batch.Size() > 1000 {
					err := index.Batch(batch)
					if err != nil {
						fmt.Fprintf(os.Stderr, "bleve.Batch(): %s\n", err.Error())
						os.Exit(1)
					}
					batch = index.NewBatch()
				}
			}
			if batch.Size() > 0 {
				err := index.Batch(batch)
				if err != nil {
					fmt.Fprintf(os.Stderr, "bleve.Batch(): %s\n", err.Error())
					os.Exit(1)
				}
			}
		}

	case "species":
		cmd := flag.NewFlagSet("species", flag.ExitOnError)
		cmd.SetOutput(os.Stdout)
		cmd.Usage = func() {
			fmt.Fprintf(cmd.Output(),
				"usage: %s %s [options] [filename ...]\n", os.Args[0], cmd.Name())
			fmt.Fprintf(cmd.Output(), "options:\n")
			cmd.PrintDefaults()
		}

		output_yaml := cmd.Bool("yaml", false, "output yaml")
		cmd.Parse(os.Args[2:])

		if cmd.NArg() < 1 {
			fmt.Fprintf(os.Stderr, "error: no files specified\n")
			cmd.Usage()
			os.Exit(1)
		}

		var species []*SpeciesDetail
		for _, fn := range cmd.Args() {
			err := ParseSpecies(fn, &species)
			if err != nil {
				fmt.Fprintf(os.Stderr, "ParseSpecies(): %s\n", err.Error())
				os.Exit(1)
			}
		}

		if *output_yaml {
			b, err := yaml.Marshal(species)
			if err != nil {
				fmt.Fprintf(os.Stderr, "yaml.Marshal(): %s\n", err.Error())
				os.Exit(1)
			}
			fmt.Printf("%s", string(b))
		} else {
			index, err := bleve.Open("species.bleve")
			if err == bleve.ErrorIndexPathDoesNotExist {
				imap := bleve.NewIndexMapping()
				id_f := bleve.NewNumericFieldMapping()
				id_f.Store = false
				id_f.Index = false
				imap.DefaultMapping.AddFieldMappingsAt("ID", id_f)
				index, err = bleve.New("species.bleve", imap)
			}
			if err != nil {
				fmt.Fprintf(os.Stderr, "bleve.New(): %s\n", err.Error())
				os.Exit(1)
			}
			defer index.Close()

			batch := index.NewBatch()
			for _, sp := range species {
				batch.Index(fmt.Sprintf("%d", sp.ID), sp)
				if batch.Size() > 1000 {
					err := index.Batch(batch)
					if err != nil {
						fmt.Fprintf(os.Stderr, "bleve.Batch(): %s\n", err.Error())
						os.Exit(1)
					}
					batch = index.NewBatch()
				}
			}
			if batch.Size() > 0 {
				err := index.Batch(batch)
				if err != nil {
					fmt.Fprintf(os.Stderr, "bleve.Batch(): %s\n", err.Error())
					os.Exit(1)
				}
			}
		}

	case "genera":
		cmd := flag.NewFlagSet("species", flag.ExitOnError)
		cmd.SetOutput(os.Stdout)
		cmd.Usage = func() {
			fmt.Fprintf(cmd.Output(),
				"usage: %s %s [options] [filename ...]\n", os.Args[0], cmd.Name())
			fmt.Fprintf(cmd.Output(), "options:\n")
			cmd.PrintDefaults()
		}
		output_yaml := cmd.Bool("yaml", false, "output yaml")

		cmd.Parse(os.Args[2:])

		if cmd.NArg() < 1 {
			fmt.Fprintf(os.Stderr, "error: no files specified\n")
			cmd.Usage()
			os.Exit(1)
		}

		var genera []*GenusDetail
		for _, fn := range cmd.Args() {
			err := ParseGenera(fn, &genera)
			if err != nil {
				fmt.Fprintf(os.Stderr, "ParseGenera(%s): %s\n", fn, err.Error())
			}
		}

		if *output_yaml {
			b, err := yaml.Marshal(genera)
			if err != nil {
				fmt.Fprintf(os.Stderr, "yaml.Marshal(): %s\n", err.Error())
				os.Exit(1)
			}
			fmt.Printf("%s", string(b))
		} else {
			index, err := bleve.Open("genera.bleve")
			if err == bleve.ErrorIndexPathDoesNotExist {
				imap := bleve.NewIndexMapping()
				line_f := bleve.NewNumericFieldMapping()
				line_f.Store = false
				line_f.Index = false
				file_f := bleve.NewTextFieldMapping()
				file_f.Store = false
				file_f.Index = false
				sheet_f := bleve.NewTextFieldMapping()
				sheet_f.Store = false
				sheet_f.Index = false
				imap.DefaultMapping.AddFieldMappingsAt("Line", line_f)
				imap.DefaultMapping.AddFieldMappingsAt("File", file_f)
				imap.DefaultMapping.AddFieldMappingsAt("Sheet", sheet_f)
				index, err = bleve.New("genera.bleve", imap)
			}
			if err != nil {
				fmt.Fprintf(os.Stderr, "bleve.New(): %s\n", err.Error())
				os.Exit(1)
			}
			defer index.Close()

			batch := index.NewBatch()
			for _, gr := range genera {
				batch.Index(fmt.Sprintf("%s-%s-%d", gr.File, gr.Sheet, gr.Line), gr)
				if batch.Size() > 1000 {
					err := index.Batch(batch)
					if err != nil {
						fmt.Fprintf(os.Stderr, "bleve.Batch(): %s\n", err.Error())
						os.Exit(1)
					}
					batch = index.NewBatch()
				}
			}
			if batch.Size() > 0 {
				err := index.Batch(batch)
				if err != nil {
					fmt.Fprintf(os.Stderr, "bleve.Batch(): %s\n", err.Error())
					os.Exit(1)
				}
			}
		}

	default:
		fmt.Fprintf(os.Stderr, "invalid action: %s\n", os.Args[1])
		PrintHelp()
		os.Exit(1)
	}
}

func PrintHelp() {
	fmt.Printf("usage: %s [action] [options] [file ...]\n", os.Args[0])
	fmt.Printf("available actions:\n")
	fmt.Printf("       species (parse species files)\n")
	fmt.Printf("       genera (parse genera files)\n")
	fmt.Printf("       taxon (parse taxon reference files)\n")
}
