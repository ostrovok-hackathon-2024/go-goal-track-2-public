package model

import (
	"C"
	"fmt"
	"math"
	"path/filepath"
	"sync"
	"unsafe"

	cb "github.com/go-goal/tagger/internal/catboost"
	"github.com/go-goal/tagger/pkg/utils"
)

const Eps float64 = 1e-8

func LoadModelAndPredict(model cb.Model, floatsC unsafe.Pointer, num int) ([]float64, error) {
	return model.Predict(floatsC, num)
}

func PredictAll(floats [][]float32, namesToPredict []string, labels []string, modelsDir, labelsDir string) (map[string]map[string]string, error) {
	floatsC := cb.MakeFloatArray2D(floats)

	results := make(map[string]map[string]string, len(namesToPredict))
	for _, rateName := range namesToPredict {
		results[rateName] = make(map[string]string, len(labels))
		for _, label := range labels {
			results[rateName][label] = ""
		}
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	var loadingErr error

	for _, label := range labels {
		wg.Add(1)
		go func(label string) {
			defer wg.Done()

			modelPath := filepath.Join(modelsDir, fmt.Sprintf("catboost_model_%s.cbm", label))
			model, err := cb.LoadFullModelFromFile(modelPath)
			if err != nil {
				mu.Lock()
				loadingErr = fmt.Errorf("error loading model for %s: %v", label, err)
				mu.Unlock()
				return
			}

			labelsPath := filepath.Join(labelsDir, fmt.Sprintf("labels_%s.json", label))
			loadedLabels, err := utils.ReadJsonStringArray(labelsPath)
			if err != nil {
				mu.Lock()
				loadingErr = fmt.Errorf("error loading labels for %s: %v", label, err)
				mu.Unlock()
				return
			}

			predicted, err := LoadModelAndPredict(*model, unsafe.Pointer(floatsC), len(floats))
			if err != nil {
				mu.Lock()
				loadingErr = fmt.Errorf("error predicting for %s: %v", label, err)
				mu.Unlock()
				return
			}

			predictions := makePredictions(predicted, loadedLabels, len(namesToPredict))

			mu.Lock()
			for i, pred := range predictions {
				results[namesToPredict[i]][label] = pred
			}
			mu.Unlock()
		}(label)
	}

	wg.Wait()

	if loadingErr != nil {
		return nil, loadingErr
	}

	return results, nil
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

func argmax(arr []float64) int {
	maxIdx := 0
	maxVal := arr[0]
	for i, val := range arr[1:] {
		if val > maxVal {
			maxIdx = i + 1
			maxVal = val
		}
	}
	return maxIdx
}

func makePredictions(predicted []float64, labels []string, numRateNames int) []string {
	predictions := make([]string, numRateNames)
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
		for i := 0; i < numRateNames; i++ {
			start := i * numClasses
			end := start + numClasses
			if end > len(predicted) {
				predictions[i] = ""
				continue
			}
			logits := predicted[start:end]
			probs := softmax(logits)
			predictions[i] = labels[argmax(probs)]
		}
	}
	return predictions
}
