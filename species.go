package main

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/xuri/excelize/v2"
)

var Year_rx *regexp.Regexp = regexp.MustCompile(`\d{4}`)
var Dfgen_rx *regexp.Regexp = regexp.MustCompile(`\(\*{0,1}T\)`)

type SpeciesDetail struct {
	ID           int
	Name         string
	Origin       string             `yaml:",omitempty"`
	Author       string             `yaml:",omitempty"`
	Reference    string             `yaml:",omitempty"`
	DefinesGenus bool               `yaml:",omitempty"`
	Year         int                `yaml:",omitempty"`
	AltNames     []SpeciesAltName   `yaml:",omitempty"`
	Occurances   []SpeciesOccurance `yaml:",omitempty"`
	Comments     []string           `yaml:",omitempty"`
}

type SpeciesAltName struct {
	Name         string `yaml:",omitempty"`
	Year         int    `yaml:",omitempty"`
	Author       string `yaml:",omitempty"`
	Reference    string `yaml:",omitempty"`
	DefinesGenus bool   `yaml:",omitempty"`
}

type SpeciesOccurance struct {
	Author    string `yaml:",omitempty"`
	Years     []int  `yaml:",flow,omitempty"`
	Reference string `yaml:",omitempty"`
}

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
			var author string
			var reference string
			var definesgenus bool
			var year int
			var years []int
			var comment string

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

			curcol := 0
			if len(row[2]) > 0 {
				curcol = 2
			} else if len(row[3]) > 0 {
				curcol = 3
			} else if len(row[4]) > 0 {
				curcol = 4
			} else if len(row[5]) > 0 {
				curcol = 5
			}

			switch curcol {
			// Column C - species details
			// Column D - species alt names
			case 2, 3:
				n, err := excelize.CoordinatesToCellName(curcol+1, y+1)
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
				rb := strings.Builder{}
				pn := true
				for _, r := range rt {
					if pn == true && ((r.Font != nil && r.Font.Italic) || (r.Font == nil && cellItalic)) {
						nb.WriteString(r.Text)
					} else {
						pn = false
						rb.WriteString(r.Text)
					}
				}

				name = strings.Trim(nb.String(), " ")
				if len(name) < 1 {
					fmt.Fprintf(os.Stderr,
						"%s %s row %d missing italicized name at start\n",
						fn, sheet, y+1,
					)
					reference = row[curcol]
				} else {
					reference = strings.Trim(rb.String(), " ")
				}

				if nr := Dfgen_rx.ReplaceAllString(reference, ""); len(nr) != len(reference) {
					definesgenus = true
					reference = nr
				}

				if si := strings.Index(reference, ";"); si >= 1 {
					author = strings.Trim(reference[:si], " ,'")
					ai := Year_rx.FindStringIndex(author)
					if len(ai) > 0 {
						author = strings.Trim(author[:ai[0]], " ,")
					}

					if len(reference) > si+1 {
						reference = reference[si+1:]
					} else {
						reference = ""
					}

					reference = strings.Trim(reference, " ;,*▓¢.&'")
				} else {
					reference = strings.Trim(reference, " *▓¢.&'")
				}

				yrst := Year_rx.FindString(reference)
				if yrst != "" {
					year, _ = strconv.Atoi(yrst)
				}

			// Column E - species occurances
			case 4:
				firstyr := -1
				lastyr := -1

				vals := strings.Split(row[4], ",")
				for i, val := range vals {
					yrst := Year_rx.FindString(val)
					if yrst != "" {
						year, _ = strconv.Atoi(yrst)
						if year > 0 {
							years = append(years, year)

							if firstyr < 0 {
								firstyr = i
							}
							lastyr = i
						}
					}
				}

				if firstyr >= 0 {
					author = strings.Trim(strings.Join(vals[:firstyr], ","), " ")
				}

				if lastyr >= 0 && len(vals) > lastyr+1 {
					reference = strings.Trim(strings.Join(vals[lastyr+1:], ","), " ")
				}

				if lastyr < 0 && firstyr < 0 {
					fmt.Fprintf(os.Stderr,
						"%s %s row %d no year in occurance\n",
						fn, sheet, y+1,
					)
					reference = row[4]
				}

			// Column F - comments
			case 5:
				comment = strings.Trim(row[5], "<> ")

			default:
				continue
			}

			// Row is a new ID, and there's a previous row to compare against
			if sp != nil && sp.ID != id {
				// If it's a new ID, and there's no species name, try
				// calculating Levenshtein distance. If it's "close enough"
				// just assume it's the same ID
				if curcol != 2 {
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

			// Row is a continuation of the previous row's ID
			if sp != nil && sp.ID == id {
				// If this row contains a species name and there's already
				// one specified.
				if curcol == 2 && name != "" && sp.Name != "" {
					fmt.Fprintf(os.Stderr,
						"%s %s row %d more than one published species name\n",
						fn, sheet, y+1,
					)
				}

				switch curcol {
				case 3:
					alt := SpeciesAltName{
						Name: name, Author: author, Reference: reference,
						DefinesGenus: definesgenus, Year: year,
					}
					sp.AltNames = append(sp.AltNames, alt)
				case 4:
					occ := SpeciesOccurance{
						Author: author, Years: years, Reference: reference,
					}
					sp.Occurances = append(sp.Occurances, occ)
				case 5:
					sp.Comments = append(sp.Comments, comment)
				}
			}

			// Row is the first, or it's a new ID
			if sp == nil || sp.ID != id {
				if name == "" {
					fmt.Fprintf(os.Stderr,
						"%s %s row %d new id with no species name\n",
						fn, sheet, y+1,
					)
				}

				if curcol == 2 {
					sp = &SpeciesDetail{
						ID: id, Name: name, Origin: fn, Author: author,
						Reference: reference, DefinesGenus: definesgenus,
						Year: year,
					}

					*species = append(*species, sp)
				}
			}
		}
	}

	return nil
}
