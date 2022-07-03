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
		laddr := cmd.String("listen", "127.0.0.1:8080", "start listening on address")
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

		var taxons []map[string][]interface{}
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

				yr_f := bleve.NewNumericFieldMapping()
				yr_f.Store = false
				yr_f.Index = true
				yr_f.IncludeInAll = false
				imap.DefaultMapping.AddFieldMappingsAt("year", yr_f)

				ath_f := bleve.NewTextFieldMapping()
				ath_f.Store = false
				ath_f.Index = true
				ath_f.IncludeInAll = false
				imap.DefaultMapping.AddFieldMappingsAt("author", ath_f)

				index, err = bleve.New("taxon.bleve", imap)
			}
			if err != nil {
				fmt.Fprintf(os.Stderr, "bleve.New(): %s\n", err.Error())
				os.Exit(1)
			}
			defer index.Close()

			batch := index.NewBatch()
			for _, tx := range taxons {
				id := tx["ID"][0].(string)
				delete(tx, "ID")
				batch.Index(id, tx)
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

		var genera []map[string][]interface{}
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

				yr_f := bleve.NewNumericFieldMapping()
				yr_f.Store = false
				yr_f.Index = true
				yr_f.IncludeInAll = false
				imap.DefaultMapping.AddFieldMappingsAt("year", yr_f)

				spc_f := bleve.NewTextFieldMapping()
				spc_f.Store = false
				spc_f.Index = true
				spc_f.IncludeInAll = false
				imap.DefaultMapping.AddFieldMappingsAt("species", spc_f)

				ath_f := bleve.NewTextFieldMapping()
				ath_f.Store = false
				ath_f.Index = true
				ath_f.IncludeInAll = false
				imap.DefaultMapping.AddFieldMappingsAt("author", ath_f)

				genus_f := bleve.NewTextFieldMapping()
				genus_f.Store = false
				genus_f.Index = true
				genus_f.IncludeInAll = false
				imap.DefaultMapping.AddFieldMappingsAt("genus", genus_f)

				spcyr_f := bleve.NewNumericFieldMapping()
				spcyr_f.Store = false
				spcyr_f.Index = true
				spcyr_f.IncludeInAll = false
				imap.DefaultMapping.AddFieldMappingsAt("speciesyear", spc_f)

				spcath_f := bleve.NewTextFieldMapping()
				spcath_f.Store = false
				spcath_f.Index = true
				spcath_f.IncludeInAll = false
				imap.DefaultMapping.AddFieldMappingsAt("speciesauthor", spc_f)

				index, err = bleve.New("genera.bleve", imap)
			}
			if err != nil {
				fmt.Fprintf(os.Stderr, "bleve.New(): %s\n", err.Error())
				os.Exit(1)
			}
			defer index.Close()

			batch := index.NewBatch()
			for _, gr := range genera {
				id := gr["ID"][0].(string)
				delete(gr, "ID")
				batch.Index(id, gr)
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
	fmt.Printf("       start (start web server for app)\n")
}
