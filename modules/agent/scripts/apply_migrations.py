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


def _import_psycopg2():
    """Import psycopg2, return None if not available."""
    try:
        import psycopg2
        return psycopg2
    except ImportError:
        print("\nNote: For automatic migration application, install psycopg2:")
        print("  pip install psycopg2-binary")
        return None


def _parse_pooler_url(pooler_url: str, project_ref: str):
    """Parse Supabase pooler URL into connection params, return None if invalid."""
    import urllib.parse
    try:
        parsed = urllib.parse.urlparse(pooler_url)
        if not parsed.hostname:
            return None
        return {
            "method": "pooler",
            "host": parsed.hostname,
            "port": parsed.port or 6543,
            "user": parsed.username or f"postgres.{project_ref}",
        }
    except Exception:
        return None


def _build_connection_methods(project_ref: str):
    """Build list of connection methods to try."""
    methods = [{
        "method": "hostname",
        "host": f"db.{project_ref}.supabase.co",
        "port": 5432,
        "user": "postgres",
    }]
    
    pooler_url = os.getenv("SUPABASE_POOLER_URL", "")
    if not pooler_url:
        return methods
    
    pooler_config = _parse_pooler_url(pooler_url, project_ref)
    if pooler_config:
        return [pooler_config] + methods
    
    return methods


def _try_connect(psycopg2, method_config: dict, db_password: str):
    """Attempt connection, return connection or None."""
    print(f"\nTrying {method_config['method']} connection to {method_config['host']}:{method_config['port']}...")
    try:
        conn = psycopg2.connect(
            host=method_config["host"],
            port=method_config["port"],
            database="postgres",
            user=method_config["user"],
            password=db_password,
            connect_timeout=10
        )
        print(f"✅ Connected via {method_config['method']}!")
        return conn
    except psycopg2.OperationalError as e:
        print(f"❌ {method_config['method']} connection failed: {e}")
        return None
    except Exception as e:
        print(f"❌ Unexpected error with {method_config['method']}: {e}")
        return None


def _connect_to_database(psycopg2, project_ref: str, db_password: str):
    """Connect to database using available methods."""
    methods = _build_connection_methods(project_ref)
    
    for method_config in methods:
        conn = _try_connect(psycopg2, method_config, db_password)
        if conn:
            return conn
    
    return None


def _apply_migration_files(cursor, conn, migration_files: List[Path]):
    """Apply all migration files to database."""
    for migration_file in migration_files:
        print(f"\nApplying: {migration_file.name}")
        sql_content = migration_file.read_text()
        cursor.execute(sql_content)
        conn.commit()
        print(f"✓ Successfully applied {migration_file.name}")


def apply_migrations_with_psycopg2(migrations_dir: Path) -> bool:
    """Apply migrations using psycopg2 (direct PostgreSQL connection)."""
    psycopg2 = _import_psycopg2()
    if not psycopg2:
        return False

    db_password = os.getenv("SUPABASE_DB_PASSWORD")
    if not db_password:
        print("\nNote: Set SUPABASE_DB_PASSWORD for automatic migration application")
        return False

    supabase_url = os.getenv("SUPABASE_URL", "")
    if "supabase.co" not in supabase_url:
        print("Invalid Supabase URL format")
        return False

    project_ref = supabase_url.split("//")[1].split(".")[0]
    print(f"\nAttempting connection to db.{project_ref}.supabase.co...")

    conn = _connect_to_database(psycopg2, project_ref, db_password)
    if not conn:
        print("\n❌ All connection methods failed")
        print("\n💡 Solutions:")
        print("   1. Use Supabase Dashboard SQL Editor to run migrations manually")
        print("   2. Set SUPABASE_POOLER_URL env var with connection pooling URL")
        print("   3. Enable IPv6 in Docker (requires Docker daemon configuration)")
        print("   4. Run migrations from host machine instead of Docker")
        return False

    migration_files = get_migration_files(migrations_dir)
    if not migration_files:
        conn.close()
        print("No migration files found")
        return True

    cursor = conn.cursor()
    _apply_migration_files(cursor, conn, migration_files)
    cursor.close()
    conn.close()

    print("\n✓ All migrations applied successfully!")
    return True


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
    if not docker_migrations_dir.exists():
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
    psycopg2 = _import_psycopg2()
    if psycopg2:
        print("\n✅ psycopg2 is available - applying migrations automatically")
        success = apply_migrations_with_psycopg2(migrations_dir)
        if not success:
            print("\n❌ Migration application failed")
            sys.exit(1)
        return

    print("\n⚠️  psycopg2 not installed - showing manual instructions")
    client = get_supabase_client()
    print("✓ Supabase credentials validated")

    for migration_file in migration_files:
        if not apply_migration(client, migration_file):
            print("\n❌ Migration application failed")
            sys.exit(1)


if __name__ == "__main__":
    main()
