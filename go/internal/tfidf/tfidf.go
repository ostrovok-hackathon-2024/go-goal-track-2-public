package tfidf

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type TfIdfData struct {
	Vocabulary map[string]int `json:"vocabulary"`
	IdfValues  []float64      `json:"idf_values"`
}

func LoadTfIdfData(filePath string) (TfIdfData, error) {
	data := TfIdfData{}
	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		return data, fmt.Errorf("failed to read TF-IDF file: %v", err)
	}
	err = json.Unmarshal(fileContent, &data)
	if err != nil {
		return data, fmt.Errorf("failed to unmarshal TF-IDF data: %v", err)
	}
	return data, nil
}

func TermFrequency(term string, document string) float64 {
	words := strings.Fields(document)
	termCount := 0
	for _, word := range words {
		if word == term {
			termCount++
		}
	}
	return float64(termCount) / float64(len(words))
}

func CalculateTfIdfVector(rateName string, tfidfData TfIdfData) []float32 {
	vector := make([]float32, len(tfidfData.Vocabulary))
	words := strings.Fields(strings.ToLower(rateName))

	for _, word := range words {
		if index, exists := tfidfData.Vocabulary[word]; exists {
			tf := float64(1) // Term frequency is always 1 for binary representation
			idf := tfidfData.IdfValues[index]
			vector[index] = float32(tf * idf)
		}
	}

	return vector
}
