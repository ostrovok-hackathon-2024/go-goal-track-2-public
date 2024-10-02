package tfidf

import (
	"encoding/json"
	"fmt"
	"math"
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
			tf := float32(1 + math.Log(float64(count)))
			vector[index] = tf * tfidfData.IdfValues[index]
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
