#!/usr/bin/env python3
"""Run database seeders for local Postgres and test environments.

Seeders run after migrations and only insert data if it doesn't already exist.
Set RUN_SEEDERS=false to skip seeders (e.g., in staging/production).
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


def _import_psycopg2():
    """Import psycopg2, return None if not available."""
    try:
        import psycopg2
        return psycopg2
    except ImportError:
        return None


def _get_seeders_dir():
    """Get seeders directory path."""
    script_dir = Path(__file__).parent
    docker_seeders_dir = Path("/app/seeders")
    if docker_seeders_dir.exists():
        return docker_seeders_dir
    return script_dir.parent / "seeders"


def _ensure_seeders_table(cursor):
    """Create seeders tracking table if it doesn't exist."""
    cursor.execute("""
        CREATE TABLE IF NOT EXISTS schema_seeders (
            seeder_name TEXT PRIMARY KEY,
            applied_at TIMESTAMP DEFAULT NOW()
        );
    """)


def _get_applied_seeders(cursor) -> set:
    """Get set of already-applied seeder names."""
    cursor.execute("SELECT seeder_name FROM schema_seeders;")
    return {row[0] for row in cursor.fetchall()}


def _record_seeder(cursor, conn, seeder_name: str):
    """Record that a seeder has been applied."""
    cursor.execute(
        "INSERT INTO schema_seeders (seeder_name) VALUES (%s) ON CONFLICT DO NOTHING;",
        (seeder_name,)
    )
    conn.commit()


def _get_postgres_config():
    """Get Postgres connection configuration.
    
    Prioritizes local Postgres when POSTGRES_HOST is set to local values.
    Falls back to Supabase if explicitly configured and not running locally.
    """
    postgres_host = os.getenv("POSTGRES_HOST")
    postgres_port = int(os.getenv("POSTGRES_PORT"))
    postgres_user = os.getenv("POSTGRES_USER")
    postgres_password = os.getenv("POSTGRES_PASSWORD")
    postgres_db = os.getenv("POSTGRES_DB")
    
    # If POSTGRES_HOST is set to local values, use local Postgres (don't check Supabase)
    local_hosts = ["postgres", "localhost", "127.0.0.1"]
    if postgres_host in local_hosts:
        return postgres_host, postgres_port, postgres_user, postgres_password, postgres_db
    
    # Otherwise, check for Supabase configuration
    supabase_db_password = os.getenv("SUPABASE_DB_PASSWORD")
    if not supabase_db_password:
        return postgres_host, postgres_port, postgres_user, postgres_password, postgres_db
    
    postgres_password = supabase_db_password
    supabase_url = os.getenv("SUPABASE_URL", "")
    if not supabase_url or "supabase.co" not in supabase_url:
        return postgres_host, postgres_port, postgres_user, postgres_password, postgres_db
    
    try:
        project_ref = supabase_url.split("//")[1].split(".")[0]
        postgres_host = f"db.{project_ref}.supabase.co"
        postgres_db = "postgres"
        print(f"\n✓ Detected Supabase connection")
    except Exception:
        pass
    
    return postgres_host, postgres_port, postgres_user, postgres_password, postgres_db


def run_seeders():
    """Run all seeder files in seeders directory."""
    run_seeders_flag = os.getenv("RUN_SEEDERS", "").lower()
    if run_seeders_flag != "true":
        print(f"⏭️  Skipping seeders (RUN_SEEDERS not set or not 'true')")
        return True
    
    psycopg2 = _import_psycopg2()
    if not psycopg2:
        print("⚠️  psycopg2 not available - skipping seeders")
        return True

    seeders_dir = _get_seeders_dir()
    if not seeders_dir.exists():
        print(f"ℹ️  No seeders directory found at {seeders_dir}")
        return True

    seeder_files = sorted(seeders_dir.glob("*.sql"))
    if not seeder_files:
        print(f"ℹ️  No seeder files found in {seeders_dir}")
        return True

    print(f"\n📁 Found {len(seeder_files)} seeder file(s):")
    for sf in seeder_files:
        print(f"  - {sf.name}")

    postgres_host, postgres_port, postgres_user, postgres_password, postgres_db = _get_postgres_config()

    try:
        print(f"\nConnecting to Postgres: {postgres_host}:{postgres_port}/{postgres_db}")
        conn = psycopg2.connect(
            host=postgres_host,
            port=postgres_port,
            database=postgres_db,
            user=postgres_user,
            password=postgres_password,
            connect_timeout=10
        )
    except psycopg2.OperationalError as e:
        print(f"\n❌ Failed to connect to Postgres: {e}")
        return False

    cursor = conn.cursor()
    try:
        # Ensure seeders tracking table exists
        _ensure_seeders_table(cursor)
        conn.commit()
        
        applied_seeders = _get_applied_seeders(cursor)
        
        unapplied_seeders = [
            sf for sf in seeder_files
            if sf.name not in applied_seeders
        ]
        
        if not unapplied_seeders:
            print("✓ All seeders already applied, nothing to run")
            cursor.close()
            conn.close()
            return True
        
        print(f"\n📋 Found {len(unapplied_seeders)} seeder(s) to apply:")
        for sf in unapplied_seeders:
            print(f"  - {sf.name}")
        
        for seeder_file in unapplied_seeders:
            print(f"\n🌱 Running seeder: {seeder_file.name}")
            sql_content = seeder_file.read_text()
            cursor.execute(sql_content)

            # Try to fetch results if the seeder returns data
            rows_affected = cursor.rowcount
            result_summary = None
            try:
                if cursor.description:  # Check if there are results
                    results = cursor.fetchall()
                    is_single_value = results and len(results) == 1 and len(results[0]) == 1
                    if is_single_value:
                        result_summary = f"{results[0][0]} rows"
                    if results and not is_single_value:
                        result_summary = ", ".join(
                            f"{cursor.description[i][0]}: {val}"
                            for i, val in enumerate(results[0])
                        )
            except:
                pass

            _record_seeder(cursor, conn, seeder_file.name)

            summary = result_summary or f"{rows_affected} rows affected"
            print(f"✓ Seeder {seeder_file.name} completed ({summary})")

        cursor.close()
        conn.close()
        print("\n✓ All seeders completed successfully!")
        return True
    except Exception as e:
        print(f"\n❌ Error running seeders: {e}")
        conn.rollback()
        cursor.close()
        conn.close()
        return False


def main():
    """Main entry point."""
    print("="*60)
    print("Database Seeder Tool")
    print("="*60)
    
    success = run_seeders()
    if not success:
        sys.exit(1)


if __name__ == "__main__":
    main()

