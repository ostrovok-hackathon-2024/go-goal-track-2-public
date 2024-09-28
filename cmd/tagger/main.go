package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"strings"

	"github.com/cdipaolo/goml/base"
	"github.com/cdipaolo/goml/text"
)

// Define RoomCharacteristics with the new fields
type RoomCharacteristics struct {
	Class     string
	Quality   string
	Bathroom  string
	Bedding   string
	Capacity  string
	Club      string
	Bedrooms  string
	Balcony   string
	View      string
}

type Classifier struct {
	Class     *text.NaiveBayes
	Quality   *text.NaiveBayes
	Bathroom  *text.NaiveBayes
	Bedding   *text.NaiveBayes
	Capacity  *text.NaiveBayes
	View      *text.NaiveBayes
}

func preprocessText(input string) string {
	return strings.Join(strings.Fields(strings.ToLower(input)), " ")
}

// Training classifier for a given label index
func trainClassifier(data [][]string, labelIndex int) *text.NaiveBayes {
	stream := make(chan base.TextDatapoint, 100)
	classifier := text.NewNaiveBayes(stream, 5, base.OnlyWordsAndNumbers)

	go func() {
		for _, row := range data {
			if len(row) > labelIndex {
				stream <- base.TextDatapoint{
					X: preprocessText(row[0]),
					Y: uint8(labelToInt(row[labelIndex])),
				}
			}
		}
		close(stream)
	}()

	err := make(chan error)
	classifier.OnlineLearn(err)
	return classifier
}

// Updated label to integer mappings
func labelToInt(label string) int {
	switch strings.ToLower(label) {
	case "room", "studio", "suite", "junior-suite":
		return 1
	case "deluxe", "premium", "standard":
		return 2
	case "private bathroom", "shared bathroom":
		return 3
	case "queen", "king", "double":
		return 4
	case "sea view", "city view", "garden view", "pool view":
		return 5
	default:
		return 0
	}
}

func classifyRoom(classifier *Classifier, description string) RoomCharacteristics {
	processed := preprocessText(description)

	classClass := classifier.Class.Predict(processed)
	qualityClass := classifier.Quality.Predict(processed)
	bathroomClass := classifier.Bathroom.Predict(processed)
	beddingClass := classifier.Bedding.Predict(processed)
	capacityClass := classifier.Capacity.Predict(processed)
	viewClass := classifier.View.Predict(processed)

	return RoomCharacteristics{
		Class:    getClassLabel(classClass),
		Quality:  getQualityLabel(qualityClass),
		Bathroom: getBathroomLabel(bathroomClass),
		Bedding:  getBeddingLabel(beddingClass),
		Capacity: getCapacityLabel(capacityClass),
		View:     getViewLabel(viewClass),
	}
}

func getClassLabel(class uint8) string {
	labels := []string{"undefined", "room", "studio", "suite", "junior-suite"}
	if int(class) < len(labels) {
		return labels[class]
	}
	return "undefined"
}

func getQualityLabel(class uint8) string {
	labels := []string{"undefined", "standard", "deluxe", "premium"}
	if int(class) < len(labels) {
		return labels[class]
	}
	return "undefined"
}

func getBathroomLabel(class uint8) string {
	labels := []string{"undefined", "private bathroom", "shared bathroom"}
	if int(class) < len(labels) {
		return labels[class]
	}
	return "undefined"
}

func getBeddingLabel(class uint8) string {
	labels := []string{"undefined", "single", "double", "queen", "king"}
	if int(class) < len(labels) {
		return labels[class]
	}
	return "undefined"
}

func getCapacityLabel(class uint8) string {
	labels := []string{"undefined", "single", "double", "triple", "quadruple"}
	if int(class) < len(labels) {
		return labels[class]
	}
	return "undefined"
}

func getViewLabel(class uint8) string {
	labels := []string{"undefined", "sea view", "city view", "pool view", "garden view"}
	if int(class) < len(labels) {
		return labels[class]
	}
	return "undefined"
}

func main() {
	// Load training data
	trainingData, err := loadCSV("training_data.csv")
	if err != nil {
		fmt.Println("Error loading training data:", err)
		return
	}

	// Initialize classifiers
	classifier := &Classifier{
		Class:    trainClassifier(trainingData, 1),
		Quality:  trainClassifier(trainingData, 2),
		Bathroom: trainClassifier(trainingData, 3),
		Bedding:  trainClassifier(trainingData, 4),
		Capacity: trainClassifier(trainingData, 5),
		View:     trainClassifier(trainingData, 6),
	}

	// Input/Output setup
	inputFile := "input.csv"
	outputFile := "output.csv"

	input, err := os.Open(inputFile)
	if err != nil {
		fmt.Println("Error opening input file:", err)
		return
	}
	defer input.Close()

	output, err := os.Create(outputFile)
	if err != nil {
		fmt.Println("Error creating output file:", err)
		return
	}
	defer output.Close()

	reader := csv.NewReader(input)
	writer := csv.NewWriter(output)

	// Write the header
	writer.Write([]string{"rate_name", "class", "quality", "bathroom", "bedding", "capacity", "club", "bedrooms", "balcony", "view"})

	// Process each row in input.csv
	for {
		record, err := reader.Read()
		if err != nil {
			break
		}

		description := record[0]
		characteristics := classifyRoom(classifier, description)

		// Write to output.csv
		writer.Write([]string{
			description,
			characteristics.Class,
			characteristics.Quality,
			characteristics.Bathroom,
			characteristics.Bedding,
			characteristics.Capacity,
			"not club", // assuming club is "not club" for all
			"",         // bedrooms not specified
			"no balcony",
			characteristics.View,
		})
	}

	writer.Flush()
	fmt.Println("Processing complete. Output written to", outputFile)
}

// Function to load CSV data
func loadCSV(filename string) ([][]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	return reader.ReadAll()
}
