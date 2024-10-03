# Tagger CLI

## Features

- Process individual strings or CSV files containing rate names
- Predict multiple categories for each input
- Configurable via YAML configuration file
- Output results in various formats (CSV, JSON, TSV, YAML)
- Utilizes TF-IDF and machine learning models for predictions

## Configuration

Tagger uses a configuration file (default: `config.yaml`) to set up various parameters. You can specify a custom configuration file using the `--config` flag.

Example `config.yaml`:

```yaml
modelsDir: "../artifacts"
inputCol: "RateName"
categories:
  - "Category1"
  - "Category2"
  - "Category3"
```

Only specified categories will be used for prediction.

## Usage

### Basic Usage

```bash
tagger --input <input_file_or_string> [flags]
```

### Flags

- `--input`, `-i`: Input CSV file containing strings to classify or a single string to classify (required)
- `--output`, `-o`: Output file for predictions (optional)
- `--category`, `-c`: Categories to predict (can be specified multiple times, optional)
- `--format`, `-f`: Output format (csv, json, tsv, yaml) (default: csv)
- `--config`: Config file (default is ./config.yaml)

### Examples

1. Classify a single string:

```bash
tagger --input "Example Rate Name"
```

2. Process a CSV file and output results to another file:

```bash
tagger --input input.csv --output results.csv
```

3. Specify custom categories and output format:

```bash
tagger --input input.csv --category Category1 --category Category2 --format json
```

4. Use a custom configuration file:

```bash
tagger --config custom_config.yaml --input input.csv
```

## Output

The tool will output the results in the specified format (CSV, JSON, TSV, or YAML), either to the specified output file or to the console if no output file is provided. The output will include the input string/rate name and the predicted categories.

## Models and Data

The tool expects the following directory structure for models and data:

- `<modelsDir>/tfidf/tfidf_data.json`: TF-IDF data
- `<modelsDir>/cbm/catboost_model_XXXXXX.cbm`: Directory containing CBM models
- `<modelsDir>/labels/json/labels_XXXXXX.json`: Directory containing label data

Ensure that these directories and files are present and properly configured in your `config.yaml` file.
