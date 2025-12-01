"""Export production data to CSV files for ETL migration.

Usage:
    python3 scripts/etl.py

Exports business tables to CSV files in ./exports/ using psql.
"""

import subprocess
import os
from pathlib import Path

# Database connection
DB_HOST = os.getenv("DB_HOST", "localhost")
DB_PORT = "5432" #6543 for test and prod
DB_NAME = os.getenv("DB_NAME", "")
DB_USER = os.getenv("DB_USER", "postgres")
DB_PASS = os.getenv("DB_PASS", "postgres")

# Tables to export (in dependency order for later import)
# Only business data - Status/Priority tables are seeded by migrations
TABLES_TO_EXPORT = [
    "Tenant",
    "User",
    "Organization",
    "Lead",
    "Job",
    "Deal",
]

EXPORT_DIR = Path(__file__).parent.parent / "exports"


def export_table(table_name: str) -> int:
    """Export a single table to CSV file using psql. Returns row count."""
    output_file = EXPORT_DIR / f"{table_name}.csv"

    # Build psql command to export as CSV
    query = f"\\COPY \"{table_name}\" TO STDOUT WITH CSV HEADER"

    env = os.environ.copy()
    env["PGPASSWORD"] = DB_PASS

    try:
        result = subprocess.run(
            [
                "psql",
                "-h", DB_HOST,
                "-p", DB_PORT,
                "-U", DB_USER,
                "-d", DB_NAME,
                "-c", query,
            ],
            env=env,
            capture_output=True,
            text=True,
        )

        if result.returncode != 0:
            print(f"  {table_name}: ERROR - {result.stderr.strip()}")
            return 0

        # Write output to file
        with open(output_file, "w") as f:
            f.write(result.stdout)

        # Count rows (subtract 1 for header)
        row_count = len(result.stdout.strip().split("\n")) - 1 if result.stdout.strip() else 0

        if row_count > 0:
            print(f"  {table_name}: {row_count} rows -> {output_file.name}")
        else:
            print(f"  {table_name}: 0 rows (skipped)")

        return row_count

    except FileNotFoundError:
        print(f"  ERROR: psql not found. Install postgresql-client.")
        return 0
    except Exception as e:
        print(f"  {table_name}: ERROR - {e}")
        return 0


def main():
    # Create export directory
    EXPORT_DIR.mkdir(exist_ok=True)

    print(f"Exporting tables to {EXPORT_DIR}/")
    print("-" * 50)

    total_rows = 0
    for table in TABLES_TO_EXPORT:
        count = export_table(table)
        total_rows += count

    print("-" * 50)
    print(f"Export complete: {total_rows} total rows")


if __name__ == "__main__":
    main()
