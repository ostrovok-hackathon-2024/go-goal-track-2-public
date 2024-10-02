package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

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
	rootCmd.Flags().StringP("format", "f", "csv", "Output format (csv, json, tsv, yaml)")
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
	outputFormat, _ := cmd.Flags().GetString("format")

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
		err := utils.WriteOutput(outputFile, outputFormat, headers, inputStrings, results)
		if err != nil {
			fmt.Printf("Error writing output: %v\n", err)
			return
		}
	} else {
		printResult(outputFormat, headers, inputStrings, results)
	}
}

func printResult(format string, headers []string, firstColData []string, rowData map[string]map[string]string) {
	switch strings.ToLower(format) {
	case "csv":
		utils.PrintCSV(os.Stdout, headers, firstColData, rowData)
	case "json":
		utils.PrintJSON(os.Stdout, headers, firstColData, rowData)
	case "tsv":
		utils.PrintTSV(os.Stdout, headers, firstColData, rowData)
	case "yaml":
		utils.PrintYAML(os.Stdout, headers, firstColData, rowData)
	default:
		fmt.Printf("Unsupported format: %s\n", format)
	}
}
