package utils

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
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

// WriteOutput writes the output in the specified format
func WriteOutput(outputFile, format string, headers []string, firstColData []string, rowData map[string]map[string]string) error {
	switch strings.ToLower(format) {
	case "csv":
		return WriteCSV(outputFile, headers, firstColData, rowData)
	case "json":
		return WriteJSON(outputFile, headers, firstColData, rowData)
	case "tsv":
		return WriteTSV(outputFile, headers, firstColData, rowData)
	case "yaml":
		return WriteYAML(outputFile, headers, firstColData, rowData)
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}
}

// PrintCSV prints CSV data to the given writer
func PrintCSV(w io.Writer, headers []string, firstColData []string, rowData map[string]map[string]string) {
	writer := csv.NewWriter(w)
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
}

// PrintJSON prints JSON data to the given writer
func PrintJSON(w io.Writer, headers []string, firstColData []string, rowData map[string]map[string]string) {
	data := prepareOutputData(headers, firstColData, rowData)
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	encoder.Encode(data)
}

// PrintTSV prints TSV data to the given writer
func PrintTSV(w io.Writer, headers []string, firstColData []string, rowData map[string]map[string]string) {
	writer := csv.NewWriter(w)
	writer.Comma = '\t'
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
}

// PrintYAML prints YAML data to the given writer
func PrintYAML(w io.Writer, headers []string, firstColData []string, rowData map[string]map[string]string) {
	data := prepareOutputData(headers, firstColData, rowData)
	encoder := yaml.NewEncoder(w)
	encoder.Encode(data)
}

// WriteJSON writes JSON data to a file
func WriteJSON(outputFile string, headers []string, firstColData []string, rowData map[string]map[string]string) error {
	file, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("error creating output file: %v", err)
	}
	defer file.Close()

	data := prepareOutputData(headers, firstColData, rowData)
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

// WriteTSV writes TSV data to a file
func WriteTSV(outputFile string, headers []string, firstColData []string, rowData map[string]map[string]string) error {
	file, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("error creating output file: %v", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	writer.Comma = '\t'
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

// WriteYAML writes YAML data to a file
func WriteYAML(outputFile string, headers []string, firstColData []string, rowData map[string]map[string]string) error {
	file, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("error creating output file: %v", err)
	}
	defer file.Close()

	data := prepareOutputData(headers, firstColData, rowData)
	encoder := yaml.NewEncoder(file)
	return encoder.Encode(data)
}


// prepareOutputData prepares the data for JSON and YAML output
func prepareOutputData(headers []string, firstColData []string, rowData map[string]map[string]string) []map[string]string {
	var data []map[string]string
	for _, inputValue := range firstColData {
		row := make(map[string]string)
		row[headers[0]] = inputValue
		for j := 1; j < len(headers); j++ {
			row[headers[j]] = rowData[inputValue][headers[j]]
		}
		data = append(data, row)
	}
	return data
}
