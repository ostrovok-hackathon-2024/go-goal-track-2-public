package utils

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
)

// ReadFirstCSVColumn reads a CSV file and returns a slice of strings for a specified column
func ReadFirstCSVColumn(filePath string, columnName string) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("error opening file: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("error reading CSV: %v", err)
	}

	if len(records) == 0 {
		return nil, fmt.Errorf("no records found in CSV")
	}

	if len(records[0]) == 0 {
		return nil, fmt.Errorf("provide a valid CSV file")
	}

	// Find the index of the specified column
	columnIndex := -1
	for i, header := range records[0] {
		if header == columnName {
			columnIndex = i
			break
		}
	}

	if columnIndex == -1 {
		return nil, fmt.Errorf("column '%s' not found in CSV", columnName)
	}

	columnRecords := make([]string, len(records)-1)
	for i, record := range records[1:] {
		if len(record) > columnIndex {
			columnRecords[i] = record[columnIndex]
		} else {
			return nil, fmt.Errorf("record %d does not have enough columns", i+1)
		}
	}

	return columnRecords, nil
}

// ReadJsonStringArray loads a JSON file and returns a slice of strings
func ReadJsonStringArray(filePath string) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("error opening file: %v", err)
	}
	defer file.Close()

	var labels []string
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&labels); err != nil {
		return nil, fmt.Errorf("error decoding JSON: %v", err)
	}

	return labels, nil
}

// WriteCSV writes a CSV file with the provided headers, first column data, and row data.
//
// rowData is a map of the form map[string]map[string]string
// where the keys of the outer map are the first column data and the keys of the inner map are the headers
func WriteCSV(outputFile string, headers []string, firstColData []string, rowData map[string]map[string]string) error {
	file, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("error creating output file: %v", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()


	writer.Write(headers)

	for _, inputValue := range firstColData {
		row := make([]string, len(headers))
		row[0] = inputValue
		for j := 1; j < len(headers); j++ {
			row[j] = rowData[inputValue][headers[j]]
		}
		writer.Write(row)
	}

	return nil
}

// IsFile checks if a file exists and is not a directory
func IsFile(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
