package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"regexp"
	"strings"
	"unicode"

	"github.com/cdipaolo/goml/base"
	"github.com/cdipaolo/goml/text"
)

type RoomCharacteristics struct {
	Class    string
	Quality  string
	Bathroom string
	Bedding  string
	Capacity string
	Bedrooms string
	Club     string
	Balcony  string
	View     string
	Floor    string
}

type Classifier struct {
	Class    *text.NaiveBayes
	Quality  *text.NaiveBayes
	Bathroom *text.NaiveBayes
	View     *text.NaiveBayes
}

func preprocessText(input string) string {
	lower := strings.ToLower(input)
	cleaned := strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsNumber(r) || r == ' ' || r == '-' {
			return r
		}
		return ' '
	}, lower)
	return strings.Join(strings.Fields(cleaned), " ")
}

func trainClassifier(data [][]string, labelIndex int, labels []string) *text.NaiveBayes {
	stream := make(chan base.TextDatapoint, 100)
	classifier := text.NewNaiveBayes(stream, uint8(len(labels)), base.OnlyWordsAndNumbers)

	go func() {
		for _, row := range data {
			if len(row) > labelIndex {
				stream <- base.TextDatapoint{
					X: preprocessText(row[0]),
					Y: uint8(labelToInt(row[labelIndex], labels)),
				}
			}
		}
		close(stream)
	}()

	err := make(chan error)
	classifier.OnlineLearn(err)
	return classifier
}

func labelToInt(label string, labels []string) int {
	for i, l := range labels {
		if strings.ToLower(label) == l {
			return i
		}
	}
	return 0
}

func classifyRoom(classifier *Classifier, description string) RoomCharacteristics {
	processed := preprocessText(description)

	return RoomCharacteristics{
		Class:    getLabel(classifier.Class.Predict(processed), classLabels),
		Quality:  getLabel(classifier.Quality.Predict(processed), qualityLabels),
		Bathroom: getLabel(classifier.Bathroom.Predict(processed), bathroomLabels),
		Bedding:  inferBedding(description),
		Capacity: inferCapacity(description),
		Bedrooms: inferBedrooms(description),
		Club:     inferClub(description),
		Balcony:  inferBalcony(description),
		View:     getLabel(classifier.View.Predict(processed), viewLabels),
		Floor:    inferFloor(description),
	}
}

func getLabel(class uint8, labels []string) string {
	if int(class) < len(labels) {
		return labels[class]
	}
	return "undefined"
}

func inferBedding(description string) string {
	lower := strings.ToLower(description)
	if strings.Contains(lower, "bunk") {
		return "bunk bed"
	} else if strings.Contains(lower, "single") {
		return "single bed"
	} else if strings.Contains(lower, "double") || strings.Contains(lower, "twin") {
		return "double/double-or-twin"
	} else if strings.Contains(lower, "multiple") {
		return "multiple"
	}
	return "undefined"
}

func inferCapacity(description string) string {
	lower := strings.ToLower(description)
	if strings.Contains(lower, "sextuple") {
		return "sextuple"
	} else if strings.Contains(lower, "quintuple") {
		return "quintuple"
	} else if strings.Contains(lower, "quadruple") || strings.Contains(lower, "quad") {
		return "quadruple"
	} else if strings.Contains(lower, "triple") {
		return "triple"
	} else if strings.Contains(lower, "double") {
		return "double"
	} else if strings.Contains(lower, "single") {
		return "single"
	}
	return "undefined"
}

func inferBedrooms(description string) string {
	re := regexp.MustCompile(`(\d+)\s*bedroom`)
	match := re.FindStringSubmatch(strings.ToLower(description))
	if len(match) > 1 {
		return match[1] + " bedroom"
	}
	return "undefined"
}

func inferClub(description string) string {
	if strings.Contains(strings.ToLower(description), "club") {
		return "club"
	}
	return "not club"
}

func inferBalcony(description string) string {
	if strings.Contains(strings.ToLower(description), "balcony") {
		return "with balcony"
	}
	return "no balcony"
}

func inferFloor(description string) string {
	lower := strings.ToLower(description)
	if strings.Contains(lower, "penthouse") {
		return "penthouse floor"
	} else if strings.Contains(lower, "duplex") {
		return "duplex floor"
	} else if strings.Contains(lower, "basement") {
		return "basement floor"
	} else if strings.Contains(lower, "attic") {
		return "attic floor"
	}
	return "undefined"
}

func postProcessCharacteristics(ch RoomCharacteristics) RoomCharacteristics {
	if ch.Class == "suite" || ch.Class == "junior-suite" || ch.Class == "apartment" || ch.Class == "villa" {
		ch.Bathroom = "private bathroom"
	}
	if ch.Capacity == "undefined" && ch.Bedding != "undefined" {
		if ch.Bedding == "single bed" {
			ch.Capacity = "single"
		} else {
			ch.Capacity = "double"
		}
	}
	if ch.Bedding == "undefined" && ch.Capacity != "undefined" {
		if ch.Capacity == "single" {
			ch.Bedding = "single bed"
		} else {
			ch.Bedding = "double/double-or-twin"
		}
	}
	return ch
}

var classLabels = []string{"undefined", "run-of-house", "dorm", "capsule", "room", "junior-suite", "suite", "apartment", "studio", "villa", "cottage", "bungalow", "chalet", "camping", "tent"}
var qualityLabels = []string{"undefined", "economy", "standard", "comfort", "business", "superior", "deluxe", "premier", "executive", "presidential", "premium", "classic", "ambassador", "grand", "luxury", "platinum", "prestige", "privilege", "royal"}
var bathroomLabels = []string{"undefined", "shared bathroom", "private bathroom", "external private bathroom"}
var viewLabels = []string{"undefined", "bay view", "bosphorus view", "burj-khalifa view", "canal view", "city view", "courtyard view", "dubai-marina view", "garden view", "golf view", "harbour view", "inland view", "kremlin view", "lake view", "land view", "mountain view", "ocean view", "panoramic view", "park view", "partial-ocean view", "partial-sea view", "partial view", "pool view", "river view", "sea view", "sheikh-zayed view", "street view", "sunrise view", "sunset view", "water view", "with view", "beachfront", "ocean front", "sea front"}

func main() {
	trainingData, err := loadCSV("training_data.csv")
	if err != nil {
		fmt.Println("Error loading training data:", err)
		return
	}

	classifier := &Classifier{
		Class:    trainClassifier(trainingData, 1, classLabels),
		Quality:  trainClassifier(trainingData, 2, qualityLabels),
		Bathroom: trainClassifier(trainingData, 3, bathroomLabels),
		View:     trainClassifier(trainingData, 9, viewLabels),
	}

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

	writer.Write([]string{"rate_name", "class", "quality", "bathroom", "bedding", "capacity", "bedrooms", "club", "balcony", "view", "floor"})

	for {
		record, err := reader.Read()
		if err != nil {
			break
		}

		description := record[0]
		characteristics := classifyRoom(classifier, description)
		characteristics = postProcessCharacteristics(characteristics)

		writer.Write([]string{
			description,
			characteristics.Class,
			characteristics.Quality,
			characteristics.Bathroom,
			characteristics.Bedding,
			characteristics.Capacity,
			characteristics.Bedrooms,
			characteristics.Club,
			characteristics.Balcony,
			characteristics.View,
			characteristics.Floor,
		})
	}

	writer.Flush()
	fmt.Println("Processing complete. Output written to", outputFile)
}

func loadCSV(filename string) ([][]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	return reader.ReadAll()
}
