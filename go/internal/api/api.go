package api

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"path/filepath"

	"github.com/gofiber/fiber/v2"

	"github.com/go-goal/tagger/internal/config"
	"github.com/go-goal/tagger/internal/model"
	"github.com/go-goal/tagger/internal/tfidf"
)

var (
	cfg       *config.Config
	tfidfData tfidf.TfIdfData
)

func init() {
	var err error
	cfg, err = config.LoadConfig("config.yaml")
	if err != nil {
		panic(fmt.Sprintf("Error loading config: %v", err))
	}

	tfidfFile := filepath.Join(cfg.ModelsDir, "tfidf", "tfidf_data.json")
	tfidfData, err = tfidf.LoadTfIdfData(tfidfFile)
	if err != nil {
		panic(fmt.Sprintf("Error loading TF-IDF data: %v", err))
	}
}

func SetupRoutes(app *fiber.App) {
	app.Post("/predict", predictRateNames)
	app.Post("/predict_csv", predictRateNamesCSV)
}

type RateNameInput struct {
	RateNames  []string `json:"rate_names"`
	InputCol   string   `json:"input_col"`
	Categories []string `json:"categories"`
}

func predictRateNames(c *fiber.Ctx) error {
	var input RateNameInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	// Clean rate names
	cleanedRateNames := make([]string, len(input.RateNames))
	for i, name := range input.RateNames {
		if name != "" {
			cleanedRateNames[i] = name
		}
	}

	// Calculate TF-IDF vectors
	floats := tfidf.CalculateTfIdfVectors(cleanedRateNames, &tfidfData)

	if len(input.Categories) == 0 {
		input.Categories = cfg.Categories
	}

	// Load models and make predictions
	cbmDir := filepath.Join(cfg.ModelsDir, "cbm")
	labelsDir := filepath.Join(cfg.ModelsDir, "labels/json")
	results, err := model.PredictAll(floats, cleanedRateNames, input.Categories, cbmDir, labelsDir)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(results)
}

func predictRateNamesCSV(c *fiber.Ctx) error {
	file, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "File upload failed"})
	}

	fileContent, err := file.Open()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to open file"})
	}
	defer fileContent.Close()

	csvReader := csv.NewReader(fileContent)
	records, err := csvReader.ReadAll()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Failed to parse CSV"})
	}

	if len(records) < 2 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "CSV file must contain at least a header row and one data row"})
	}

	headers := records[0]
	categories := headers[1:]
	rateNames := make([]string, len(records)-1)
	for i, record := range records[1:] {
		if len(record) > 0 {
			rateNames[i] = record[0]
		}
	}

	// Calculate TF-IDF vectors
	floats := tfidf.CalculateTfIdfVectors(rateNames, &tfidfData)

	// Load models and make predictions
	results, err := model.PredictAll(floats, rateNames, categories, filepath.Join(cfg.ModelsDir, "cbm"), filepath.Join(cfg.ModelsDir, "labels"))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	// Create CSV from predictions
	var buf bytes.Buffer
	csvWriter := csv.NewWriter(&buf)

	// Write header
	csvWriter.Write(headers)

	// Write data
	for _, rateName := range rateNames {
		predictions := results[rateName]
		row := make([]string, len(headers))
		row[0] = rateName
		for i, category := range categories {
			row[i+1] = predictions[category]
		}
		csvWriter.Write(row)
	}
	csvWriter.Flush()

	c.Set("Content-Type", "text/csv")
	c.Set("Content-Disposition", "attachment; filename=predictions.csv")
	return c.Send(buf.Bytes())
}
