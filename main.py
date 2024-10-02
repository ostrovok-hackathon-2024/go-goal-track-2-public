import numpy as np
import pandas as pd
import json
from datetime import datetime
from difflib import SequenceMatcher
from collections import defaultdict

categories = [
    "capacity",
    "quality",
    "view",
    "bedding",
    "balcony",
    "bedrooms",
    "club",
    "floor",
    "bathroom",
    "class",
]


def calculate_similarity(a, b):
    return SequenceMatcher(None, str(a), str(b)).ratio()


def compare_csv_files(file1_path, file2_path, threshold=0.8):
    # Read CSV files into pandas DataFrames
    df1 = pd.read_csv(file1_path, keep_default_na=False)
    df2 = pd.read_csv(file2_path, keep_default_na=False)

    # Sort columns alphabetically
    df1 = df1.reindex(sorted(df1.columns), axis=1)
    df2 = df2.reindex(sorted(df2.columns), axis=1)

    # Sort rows based on 'rate_name' column
    df1 = df1.sort_values("rate_name").reset_index(drop=True)
    df2 = df2.sort_values("rate_name").reset_index(drop=True)

    total_rows = len(df1)
    error_rows = 0
    column_errors = defaultdict(int)
    mismatches = []

    for index, (_, row1) in enumerate(df1.iterrows()):
        row2 = df2.iloc[index]
        row_errors = 0
        row_mismatches = {}

        for cat in categories:
            if cat in df1.columns and cat in df2.columns:
                val1 = str(row1.get(cat, ""))
                val2 = str(row2.get(cat, ""))
                val1 = "" if val1 in ["undefined", "nan", "None"] else val1
                val2 = "" if val2 in ["undefined", "nan", "None"] else val2

                similarity = calculate_similarity(val1, val2)
                if similarity < threshold:
                    row_errors += 1
                    column_errors[cat] += 1
                    row_mismatches[cat] = {"expected": val2, "received": val1}

        if row_errors > 0:
            error_rows += 1
            mismatches.append(
                {
                    "row": index + 1,
                    "rate_name": row1.get("rate_name", "N/A"),
                    "mismatches": row_mismatches,
                }
            )

    overall_error_rate = error_rows / total_rows
    print(f"\nOverall error rate: {overall_error_rate:.2f}")
    print(f"Total rows: {total_rows}")
    print(f"Rows with errors: {error_rows}")

    print("\nColumn Error Statistics:")
    for cat, error_count in column_errors.items():
        column_error_rate = error_count / total_rows
        print(
            f"Column {cat}: Error rate = {column_error_rate:.2f}, Errors = {error_count}"
        )

    with open("mismatches.json", "w") as f:
        json.dump(mismatches, f, indent=2)

    print("\nMismatches have been saved to 'mismatches.json'")


# Usage
file1_path = "rates_dirty.csv"
file2_path = "predictions.csv"

compare_csv_files(file1_path, file2_path)
