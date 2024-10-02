package cli

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/go-goal/tagger/internal/config"
	"github.com/go-goal/tagger/internal/model"
	"github.com/go-goal/tagger/internal/tfidf"
	"github.com/go-goal/tagger/pkg/utils"
)

var (
	cfgFile    string
	cfg        *config.Config
	categories []string
)

var rootCmd = &cobra.Command{
	Use:   "tagger",
	Short: "A CLI tool for tagging rate names",
	Long:  `Tagger is a CLI tool that uses machine learning models to tag rate names with various attributes.`,
	Run:   runTagger,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ./config.yaml)")
	rootCmd.Flags().StringP("input", "i", "", "Input CSV file containing strings to classify or a single string to classify")
	rootCmd.Flags().StringP("output", "o", "", "Output CSV file for predictions")
	rootCmd.Flags().StringSliceVarP(&categories, "category", "c", []string{}, "Categories to predict (can be specified multiple times)")
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.AddConfigPath(".")
		viper.SetConfigName("config")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		fmt.Printf("Error reading config: %v\n", err)
		os.Exit(1)
	}

	var err error
	cfg, err = config.LoadConfig(viper.ConfigFileUsed())
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		return
	}
}

func runTagger(cmd *cobra.Command, args []string) {
	inputFile, _ := cmd.Flags().GetString("input")
	outputFile, _ := cmd.Flags().GetString("output")

	if inputFile == "" {
		fmt.Println("Error: input is required")
		return
	}

	tfidfFile := filepath.Join(cfg.ModelsDir, "tfidf", "tfidf_data.json")
	tfidfData, err := tfidf.LoadTfIdfData(tfidfFile)
	if err != nil {
		fmt.Printf("Error loading TF-IDF data: %v\n", err)
		return
	}

	// Determine mode based on input
	isFileMode := utils.IsFile(inputFile)

	var inputStrings []string
	if isFileMode {
		// File mode: Read rate names from file
		inputStrings, err = utils.ReadFirstCSVColumn(inputFile, cfg.InputCol)
		if err != nil {
			fmt.Printf("Error reading rate names: %v\n", err)
			return
		}
	} else {
		// Text input mode: Use the input directly
		inputStrings = []string{inputFile}
	}

	// If no categories are specified, use the default categories from the config
	if len(categories) == 0 {
		categories = cfg.Categories
	}

	// Load models and make predictions
	cbmDir := filepath.Join(cfg.ModelsDir, "cbm")
	labelsDir := filepath.Join(cfg.ModelsDir, "labels/json")
	predictor := model.NewPredictor(&tfidfData, cbmDir, labelsDir, categories)
	err = predictor.LoadModels()
	if err != nil {
		fmt.Printf("Error loading models: %v\n", err)
		return
	}

	results, err := predictor.PredictAll(inputStrings)
	if err != nil {
		fmt.Printf("Error making predictions: %v\n", err)
		return
	}

	headers := append([]string{cfg.InputCol}, categories...)
	if outputFile != "" {
		err := utils.WriteCSV(outputFile, headers, inputStrings, results)
		if err != nil {
			fmt.Printf("Error writing CSV: %v\n", err)
			return
		}
	} else {
		printCSVResult(headers, inputStrings, results)
	}
}

func printCSVResult(headers []string, firstColData []string, rowData map[string]map[string]string) {
	writer := csv.NewWriter(os.Stdout)
	defer writer.Flush()

	// Write headers
	writer.Write(headers)

	// Write data rows
	for _, inputValue := range firstColData {
		row := make([]string, len(headers))
		row[0] = inputValue
		for j := 1; j < len(headers); j++ {
			row[j] = rowData[inputValue][headers[j]]
		}
		writer.Write(row)
	}

	// Check for any errors during writing
	if err := writer.Error(); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing CSV to stdout: %v\n", err)
	}
}
