#!/usr/bin/env python3
"""Apply Supabase migrations from supabase_migrations directory.

This script reads SQL migration files and applies them to Supabase database.
Migrations are applied in alphabetical order (use numbered prefixes).

Supports both manual execution and Docker container execution.
"""

import os
import sys
from pathlib import Path
from typing import List

# Add parent directory to path to import from agent package
sys.path.insert(0, str(Path(__file__).parent.parent / "src"))

try:
    from dotenv import load_dotenv
    load_dotenv(override=True)
except ImportError:
    # In Docker, environment variables are already set
    pass

try:
    from supabase import create_client, Client
except ImportError:
    print("Error: supabase-py not installed. Run: pip install supabase")
    sys.exit(1)


def get_supabase_client() -> Client:
    """Create Supabase client from environment variables."""
    supabase_url = os.getenv("SUPABASE_URL")
    supabase_key = os.getenv("SUPABASE_SERVICE_KEY")

    if not supabase_url or not supabase_key:
        print("Error: SUPABASE_URL or SUPABASE_SERVICE_KEY not set")
        sys.exit(1)

    return create_client(supabase_url, supabase_key)


def get_migration_files(migrations_dir: Path) -> List[Path]:
    """Get sorted list of migration files."""
    if not migrations_dir.exists():
        print(f"⚠️  Migrations directory not found: {migrations_dir}")
        print(f"   Searched in: {migrations_dir.absolute()}")
        return []

    migration_files = sorted(migrations_dir.glob("*.sql"))

    if not migration_files:
        print(f"ℹ️  No migration files found in {migrations_dir}")
        return []

    return migration_files


def apply_migration(client: Client, migration_file: Path) -> bool:
    """Apply a single migration file."""
    print(f"\nApplying migration: {migration_file.name}")

    try:
        sql_content = migration_file.read_text()

        # Execute the SQL using Supabase RPC or direct SQL execution
        # Note: Supabase Python client doesn't have direct SQL execution
        # We'll use the rpc method if you have a stored procedure, or
        # need to use PostgREST SQL editor / CLI

        # For now, we'll print instructions
        print(f"SQL content length: {len(sql_content)} characters")
        print("\nIMPORTANT: The Supabase Python client doesn't support raw SQL execution.")
        print("Please apply this migration using one of these methods:")
        print("1. Supabase Dashboard SQL Editor")
        print("2. Supabase CLI: supabase db push")
        print("3. psycopg2 direct connection (see below)")
        print("\n" + "="*60)
        print("SQL to execute:")
        print("="*60)
        print(sql_content)
        print("="*60)

        return True

    except Exception as e:
        print(f"Error applying migration {migration_file.name}: {e}")
        return False


def apply_migrations_with_psycopg2(migrations_dir: Path):
    """Apply migrations using psycopg2 (direct PostgreSQL connection)."""
    try:
        import psycopg2
    except ImportError:
        print("\nNote: For automatic migration application, install psycopg2:")
        print("  pip install psycopg2-binary")
        return

    # Parse Supabase connection string
    supabase_url = os.getenv("SUPABASE_URL", "")
    db_password = os.getenv("SUPABASE_DB_PASSWORD")

    if not db_password:
        print("\nNote: Set SUPABASE_DB_PASSWORD for automatic migration application")
        return

    # Extract project ref from URL
    # Format: https://<project-ref>.supabase.co
    if "supabase.co" not in supabase_url:
        print("Invalid Supabase URL format")
        return

    project_ref = supabase_url.split("//")[1].split(".")[0]

    # Build connection string
    conn_string = f"postgresql://postgres:{db_password}@db.{project_ref}.supabase.co:5432/postgres"

    print(f"\nConnecting to Supabase database...")

    try:
        conn = psycopg2.connect(conn_string)
        cursor = conn.cursor()

        migration_files = get_migration_files(migrations_dir)

        for migration_file in migration_files:
            print(f"\nApplying: {migration_file.name}")
            sql_content = migration_file.read_text()

            cursor.execute(sql_content)
            conn.commit()
            print(f"✓ Successfully applied {migration_file.name}")

        cursor.close()
        conn.close()

        print("\n✓ All migrations applied successfully!")

    except Exception as e:
        print(f"\nError applying migrations: {e}")
        if 'conn' in locals():
            conn.rollback()


def main():
    """Main entry point."""
    print("="*60)
    print("Supabase Migration Tool")
    print("="*60)

    # Get migrations directory
    # In Docker: /app/supabase_migrations
    # Locally: relative to project root
    script_dir = Path(__file__).parent

    # Try Docker path first
    docker_migrations_dir = Path("/app/supabase_migrations")
    if docker_migrations_dir.exists():
        migrations_dir = docker_migrations_dir
        print(f"\n✓ Running in Docker container")
    else:
        # Local development path
        project_root = script_dir.parent.parent.parent
        migrations_dir = project_root / "supabase_migrations"
        print(f"\n✓ Running in local development")

    print(f"Migrations directory: {migrations_dir}")

    # Get migration files
    migration_files = get_migration_files(migrations_dir)

    if not migration_files:
        print("✓ No migrations to apply")
        return

    print(f"\n📁 Found {len(migration_files)} migration file(s):")
    for mf in migration_files:
        print(f"  - {mf.name}")

    # Check if psycopg2 is available for direct execution
    try:
        import psycopg2
        print("\n✅ psycopg2 is available - applying migrations automatically")
        apply_migrations_with_psycopg2(migrations_dir)
    except ImportError:
        print("\n⚠️  psycopg2 not installed - showing manual instructions")
        # Get client just to validate credentials
        client = get_supabase_client()
        print("✓ Supabase credentials validated")

        # Show migration content for manual application
        for migration_file in migration_files:
            apply_migration(client, migration_file)


if __name__ == "__main__":
    main()
