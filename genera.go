package main

import (
	"fmt"

	"github.com/xuri/excelize/v2"
)

type GenusDetail struct {
	Name     string
	Author   string         `yaml:",omitempty"`
	Year     int            `yaml:",omitempty"`
	AltNames []GenusAltName `yaml:",omitempty"`
	Species  []GenusSpecies `yaml:",omitempty"`
	Comments []string       `yaml:",omitempty"`
}

type GenusSpecies struct {
	Name         string `yaml:",omitempty"`
	Year         int    `yaml:",omitempty"`
	Author       string `yaml:",omitempty"`
	Reference    string `yaml:",omitempty"`
	DefinesGenus bool   `yaml:",omitempty"`
}

type GenusAltName struct {
	Name      string `yaml:",omitempty"`
	Author    string `yaml:",omitempty"`
	Year      int    `yaml:",flow,omitempty"`
	Reference string `yaml:",omitempty"`
}

func ParseGenera(fn string, genera *[]*GenusDetail) error {
	//var gd *GenusDetail

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
			if y == 0 || y == 1 {
				continue
			}

			switch {
			case len(row[1]) > 0:
				var name string

				name = Nupper_rx.FindString(row[1])
				fmt.Printf("%d [%s]\n", y+1, name)
			}
		}
	}
	return nil
}
