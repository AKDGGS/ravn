package main

import (
	"fmt"
	"os"
	"strings"
)

type SpeciesDetail struct {
	ID           int
	Name         string
	DefinesGenus bool
	Year         int
	AltNames     []AltName
	Occurances   []string
}

type AltName struct {
	Name      string
	Reference string
}

type Occurance struct {
	Authors   string
	Year      int
	Reference string
}

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintf(os.Stderr, "action and/or filename(s) missing\n")
		PrintHelp()
		os.Exit(1)
	}

	switch strings.ToLower(os.Args[1]) {
	case "species":
		var species []SpeciesDetail
		for _, fn := range os.Args[2:] {
			err := ParseSpecies(fn, &species)
			if err != nil {
				fmt.Fprintf(os.Stderr, "ParseSpecies(): %s\n", err.Error())
				os.Exit(1)
			}
		}
		fmt.Printf("Parsed %d species\n", len(species))
	default:
		fmt.Fprintf(os.Stderr, "invalid action: %s\n", os.Args[1])
		PrintHelp()
		os.Exit(1)
	}
}

func PrintHelp() {
	fmt.Printf("usage: %s [action] [file ...]\n", os.Args[0])
	fmt.Printf("available actions:\n")
	fmt.Printf("       species (parse species files)\n")
}
