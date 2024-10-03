package model

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sync"
	"unsafe"

	cb "github.com/go-goal/tagger/internal/catboost"
	"github.com/go-goal/tagger/internal/tfidf"
)

// #include <stdlib.h>
import "C"

const Eps float64 = 1e-8

type Predictor struct {
	TfidfData    *tfidf.TfIdfData
	ModelsDir    string
	LabelsDir    string
	Categories   []string
	loadedModels map[string]*cb.Model
	loadedLabels map[string][]string
}

func NewPredictor(tfidfData *tfidf.TfIdfData, modelsDir, labelsDir string, categories []string) *Predictor {
	return &Predictor{
		TfidfData:    tfidfData,
		ModelsDir:    modelsDir,
		LabelsDir:    labelsDir,
		Categories:   categories,
		loadedModels: make(map[string]*cb.Model),
		loadedLabels: make(map[string][]string),
	}
}

func (p *Predictor) LoadModels() error {
	var wg sync.WaitGroup
	errChan := make(chan error, len(p.Categories))
	var mu sync.Mutex

	for _, category := range p.Categories {
		wg.Add(1)
		go func(cat string) {
			defer wg.Done()
			modelPath := filepath.Join(p.ModelsDir, fmt.Sprintf("catboost_model_%s.cbm", cat))
			model, err := cb.LoadFullModelFromFile(modelPath)
			if err != nil {
				errChan <- fmt.Errorf("error loading model for %s: %v", cat, err)
				return
			}

			labelsPath := filepath.Join(p.LabelsDir, fmt.Sprintf("labels_%s.json", cat))
			labels, err := loadLabels(labelsPath)
			if err != nil {
				errChan <- fmt.Errorf("error loading labels for %s: %v", cat, err)
				return
			}

			mu.Lock()
			p.loadedModels[cat] = model
			p.loadedLabels[cat] = labels
			mu.Unlock()
		}(category)
	}

	wg.Wait()
	close(errChan)

	for err := range errChan {
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *Predictor) PredictAll(inputStrings []string) (map[string]map[string]string, error) {
	floats := tfidf.CalculateTfIdfVectors(inputStrings, p.TfidfData)

	results := make(map[string]map[string]string)
	for _, input := range inputStrings {
		results[input] = make(map[string]string)
	}

	var wg sync.WaitGroup
	errChan := make(chan error, len(p.Categories))
	
	// Create a mutex to protect concurrent writes to the results map
	var resultsMutex sync.Mutex

	for _, category := range p.Categories {
		wg.Add(1)
		go func(cat string) {
			defer wg.Done()
			model, exists := p.loadedModels[cat]
			if !exists {
				errChan <- fmt.Errorf("model not loaded for category: %s", cat)
				return
			}

			labels, exists := p.loadedLabels[cat]
			if !exists {
				errChan <- fmt.Errorf("labels not loaded for category: %s", cat)
				return
			}

			predictions, err := predictCategory(model, floats, labels)
			if err != nil {
				errChan <- fmt.Errorf("error predicting for %s: %v", cat, err)
				return
			}

			// Use the mutex to safely write to the results map
			resultsMutex.Lock()
			for i, prediction := range predictions {
				results[inputStrings[i]][cat] = prediction
			}
			resultsMutex.Unlock()
		}(category)
	}

	wg.Wait()
	close(errChan)

	for err := range errChan {
		if err != nil {
			return nil, err
			}
	}

	return results, nil
}

func predictCategory(model *cb.Model, floats [][]float32, labels []string) ([]string, error) {
	floatsC := cb.MakeFloatArray2D(floats)
	defer C.free(unsafe.Pointer(floatsC))

	predicted, err := model.Predict(unsafe.Pointer(floatsC), len(floats))
	if err != nil {
		return nil, fmt.Errorf("error predicting: %v", err)
	}

	predictions := make([]string, len(floats))
	if len(labels) == 2 {
		for i, logit := range predicted {
			probability := 1.0 / (1.0 + math.Exp(-logit))
			if probability >= 0.5 {
				predictions[i] = labels[1]
			} else {
				predictions[i] = labels[0]
			}
		}
	} else {
		numClasses := len(labels)
		for i := 0; i < len(floats); i++ {
			start := i * numClasses
			end := start + numClasses
			if end > len(predicted) {
				return nil, fmt.Errorf("insufficient logits for input %v", i)
			}
			logits := predicted[start:end]
			probs := softmax(logits)
			predictions[i] = labels[argmax(probs)]
		}
	}

	return predictions, nil
}

func loadLabels(filePath string) ([]string, error) {
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
