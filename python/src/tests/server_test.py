import pytest
from fastapi.testclient import TestClient
from unittest.mock import patch, MagicMock
import pandas as pd
from io import StringIO

from server.main import app
from server.api.schemas import RateNameInput
from shared.models.registry import ModelRegistry

client = TestClient(app)


@pytest.fixture
def mock_model_registry():
    with patch("server.api.routes.ModelRegistry") as mock:
        mock_instance = MagicMock()
        mock_instance.predict.return_value = [
            {
                "input": "Deluxe Ocean View",
                "predicted_category": "view",
                "confidence": 0.9,
            },
            {
                "input": "Standard Room",
                "predicted_category": "capacity",
                "confidence": 0.8,
            },
        ]
        mock.return_value = mock_instance
        yield mock_instance


def test_predict_rate_names(mock_model_registry):
    response = client.post(
        "/predict", json={"rate_names": ["Deluxe Ocean View", "Standard Room"]}
    )
    assert response.status_code == 200
    assert len(response.json()) == 2
    assert response.json()[0]["input"] == "Deluxe Ocean View"
    assert response.json()[1]["input"] == "Standard Room"


def test_predict_rate_names_with_categories(mock_model_registry):
    response = client.post(
        "/predict", json={"rate_names": ["Deluxe Ocean View"], "categories": ["view"]}
    )
    assert response.status_code == 200
    mock_model_registry.predict.assert_called_once_with(["Deluxe Ocean View"], ["view"])


def test_predict_rate_names_with_invalid_input():
    response = client.post("/predict", json={"rate_names": []})
    assert response.status_code == 422  # Validation error


def test_predict_rate_names_with_nan_values(mock_model_registry):
    response = client.post(
        "/predict", json={"rate_names": ["Deluxe Ocean View", None, ""]}
    )
    assert response.status_code == 200
    mock_model_registry.predict.assert_called_once_with(
        ["Deluxe Ocean View", "", ""], None
    )


def test_predict_csv(mock_model_registry):
    csv_content = "rate_name\nDeluxe Ocean View\nStandard Room"
    response = client.post("/predict_csv", files={"file": ("test.csv", csv_content)})
    assert response.status_code == 200
    assert response.headers["Content-Type"] == "text/csv"
    assert (
        "attachment; filename=predictions.csv"
        in response.headers["Content-Disposition"]
    )

    # Parse the returned CSV
    result_df = pd.read_csv(StringIO(response.content.decode()))
    assert len(result_df) == 2
    assert "input" in result_df.columns
    assert "predicted_category" in result_df.columns
    assert "confidence" in result_df.columns


def test_predict_csv_invalid_file():
    response = client.post(
        "/predict_csv", files={"file": ("test.csv", "invalid,csv,content")}
    )
    assert response.status_code == 400


def test_predict_csv_missing_column():
    csv_content = "wrong_column\nDeluxe Ocean View\nStandard Room"
    response = client.post("/predict_csv", files={"file": ("test.csv", csv_content)})
    assert response.status_code == 400
    assert "CSV file must contain a 'rate_name' column" in response.json()["detail"]


# Tests for schemas.py
def test_rate_name_input_schema_valid():
    input_data = RateNameInput(rate_names=["Deluxe Ocean View", "Standard Room"])
    assert input_data.rate_names == ["Deluxe Ocean View", "Standard Room"]
    assert input_data.categories is None


def test_rate_name_input_schema_with_categories():
    input_data = RateNameInput(rate_names=["Deluxe Ocean View"], categories=["view"])
    assert input_data.rate_names == ["Deluxe Ocean View"]
    assert input_data.categories == ["view"]


def test_rate_name_input_schema_invalid():
    with pytest.raises(ValueError):
        RateNameInput(rate_names=[])


def test_rate_name_input_schema_invalid_category():
    with pytest.raises(ValueError):
        RateNameInput(rate_names=["Deluxe Ocean View"], categories=["invalid_category"])
