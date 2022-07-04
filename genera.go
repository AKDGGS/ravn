package main

import (
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/xuri/excelize/v2"
)

func ParseGenera(fn string, genera *[]map[string]interface{}) error {
	var gd map[string]interface{}

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
			case len(row) > 1 && len(row[1]) > 0:
				gd = make(map[string]interface{})
				colv := strings.TrimSpace(row[1])

				nidx := Nupper_rx.FindStringIndex(colv)
				if len(nidx) < 1 {
					fmt.Fprintf(os.Stderr,
						"%s %s row %d no uppercase genus name\n",
						fn, sheet, y+1,
					)
					continue
				}

				gd["ID"] = fmt.Sprintf("%s.%d", path.Base(fn)[0:1], y+1)
				gd["source"] = colv
				gd["genus"] = colv[nidx[0]:nidx[1]]

				if len(colv) > (nidx[1] + 1) {
					colv = colv[nidx[1]+1:]

					yidx := YearAB_rx.FindStringSubmatchIndex(colv)
					if len(yidx) < 1 {
						sidx := strings.Index(colv, ";")
						if sidx == -1 {
							fmt.Fprintf(os.Stderr,
								"%s line %d missing both year and semicolon\n",
								fn, y+1,
							)
							return nil
						}
						appendMap(gd, "author", strings.Trim(colv[0:sidx], ", "))
					} else {
						appendMap(gd, "author", strings.Trim(colv[0:yidx[0]], ", "))

						yr, _ := strconv.Atoi(colv[yidx[2]:yidx[3]])
						if yr > 2022 || yr < 1800 {
							fmt.Fprintf(os.Stderr,
								"%s line %d invalid year (%d)\n",
								fn, y+1, yr,
							)
						} else {
							appendMap(gd, "year", yr)
						}
					}
				}
				*genera = append(*genera, gd)

			// Column C - Alt Names
			case len(row) > 2 && len(row[2]) > 0:
				if _, ok := gd["ID"]; !ok {
					fmt.Fprintf(os.Stderr,
						"%s %s row %d alt name before genus definition\n",
						fn, sheet, y+1,
					)
					continue
				}

				appendMap(gd, "alt_source", strings.TrimSpace(row[2]))

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
					break
				}

				name := strings.Trim(nb.String(), " ")
				if name == "" {
					fmt.Fprintf(os.Stderr,
						"%s %s row %d alt genus name does not begin with italics\n",
						fn, sheet, y+1,
					)
					continue
				}
				appendMap(gd, "genus", name)

				if len(row[2]) > nb.Len() {
					colv := row[2][nb.Len():]

					yidx := YearAB_rx.FindStringSubmatchIndex(colv)
					if len(yidx) < 1 {
						sidx := strings.Index(colv, ";")
						if sidx == -1 {
							fmt.Fprintf(os.Stderr,
								"%s line %d missing both year and semicolon\n",
								fn, y+1,
							)
							return nil
						}
						appendMap(gd, "author", strings.Trim(colv[0:sidx], ", "))
					} else {
						appendMap(gd, "author", strings.Trim(colv[0:yidx[0]], ", "))

						yr, _ := strconv.Atoi(colv[yidx[2]:yidx[3]])
						if yr > 2022 || yr < 1800 {
							fmt.Fprintf(os.Stderr,
								"%s line %d invalid year (%d)\n",
								fn, y+1, yr,
							)
						} else {
							appendMap(gd, "year", yr)
						}
					}
				}

			// Column D - Comments
			case len(row) > 3 && len(row[3]) > 0:
				if _, ok := gd["ID"]; !ok {
					fmt.Fprintf(os.Stderr,
						"%s %s row %d comment before genus definition\n",
						fn, sheet, y+1,
					)
					continue
				}
				appendMap(gd, "comment", strings.Trim(row[3], "<> "))

			// Column E - Species
			case len(row) > 4 && len(row[4]) > 0:
				if _, ok := gd["ID"]; !ok {
					fmt.Fprintf(os.Stderr,
						"%s %s row %d species before genus definition\n",
						fn, sheet, y+1,
					)
					continue
				}
				appendMap(gd, "species_source", strings.TrimSpace(row[4]))

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
					break
				}
				name := strings.Trim(nb.String(), " ")
				if len(name) < 1 {
					fmt.Fprintf(os.Stderr,
						"%s %s row %d missing italicized species name at start\n",
						fn, sheet, y+1,
					)
					continue
				}
				appendMap(gd, "species", name)

				colv := row[4][nb.Len():]

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
							appendMap(gd, "speciesyear", year)
						}
					}
					colv = colv[:yidx[0][0]]
				}
				appendMap(gd, "speciesauthor", strings.Trim(colv, ", "))
			}
		}
	}
	return nil
}
