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
	Year      Year   `yaml:",flow,omitempty"`
	Reference string `yaml:",omitempty"`
	Origin    string `yaml:",omitempty"`
}

func ParseTaxonReference(fn string, refs *[]*TaxonReference) error {
	f, err := os.Open(fn)
	if err != nil {
		return err
	}

	scan := bufio.NewScanner(f)
	scan.Split(bufio.ScanLines)

	for ln := 1; scan.Scan(); ln++ {
		line := scan.Text()
		yidx := YearAB_rx.FindStringSubmatchIndex(line)
		if len(yidx) < 1 {
			fmt.Fprintf(os.Stderr,
				"%s line %d missing year\n",
				fn, ln,
			)
			continue
		}

		year, _ := strconv.Atoi(line[yidx[2]:yidx[3]])
		if year > 2022 || year < 1800 {
			fmt.Fprintf(os.Stderr,
				"%s line %d invalid year (%d)\n",
				fn, ln, year,
			)
		}

		tx := &TaxonReference{
			Origin:  fmt.Sprintf("%s line %d", fn, ln),
			Authors: strings.TrimRight(line[:yidx[0]], " "),
			Reference: strings.TrimRight(
				strings.TrimLeft(line[yidx[1]:], " ,.:"), " ",
			),
			Year: Year{Year: year, Ref: line[yidx[5]:yidx[5]]},
		}

		*refs = append(*refs, tx)
	}
	return nil
}
