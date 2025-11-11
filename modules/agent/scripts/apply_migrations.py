#!/usr/bin/env python3
"""Apply Supabase migrations from migrations directory.

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


def _ensure_migrations_table(cursor):
    """Create migrations tracking table if it doesn't exist."""
    cursor.execute("""
        CREATE TABLE IF NOT EXISTS schema_migrations (
            migration_name TEXT PRIMARY KEY,
            applied_at TIMESTAMP DEFAULT NOW()
        );
    """)


def _ensure_seeders_table(cursor):
    """Create seeders tracking table if it doesn't exist."""
    cursor.execute("""
        CREATE TABLE IF NOT EXISTS schema_seeders (
            seeder_name TEXT PRIMARY KEY,
            applied_at TIMESTAMP DEFAULT NOW()
        );
    """)


def _get_applied_migrations(cursor) -> set:
    """Get set of already-applied migration names."""
    _ensure_migrations_table(cursor)
    cursor.execute("SELECT migration_name FROM schema_migrations;")
    return {row[0] for row in cursor.fetchall()}


def _record_migration(cursor, conn, migration_name: str):
    """Record that a migration has been applied."""
    _ensure_migrations_table(cursor)
    cursor.execute(
        "INSERT INTO schema_migrations (migration_name) VALUES (%s) ON CONFLICT DO NOTHING;",
        (migration_name,)
    )
    conn.commit()


def apply_migrations_to_local_postgres(migrations_dir: Path) -> bool:
    """Apply migrations to local Postgres database."""
    psycopg2 = _import_psycopg2()
    if not psycopg2:
        return False

    # Get local Postgres connection details
    postgres_host = os.getenv("POSTGRES_HOST", "postgres")
    postgres_port = int(os.getenv("POSTGRES_PORT", "5432"))
    postgres_user = os.getenv("POSTGRES_USER", "postgres")
    postgres_password = os.getenv("POSTGRES_PASSWORD", "postgres")
    postgres_db = os.getenv("POSTGRES_DB", "development")

    try:
        print(f"\nConnecting to local Postgres: {postgres_host}:{postgres_port}/{postgres_db}")
        conn = psycopg2.connect(
            host=postgres_host,
            port=postgres_port,
            database=postgres_db,
            user=postgres_user,
            password=postgres_password,
            connect_timeout=10
        )
    except psycopg2.OperationalError as e:
        print(f"\n❌ Failed to connect to local Postgres: {e}")
        return False

    cursor = conn.cursor()
    try:
        # Ensure tracking tables exist (migrations and seeders)
        _ensure_migrations_table(cursor)
        _ensure_seeders_table(cursor)
        conn.commit()
        
        # Get list of already-applied migrations
        applied_migrations = _get_applied_migrations(cursor)
        
        # Get all migration files
        migration_files = get_migration_files(migrations_dir)
        if not migration_files:
            cursor.close()
            conn.close()
            print("✓ No migration files found")
            return True

        # Filter to only unapplied migrations
        unapplied_migrations = [
            mf for mf in migration_files
            if mf.name not in applied_migrations
        ]

        if not unapplied_migrations:
            print("✓ Database schema up to date, no migrations needed")
            cursor.close()
            conn.close()
            return True

        print(f"\n📋 Found {len(unapplied_migrations)} migration(s) to apply:")
        for mf in unapplied_migrations:
            print(f"  - {mf.name}")

        # Apply each unapplied migration
        for migration_file in unapplied_migrations:
            print(f"\nApplying: {migration_file.name}")
            sql_content = migration_file.read_text()
            cursor.execute(sql_content)
            _record_migration(cursor, conn, migration_file.name)
            print(f"✓ Successfully applied {migration_file.name}")

        cursor.close()
        conn.close()

        print("\n✓ All migrations applied successfully!")
        return True
    except Exception as e:
        print(f"\n❌ Error applying migrations: {e}")
        conn.rollback()
        cursor.close()
        conn.close()
        return False


def apply_migrations_with_psycopg2(migrations_dir: Path) -> bool:
    """Apply migrations using psycopg2 (direct PostgreSQL connection) - Supabase only."""
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
    """Main entry point for local development migrations.
    
    Note: Supabase migrations are handled in Azure Pipeline for production.
    This script only applies migrations to local Postgres.
    """
    print("="*60)
    print("Local Postgres Migration Tool")
    print("="*60)

    # Get migrations directory
    # In Docker: /app/migrations
    # Locally: modules/agent/migrations/
    script_dir = Path(__file__).parent

    # Try Docker path first
    docker_migrations_dir = Path("/app/migrations")
    if docker_migrations_dir.exists():
        migrations_dir = docker_migrations_dir
        print(f"\n✓ Running in Docker container")
    else:
        # Local development path - migrations are in modules/agent/migrations/
        agent_dir = script_dir.parent
        migrations_dir = agent_dir / "migrations"
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
        print("\n✅ psycopg2 is available - applying migrations to local Postgres")
        success = apply_migrations_to_local_postgres(migrations_dir)
        if not success:
            print("\n❌ Migration application failed")
            sys.exit(1)
        return

    print("\n❌ psycopg2 not installed - cannot apply migrations automatically")
    print("   Please install psycopg2 or apply migrations manually via SQL")
    sys.exit(1)


if __name__ == "__main__":
    main()
