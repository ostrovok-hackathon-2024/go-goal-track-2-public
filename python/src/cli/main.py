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
from tenacity import retry, stop_after_attempt, wait_exponential, retry_if_exception_type

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
    reraise=True
)
def make_prediction(model_registry: ModelRegistry, inputs: List[str], categories: Optional[List[str]]) -> List[Dict[str, Any]]:
    try:
        return model_registry.predict(inputs, categories)
    except Exception as e:
        raise PredictionError(f"Prediction failed: {str(e)}")

@click.group()
@click.option(
    "--config",
    "-cfg",
    type=click.Path(exists=True),
    help="Path to the configuration file",
    default="config.yaml"
)
@click.option(
    "--verbose",
    "-v",
    is_flag=True,
    help="Enable verbose output"
)
@click.pass_context
def cli(ctx: click.Context, config: str, verbose: bool) -> None:
    """Hotel room rate classification CLI"""
    ctx.ensure_object(dict)
    ctx.obj['settings'] = load_settings(config)
    ctx.obj['model_registry'] = ModelRegistry(ctx.obj['settings'].MODELS_DIR, ctx.obj['settings'].CATEGORIES)
    ctx.obj['verbose'] = verbose

    if verbose:
        console.print(f"[blue]Loaded configuration from: {config}[/blue]")
        console.print(f"[blue]Models directory: {ctx.obj['settings'].MODELS_DIR}[/blue]")
        console.print(f"[blue]Categories: {ctx.obj['settings'].CATEGORIES}[/blue]")

@cli.command()
@click.argument("input")
@click.option(
    "--categories", "-c", multiple=True, help="Specify categories for prediction"
)
@click.option(
    "--output",
    "-o",
    type=click.Path(),
    help="Output file path (without extension)",
)
@click.option(
    "--format",
    "-f",
    type=click.Choice(["json", "csv", "tsv", "yaml", "parquet"]),
    default="json",
    help="Output file format",
)
@click.pass_context
def predict(ctx: click.Context, input: str, categories: Tuple[str, ...], output: Optional[str], format: str) -> None:
    """Predict categories for input (string or CSV file)"""
    categories_list: Optional[List[str]] = list(categories) if categories else None
    verbose: bool = ctx.obj['verbose']

    # Read input
    try:
        inputs: List[str] = _read_input(input, verbose)
    except Exception as e:
        console.print(f"[red]Error reading input: {e}[/red]")
        os._exit(1)

    # Make predictions
    try:
        if verbose:
            console.print("[blue]Starting prediction process...[/blue]")
        results: List[Dict[str, Any]] = make_prediction(ctx.obj['model_registry'], inputs, categories_list)
        console.print(f"[green]Successfully made predictions[/green]")
        if verbose:
            console.print(f"[blue]Number of predictions: {len(results)}[/blue]")
    except PredictionError as e:
        console.print(f"[red]Error during prediction after retries: {e}[/red]")
        console.print("[yellow]Attempting to proceed with partial results...[/yellow]")
        results = _get_partial_results(inputs)
    except Exception as e:
        console.print(f"[red]Unexpected error during prediction: {e}[/red]")
        return

    # Output results
    if output:
        output_path: str = _prepare_output_path(output, format)
        try:
            _save_results(results, output_path, format)
            console.print(f"[green]Results saved to {output_path}[/green]")
            if verbose:
                console.print(f"[blue]Saved {len(results)} results in {format.upper()} format[/blue]")
        except Exception as e:
            console.print(f"[red]Error saving results: {e}[/red]")
    else:
        _display_results(results)

def _read_input(input: str, verbose: bool) -> List[str]:
    if input.endswith(".csv"):
        try:
            df: pd.DataFrame = pd.read_csv(input)
            df["rate_name"] = df["rate_name"].fillna("")
            inputs: List[str] = df["rate_name"].tolist()
            console.print(f"[blue]Loaded {len(inputs)} entries from {input}[/blue]")
            if verbose:
                console.print(f"[blue]First 5 entries: {inputs[:5]}[/blue]")
            return inputs
        except Exception as e:
            raise Exception(f"Error reading CSV file: {e}")
    else:
        inputs: List[str] = [input.strip()]
        console.print("[blue]Loaded input string for prediction[/blue]")
        if verbose:
            console.print(f"[blue]Input: {inputs[0]}[/blue]")
        return inputs

def _get_partial_results(inputs: List[str]) -> List[Dict[str, Any]]:
    return [{"input": input, "predicted_category": "UNKNOWN", "confidence": 0.0} for input in inputs]

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
            writer = csv.DictWriter(f, fieldnames=results[0].keys(), delimiter='\t')
            writer.writeheader()
            writer.writerows(results)
    elif format == "yaml":
        with open(output_path, "w") as f:
            yaml.dump(results, f)
    elif format == "parquet":
        df: pd.DataFrame = pd.DataFrame(results)
        table: pa.Table = pa.Table.from_pandas(df)
        pq.write_table(table, output_path)

def _display_results(results: List[Dict[str, Any]]) -> None:
    """Display results in a rich table."""
    if not results:
        console.print("[yellow]No results to display.[/yellow]")
        return

    table = Table(show_header=True, header_style="bold magenta")
    for key in results[0].keys():
        table.add_column(key)

    for result in results:
        row = [str(value) for value in result.values()]
        table.add_row(*row)

    console.print(table)

if __name__ == "__main__":
    cli()
