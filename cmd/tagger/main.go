package main

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"strings"

	cb "github.com/mirecl/catboost-cgo/catboost"
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
	for term, index := range tfidfData.Vocabulary {
		tf := TermFrequency(term, rateName)
		idf := tfidfData.IdfValues[index]
		vector[index] = float32(tf * idf)
	}
	return vector
}

func LoadModelAndPredict(rateName string, modelPath string, tfidfData TfIdfData, labels []string) (string, error) {
	model, err := cb.LoadFullModelFromFile(modelPath)

	if err != nil {
		return "", fmt.Errorf("failed to load model: %v", err)
	}

	vector := CalculateTfIdfVector(rateName, tfidfData)
	res, err := model.Predict([][]float32{vector}, [][]string{})

	if err != nil {
		return "", fmt.Errorf("failed to predict: %v", err)
	}

	predicted := labels[argmax(res)]

	return predicted, nil
}

func LoadLabels(filePath string) ([]string, error) {
	var labels []string

	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read labels file: %v", err)
	}

	err = json.Unmarshal(fileContent, &labels)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal labels: %v", err)
	}

	return labels, nil
}

func argmax(values []float64) int {
	maxIndex := 0
	maxValue := math.Inf(-1)
	for i, v := range values {
		if v > maxValue {
			maxValue = v
			maxIndex = i
		}
	}
	return maxIndex
}

func main() {
	tfidfData, err := LoadTfIdfData("tfidf_data.json")
	if err != nil {
		fmt.Printf("Error loading TF-IDF data: %v\n", err)
		return
	}

	rateName := "deluxe triple room"
	modelPath := "catboost_model_class.cbm"
	classLabels, err := LoadLabels("class_labels.json")

	if err != nil {
		fmt.Printf("Error loading class labels: %v\n", err)
		return
	}

	predictedClass, err := LoadModelAndPredict(rateName, modelPath, tfidfData, classLabels)
	if err != nil {
		fmt.Printf("Error predicting class: %v\n", err)
		return
	}

	results := map[string]string{
		"class": predictedClass,
	}

	resultJSON, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		fmt.Printf("Error marshalling results: %v\n", err)
		return
	}

	fmt.Println(string(resultJSON))
}
