#!/usr/bin/env python3
"""Reset test database - drops all tables and reruns migrations + seeders.

This script is intended for test/development environments ONLY.
It will refuse to run in production.

Usage:
    python scripts/reset_test_db.py

    # Or with explicit confirmation bypass (for CI/CD):
    python scripts/reset_test_db.py --yes
"""

import os
import sys
from pathlib import Path

# Add parent directory to path to import from agent package
sys.path.insert(0, str(Path(__file__).parent.parent / "src"))

try:
    from dotenv import load_dotenv
    load_dotenv(override=True)
except ImportError:
    pass


ALLOWED_ENVIRONMENTS = ["development", "test", "testing", "local", "ci"]
DANGEROUS_HOSTS = ["prod", "production", "live", "main"]


def _import_psycopg2():
    """Import psycopg2, return None if not available."""
    try:
        import psycopg2
        return psycopg2
    except ImportError:
        print("Error: psycopg2 not installed. Run: pip install psycopg2-binary")
        return None


def _is_safe_environment() -> bool:
    """Check if we're in a safe (non-production) environment."""
    environment = os.getenv("ENVIRONMENT", "").lower()
    postgres_host = os.getenv("POSTGRES_HOST", "").lower()
    postgres_db = os.getenv("POSTGRES_DB", "").lower()

    # Check environment variable
    if environment and environment not in ALLOWED_ENVIRONMENTS:
        print(f"Error: ENVIRONMENT '{environment}' is not in allowed list: {ALLOWED_ENVIRONMENTS}")
        return False

    # Check for production indicators in host
    for danger in DANGEROUS_HOSTS:
        if danger in postgres_host:
            print(f"Error: POSTGRES_HOST '{postgres_host}' contains '{danger}' - refusing to reset")
            return False

    # Check for production indicators in database name
    for danger in DANGEROUS_HOSTS:
        if danger in postgres_db:
            print(f"Error: POSTGRES_DB '{postgres_db}' contains '{danger}' - refusing to reset")
            return False

    # Check for Supabase production (non-local)
    supabase_url = os.getenv("SUPABASE_URL", "")
    if supabase_url and "supabase.co" in supabase_url:
        print("Error: Detected Supabase cloud connection - refusing to reset")
        print("       This script only works with local Postgres")
        return False

    return True


def _get_postgres_connection(psycopg2):
    """Get connection to Postgres database."""
    postgres_host = os.getenv("POSTGRES_HOST")
    postgres_port = int(os.getenv("POSTGRES_PORT", 5432))
    postgres_user = os.getenv("POSTGRES_USER")
    postgres_password = os.getenv("POSTGRES_PASSWORD")
    postgres_db = os.getenv("POSTGRES_DB")

    print(f"Connecting to: {postgres_host}:{postgres_port}/{postgres_db}")

    return psycopg2.connect(
        host=postgres_host,
        port=postgres_port,
        database=postgres_db,
        user=postgres_user,
        password=postgres_password,
        connect_timeout=10
    )


def _drop_all_tables(cursor, conn):
    """Drop all tables in the public schema."""
    print("\nDropping all tables in public schema...")

    # Get all table names
    cursor.execute("""
        SELECT tablename FROM pg_tables
        WHERE schemaname = 'public';
    """)
    tables = [row[0] for row in cursor.fetchall()]

    if not tables:
        print("  No tables found to drop")
        return

    print(f"  Found {len(tables)} table(s) to drop:")
    for table in tables:
        print(f"    - {table}")

    # Drop all tables with CASCADE
    cursor.execute("DROP SCHEMA public CASCADE;")
    cursor.execute("CREATE SCHEMA public;")
    cursor.execute("GRANT ALL ON SCHEMA public TO postgres;")
    cursor.execute("GRANT ALL ON SCHEMA public TO public;")
    conn.commit()

    print("  All tables dropped successfully")


