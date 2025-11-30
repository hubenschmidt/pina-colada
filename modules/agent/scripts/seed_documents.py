"""Upload seed document files to storage.

Run after 012_documents_seed.sql to upload actual PDF files.

Usage:
  docker compose exec agent python /app/scripts/seed_documents.py

This script uploads PDF files from seeders/documents/ to storage.
The SQL seeder (012_documents_seed.sql) creates the DB records.
"""

import asyncio
import logging
from pathlib import Path

from sqlalchemy import text

from lib.db import async_get_session
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

    # Get all active tenants
    async with async_get_session() as session:
        result = await session.execute(
            text('SELECT id FROM "Tenant" WHERE status = \'active\' ORDER BY id')
        )
        tenant_ids = [row[0] for row in result.fetchall()]

    if not tenant_ids:
        logger.warning("No active tenants found")
        return

    for tenant_id in tenant_ids:
        logger.info(f"Uploading seed files for tenant {tenant_id}...")

        for filename in SEED_FILES:
            filepath = SEEDS_DIR / filename
            if not filepath.exists():
                logger.warning(f"File not found: {filepath}")
                continue

            storage_path = f"{tenant_id}/seed/{filename}"

            with open(filepath, "rb") as f:
                content = f.read()

            await storage.upload(storage_path, content, "application/pdf")
            logger.info(f"  Uploaded: {storage_path}")

    logger.info("Seed file upload complete!")


async def main():
    """Main entry point."""
    logger.info("Starting seed file upload...")
    await upload_seed_files()


if __name__ == "__main__":
    asyncio.run(main())
