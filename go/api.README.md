# Golang Prediction API

This API provides endpoints for predicting rate names based on categories using TF-IDF and machine learning models.

## Endpoints

### 1. Predict Rate Names

**Endpoint:** `POST /predict`

This endpoint accepts JSON input containing rate names and categories, and returns predictions for each rate name.

#### Request Body

```json
{
  "inputs": ["rate_name_1", "rate_name_2", "..."],
  "categories": ["category_1", "category_2", "..."]
}
```

- `inputs`: An array of rate names to predict.
- `categories`: (Optional) An array of categories to use for prediction. If not provided, default categories from the configuration will be used.

#### Response

The response is a JSON object where keys are the input rate names, and values are objects containing predictions for each category.

```json
{
  "rate_name_1": {
    "category_1": "prediction_1",
    "category_2": "prediction_2",
    "...": "..."
  },
  "rate_name_2": {
    "...": "..."
  }
}
```

### 2. Predict Rate Names from CSV

**Endpoint:** `POST /predict_csv`

This endpoint accepts a CSV file upload containing rate names and categories, processes it, and returns predictions as a downloadable CSV file.

#### Request

- Method: POST
- Content-Type: multipart/form-data
- Form field: `file` (CSV file)

The CSV file should have the following format:
- First column: Rate names
- Subsequent columns: Category names (used as headers)

Example:
```
Rate Name,Category1,Category2,Category3
rate_name_1,,,
rate_name_2,,,
...
```

#### Response

- Content-Type: text/csv
- Content-Disposition: attachment; filename=predictions.csv

The response is a CSV file containing the original rate names and the predicted categories.

## Error Handling

Both endpoints return appropriate HTTP status codes and error messages in case of failures:

- 400 Bad Request: Invalid input data or file format
- 500 Internal Server Error: Server-side errors (e.g., model loading failures)

Error responses are in JSON format:

```json
{
  "error": "Error message description"
}
```

## Configuration

The API uses a configuration file (`config.yaml`) to set up various parameters, including:

- Model directories
- Default categories
- TF-IDF data file location

Ensure that the configuration file and all necessary model files are properly set up before running the API.
