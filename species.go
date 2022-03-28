package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/xuri/excelize/v2"
)

func ParseSpecies(fn string, species *[]SpeciesDetail) error {
	var sp SpeciesDetail

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
			if y == 0 || y == 1 || len(row) < 3 || len(row[1]) < 1 {
				continue
			}

			rawid := row[1]
			if len(rawid) < 3 {
				// Ignore rows without an ID
				continue
			}

			id, err := strconv.Atoi(strings.Trim(rawid, "[]"))
			if err != nil {
				return fmt.Errorf(
					"Atoi %s %s row %d: %s", fn, sheet, y+1, err.Error(),
				)
			}

			if id != sp.ID {
				if sp.ID != 0 {
					*species = append(*species, sp)
				}
				sp = SpeciesDetail{ID: id}
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

				if sp.Name != "" {
					fmt.Fprintf(os.Stderr,
						"%s %s row %d more than one published species name\n",
						fn, sheet, y+1,
					)
					continue
				}
				sp.Name = b.String()
			}

			if sp.Name == "" && sp.ID > 0 {
				fmt.Fprintf(os.Stderr,
					"%s %s row %d id with no proceeding species name\n",
					fn, sheet, y+1,
				)
				continue
			}
		}
	}

	return nil
}
