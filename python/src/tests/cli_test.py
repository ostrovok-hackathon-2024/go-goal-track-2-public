import csv
import json
import time
import pytest
from click.testing import CliRunner
from typing import List, Dict, Any
import pandas as pd
import pyarrow.parquet as pq
from unittest.mock import MagicMock, patch

import yaml

# Update this import statement
from cli.main import cli, make_prediction, PredictionError
from shared.config.config import load_settings
from shared.models.registry import ModelRegistry


@pytest.fixture
def mock_model_registry():
    mock_registry = MagicMock(spec=ModelRegistry)
    mock_registry.predict.return_value = [
        {
            "input": "test input",
            "predicted_category": "test category",
            "confidence": 0.9,
        }
    ]
    return mock_registry


@pytest.fixture
def test_config(tmp_path):
    config_path = tmp_path / "test_config.yaml"
    config_content = """
    models_dir: "test_models"
    categories:
      - test_category1
      - test_category2
    """
    config_path.write_text(config_content)
    return str(config_path)


def test_cli_predict(mock_model_registry, test_config):
    runner = CliRunner()
    # Update the patch target
    with patch("cli.main.ModelRegistry", return_value=mock_model_registry):
        result = runner.invoke(cli, ["--config", test_config, "predict", "test input"])

    assert result.exit_code == 0
    assert "Successfully made predictions" in result.output


def test_make_prediction(mock_model_registry):
    inputs = ["test input 1", "test input 2"]
    categories = ["category1", "category2"]

    results = make_prediction(mock_model_registry, inputs, categories)

    assert len(results) == 1
    assert results[0]["input"] == "test input"
    assert results[0]["predicted_category"] == "test category"
    assert results[0]["confidence"] == 0.9


def test_make_prediction_error(mock_model_registry):
    mock_model_registry.predict.side_effect = Exception("Test error")

    with pytest.raises(PredictionError):
        make_prediction(mock_model_registry, ["test input"], None)


def test_load_settings(test_config):
    settings = load_settings(test_config)
    assert settings.MODELS_DIR == "test_models"
    assert settings.CATEGORIES == ["test_category1", "test_category2"]


@pytest.mark.parametrize("input_format", ["json", "csv", "tsv", "yaml", "parquet"])
def test_predict_output_formats(
    mock_model_registry, test_config, tmp_path, input_format
):
    output_file = tmp_path / f"test_output.{input_format}"
    runner = CliRunner()
    # Update the patch target
    with patch("cli.main.ModelRegistry", return_value=mock_model_registry):
        result = runner.invoke(
            cli,
            [
                "--config",
                test_config,
                "predict",
                "test input",
                "--output",
                str(output_file),
                "--format",
                input_format,
            ],
        )

    assert result.exit_code == 0
    assert output_file.exists()

    # Verify content of the output file
    if input_format == "json":
        with open(output_file, "r") as f:
            data = json.load(f)
    elif input_format in ["csv", "tsv"]:
        with open(output_file, "r") as f:
            reader = csv.DictReader(f, delimiter="," if input_format == "csv" else "\t")
            data = list(reader)
    elif input_format == "yaml":
        with open(output_file, "r") as f:
            data = yaml.safe_load(f)
    elif input_format == "parquet":
        table = pq.read_table(output_file)
        data = table.to_pylist()

    assert len(data) == 1
    assert data[0]["input"] == "test input"
    assert data[0]["predicted_category"] == "test category"
    assert float(data[0]["confidence"]) == 0.9


def test_predict_with_csv_input(mock_model_registry, test_config, tmp_path):
    input_file = tmp_path / "test_input.csv"
    input_data = pd.DataFrame({"rate_name": ["test input 1", "test input 2"]})
    input_data.to_csv(input_file, index=False)

    runner = CliRunner()
    with patch("cli.main.ModelRegistry", return_value=mock_model_registry):
        result = runner.invoke(
            cli,
            [
                "--config",
                test_config,
                "predict",
                str(input_file),
            ],
        )

    assert result.exit_code == 0
    assert "Successfully made predictions" in result.output
    assert "Loaded 2 entries from" in result.output


def test_predict_with_verbose_flag(mock_model_registry, test_config):
    runner = CliRunner()
    with patch("cli.main.ModelRegistry", return_value=mock_model_registry):
        result = runner.invoke(
            cli,
            [
                "--config",
                test_config,
                "--verbose",
                "predict",
                "test input",
            ],
        )

    assert result.exit_code == 0
    assert "Loaded configuration from:" in result.output
    assert "Models directory:" in result.output
    assert "Categories:" in result.output
    assert "Starting prediction process..." in result.output
    assert "Number of predictions:" in result.output


def test_predict_with_categories(mock_model_registry, test_config):
    runner = CliRunner()
    with patch("cli.main.ModelRegistry", return_value=mock_model_registry):
        result = runner.invoke(
            cli,
            [
                "--config",
                test_config,
                "predict",
                "test input",
                "--categories",
                "category1",
                "--categories",
                "category2",
            ],
        )

    assert result.exit_code == 0
    assert "Successfully made predictions" in result.output
    # You might want to add more assertions here to verify that the categories were used correctly


def test_predict_with_prediction_error(mock_model_registry, test_config):
    mock_model_registry.predict.side_effect = Exception("Test error")

    runner = CliRunner()
    with patch("cli.main.ModelRegistry", return_value=mock_model_registry):
        result = runner.invoke(
            cli,
            [
                "--config",
                test_config,
                "predict",
                "test input",
            ],
        )

    assert result.exit_code == 0
    assert "Error during prediction after retries:" in result.output
    assert "Attempting to proceed with partial results..." in result.output


def test_predict_uses_specified_categories(mock_model_registry, test_config):
    runner = CliRunner()
    with patch("cli.main.ModelRegistry", return_value=mock_model_registry) as mock_registry:
        result = runner.invoke(
            cli,
            [
                "--config",
                test_config,
                "predict",
                "test input",
                "--categories",
                "category1",
                "--categories",
                "category2",
            ],
        )

    assert result.exit_code == 0
    mock_registry.return_value.predict.assert_called_once_with(
        ["test input"], ["category1", "category2"]
    )


if __name__ == "__main__":
    pytest.main()
