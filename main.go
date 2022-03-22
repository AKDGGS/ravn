package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/xuri/excelize/v2"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "no filename specified\n")
		os.Exit(1)
	}

	fn := os.Args[1]
	f, err := excelize.OpenFile(fn)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		return
	}

	for _, sheet := range f.GetSheetList() {
		fmt.Printf("Found sheet: %s\n", sheet)

		rows, err := f.GetRows(sheet)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err)
			return
		}

		last_id := 0
		for y, row := range rows {
			if y == 0 || y == 1 || len(row) < 3 || len(row[1]) < 1 {
				continue
			}

			rawid := row[1]
			if len(rawid) < 3 || !strings.HasPrefix(rawid, "[") || !strings.HasSuffix(rawid, "]") {
				// Ignore rows without an ID
				continue
			}

			id, err := strconv.Atoi(strings.TrimSuffix(strings.TrimPrefix(rawid, "["), "]"))
			if err != nil {
				fmt.Fprintf(os.Stderr, "Atoi %s/%d: %s", sheet, y, err.Error())
				return
			}

			n, err := excelize.CoordinatesToCellName(3, y+1)
			if err != nil {
				fmt.Fprintf(os.Stderr, "CoordinatesToCellName %s/%d: %s\n", sheet, y, err.Error())
				return
			}

			sid, err := f.GetCellStyle(sheet, n)
			if err != nil {
				fmt.Fprintf(os.Stderr, "CoordinatesToCellName %s/%d: %s\n", sheet, y, err.Error())
				return
			}

			fid := *f.Styles.CellXfs.Xf[sid].FontID
			cellItalic := f.Styles.Fonts.Font[fid].I != nil

			var species string
			if len(row[2]) > 0 {
				rt, err := f.GetCellRichText(sheet, n)
				if err != nil {
					fmt.Fprintf(os.Stderr, "GetCellRichText %s/%d: %s\n", sheet, y, err.Error())
					return
				}

				b := strings.Builder{}
				for _, r := range rt {
					if (r.Font != nil && r.Font.Italic) || (r.Font == nil && cellItalic) {
						b.WriteString(r.Text)
					}
				}
				species = b.String()
			}

			fmt.Printf("%s %d: \"%s\"\n", n, id, species)

			if last_id != id {
				if last_id > 0 && species == "" {
					fmt.Printf("%s row %d new id without species\n", fn, y+1)
				}
				last_id = id
			}
		}
	}
}
