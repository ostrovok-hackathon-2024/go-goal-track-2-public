package utils

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
)

func ReadRateNames(filePath string) []string {
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Printf("Error opening file: %v\n", err)
		return nil
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		fmt.Printf("Error reading CSV: %v\n", err)
		return nil
	}

	rateNames := make([]string, len(records))
	for i, record := range records {
		rateNames[i] = record[0]
	}

	return rateNames
}

func LoadLabels(filePath string) ([]string, error) {
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

func WriteResultsToCSV(outputFile string, results map[string]map[string]string, rateNames []string, labels []string) error {
	file, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("error creating output file: %v", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	header := append([]string{"rate_name"}, labels...)
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("error writing header to CSV: %v", err)
	}

	for _, rateName := range rateNames {
		row := make([]string, 1+len(labels))
		row[0] = rateName
		for i, label := range labels {
			row[i+1] = results[rateName][label]
		}
		if err := writer.Write(row); err != nil {
			return fmt.Errorf("error writing record to CSV: %v", err)
		}
	}

	return nil
}