def _run_migrations(migrations_dir: Path, cursor, conn):
    """Run all migration files."""
    print("\nRunning migrations...")

    migration_files = sorted(migrations_dir.glob("*.sql"))
    if not migration_files:
        print("  No migration files found")
        return

    print(f"  Found {len(migration_files)} migration(s):")

    # Create tracking table
    cursor.execute("""
        CREATE TABLE IF NOT EXISTS schema_migrations (
            migration_name TEXT PRIMARY KEY,
            applied_at TIMESTAMP DEFAULT NOW()
        );
    """)
    conn.commit()

    for migration_file in migration_files:
        print(f"    Applying: {migration_file.name}")
        sql_content = migration_file.read_text()
        cursor.execute(sql_content)
        cursor.execute(
            "INSERT INTO schema_migrations (migration_name) VALUES (%s);",
            (migration_file.name,)
        )
        conn.commit()

    print("  All migrations applied successfully")


def _run_seeders(seeders_dir: Path, cursor, conn):
    """Run all seeder files."""
    print("\nRunning seeders...")

    seeder_files = sorted(seeders_dir.glob("*.sql"))
    if not seeder_files:
        print("  No seeder files found")
        return

    print(f"  Found {len(seeder_files)} seeder(s):")

    # Create tracking table
    cursor.execute("""
        CREATE TABLE IF NOT EXISTS schema_seeders (
            seeder_name TEXT PRIMARY KEY,
            applied_at TIMESTAMP DEFAULT NOW()
        );
    """)
    conn.commit()

    for seeder_file in seeder_files:
        print(f"    Running: {seeder_file.name}")
        sql_content = seeder_file.read_text()
        cursor.execute(sql_content)
        cursor.execute(
            "INSERT INTO schema_seeders (seeder_name) VALUES (%s);",
            (seeder_file.name,)
        )
        conn.commit()

    print("  All seeders applied successfully")


def reset_database(skip_confirmation: bool = False):
    """Reset the test database."""
    print("=" * 60)
    print("Test Database Reset Tool")
    print("=" * 60)

    # Safety check
    if not _is_safe_environment():
        print("\nAborting: Environment safety check failed")
        return False

    psycopg2 = _import_psycopg2()
    if not psycopg2:
        return False

    # Get directories
    script_dir = Path(__file__).parent
    docker_base = Path("/app")

    if docker_base.exists():
        migrations_dir = docker_base / "migrations"
        seeders_dir = docker_base / "seeders"
    else:
        agent_dir = script_dir.parent
        migrations_dir = agent_dir / "migrations"
        seeders_dir = agent_dir / "seeders"

    print(f"\nMigrations: {migrations_dir}")
    print(f"Seeders: {seeders_dir}")

    # Confirmation
    if not skip_confirmation:
        env = os.getenv("ENVIRONMENT", "unknown")
        db = os.getenv("POSTGRES_DB", "unknown")
        print(f"\n*** WARNING ***")
        print(f"This will DROP ALL TABLES in database '{db}' (env: {env})")
        print(f"All data will be permanently lost!")
        response = input("\nType 'yes' to confirm: ")
        if response.lower() != "yes":
            print("Aborted by user")
            return False

    try:
        conn = _get_postgres_connection(psycopg2)
        cursor = conn.cursor()

        # Drop all tables
        _drop_all_tables(cursor, conn)

        # Run migrations
        _run_migrations(migrations_dir, cursor, conn)

        # Run seeders
        _run_seeders(seeders_dir, cursor, conn)

        cursor.close()
        conn.close()

        print("\n" + "=" * 60)
        print("Database reset complete!")
        print("=" * 60)
        return True

    except Exception as e:
        print(f"\nError during reset: {e}")
        return False


def main():
    """Main entry point."""
    skip_confirmation = "--yes" in sys.argv or "-y" in sys.argv

    success = reset_database(skip_confirmation=skip_confirmation)
    if not success:
        sys.exit(1)


if __name__ == "__main__":
    main()
