# Tagger CLI

This is a command-line interface (CLI) application that makes predictions using a ModelRegistry. It allows users to input either a single string or a CSV file containing multiple inputs and predicts categories for them.

## Features

- Predict categories for a single input string or multiple inputs from a CSV file
- Specify custom categories for prediction
- Output results to the console or save them to a file (JSON for single input, CSV for multiple inputs)
- Utilizes a ModelRegistry for making predictions

## Installation

To install this CLI app, follow these steps:

1. Ensure you have Python 3.6 or higher installed.
2. Clone this repository:
   ```
   git clone <repository-url>
   cd <repository-name>
   ```
3. Install the required dependencies:
   ```
   pip install -r requirements.txt
   ```

## Usage

The basic syntax for using the CLI is:

```sh
python python/src/cli/main.py predict [OPTIONS] INPUT
```

### Options

- `--categories`, `-c`: Specify categories for prediction (can be used multiple times)
- `--output`, `-o`: Output file path (JSON or CSV)

### Examples

1. Predict categories for a single input string:

   ```
   cli predict "Your input string here"
   ```

2. Predict categories for a single input string with specific categories:

   ```
   cli predict "Your input string here" -c category1 -c category2
   ```

3. Predict categories for inputs in a CSV file:

   ```
   cli predict input.csv
   ```

4. Predict categories and save output to a file:
   ```
   cli predict input.csv -o output.csv
   ```

## Configuration

For the CLI to work, you need to set the `MODELS_DIR` and `CATEGORIES` in the `config.yaml` file.
