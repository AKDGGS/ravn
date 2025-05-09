package main

import (
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
)

const (
	START = iota
	FOUND
	PASSED
)

func ParseSpecies(fn, idir string, species *[]map[string]interface{}) error {
	filelist := make(map[string][]string, 0)
	if idir != "" {
		ents, err := os.ReadDir(idir)
		if err != nil {
			return err
		}
		for _, ent := range ents {
			name := ent.Name()
			ext := strings.ToLower(path.Ext(name))
			if !ent.IsDir() && (ext == ".jpg" || ext == ".png" || ext == ".jpeg") {
				i := strings.Index(name, " ")
				if i >= 0 {
					filelist[name[:i]] = append(filelist[name[:i]], name)
				}
			}
		}
	}

	var sp map[string]interface{}

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

			if len(row[1]) < 3 {
				// Ignore rows without an ID
				continue
			}

			id := strings.Trim(row[1], "[]")

			switch {
			// Column C - species details
			case len(row) > 2 && len(row[2]) > 0:
				if oid, ok := sp["ID"].(string); ok {
					if oid == id {
						fmt.Fprintf(os.Stderr,
							"%s %s row %d more than one published species name\n",
							fn, sheet, y+1,
						)
						continue
					}
				}

				sp = make(map[string]interface{})
				sp["ID"] = id
				sp["source"] = strings.TrimSpace(row[2])
				if imgs, ok := filelist[id]; ok {
					sp["images"] = imgs
				}
				if err := parse_species(f, fn, sheet, 3, y+1, &sp); err != nil {
					return err
				}
				*species = append(*species, sp)

			// Column D - species alt names
			case len(row) > 3 && len(row[3]) > 0:
				// If the new id does not match the last id
				if oid, ok := sp["ID"].(string); ok && oid != id {
					// Check the distance from this id to the last one and
					// assume it's the last id if they're close enough
					if Levenshtein(id, oid) > 2 {
						fmt.Fprintf(os.Stderr,
							"%s %s row %d new id with no species name\n",
							fn, sheet, y+1,
						)
						continue
					} else {
						fmt.Fprintf(os.Stderr,
							"%s %s row %d assuming id %s == %s\n",
							fn, sheet, y+1, id, oid,
						)
						id = oid
					}
				}
				appendMap(sp, "alt_source", strings.TrimSpace(row[3]))

				if err := parse_species(f, fn, sheet, 4, y+1, &sp); err != nil {
					return err
				}

			// Column E - occurances
			case len(row) > 4 && len(row[4]) > 0:
				appendMap(sp, "occurance", strings.TrimSpace(row[4]))

			// Column F - comments
			case len(row) > 5 && len(row[5]) > 0:
				appendMap(sp, "comment", strings.Trim(row[5], "<> "))

			default:
				continue
			}
		}
	}
	return nil
}

func parse_species(f *excelize.File, fn, sheet string, col, row int, ps *map[string]interface{}) error {
	sp := *ps

	n, err := excelize.CoordinatesToCellName(col, row)
	if err != nil {
		return fmt.Errorf(
			"CoordinatesToCellName %s %s row %d: %s",
			fn, sheet, row, err.Error(),
		)
	}

	sid, err := f.GetCellStyle(sheet, n)
	if err != nil {
		return fmt.Errorf(
			"CoordinatesToCellName %s %s row %d: %s",
			fn, sheet, row, err.Error(),
		)
	}

	fid := *f.Styles.CellXfs.Xf[sid].FontID
	cellItalic := f.Styles.Fonts.Font[fid].I != nil

	rt, err := f.GetCellRichText(sheet, n)
	if err != nil {
		return fmt.Errorf(
			"GetCellRichText %s %s row %d: %s",
			fn, sheet, row, err.Error(),
		)
	}

	nb := strings.Builder{}
	rb := strings.Builder{}
	state := START
	for _, r := range rt {
		italic := ((r.Font != nil && r.Font.Italic) || (r.Font == nil && cellItalic))
		switch state {
		case START:
			if italic {
				state = FOUND
			}
		case FOUND:
			if !italic {
				state = PASSED
			}
		}

		if state == FOUND || state == START {
			nb.WriteString(r.Text)
		} else {
			rb.WriteString(r.Text)
		}
	}
	if nb.Len() > 0 {
		appendMap(sp, "species", strings.TrimSpace(nb.String()))
	}

	if rb.Len() > 0 {
		re := rb.String()

		yidx := YearAB_rx.FindStringSubmatchIndex(re)
		if len(yidx) < 1 {
			sidx := strings.Index(re, ";")
			if sidx == -1 {
				fmt.Fprintf(os.Stderr,
					"%s line %d missing both year and semicolon\n",
					fn, row,
				)
				return nil
			}
			appendMap(sp, "author", strings.TrimSpace(re[0:sidx]))
		} else {
			appendMap(sp, "author", strings.TrimSpace(re[0:yidx[0]]))

			yr, _ := strconv.Atoi(re[yidx[2]:yidx[3]])
			if yr > time.Now().Year() || yr < 1800 {
				fmt.Fprintf(os.Stderr,
					"%s line %d invalid year (%d)\n",
					fn, row, yr,
				)
				return nil
			}
			appendMap(sp, "year", yr)
		}
	}
	return nil
}
