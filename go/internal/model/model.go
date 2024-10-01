package model

import (
	"encoding/json"
	"fmt"
	"math"
	"os"

	cb "github.com/mirecl/catboost-cgo/catboost"

	"github.com/go-goal/tagger/internal/tfidf"
)

type CatBoostModel struct {
	model *cb.Model
}

func LoadCatBoostModel(modelPath string) (*CatBoostModel, error) {
	model, err := cb.LoadFullModelFromFile(modelPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load model: %v", err)
	}
	return &CatBoostModel{model: model}, nil
}

func (m *CatBoostModel) Predict(rateName string, tfidfData tfidf.TfIdfData, labels []string) (string, error) {
	vector := tfidf.CalculateTfIdfVector(rateName, tfidfData)
	res, err := m.model.Predict([][]float32{vector}, [][]string{})
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
