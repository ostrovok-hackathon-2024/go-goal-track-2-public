package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/go-goal/tagger/internal/prediction"
	"github.com/go-goal/tagger/internal/tui"
)

var (
	useTUI      bool
	modelsDir   string
	labelsDir   string
	tfidfPath   string
	inputString string

	rootCmd = &cobra.Command{
		Use:   "tagger [input string]",
		Short: "Tagger is a tool for managing tags",
		Long:  `A longer description of the Tagger application...`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			inputString = args[0]
			if useTUI {
				app, err := tui.NewTaggerApp(tfidfPath, labelsDir, modelsDir)
				if err != nil {
					return err
				}
				return app.Run()
			}

			// Use the prediction package for CLI mode
			predictor, err := prediction.NewPredictor(tfidfPath, labelsDir, modelsDir)
			if err != nil {
				return err
			}

			results, err := predictor.Predict(inputString)
			if err != nil {
				return err
			}

			// Print results
			for model, result := range results {
				fmt.Printf("%s: %s\n", model, result)
			}

			return nil
		},
	}
)

func init() {
	rootCmd.PersistentFlags().BoolVarP(&useTUI, "tui", "t", false, "Start the Text User Interface (TUI)")
	rootCmd.PersistentFlags().StringVar(&modelsDir, "models", "models_better/cbm", "Directory containing models")
	rootCmd.PersistentFlags().StringVar(&labelsDir, "labels", "models_better/labels", "Directory containing labels")
	rootCmd.PersistentFlags().StringVar(&tfidfPath, "tfidf", "models_better/tfidf/tfidf_data.json", "Path to TF-IDF data file")
}

func Execute() error {
	return rootCmd.Execute()
}
