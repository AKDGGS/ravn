package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/xuri/excelize/v2"
)

func ParseSpecies(fn string, species *[]*SpeciesDetail) error {
	var sp *SpeciesDetail

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
			var id int
			var name string

			if y == 0 || y == 1 || len(row) < 3 || len(row[1]) < 1 {
				continue
			}

			rawid := row[1]
			if len(rawid) < 3 {
				// Ignore rows without an ID
				continue
			}

			id, err = strconv.Atoi(strings.Trim(rawid, "[]"))
			if err != nil {
				return fmt.Errorf(
					"Atoi %s %s row %d: %s", fn, sheet, y+1, err.Error(),
				)
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

			if len(row[2]) > 0 {
				rt, err := f.GetCellRichText(sheet, n)
				if err != nil {
					return fmt.Errorf(
						"GetCellRichText %s/%d: %s", sheet, y, err.Error(),
					)
				}

				b := strings.Builder{}
				for _, r := range rt {
					if (r.Font != nil && r.Font.Italic) || (r.Font == nil && cellItalic) {
						b.WriteString(r.Text)
					}
				}

				name = b.String()
			}

			// Row is a continuation of the previous row's ID
			if sp != nil && sp.ID == id {
				// If this row contains a species name and there's already
				// one specified.
				if name != "" && sp.Name != "" {
					fmt.Fprintf(os.Stderr,
						"%s %s row %d more than one published species name\n",
						fn, sheet, y+1,
					)
				}
			}

			// Row is a new ID, and there's a previous row to compare against
			if sp != nil && sp.ID != id {
				// If it's a new ID, and there's no species name, try
				// calculating Levenshtein distance. If it's "close enough"
				// just assume it's the same ID
				if name == "" {
					distance := Levenshtein(strconv.Itoa(id), strconv.Itoa(sp.ID))
					if distance <= 2 {
						fmt.Fprintf(os.Stderr,
							"%s %s row %d assuming id %d == %d\n",
							fn, sheet, y+1, id, sp.ID,
						)
						id = sp.ID
					}
				}
			}

			// Save this row as the last species
			if sp == nil || sp.ID != id {
				if name == "" {
					fmt.Fprintf(os.Stderr,
						"%s %s row %d new id with no species name\n",
						fn, sheet, y+1,
					)
				}

				sp = &SpeciesDetail{ID: id, Name: name}
				*species = append(*species, sp)
			}
		}
	}

	return nil
}
