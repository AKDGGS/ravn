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
	Reworked  bool   `yaml:",omitempty"`
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

		var years []Year
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
			year := Year{Year: yr, Ref: line[yidx[i+2]:yidx[i+3]]}
			years = append(years, year)
		}

		var reworked bool
		line = Flags_rx.ReplaceAllStringFunc(line, func(m string) string {
			for _, c := range m {
				switch c {
				case 'r', 'R':
					reworked = true
					return ""
				}
			}
			fmt.Fprintf(os.Stderr, "%s line %d unknown flag\n", fn, ln)
			return ""
		})

		tx := &TaxonReference{
			Origin:  fmt.Sprintf("%s line %d", fn, ln),
			Authors: strings.TrimRight(line[:yidx[0]], " "),
			Reference: strings.TrimRight(
				strings.TrimLeft(line[yidx[1]:], " ,.:"), " ",
			),
			Years:    years,
			Reworked: reworked,
		}

		*refs = append(*refs, tx)
	}
	return nil
}
