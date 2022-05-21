package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type TaxonReference struct {
	Authors   string `yaml:",omitempty"`
	Years     []Year `yaml:",flow,omitempty"`
	Reference string `yaml:",omitempty"`
	Origin    string `yaml:",omitempty"`
	Reworked  bool
}

func ParseTaxonReference(fn string, refs *[]*TaxonReference) error {
	f, err := os.Open(fn)
	if err != nil {
		return err
	}

	scan := bufio.NewScanner(f)
	scan.Split(bufio.ScanLines)

	for ln := 1; scan.Scan(); ln++ {
		var years []Year

		line := scan.Text()
		yidx := YearAB_rx.FindStringSubmatchIndex(line)
		if len(yidx) < 1 {
			fmt.Fprintf(os.Stderr,
				"%s line %d missing year\n",
				fn, ln,
			)
			continue
		}

		for i := 2; i < len(yidx); i += 4 {
			yr, _ := strconv.Atoi(line[yidx[i]:yidx[i+1]])
			if yr > 2022 || yr < 1800 {
				fmt.Fprintf(os.Stderr,
					"%s line %d invalid year (%d)\n",
					fn, ln, yr,
				)
				continue
			}
			year := Year{ Year: yr, Ref: line[yidx[i+2]:yidx[i+3]] }
			years = append(years, year)
		}

		tx := &TaxonReference{
			Origin:  fmt.Sprintf("%s line %d", fn, ln),
			Authors: strings.TrimRight(line[:yidx[0]], " "),
			Reference: strings.TrimRight(
				strings.TrimLeft(line[yidx[1]:], " ,.:"), " ",
			),
			Years: years,
		}

		*refs = append(*refs, tx)
	}
	return nil
}
