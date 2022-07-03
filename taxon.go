package main

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"
)

func ParseTaxonReference(fn string, taxons *[]map[string][]interface{}) error {
	f, err := os.Open(fn)
	if err != nil {
		return err
	}

	scan := bufio.NewScanner(f)
	scan.Split(bufio.ScanLines)

	for ln := 1; scan.Scan(); ln++ {
		tx := make(map[string][]interface{}, 0)

		line := scan.Text()
		yidx := YearAB_rx.FindStringSubmatchIndex(line)
		if len(yidx) < 1 {
			fmt.Fprintf(os.Stderr,
				"%s line %d missing year\n",
				fn, ln,
			)
			continue
		}

		tx["source"] = append(tx["source"], line)

		for i := 2; i < len(yidx); i += 4 {
			yr, _ := strconv.Atoi(line[yidx[i]:yidx[i+1]])
			if yr > 2022 || yr < 1800 {
				fmt.Fprintf(os.Stderr,
					"%s line %d invalid year (%d)\n",
					fn, ln, yr,
				)
				continue
			}
			tx["year"] = append(tx["year"], yr)
		}

		tx["author"] = append(tx["author"], strings.TrimSpace(line[:yidx[0]]))
		tx["ID"] = append(tx["ID"], fmt.Sprintf("%s/%d", path.Base(fn), ln))
		*taxons = append(*taxons, tx)
	}
	return nil
}
