package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
)

func parseCSV(walletsFile string) []string {
	f, err := os.Open(walletsFile)
	if err != nil {
		fmt.Println(err)
	}

	var pvtKeys []string

	csvReader := csv.NewReader(f)
	for {
		rec, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Println(err)
		}
		if len(rec[1]) > 60 {
			pvtKeys = append(pvtKeys, rec[1])
		}
	}
	f.Close()

	return pvtKeys
}
