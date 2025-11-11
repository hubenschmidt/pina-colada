#!/usr/bin/env python3
"""Check database migration status.

For local Postgres: Checks if tables exist.
For Supabase: Checks supabase_migrations.schema_migrations table.
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


def check_supabase_migrations():
    """Check migration status using Supabase migration tracking table."""
    try:
        from supabase import create_client, Client
        
        supabase_url = os.getenv("SUPABASE_URL")
        supabase_key = os.getenv("SUPABASE_SERVICE_KEY")
        
        if not supabase_url or not supabase_key:
            return None
        
        client = create_client(supabase_url, supabase_key)
        
        # Check if migrations have been applied by checking if the main table exists
        # Supabase CLI tracks migrations in supabase_migrations.schema_migrations,
        # but we can't easily query that via the Python client, so we check if
        # the main application table exists as a proxy
        try:
            # Try to query the main table - if it exists, migrations have been applied
            result = client.table("applied_jobs").select("id").limit(1).execute()
            migration_files = get_migration_files()
            return {"type": "supabase", "migrations_applied": True, "migration_files": migration_files}
        except Exception:
            # Table doesn't exist - migrations not applied yet
            migration_files = get_migration_files()
            return {"type": "supabase", "migrations_applied": False, "migration_files": migration_files}
    except ImportError:
        return None
    except Exception:
        return None


def check_local_postgres():
    """Check migration status by checking if tables exist."""
    try:
        import psycopg2
        
        # Check for local Postgres connection
        postgres_host = os.getenv("POSTGRES_HOST", "postgres")
        postgres_port = int(os.getenv("POSTGRES_PORT", "5432"))
        postgres_user = os.getenv("POSTGRES_USER", "postgres")
        postgres_password = os.getenv("POSTGRES_PASSWORD", "postgres")
        postgres_db = os.getenv("POSTGRES_DB", "pina_colada")
        
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
            
            # Check if applied_jobs table exists (main table from migrations)
            cursor.execute("""
                SELECT EXISTS (
                    SELECT FROM information_schema.tables 
                    WHERE table_schema = 'public' 
                    AND table_name = 'applied_jobs'
                );
            """)
            table_exists = cursor.fetchone()[0]
            
            cursor.close()
            conn.close()
            
            return {"type": "local_postgres", "migrations_applied": table_exists}
        except psycopg2.OperationalError:
            # Can't connect to Postgres
            return None
    except ImportError:
        return None
    except Exception:
        return None


def get_migration_files():
    """Get list of migration files."""
    # Try Docker path first
    docker_migrations_dir = Path("/app/supabase_migrations")
    if docker_migrations_dir.exists():
        migrations_dir = docker_migrations_dir
    else:
        # Local development path
        script_dir = Path(__file__).parent
        project_root = script_dir.parent.parent.parent
        migrations_dir = project_root / "supabase_migrations"
    
    if not migrations_dir.exists():
        return []
    
    migration_files = sorted(migrations_dir.glob("*.sql"))
    return [f.name for f in migration_files]


def main():
    """Main entry point."""
    # Check Supabase first (production)
    supabase_status = check_supabase_migrations()
    if supabase_status:
        migrations_applied = supabase_status.get("migrations_applied", False)
        migration_files = supabase_status.get("migration_files", [])
        
        if not migration_files:
            print("✓ No migration files found")
            return
        
        if migrations_applied:
            print(f"✓ Database schema up to date, migrations already applied")
            print(f"  Found {len(migration_files)} migration file(s) in repository")
        else:
            print(f"⚠️  {len(migration_files)} migration file(s) found but not yet applied")
            print("   Run migrations via Supabase CLI or Azure Pipeline")
        return
    
    # Check local Postgres (development)
    postgres_status = check_local_postgres()
    if postgres_status:
        if postgres_status["migrations_applied"]:
            print("✓ Database schema up to date, migrations already applied")
        else:
            migration_files = get_migration_files()
            if migration_files:
                print(f"⚠️  {len(migration_files)} migration file(s) found but not yet applied")
                print("   Run migrations manually via Supabase CLI or SQL Editor")
            else:
                print("ℹ️  No migration files found")
        return
    
    # Can't determine status
    print("ℹ️  Could not determine migration status (no database connection available)")


if __name__ == "__main__":
    main()

