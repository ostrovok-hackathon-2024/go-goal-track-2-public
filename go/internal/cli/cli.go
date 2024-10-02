package cli

import (
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
	rootCmd.Flags().StringP("input", "i", "", "Input CSV file containing strings to classify")
	rootCmd.Flags().StringP("output", "o", "predictions.csv", "Output CSV file for predictions")
	rootCmd.Flags().StringP("tfidf", "t", "", "TF-IDF data file")
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
	tfidfFile, _ := cmd.Flags().GetString("tfidf")

	if inputFile == "" {
		fmt.Println("Error: input file is required")
		return
	}

	// Load TF-IDF data
	if tfidfFile == "" {
		tfidfFile = filepath.Join(cfg.ModelsDir, "tfidf", "tfidf_data.json")

	}
	tfidfData, err := tfidf.LoadTfIdfData(tfidfFile)
	if err != nil {
		fmt.Printf("Error loading TF-IDF data: %v\n", err)
		return
	}

	// Read rate names
	rateNames := utils.ReadRateNames(inputFile)[1:]

	// Calculate TF-IDF vectors
	floats := tfidf.CalculateTfIdfVectors(rateNames, &tfidfData)

	// Load models and make predictions
	cbmDir := filepath.Join(cfg.ModelsDir, "cbm")
	labelsDir := filepath.Join(cfg.ModelsDir, "labels")
	results, err := model.PredictAll(floats, rateNames, categories, cbmDir, labelsDir)
	if err != nil {
		fmt.Printf("Error making predictions: %v\n", err)
		return
	}

	// Write results to CSV
	err = utils.WriteResultsToCSV(outputFile, results, rateNames, categories)
	if err != nil {
		fmt.Printf("Error writing results to CSV: %v\n", err)
		return
	}

	fmt.Printf("Predictions written to %s\n", outputFile)
}
