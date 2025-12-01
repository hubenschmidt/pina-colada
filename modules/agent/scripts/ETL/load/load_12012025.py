"""Import production data from CSV files for ETL migration.

Run this AFTER:
1. Running all database migrations
2. Running transform_prod_csv.py to transform old schema CSVs to new schema

Usage:
    cd modules/agent/scripts
    python3 transform_prod_csv.py   # Transform CSVs first
    python3 import_prod_12012025.py  # Then import
"""
import subprocess
import os
from pathlib import Path

# Target database (update these for your target environment)
DB_HOST = "aws-1-us-east-1.pooler.supabase.com"
DB_PORT = "6543"
DB_NAME = "postgres"
DB_USER = "postgres.ckyusrrjwwcgwrbcfsal"
DB_PASS = "ShortTimCrook123"

# DB_HOST = "localhost"
# DB_PORT = "5432"
# DB_NAME = "pina_colada"
# DB_USER = "postgres"
# DB_PASS = "postgres"

# Import order matters for foreign key constraints
TABLES_TO_IMPORT = [
    "Tenant",       # No dependencies
    "Account",      # Depends on Tenant
    "Individual",   # Depends on Account
    "User",         # Depends on Tenant, Individual
    "Role",         # Depends on Tenant
    "User_Role",    # Depends on User, Role
    "Project",      # Depends on Tenant, Status
    "Organization", # Depends on Account
    "Deal",         # Depends on Tenant, Status
    "Lead",         # Depends on Deal, Account, Status
    "Lead_Project", # Depends on Lead, Project
    "Job",          # Depends on Lead (id)
]

# Use transformed exports (run transform_prod_csv.py first)
IMPORT_DIR = Path(__file__).parent.parent / "exports/transformed"


def run_sql(sql: str) -> tuple[bool, str]:
    """Run a SQL command via psql."""
    env = os.environ.copy()
    env["PGPASSWORD"] = DB_PASS
    result = subprocess.run(
        ["psql", "-h", DB_HOST, "-p", DB_PORT, "-U", DB_USER, "-d", DB_NAME, "-c", sql],
        env=env,
        capture_output=True,
        text=True,
    )
    success = result.returncode == 0
    output = result.stdout + result.stderr
    return success, output


def import_table(table_name: str) -> tuple[int, str]:
    """Import a single table from CSV using psql \copy command."""
    csv_file = IMPORT_DIR / f"{table_name}.csv"

    if not csv_file.exists():
        return 0, f"CSV file not found: {csv_file}"

    # Read CSV to get column headers
    with open(csv_file, "r") as f:
        header_line = f.readline().strip()
        columns = header_line.split(",")
        row_count = sum(1 for _ in f)  # Count data rows

    if row_count == 0:
        return 0, "No data rows in CSV"

    # Build the \copy command
    columns_str = ", ".join(f'"{col}"' for col in columns)
    copy_cmd = f"\\COPY \"{table_name}\" ({columns_str}) FROM '{csv_file.absolute()}' WITH CSV HEADER"

    env = os.environ.copy()
    env["PGPASSWORD"] = DB_PASS

    result = subprocess.run(
        ["psql", "-h", DB_HOST, "-p", DB_PORT, "-U", DB_USER, "-d", DB_NAME, "-c", copy_cmd],
        env=env,
        capture_output=True,
        text=True,
    )

    if result.returncode != 0:
        return 0, f"Error: {result.stderr}"

    return row_count, "OK"


# Tables without an 'id' column (junction tables with composite keys)
TABLES_WITHOUT_ID = ["User_Role", "Lead_Project"]


def reset_sequence(table_name: str) -> None:
    """Reset the sequence for a table to max(id) + 1."""
    if table_name in TABLES_WITHOUT_ID:
        return  # Skip tables without id column

    sql = f"""
    SELECT setval(
        pg_get_serial_sequence('"{table_name}"', 'id'),
        COALESCE((SELECT MAX(id) FROM "{table_name}"), 1)
    );
    """
    run_sql(sql)


def main():
    print("=" * 60)
    print("Production Data Import - December 1, 2025")
    print("=" * 60)
    print(f"Target: {DB_USER}@{DB_HOST}:{DB_PORT}/{DB_NAME}")
    print(f"Source: {IMPORT_DIR}")
    print()

    # Check if transformed exports directory exists
    if not IMPORT_DIR.exists():
        print(f"Error: Transformed exports not found: {IMPORT_DIR}")
        print()
        print("Run the transform script first:")
        print("  python3 transform_prod_csv.py")
        return

    # Verify connection
    success, output = run_sql("SELECT 1;")
    if not success:
        print(f"Failed to connect to database: {output}")
        return

    print("Connected to database successfully.")
    print()

    # Clear existing data in reverse order (to handle FK constraints)
    print("Clearing existing data...")
    for table in reversed(TABLES_TO_IMPORT):
        success, output = run_sql(f'DELETE FROM "{table}";')
        if success:
            print(f"  Cleared {table}")
        else:
            print(f"  Warning clearing {table}: {output}")
    print()

    # Import tables in order
    print("Importing data...")
    total_rows = 0
    for table in TABLES_TO_IMPORT:
        count, status = import_table(table)
        total_rows += count
        print(f"  {table}: {count} rows - {status}")

        # Reset sequence after import
        if count > 0:
            reset_sequence(table)

    print()
    print(f"Total rows imported: {total_rows}")
    print("=" * 60)
    print("Import complete!")
    print()
    print("Note: Verify the data with:")
    print('  SELECT table_name, COUNT(*) FROM (')
    print('    SELECT \'Tenant\' as table_name, COUNT(*) as cnt FROM "Tenant"')
    print('    UNION ALL SELECT \'Account\', COUNT(*) FROM "Account"')
    print('    UNION ALL SELECT \'Individual\', COUNT(*) FROM "Individual"')
    print('    UNION ALL SELECT \'User\', COUNT(*) FROM "User"')
    print('    UNION ALL SELECT \'Organization\', COUNT(*) FROM "Organization"')
    print('    UNION ALL SELECT \'Deal\', COUNT(*) FROM "Deal"')
    print('    UNION ALL SELECT \'Lead\', COUNT(*) FROM "Lead"')
    print('    UNION ALL SELECT \'Job\', COUNT(*) FROM "Job"')
    print('  ) t GROUP BY table_name ORDER BY table_name;')


if __name__ == "__main__":
    main()
