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
from rich.table import Table
from tenacity import (
    retry,
    stop_after_attempt,
    wait_exponential,
    retry_if_exception_type,
)
from io import StringIO
import sys

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
    type=click.Path(exists=True),
    required=True,
    help="Input file path (CSV) or '-' for stdin",
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
@click.option(
    "--output-categories",
    "-oc",
    multiple=True,
    help="Specify categories to include in the output",
)
def cli(
    input: str,
    categories: Tuple[str, ...],
    output: Optional[str],
    format: str,
    config: str,
    verbose: bool,
    output_categories: Tuple[str, ...],
) -> None:
    """Hotel room rate classification CLI"""
    settings = load_settings(config)
    model_registry = ModelRegistry(settings.MODELS_DIR, settings.CATEGORIES)

    if verbose:
        console.print(f"[blue]Loaded configuration from: {config}[/blue]")
        console.print(f"[blue]Models directory: {settings.MODELS_DIR}[/blue]")
        console.print(f"[blue]Categories: {settings.CATEGORIES}[/blue]")

    categories_list: Optional[List[str]] = list(categories) if categories else None
    output_categories_list: Optional[List[str]] = (
        list(output_categories) if output_categories else None
    )

    input_stream = sys.stdin if input == "-" else open(input, "r")
    output_stream = open(output, "w") if output else sys.stdout

    try:
        process_data(
            input_stream,
            output_stream,
            model_registry,
            categories_list,
            format,
            verbose,
            output_categories_list,
        )
    finally:
        if input != "-":
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
    output_categories: Optional[List[str]],
):
    reader = csv.DictReader(input_stream)
    writer = None

    inputs = [row["rate_name"] for row in reader]

    try:
        if verbose:
            console.print(f"[blue]Processing {len(inputs)} entries...[/blue]")
        results = make_prediction(model_registry, inputs, categories)

        if output_categories:
            results = [
                {k: v for k, v in result.items() if k in output_categories}
                for result in results
            ]

        if not writer:
            writer = _get_writer(output_stream, format, results[0].keys())

        _write_results(writer, results, format)

        if verbose:
            console.print(
                f"[green]Successfully processed {len(results)} predictions[/green]"
            )
    except PredictionError as e:
        if verbose:
            console.print(f"[red]Error during prediction: {e}[/red]")
        results = _get_partial_results(inputs)
        _write_results(writer, results, format)
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


def _write_results(writer, results: List[Dict[str, Any]], format: str):
    if format == "json":
        for result in results:
            json.dump(result, sys.stdout, cls=writer)
            sys.stdout.write("\n")
    elif format in ["csv", "tsv"]:
        writer.writerows(results)
    elif format == "yaml":
        yaml.dump_all(results, sys.stdout, Dumper=writer)
    elif format == "parquet":
        table = writer(pd.DataFrame(results))
        pq.write_table(table, sys.stdout.buffer)
    sys.stdout.flush()


def _get_partial_results(inputs: List[str]) -> List[Dict[str, Any]]:
    return [
        {"input": input, "predicted_category": "UNKNOWN", "confidence": 0.0}
        for input in inputs
    ]


def _prepare_output_path(output: str, format: str) -> str:
    file_name, file_ext = os.path.splitext(output)
    if file_ext.lower() != f".{format}":
        return f"{file_name}.{format}"
    return output


def _save_results(results: List[Dict[str, Any]], output_path: str, format: str) -> None:
    """Save results to a file in the specified format."""
    if format == "json":
        with open(output_path, "w") as f:
            json.dump(results, f, indent=2)
    elif format == "csv":
        with open(output_path, "w", newline="") as f:
            writer = csv.DictWriter(f, fieldnames=results[0].keys())
            writer.writeheader()
            writer.writerows(results)
    elif format == "tsv":
        with open(output_path, "w", newline="") as f:
            writer = csv.DictWriter(f, fieldnames=results[0].keys(), delimiter="\t")
            writer.writeheader()
            writer.writerows(results)
    elif format == "yaml":
        with open(output_path, "w") as f:
            yaml.dump(results, f)
    elif format == "parquet":
        df: pd.DataFrame = pd.DataFrame(results)
        table: pa.Table = pa.Table.from_pandas(df)
        pq.write_table(table, output_path)


def _display_results(results: List[Dict[str, Any]], format: str) -> None:
    """Display results in the specified format to stdout."""
    if not results:
        console.print("[yellow]No results to display.[/yellow]")
        return

    if format == "json":
        console.print(json.dumps(results, indent=2))
    elif format in ["csv", "tsv"]:
        delimiter = "," if format == "csv" else "\t"
        output = StringIO()
        writer = csv.DictWriter(
            output, fieldnames=results[0].keys(), delimiter=delimiter
        )
        writer.writeheader()
        writer.writerows(results)
        console.print(output.getvalue())
    elif format == "yaml":
        console.print(yaml.dump(results))
    elif format == "parquet":
        console.print(
            "[yellow]Parquet format is not supported for stdout display.[/yellow]"
        )
        console.print(
            "[yellow]Please specify an output file to save in Parquet format.[/yellow]"
        )
    else:
        _display_results_table(results)


def _display_results_table(results: List[Dict[str, Any]]) -> None:
    """Display results in a rich table."""
    table = Table(show_header=True, header_style="bold magenta")
    for key in results[0].keys():
        table.add_column(key)

    for result in results:
        row = [str(value) for value in result.values()]
        table.add_row(*row)

    console.print(table)


if __name__ == "__main__":
    cli()
