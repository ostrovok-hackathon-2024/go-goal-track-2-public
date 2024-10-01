import click
import csv
import json

from shared.models.registry import ModelRegistry

from shared.config.config import settings
import pandas as pd


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
        df = pd.read_csv(input)
        # Replace NaN in 'rate_name' with empty string
        df["rate_name"] = df["rate_name"].fillna("")
        inputs = df["rate_name"].tolist()
    else:
        inputs = [input.strip()]

    # Make predictions
    results = model_registry.predict(inputs, categories)

    # Output results
    if output:
        if input.endswith(".csv"):
            _save_csv(results, output)
        else:
            _save_json(results, output)
        click.echo(f"Results saved to {output}")
    else:
        click.echo(click.style("Prediction Results:", fg="green", bold=True))
        for result in results:
            click.echo(result)


def _save_csv(results, output_path):
    with open(output_path, "w", newline="") as f:
        writer = csv.DictWriter(f, fieldnames=results[0].keys())
        writer.writeheader()
        writer.writerows(results)


def _save_json(results, output_path):
    with open(output_path, "w") as f:
        json.dump(results, f, indent=2)


if __name__ == "__main__":
    cli()
