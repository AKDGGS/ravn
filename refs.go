package main

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
)

func ParseReferences(fn string, refs *[]map[string]interface{}) error {
	f, err := os.Open(fn)
	if err != nil {
		return err
	}

	scan := bufio.NewScanner(f)
	scan.Split(bufio.ScanLines)

	for ln := 1; scan.Scan(); ln++ {
		line := scan.Text()
		yidx := YearAB_rx.FindStringSubmatchIndex(line)
		if len(yidx) < 1 {
			fmt.Fprintf(os.Stderr,
				"%s line %d missing year\n",
				fn, ln,
			)
			continue
		}

		yr, _ := strconv.Atoi(line[yidx[2]:yidx[3]])
		if yr > time.Now().Year() || yr < 1800 {
			fmt.Fprintf(os.Stderr,
				"%s line %d invalid year (%d)\n",
				fn, ln, yr,
			)
			continue
		}

		*refs = append(*refs, map[string]interface{}{
			"ID": fmt.Sprintf("%s.%d",
				strings.TrimSuffix(path.Base(fn), path.Ext(fn)), ln,
			),
			"year": yr, "source": line,
			"author": strings.TrimSpace(line[:yidx[0]]),
		})
	}
	return nil
}
