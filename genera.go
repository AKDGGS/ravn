package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/xuri/excelize/v2"
)

type GenusDetail struct {
	Name      string
	Author    string `yaml:",omitempty"`
	Years     []Year `yaml:",omitempty"`
	Reference string `yaml:",omitempty"`

	/*
		AltNames []GenusAltName `yaml:",omitempty"`
		Species  []GenusSpecies `yaml:",omitempty"`
		Comments []string       `yaml:",omitempty"`
	*/
}

/*
type GenusSpecies struct {
	Name         string `yaml:",omitempty"`
	Year         int    `yaml:",omitempty"`
	Author       string `yaml:",omitempty"`
	Reference    string `yaml:",omitempty"`
	DefinesGenus bool   `yaml:",omitempty"`
}

type GenusAltName struct {
	Name      string `yaml:",omitempty"`
	Author    string `yaml:",omitempty"`
	Year      int    `yaml:",flow,omitempty"`
	Reference string `yaml:",omitempty"`
}
*/

type Year struct {
	Year int    `yaml:",omitempty"`
	Ref  string `yaml:",omitempty"`
}

func ParseGenera(fn string, genera *[]*GenusDetail) error {
	var gd *GenusDetail

	f, err := excelize.OpenFile(fn)
	if err != nil {
		return err
	}

	for _, sheet := range f.GetSheetList() {
		rows, err := f.GetRows(sheet)
		if err != nil {
			return err
		}

		for y, row := range rows {
			// Skip the first two rows
			if y < 2 {
				continue
			}

			switch {
			// Column B - species definition
			case len(row[1]) > 0:
				var name string
				var author string
				var reference string
				var years []Year

				colv := row[1]

				nidx := Nupper_rx.FindStringIndex(colv)
				if len(nidx) < 1 {
					fmt.Fprintf(os.Stderr,
						"%s %s row %d no uppercase species name\n",
						fn, sheet, y+1,
					)
				}
				name = colv[nidx[0]:nidx[1]]
				if len(colv) > (nidx[1] + 1) {
					colv = colv[nidx[1]+1:]

					spidx := strings.Index(colv, ";")
					if spidx >= 0 && len(colv) > (spidx+1) {
						reference = strings.Trim(colv[spidx+1:], " ")
						colv = colv[:spidx]
					}

					yidx := YearAB_rx.FindAllStringSubmatchIndex(colv, -1)
					if len(yidx) > 0 {
						for _, v := range yidx {
							year, _ := strconv.Atoi(colv[v[2]:v[3]])
							if year > 2022 || year < 1800 {
								fmt.Fprintf(os.Stderr,
									"%s %s row %d invalid year (%d), missing semicolon?\n",
									fn, sheet, y+1, year,
								)
							} else {
								years = append(years, Year{Year: year, Ref: colv[v[4]:v[5]]})
							}
						}
						colv = colv[:yidx[0][0]]
					}

					author = strings.Trim(colv, " ,")
				}

				gd = &GenusDetail{
					Name: name, Reference: reference, Years: years, Author: author,
				}
				*genera = append(*genera, gd)
			}
		}
	}
	return nil
}
