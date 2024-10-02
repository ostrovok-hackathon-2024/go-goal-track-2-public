package tfidf

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"strings"
	"sync"
	"unicode"

	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

type TfIdfData struct {
	Vocabulary map[string]int32 `json:"vocabulary"`
	IdfValues  []float32        `json:"idf_values"`
}

func LoadTfIdfData(filePath string) (TfIdfData, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return TfIdfData{}, fmt.Errorf("error opening file: %v", err)
	}
	defer file.Close()

	var data TfIdfData
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&data); err != nil {
		return TfIdfData{}, fmt.Errorf("error decoding JSON: %v", err)
	}

	return data, nil
}

func CalculateTfIdfVector(rateName string, tfidfData *TfIdfData) []float32 {
	vector := make([]float32, len(tfidfData.IdfValues))
	rateName = stripAccents(strings.ToLower(rateName))
	ngrams := charNGrams(rateName, 3)

	for ngram, count := range ngrams {
		if idx, ok := tfidfData.Vocabulary[ngram]; ok {
			vector[idx] = float32(count) * tfidfData.IdfValues[idx]
		}
	}

	return vector
}

func CalculateTfIdfVectors(rateNames []string, tfidfData *TfIdfData) [][]float32 {
	numWorkers := runtime.NumCPU()
	vectorChan := make(chan []float32, len(rateNames))
	var wg sync.WaitGroup

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for j := workerID; j < len(rateNames); j += numWorkers {
				vector := CalculateTfIdfVector(rateNames[j], tfidfData)
				vectorChan <- vector
			}
		}(i)
	}

	go func() {
		wg.Wait()
		close(vectorChan)
	}()

	vectors := make([][]float32, 0, len(rateNames))
	for vector := range vectorChan {
		vectors = append(vectors, vector)
	}

	return vectors
}

func stripAccents(s string) string {
	t := transform.Chain(norm.NFD, transform.RemoveFunc(func(r rune) bool {
		return unicode.Is(unicode.Mn, r)
	}), norm.NFC)
	result, _, _ := transform.String(t, s)
	return result
}

func charNGrams(s string, n int) map[string]int {
	ngrams := make(map[string]int)
	runes := []rune(s)
	for i := 0; i <= len(runes)-n; i++ {
		ngram := string(runes[i : i+n])
		ngrams[ngram]++
	}
	return ngrams
}
