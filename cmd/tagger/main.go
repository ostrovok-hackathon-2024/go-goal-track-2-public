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

	cb "github.com/go-goal/tagger/cmd/tagger/catboost"

	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

/*
#cgo LDFLAGS: -ldl
#cgo CFLAGS: -O3 -g
#include <dlfcn.h>
#include "catboost/catboost_wrapper.h"
*/
import "C"

const Eps float64 = 0.00000001

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

// New function to filter terms based on min_df and max_df
func _filterTerms(terms []string, minDf int, maxDf float64, documentCount int) []string {
	termCount := make(map[string]int)
	for _, term := range terms {
		termCount[term]++
	}

	var filteredTerms []string
	for term, count := range termCount {
		if count >= minDf && float64(count)/float64(documentCount) <= maxDf {
			filteredTerms = append(filteredTerms, term)
		}
	}
	return filteredTerms
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

func CalculateTfIdfVector(rateName string, tfidfData TfIdfData, minDf int, maxDf float64) []float32 {
	preprocessed := strings.ToLower(rateName)
	preprocessed = stripAccents(preprocessed)

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

func CalculateTfIdfVectors(rateNames []string, tfidfData TfIdfData, minDf int, maxDf float64) [][]float32 {
	vectors := make([][]float32, len(rateNames))

	for i, rateName := range rateNames {
		vectors[i] = CalculateTfIdfVector(rateName, tfidfData, minDf, maxDf)
	}

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

func LoadModelAndPredict(floatsC unsafe.Pointer, num int, modelPath string, tfidfData TfIdfData, labels []string) ([]string, error) {
	model, err := cb.LoadFullModelFromFile(modelPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load model: %v", err)
	}

	res, err := model.Predict(floatsC, num)
	if err != nil {
		return nil, fmt.Errorf("failed to predict: %v", err)
	}

	predictions := make([]string, num)

	if len(labels) == 2 {
		for i, logit := range res {
			probability := 1.0 / (1.0 + math.Exp(-float64(logit)))
			if probability >= 0.5 {
				predictions[i] = labels[1]
			} else {
				predictions[i] = labels[0]
			}
		}
	} else {
		numClasses := len(labels)
		for i := 0; i < num; i++ {
			logits := res[i*numClasses : (i+1)*numClasses]
			probs := softmax(logits)
			predictions[i] = labels[argmax(probs)]
		}
	}

	return predictions, nil
}

func softmax(logits []float64) []float64 {
	maxLogit := logits[0]
	for _, logit := range logits[1:] {
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
		log.Fatal("Unable to read input file "+filePath, err)
	}
	defer f.Close()

	csvReader := csv.NewReader(f)
	records, err := csvReader.ReadAll()
	if err != nil {
		log.Fatal("Unable to parse file as CSV for "+filePath, err)
	}

	ans := []string{}

	for _, record := range records {
		ans = append(ans, record[0])
	}

	return ans
}

func argmax(values []float64) int {
	maxIndex := 0
	maxValue := math.Inf(-1)
	for i, v := range values {
		if v-maxValue > Eps {
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

	m := sync.Mutex{}

	results := map[string]map[string]string{}

	rateNames := ReadRateNames("rates_clean.csv")
	if err != nil {
		fmt.Printf("Error loading rate names: %v\n", err)
		return
	}

	floats := CalculateTfIdfVectors(rateNames, tfidfData, 2, 0.95)

	floatsC := cb.MakeFloatArray2D(floats)
	defer C.free(unsafe.Pointer(floatsC))

	catsC := cb.MakeCharArray2D([][]string{{""}})
	defer C.freeCharArray2D((***C.char)(unsafe.Pointer(catsC)), C.int(1), C.int(1))

	for _, rateName := range rateNames {
		a := map[string]string{}

		for _, label := range labels {
			a[label] = ""
		}

		results[rateName] = a
	}

	var wg sync.WaitGroup
	wg.Add(len(labels))

	for _, label := range labels {
		go func(label string) {
			defer wg.Done()
			modelPath := fmt.Sprintf("data/cbm/catboost_model_%v.cbm", label)
			labels, err := LoadLabels(fmt.Sprintf("data/labels/labels_%v.json", label))

			if err != nil {
				fmt.Printf("Error loading class labels: %v\n", err)
				return
			}

			predicted, err := LoadModelAndPredict(unsafe.Pointer(floatsC), len(floats), modelPath, tfidfData, labels)
			if err != nil {
				fmt.Printf("Error predicting class: %v\n", err)
				return
			}

			fmt.Printf("Predicting for %v\n", label)
			m.Lock()
			for i, pred := range predicted {
				results[rateNames[i]][label] = pred
			}
			m.Unlock()
		}(label)
	}

	wg.Wait()

	resultJSON, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		fmt.Printf("Error marshalling results: %v\n", err)
		return
	}

	fmt.Println(string(resultJSON)[len(resultJSON)-1488:])
}
