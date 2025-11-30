"""Upload seed document files to storage.

Run after 012_documents_seed.sql to upload actual PDF files.

Usage:
  docker compose exec agent python /app/scripts/seed_documents.py

This script uploads PDF files from seeders/documents/ to storage.
The SQL seeder (012_documents_seed.sql) creates the DB records.
"""

import asyncio
import logging
import os
import sys
from pathlib import Path

import asyncpg

# Add src to path for lib imports without triggering circular imports
sys.path.insert(0, str(Path(__file__).parent.parent / "src"))

from lib.storage import get_storage

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

SEEDS_DIR = Path(__file__).parent.parent / "seeders" / "documents"

SEED_FILES = [
    "company_proposal.pdf",
    "meeting_notes.pdf",
    "contract_draft.pdf",
    "product_spec.pdf",
    "invoice_sample.pdf",
    "individual_resume.pdf",
]


async def upload_seed_files():
    """Upload seed PDF files to storage for all tenants."""
    storage = get_storage()

    # Get all active tenants using direct asyncpg connection
    db_host = os.environ.get("POSTGRES_HOST")
    db_port = os.environ.get("POSTGRES_PORT")
    db_user = os.environ.get("POSTGRES_USER")
    db_password = os.environ.get("POSTGRES_PASSWORD")
    db_name = os.environ.get("POSTGRES_DB")

    if not all([db_host, db_port, db_user, db_password, db_name]):
        logger.warning("POSTGRES_* env vars not set, skipping seed document upload")
        return

    conn = await asyncpg.connect(
        host=db_host,
        port=int(db_port),
        user=db_user,
        password=db_password,
        database=db_name,
    )
    try:
        rows = await conn.fetch('SELECT id FROM "Tenant" WHERE status = \'active\' ORDER BY id')
        tenant_ids = [row["id"] for row in rows]
    finally:
        await conn.close()

    if not tenant_ids:
        logger.warning("No active tenants found")
        return

    uploaded_count = 0
    skipped_count = 0

    for tenant_id in tenant_ids:
        for filename in SEED_FILES:
            filepath = SEEDS_DIR / filename
            if not filepath.exists():
                logger.warning(f"File not found: {filepath}")
                continue

            storage_path = f"{tenant_id}/seed/{filename}"

            if await storage.exists(storage_path):
                skipped_count += 1
                continue

            with open(filepath, "rb") as f:
                content = f.read()

            await storage.upload(storage_path, content, "application/pdf")
            logger.info(f"Uploaded: {storage_path}")
            uploaded_count += 1

    if uploaded_count > 0:
        logger.info(f"Seed file upload complete: {uploaded_count} uploaded, {skipped_count} skipped")
    elif skipped_count > 0:
        logger.info(f"Seed files already exist ({skipped_count} files), skipping upload")


async def main():
    """Main entry point."""
    logger.info("Starting seed file upload...")
    await upload_seed_files()


if __name__ == "__main__":
    asyncio.run(main())
