package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"os"
	"strings"
	"sync"
	"unicode"
	"unsafe"

	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"

	cb "github.com/go-goal/tagger/cmd/tagger/catboost"
)

/*
#cgo LDFLAGS: -ldl
#cgo CFLAGS: -O3 -g
#include <dlfcn.h>
#include "catboost/catboost_wrapper.h"
*/
import "C"

const Eps float64 = 1e-8

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
			if len(runes) < n {
				continue
			}
			for i := 0; i <= len(runes)-n; i++ {
				ngrams = append(ngrams, string(runes[i:i+n]))
			}
		}
	}
	return ngrams
}

// _filterTerms remains unchanged

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
	preprocessed := strings.ToLower(stripAccents(rateName))
	ngrams := charNGrams(preprocessed, [2]int{1, 3})

	termCounts := make(map[string]int, len(ngrams))
	for _, ngram := range ngrams {
		termCounts[ngram]++
	}

	vector := make([]float32, len(tfidfData.Vocabulary))

	// Compute TF-IDF
	for term, index := range tfidfData.Vocabulary {
		if count, exists := termCounts[term]; exists && count > 0 {
			vector[index] = float32((1 + math.Log(float64(count))) * tfidfData.IdfValues[index])
		} else {
			vector[index] = 0
		}
	}

	// Normalize the vector
	var normVal float32
	for _, v := range vector {
		normVal += v * v
	}
	normVal = float32(math.Sqrt(float64(normVal)))
	if normVal > 0 {
		for i := range vector {
			vector[i] /= normVal
		}
	}

	return vector
}

func CalculateTfIdfVectors(rateNames []string, tfidfData TfIdfData) [][]float32 {
	vectors := make([][]float32, len(rateNames))
	var wg sync.WaitGroup
	wg.Add(len(rateNames))

	// Limit concurrency
	workerPool := make(chan struct{}, 8) // Adjust based on CPU cores

	for i, rateName := range rateNames {
		go func(i int, rateName string) {
			defer wg.Done()
			workerPool <- struct{}{}
			vectors[i] = CalculateTfIdfVector(rateName, tfidfData)
			<-workerPool
		}(i, rateName)
	}

	wg.Wait()
	return vectors
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

func LoadModelAndPredict(model cb.Model, floatsC unsafe.Pointer, num int) ([]float64, error) {
	return model.Predict(floatsC, num)
}

func softmax(logits []float64) []float64 {
	maxLogit := math.Inf(-1)
	for _, logit := range logits {
		if logit > maxLogit {
			maxLogit = logit
		}
	}

	expSum := 0.0
	probs := make([]float64, len(logits))
	for i, logit := range logits {
		exp := math.Exp(logit - maxLogit)
		probs[i] = exp
		expSum += exp
	}

	for i := range probs {
		probs[i] /= expSum
	}

	return probs
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

func ReadRateNames(filePath string) []string {
	f, err := os.Open(filePath)
	if err != nil {
		log.Fatalf("Unable to read input file %s: %v", filePath, err)
	}
	defer f.Close()

	csvReader := csv.NewReader(f)
	records, err := csvReader.ReadAll()
	if err != nil {
		log.Fatalf("Unable to parse file as CSV for %s: %v", filePath, err)
	}

	ans := make([]string, 0, len(records))
	for _, record := range records {
		if len(record) > 0 {
			ans = append(ans, record[0])
		}
	}
	return ans
}

func argmax(values []float64) int {
	maxIndex := 0
	maxValue := math.Inf(-1)
	for i, v := range values {
		if v > maxValue+Eps {
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

	labels := []string{"balcony", "bathroom", "bedding", "bedrooms", "capacity", "class", "club", "floor", "quality", "view"}

	rateNames := ReadRateNames("rates_dirty.csv")[1:]

	floats := CalculateTfIdfVectors(rateNames, tfidfData)

	floatsC := cb.MakeFloatArray2D(floats)
	defer C.free(unsafe.Pointer(floatsC))

	catsC := cb.MakeCharArray2D([][]string{{""}})
	defer C.freeCharArray2D((***C.char)(unsafe.Pointer(catsC)), C.int(1), C.int(1))

	results := make(map[string]map[string]string, len(rateNames))
	for _, rateName := range rateNames {
		results[rateName] = make(map[string]string, len(labels))
		for _, label := range labels {
			results[rateName][label] = ""
		}
	}

	var wg sync.WaitGroup
	wg.Add(len(labels))

	fmt.Println("Starting predictions")

	// Preload all models and labels
	models := make(map[string]cb.Model, len(labels))
	labelsMap := make(map[string][]string, len(labels))
	for _, label := range labels {
		modelPath := fmt.Sprintf("data/cbm/catboost_model_%v.cbm", label)
		model, err := cb.LoadFullModelFromFile(modelPath)
		if err != nil {
			fmt.Printf("Error loading model for %v: %v\n", label, err)
			continue
		}
		models[label] = *model

		labelsPath := fmt.Sprintf("data/labels/labels_%v.json", label)
		loadedLabels, err := LoadLabels(labelsPath)
		if err != nil {
			fmt.Printf("Error loading labels for %v: %v\n", label, err)
			continue
		}
		labelsMap[label] = loadedLabels
	}

	var resultsMutex sync.Mutex

	for _, label := range labels {
		go func(label string) {
			defer wg.Done()
			model, exists := models[label]
			if !exists {
				return
			}
			classLabels, exists := labelsMap[label]
			if !exists {
				return
			}

			predicted, err := LoadModelAndPredict(model, unsafe.Pointer(floatsC), len(floats))
			if err != nil {
				fmt.Printf("Error predicting for label %v: %v\n", label, err)
				return
			}

			fmt.Printf("Predicting for %v\n", label)

			predictions := make([]string, len(rateNames))
			if len(classLabels) == 2 {
				for i, logit := range predicted {
					probability := 1.0 / (1.0 + math.Exp(-logit))
					if probability >= 0.5 {
						predictions[i] = classLabels[1]
					} else {
						predictions[i] = classLabels[0]
					}
				}
			} else {
				numClasses := len(classLabels)
				for i := 0; i < len(rateNames); i++ {
					start := i * numClasses
					end := start + numClasses
					if end > len(predicted) {
						fmt.Printf("Insufficient logits for rateName %v, label %v\n", rateNames[i], label)
						predictions[i] = ""
						continue
					}
					logits := predicted[start:end]
					probs := softmax(logits)
					predictions[i] = classLabels[argmax(probs)]
				}
			}

			resultsMutex.Lock()
			for i, pred := range predictions {
				results[rateNames[i]][label] = pred
			}
			resultsMutex.Unlock()
		}(label)
	}

	wg.Wait()

	// Write results to CSV
	outputFile, err := os.Create("predictions.csv")
	if err != nil {
		fmt.Printf("Error creating output file: %v\n", err)
		return
	}
	defer outputFile.Close()

	writer := csv.NewWriter(outputFile)
	defer writer.Flush()

	// Write header
	header := append([]string{"rate_name"}, labels...)
	if err := writer.Write(header); err != nil {
		fmt.Printf("Error writing header to CSV: %v\n", err)
		return
	}

	// Prepare data to write
	records := make([][]string, 0, len(rateNames))
	for _, rateName := range rateNames {
		row := make([]string, 1+len(labels))
		row[0] = rateName
		for i, label := range labels {
			row[i+1] = results[rateName][label]
		}
		records = append(records, row)
	}

	if err := writer.WriteAll(records); err != nil {
		fmt.Printf("Error writing records to CSV: %v\n", err)
		return
	}

	fmt.Println("Predictions written to predictions.csv")
}
