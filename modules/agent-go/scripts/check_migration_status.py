#!/usr/bin/env python3
"""Check database migration status for local Postgres.

Checks if tables exist to determine if migrations have been applied.
Note: Supabase is only used in production via Azure Pipeline.
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


def check_local_postgres():
    """Check migration status by checking if tables exist."""
    try:
        import psycopg2
        import time
    except ImportError:
        return None
    
    postgres_host = os.getenv("POSTGRES_HOST")
    postgres_port = int(os.getenv("POSTGRES_PORT"))
    postgres_user = os.getenv("POSTGRES_USER")
    postgres_password = os.getenv("POSTGRES_PASSWORD")
    postgres_db = os.getenv("POSTGRES_DB")
    
    # Retry connection up to 3 times with 2 second delays
    max_retries = 3
    for attempt in range(max_retries):
        try:
            conn = psycopg2.connect(
                host=postgres_host,
                port=postgres_port,
                database=postgres_db,
                user=postgres_user,
                password=postgres_password,
                connect_timeout=5
            )
            cursor = conn.cursor()
            
            cursor.execute("""
                SELECT EXISTS (
                    SELECT FROM information_schema.tables 
                    WHERE table_schema = 'public' 
                    AND table_name = 'Job'
                );
            """)
            table_exists = cursor.fetchone()[0]
            
            cursor.close()
            conn.close()
            
            return {"type": "local_postgres", "migrations_applied": table_exists}
        except psycopg2.OperationalError:
            if attempt >= max_retries - 1:
                return None
            time.sleep(2)
        except Exception:
            return None
    
    return None


def get_migration_files():
    """Get list of migration files."""
    docker_migrations_dir = Path("/app/migrations")
    script_dir = Path(__file__).parent
    agent_dir = script_dir.parent
    migrations_dir = docker_migrations_dir if docker_migrations_dir.exists() else agent_dir / "migrations"

    if not migrations_dir.exists():
        return []

    migration_files = sorted(migrations_dir.glob("*.sql"))
    return [f.name for f in migration_files]


def _print_migration_status(postgres_status: dict) -> None:
    """Print migration status message."""
    if postgres_status["migrations_applied"]:
        print("✓ Database schema up to date, migrations already applied")
        return
    migration_files = get_migration_files()
    if not migration_files:
        print("ℹ️  No migration files found")
        return
    print(f"⚠️  {len(migration_files)} migration file(s) found but not yet applied")
    print("   Migrations will be applied automatically on startup")


def main():
    """Main entry point."""
    postgres_status = check_local_postgres()
    if not postgres_status:
        print("ℹ️  Could not determine migration status (no database connection available)")
        return
    _print_migration_status(postgres_status)


if __name__ == "__main__":
    main()

