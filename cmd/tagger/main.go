package main

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"strings"
	"unicode"

	cb "github.com/mirecl/catboost-cgo/catboost"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

type TfIdfData struct {
	Vocabulary map[string]int `json:"vocabulary"`
	IdfValues  []float64      `json:"idf_values"`
}

func stripAccents(input string) string {
	t := transform.Chain(norm.NFD, transform.RemoveFunc(func(r rune) bool {
		return unicode.Is(unicode.Mn, r)
	}), norm.NFC)
	output, _, _ := transform.String(t, input)
	return output
}

func charNGrams(input string, ngramRange [2]int) []string {
	ngrams := []string{}
	words := strings.Fields(input)
	for _, word := range words {
		runes := []rune(" " + word + " ")
		for n := ngramRange[0]; n <= ngramRange[1]; n++ {
			for i := 0; i < len(runes)-n+1; i++ {
				ngrams = append(ngrams, string(runes[i:i+n]))
			}
		}
	}
	return ngrams
}

func sublinearTermFrequency(term string, document string) float64 {
	count := 0
	ngrams := charNGrams(document, [2]int{1, 3})
	for _, ngram := range ngrams {
		if ngram == term {
			count++
		}
	}
	if count > 0 {
		return 1 + math.Log(float64(count))
	}
	return 0
}

func CalculateTfIdfVector(rateName string, tfidfData TfIdfData) []float32 {
	preprocessed := strings.ToLower(rateName)
	preprocessed = stripAccents(preprocessed)

	_ = charNGrams(preprocessed, [2]int{1, 3})
	vector := make([]float32, len(tfidfData.Vocabulary))

	for term, index := range tfidfData.Vocabulary {
		tf := sublinearTermFrequency(term, preprocessed)
		idf := tfidfData.IdfValues[index]
		vector[index] = float32(tf * idf)
	}

	norm := float32(0.0)
	for _, v := range vector {
		norm += v * v
	}
	norm = float32(math.Sqrt(float64(norm)))
	if norm > 0 {
		for i := range vector {
			vector[i] /= norm
		}
	}

	return vector
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

func LoadModelAndPredict(rateName string, modelPath string, tfidfData TfIdfData, labels []string) (string, error) {
	model, err := cb.LoadFullModelFromFile(modelPath)

	if err != nil {
		return "", fmt.Errorf("failed to load model: %v", err)
	}

	vector := CalculateTfIdfVector(rateName, tfidfData)
	fmt.Println(vector)
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
	tfidfData, err := LoadTfIdfData("data/tfidf/tfidf_data.json")
	if err != nil {
		fmt.Printf("Error loading TF-IDF data: %v\n", err)
		return
	}

	rateName := "King Premium Mountain View no balcony"
	modelPath := "data/cbm/catboost_model_quality.cbm"
	classLabels, err := LoadLabels("data/labels/labels_quality.json")

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
