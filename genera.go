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
	Author    string         `yaml:",omitempty"`
	Years     []Year         `yaml:",omitempty"`
	Reference string         `yaml:",omitempty"`
	AltNames  []GenusAltName `yaml:",omitempty"`
	Comments  []string       `yaml:",omitempty"`
	Species   []GenusSpecies `yaml:",omitempty"`
}

type GenusAltName struct {
	Name      string `yaml:",omitempty"`
	Author    string `yaml:",omitempty"`
	Years     []Year `yaml:",omitempty"`
	Reference string `yaml:",omitempty"`
}

type GenusSpecies struct {
	Name         string `yaml:",omitempty"`
	Years        []Year `yaml:",flow,omitempty"`
	Author       string `yaml:",omitempty"`
	DefinesGenus bool   `yaml:",omitempty"`
}

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
			// Column B - Genus definition
			case len(row[1]) > 0:
				var name string
				var author string
				var reference string
				var years []Year

				colv := row[1]

				nidx := Nupper_rx.FindStringIndex(colv)
				if len(nidx) < 1 {
					fmt.Fprintf(os.Stderr,
						"%s %s row %d no uppercase genus name\n",
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

			// Column C - Alt Names
			case len(row[2]) > 0:
				var name string
				var author string
				var reference string
				var years []Year

				if gd == nil || gd.Name == "" {
					fmt.Fprintf(os.Stderr,
						"%s %s row %d alt name before genus definition\n",
						fn, sheet, y+1,
					)
					continue
				}

				n, err := excelize.CoordinatesToCellName(3, y+1)
				if err != nil {
					return fmt.Errorf(
						"CoordinatesToCellName %s %s row %d: %s",
						fn, sheet, y+1, err.Error(),
					)
				}

				sid, err := f.GetCellStyle(sheet, n)
				if err != nil {
					return fmt.Errorf(
						"CoordinatesToCellName %s %s row %d: %s",
						fn, sheet, y+1, err.Error(),
					)
				}

				fid := *f.Styles.CellXfs.Xf[sid].FontID
				cellItalic := f.Styles.Fonts.Font[fid].I != nil

				rt, err := f.GetCellRichText(sheet, n)
				if err != nil {
					return fmt.Errorf(
						"GetCellRichText %s %s row %d: %s",
						fn, sheet, y+1, err.Error(),
					)
				}

				nb := strings.Builder{}
				for _, r := range rt {
					if (r.Font != nil && r.Font.Italic) || (r.Font == nil && cellItalic) {
						nb.WriteString(r.Text)
						continue
					}
					break;
				}

				name = strings.Trim(nb.String(), " ")
				if name == "" {
					fmt.Fprintf(os.Stderr,
						"%s %s row %d alt genus name does not begin with italics\n",
						fn, sheet, y+1,
					)
					continue
				}

				if len(row[2]) > nb.Len() {
					colv := row[2][nb.Len():]
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

				gd.AltNames = append(gd.AltNames, GenusAltName{
					Name: name, Author: author, Reference: reference,
					Years: years,
				})

			// Column D - Comments
			case len(row[3]) > 0:
				if gd == nil || gd.Name == "" {
					fmt.Fprintf(os.Stderr,
						"%s %s row %d comment before genus definition\n",
						fn, sheet, y+1,
					)
					continue
				}

				gd.Comments = append(gd.Comments, strings.Trim(row[3], "<> "))

			// Column E - Species
			case len(row[4]) > 0:
				var name string
				var author string
				var definesgenus bool
				var years []Year

				if gd == nil || gd.Name == "" {
					fmt.Fprintf(os.Stderr,
						"%s %s row %d species before genus definition\n",
						fn, sheet, y+1,
					)
					continue
				}

				n, err := excelize.CoordinatesToCellName(5, y+1)
				if err != nil {
					return fmt.Errorf(
						"CoordinatesToCellName %s %s row %d: %s",
						fn, sheet, y+1, err.Error(),
					)
				}

				sid, err := f.GetCellStyle(sheet, n)
				if err != nil {
					return fmt.Errorf(
						"CoordinatesToCellName %s %s row %d: %s",
						fn, sheet, y+1, err.Error(),
					)
				}

				fid := *f.Styles.CellXfs.Xf[sid].FontID
				cellItalic := f.Styles.Fonts.Font[fid].I != nil

				rt, err := f.GetCellRichText(sheet, n)
				if err != nil {
					return fmt.Errorf(
						"GetCellRichText %s %s row %d: %s",
						fn, sheet, y+1, err.Error(),
					)
				}

				if len(rt) < 1 && cellItalic {
					fmt.Fprintf(os.Stderr,
						"%s %s row %d entire cell italic\n",
						fn, sheet, y+1,
					)
					continue
				}

				nb := strings.Builder{}
				for _, r := range rt {
					if (r.Font != nil && r.Font.Italic) || (r.Font == nil && cellItalic) {
						nb.WriteString(r.Text)
						continue
					}
					break;
				}
				name = strings.Trim(nb.String(), " ")

				if len(name) < 1 {
					fmt.Fprintf(os.Stderr,
						"%s %s row %d missing italicized species name at start\n",
						fn, sheet, y+1,
					)
					continue
				}

				colv := row[4][nb.Len():]

				// Detect if this species defines the genus
				if nr := Dfgen_rx.ReplaceAllString(colv, ""); len(nr) != len(colv) {
					definesgenus = true
					colv = nr
				}

				yidx := YearAB_rx.FindAllStringSubmatchIndex(colv, -1)
				if len(yidx) > 0 {
					for _, v := range yidx {
						year, _ := strconv.Atoi(colv[v[2]:v[3]])
						if year > 2022 || year < 1800 {
							fmt.Fprintf(os.Stderr,
								"%s %s row %d invalid year (%d)\n",
								fn, sheet, y+1, year,
							)
						} else {
							years = append(years, Year{Year: year, Ref: colv[v[4]:v[5]]})
						}
					}
					colv = colv[:yidx[0][0]]
				}

				author = strings.Trim(colv, " ")

				gd.Species = append(gd.Species, GenusSpecies{
					Name: name, Author: author, Years: years,
					DefinesGenus: definesgenus,
				})
			}
		}
	}
	return nil
}
