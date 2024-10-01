import csv
import json

import click
import pandas as pd
from rich.console import Console
from rich.table import Table

from shared.config.config import settings
from shared.models.registry import ModelRegistry

# Initialize Rich Console
console = Console()

# Initialize Model Registry
model_registry = ModelRegistry(settings.MODELS_DIR, settings.CATEGORIES)


@click.group()
def cli():
    """Hotel room rate classification CLI"""
    pass


@cli.command()
@click.argument("input")
@click.option(
    "--categories", "-c", multiple=True, help="Specify categories for prediction"
)
@click.option(
    "--output",
    "-o",
    type=click.Path(),
    help="Output file path (JSON for string, CSV for file input)",
)
def predict(input, categories, output):
    """Predict categories for input (string or CSV file)"""
    categories = list(categories) if categories else None

    # Read input
    if input.endswith(".csv"):
        try:
            df = pd.read_csv(input)
            df["rate_name"] = df["rate_name"].fillna("")
            inputs = df["rate_name"].tolist()
            console.print(f"[blue]Loaded {len(inputs)} entries from {input}[/blue]")
        except Exception as e:
            console.print(f"[red]Error reading CSV file: {e}[/red]")
            return
    else:
        inputs = [input.strip()]
        console.print(f"[blue]Loaded input string for prediction[/blue]")

    # Make predictions
    try:
        results = model_registry.predict(inputs, categories)
        console.print(f"[green]Successfully made predictions[/green]")
    except Exception as e:
        console.print(f"[red]Error during prediction: {e}[/red]")
        return

    # Output results
    if output:
        try:
            if output.endswith(".csv"):
                _save_csv(results, output)
            else:
                _save_json(results, output)
            console.print(f"[green]Results saved to {output}[/green]")
        except Exception as e:
            console.print(f"[red]Error saving results: {e}[/red]")
    else:
        _display_results(results)


def _save_csv(results, output_path):
    """Save results to a CSV file."""
    with open(output_path, "w", newline="") as f:
        writer = csv.DictWriter(f, fieldnames=results[0].keys())
        writer.writeheader()
        writer.writerows(results)


def _save_json(results, output_path):
    """Save results to a JSON file."""
    with open(output_path, "w") as f:
        json.dump(results, f, indent=2)


def _display_results(results):
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
