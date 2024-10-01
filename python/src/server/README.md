# Tagger API

## Overview

Tagger API is a FastAPI-based service that provides endpoints for predicting rate names based on input data. It supports both individual predictions and batch processing through CSV file uploads.

## Features

- Predict rate names for individual inputs
- Batch prediction using CSV file upload
- Streaming response for large CSV outputs

## Configuration

Make sure to set the following:

ENV:

- `PORT`: The port on which the API will run (default: 8080)

YAML:

- `MODELS_DIR`: Directory containing the model files
- `CATEGORIES`: List of categories for prediction

## Endpoints

### 1. Predict Rate Names

- **URL**: `/predict`
- **Method**: POST
- **Input**: JSON object with `rate_names` (list of strings) and `categories` (list of strings)
- **Output**: List of dictionaries containing predictions

Example:

```json
{
  "rate_names": ["Rate A", "Rate B"],
  "categories": ["Category1", "Category2"]
}
```

### 2. Upload CSV File

- **URL**: `/upload-csv`
- **Method**: POST
- **Input**: CSV file with a single column named `rate_name`
- **Output**: Stream of predictions in CSV format

### Error Handling

- **400 Bad Request**: If the input data is invalid (e.g., missing required fields)
- **500 Internal Server Error**: If there's an issue with the server (e.g., model loading error)
