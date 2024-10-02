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

func CalculateTfIdfVector(rateName string, tfidfData *TfIdfData) []float32 {
	preprocessed := strings.ToLower(stripAccents(rateName))
	ngrams := charNGrams(preprocessed, [2]int{1, 3})
	vector := make([]float32, len(tfidfData.Vocabulary))

	// Count term frequencies
	termFreq := make(map[string]int)
	for _, ngram := range ngrams {
		termFreq[ngram]++
	}

	// Compute TF-IDF
	for ngram, tf := range termFreq {
		if index, exists := tfidfData.Vocabulary[ngram]; exists && index >= 0 && index < int32(len(vector)) {
			vector[index] = float32(tf) * tfidfData.IdfValues[index]
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
	vectors := make([][]float32, len(rateNames))
	numWorkers := runtime.NumCPU()
	jobs := make(chan int, len(rateNames))
	var wg sync.WaitGroup

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := range jobs {
				vectors[j] = CalculateTfIdfVector(rateNames[j], tfidfData)
			}
		}()
	}

	for i := range rateNames {
		jobs <- i
	}
	close(jobs)

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
