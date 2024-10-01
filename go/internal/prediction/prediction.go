package prediction

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/go-goal/tagger/internal/model"
	"github.com/go-goal/tagger/internal/tfidf"
)

type Predictor struct {
	tfidfData   tfidf.TfIdfData
	classLabels map[string][]string
	models      map[string]*model.CatBoostModel
}

func NewPredictor(tfidfPath, labelsDir, modelsDir string) (*Predictor, error) {
	p := &Predictor{
		classLabels: make(map[string][]string),
		models:      make(map[string]*model.CatBoostModel),
	}

	// Load TF-IDF data
	tfidfData, err := tfidf.LoadTfIdfData(tfidfPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load TF-IDF data: %w", err)
	}
	p.tfidfData = tfidfData

	// Load all label files
	labelFiles, err := filepath.Glob(filepath.Join(labelsDir, "*.json"))
	if err != nil {
		return nil, fmt.Errorf("failed to find label files: %w", err)
	}
	for _, labelFile := range labelFiles {
		labelName := strings.TrimPrefix(strings.TrimSuffix(filepath.Base(labelFile), ".json"), "labels_")
		labels, err := model.LoadLabels(labelFile)
		if err != nil {
			return nil, fmt.Errorf("failed to load labels from %s: %w", labelFile, err)
		}
		p.classLabels[labelName] = labels
	}

	// Load all CatBoost model files
	modelFiles, err := filepath.Glob(filepath.Join(modelsDir, "*.cbm"))
	if err != nil {
		return nil, fmt.Errorf("failed to find model files: %w", err)
	}
	for _, modelFile := range modelFiles {
		modelName := strings.TrimSuffix(filepath.Base(modelFile), ".cbm")
		m, err := model.LoadCatBoostModel(modelFile)
		if err != nil {
			return nil, fmt.Errorf("failed to load CatBoost model from %s: %w", modelFile, err)
		}
		p.models[modelName] = m
	}

	return p, nil
}
func (p *Predictor) Predict(input string) (map[string]string, error) {
	if len(p.models) == 0 {
		return nil, fmt.Errorf("no models loaded")
	}
	results := make(map[string]string)
	for modelName, m := range p.models {
		labelKey := strings.TrimPrefix(strings.TrimSuffix(modelName, ".cbm"), "catboost_model_")
		labels, ok := p.classLabels[labelKey]
		if !ok {
			return nil, fmt.Errorf("no labels found for model %s (label key: %s)", modelName, labelKey)
		}
		result, err := m.Predict(input, p.tfidfData, labels)
		if err != nil {
			return nil, err
		}
		results[modelName] = result
	}
	return results, nil
}
