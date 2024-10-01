import time
import pytest
from click.testing import CliRunner
from typing import List, Dict, Any
import pandas as pd
from unittest.mock import MagicMock, patch

from main import cli, make_prediction
from shared.config.config import load_settings
from shared.models.registry import ModelRegistry

@pytest.fixture
def mock_model_registry():
    mock_registry = MagicMock(spec=ModelRegistry)
    mock_registry.predict.return_value = [
        {"input": "test input", "predicted_category": "test category", "confidence": 0.9}
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
    with patch('main.ModelRegistry', return_value=mock_model_registry):
        result = runner.invoke(cli, ['--config', test_config, 'predict', 'test input'])

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

def test_load_settings(test_config):
    settings = load_settings(test_config)
    assert settings.MODELS_DIR == "test_models"
    assert settings.CATEGORIES == ["test_category1", "test_category2"]

@pytest.mark.parametrize("input_format", ["json", "csv", "tsv", "yaml", "parquet"])
def test_predict_output_formats(mock_model_registry, test_config, tmp_path, input_format):
    output_file = tmp_path / f"test_output.{input_format}"
    runner = CliRunner()
    with patch('main.ModelRegistry', return_value=mock_model_registry):
        result = runner.invoke(cli, [
            '--config', test_config,
            'predict',
            'test input',
            '--output', str(output_file),
            '--format', input_format
        ])

    assert result.exit_code == 0
    assert output_file.exists()

def benchmark_prediction_speed(mock_model_registry):
    num_inputs = 1000
    inputs = [f"test input {i}" for i in range(num_inputs)]

    start_time = time.time()
    results = make_prediction(mock_model_registry, inputs, None)
    end_time = time.time()

    prediction_time = end_time - start_time
    predictions_per_second = num_inputs / prediction_time

    print(f"Made {num_inputs} predictions in {prediction_time:.2f} seconds")
    print(f"Predictions per second: {predictions_per_second:.2f}")

if __name__ == "__main__":
    mock_registry = MagicMock(spec=ModelRegistry)
    mock_registry.predict.return_value = [
        {"input": "test input", "predicted_category": "test category", "confidence": 0.9}
    ]
    benchmark_prediction_speed(mock_registry)
