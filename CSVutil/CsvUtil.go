package CSVutil

import (
	"VAA_Uebung1/pkg/Exception"
	"encoding/csv"
	"io/ioutil"
	"os"
	"path/filepath"
)

func ReadCSVRows(csvName string) [][]string {
	file, err := os.Open(canon(csvName))
	Exception.ErrorHandler(err)

	reader := csv.NewReader(file)
	rows, err := reader.ReadAll()
	Exception.ErrorHandler(err)
	return rows
}

func ReadBytes(csvName string) []byte {
	bytes, err := ioutil.ReadFile(canon(csvName))
	Exception.ErrorHandler(err)
	return bytes
}
func canon(csvName string) string {
	absPath, err := filepath.Abs(csvName)
	Exception.ErrorHandler(err)
	return absPath
}
