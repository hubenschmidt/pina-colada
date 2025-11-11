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
            
            # Check if Job table exists (main table from migrations)
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
        # Local development path - migrations are in modules/agent/supabase_migrations/
        script_dir = Path(__file__).parent
        agent_dir = script_dir.parent
        migrations_dir = agent_dir / "supabase_migrations"
    
    if not migrations_dir.exists():
        return []
    
    migration_files = sorted(migrations_dir.glob("*.sql"))
    return [f.name for f in migration_files]


def main():
    """Main entry point."""
    # Check local Postgres (development only)
    # Note: Supabase migrations are handled in Azure Pipeline for production
    postgres_status = check_local_postgres()
    if postgres_status:
        if postgres_status["migrations_applied"]:
            print("✓ Database schema up to date, migrations already applied")
        else:
            migration_files = get_migration_files()
            if migration_files:
                print(f"⚠️  {len(migration_files)} migration file(s) found but not yet applied")
                print("   Migrations will be applied automatically on startup")
            else:
                print("ℹ️  No migration files found")
        return
    
    # Can't determine status
    print("ℹ️  Could not determine migration status (no database connection available)")


if __name__ == "__main__":
    main()

