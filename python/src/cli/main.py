import csv
import json
import yaml
import pyarrow as pa
import pyarrow.parquet as pq
import os
from typing import List, Dict, Any, Optional, Tuple
import click
import pandas as pd
from rich.console import Console
from tenacity import (
    retry,
    stop_after_attempt,
    wait_exponential,
    retry_if_exception_type,
)
from io import StringIO
import sys
from io import IOBase

from shared.config.config import load_settings
from shared.models.registry import ModelRegistry

console = Console()


class PredictionError(Exception):
    """Exception raised for errors during prediction."""

    def __init__(self, message: str = "Prediction failed"):
        self.message = message
        super().__init__(self.message)


@retry(
    stop=stop_after_attempt(3),
    wait=wait_exponential(multiplier=1, min=4, max=10),
    retry=retry_if_exception_type(PredictionError),
    reraise=True,
)
def make_prediction(
    model_registry: ModelRegistry, inputs: List[str], categories: Optional[List[str]]
) -> List[Dict[str, Any]]:
    try:
        return model_registry.predict(inputs, categories)
    except Exception as e:
        raise PredictionError(f"Prediction failed: {str(e)}")


@click.command()
@click.option(
    "--input",
    "-i",
    required=True,
    help="Input string or file path (CSV)",
)
@click.option(
    "--categories", "-c", multiple=True, help="Specify categories for prediction"
)
@click.option(
    "--output",
    "-o",
    type=click.Path(),
    help="Output file path (default: write to stdout)",
)
@click.option(
    "--format",
    "-f",
    type=click.Choice(["json", "csv", "tsv", "yaml", "parquet"]),
    default="csv",
    help="Output file format (default: csv)",
)
@click.option(
    "--config",
    "-cfg",
    type=click.Path(exists=True),
    help="Path to the configuration file",
    default="config.yaml",
)
@click.option("--verbose", "-v", is_flag=True, help="Enable verbose output")
def cli(
    input: str,
    categories: Tuple[str, ...],
    output: Optional[str],
    format: str,
    config: str,
    verbose: bool,
) -> None:
    """Hotel room rate classification CLI"""
    settings = load_settings(config)
    model_registry = ModelRegistry(settings.MODELS_DIR, settings.CATEGORIES)

    if verbose:
        console.print(f"[blue]Loaded configuration from: {config}[/blue]")
        console.print(f"[blue]Models directory: {settings.MODELS_DIR}[/blue]")
        console.print(f"[blue]Categories: {settings.CATEGORIES}[/blue]")

    categories_list: Optional[List[str]] = list(categories) if categories else None

    if os.path.isfile(input):
        input_stream = open(input, "r")
    else:
        # Treat input as a string
        input_stream = StringIO(input + "\n")

    output_stream = open(output, "w") if output else sys.stdout

    try:
        process_data(
            input_stream,
            output_stream,
            model_registry,
            categories_list,
            format,
            verbose,
        )
    finally:
        if isinstance(input_stream, IOBase):
            input_stream.close()
        if output:
            output_stream.close()


def process_data(
    input_stream,
    output_stream,
    model_registry: ModelRegistry,
    categories: Optional[List[str]],
    format: str,
    verbose: bool,
):
    if isinstance(input_stream, StringIO):
        inputs = [input_stream.getvalue().strip()]
    else:
        reader = csv.DictReader(input_stream)
        inputs = [row["rate_name"] for row in reader]

    try:
        if verbose:
            console.print(f"[blue]Processing {len(inputs)} entries...[/blue]")
        results = make_prediction(model_registry, inputs, categories)

        writer = _get_writer(output_stream, format, results[0].keys())
        _write_results(writer, results, format, output_stream)

        if verbose:
            console.print(
                f"[green]Successfully processed {len(results)} predictions[/green]"
            )
    except PredictionError as e:
        if verbose:
            console.print(f"[red]Error during prediction: {e}[/red]")
        results = _get_partial_results(inputs)
        writer = _get_writer(output_stream, format, results[0].keys())
        _write_results(writer, results, format, output_stream)
    except Exception as e:
        if verbose:
            console.print(f"[red]Unexpected error during prediction: {e}[/red]")


def _get_writer(output_stream, format: str, fieldnames):
    if format == "json":
        return json.JSONEncoder(indent=2)
    elif format in ["csv", "tsv"]:
        delimiter = "," if format == "csv" else "\t"
        writer = csv.DictWriter(
            output_stream, fieldnames=fieldnames, delimiter=delimiter
        )
        writer.writeheader()
        return writer
    elif format == "yaml":
        return yaml.safe_dumper
    elif format == "parquet":
        return pa.Table.from_pandas
    else:
        raise ValueError(f"Unsupported format: {format}")


def _write_results(writer, results: List[Dict[str, Any]], format: str, output_stream):
    if format == "json":
        json.dump(results, output_stream, cls=writer, indent=2)
    elif format in ["csv", "tsv"]:
        writer.writerows(results)
    elif format == "yaml":
        yaml.dump_all(results, output_stream, Dumper=writer)
    elif format == "parquet":
        table = writer(pd.DataFrame(results))
        pq.write_table(table, output_stream.buffer)
    output_stream.flush()


def _get_partial_results(inputs: List[str]) -> List[Dict[str, Any]]:
    return [
        {"input": input, "predicted_category": "UNKNOWN", "confidence": 0.0}
        for input in inputs
    ]

if __name__ == "__main__":
    cli()
