package main

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"
)

func ParseReferences(fn string, refs *[]map[string][]interface{}) error {
	f, err := os.Open(fn)
	if err != nil {
		return err
	}

	scan := bufio.NewScanner(f)
	scan.Split(bufio.ScanLines)

	for ln := 1; scan.Scan(); ln++ {
		ref := make(map[string][]interface{}, 0)

		line := scan.Text()
		yidx := YearAB_rx.FindStringSubmatchIndex(line)
		if len(yidx) < 1 {
			fmt.Fprintf(os.Stderr,
				"%s line %d missing year\n",
				fn, ln,
			)
			continue
		}

		ref["source"] = append(ref["source"], line)

		for i := 2; i < len(yidx); i += 4 {
			yr, _ := strconv.Atoi(line[yidx[i]:yidx[i+1]])
			if yr > 2022 || yr < 1800 {
				fmt.Fprintf(os.Stderr,
					"%s line %d invalid year (%d)\n",
					fn, ln, yr,
				)
				continue
			}
			ref["year"] = append(ref["year"], yr)
		}

		ref["author"] = append(ref["author"], strings.TrimSpace(line[:yidx[0]]))
		ref["ID"] = append(ref["ID"], fmt.Sprintf("%s/%d",
			strings.TrimSuffix(path.Base(fn), path.Ext(fn)), ln,
		))
		*refs = append(*refs, ref)
	}
	return nil
}
