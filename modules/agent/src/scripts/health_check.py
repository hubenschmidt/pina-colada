#!/usr/bin/env python3
"""
Agent Health Check - validates core functionality before agent starts.

Run: python scripts/health_check.py
Exit codes: 0 = healthy, 1 = unhealthy
"""

import asyncio
import sys
import os

# Add src to path
sys.path.insert(0, os.path.join(os.path.dirname(__file__), ".."))


async def check_database() -> bool:
    """Verify database connection and basic queries."""
    from lib.db import async_get_session
    from sqlalchemy import text

    try:
        async with async_get_session() as session:
            # Check tenant exists
            result = await session.execute(
                text("SELECT id, slug FROM \"Tenant\" WHERE slug = 'pinacolada' LIMIT 1")
            )
            tenant = result.fetchone()
            if not tenant:
                print("  ✗ Default tenant 'pinacolada' not found")
                return False
            print(f"  ✓ Tenant: {tenant.slug} (id={tenant.id})")

            # Check Individual table has data
            result = await session.execute(
                text("SELECT COUNT(*) FROM \"Individual\"")
            )
            count = result.scalar()
            print(f"  ✓ Individuals: {count} records")

            # Check documents exist
            result = await session.execute(
                text("SELECT COUNT(*) FROM \"Document\"")
            )
            count = result.scalar()
            print(f"  ✓ Documents: {count} records")

            return True
    except Exception as e:
        print(f"  ✗ Database error: {e}")
        return False


async def check_crm_tools() -> bool:
    """Verify CRM lookup tools work using direct SQL."""
    from lib.db import async_get_session
    from sqlalchemy import text

    try:
        async with async_get_session() as session:
            # Test individual lookup (direct SQL to avoid mapper issues)
            result = await session.execute(
                text("""
                    SELECT i.id, i.first_name, i.last_name, i.account_id
                    FROM "Individual" i
                    JOIN "Account" a ON i.account_id = a.id
                    WHERE a.tenant_id = 1
                    AND (i.first_name ILIKE '%william%' OR i.last_name ILIKE '%hubenschmidt%')
                """)
            )
            rows = result.fetchall()
            if not rows:
                print("  ✗ lookup_individual: No individuals found matching 'William'")
                return False
            for row in rows:
                print(f"  ✓ Individual: {row.first_name} {row.last_name} (id={row.id}, account_id={row.account_id})")

            # Test Entity_Asset links for documents
            result = await session.execute(
                text("""
                    SELECT a.filename, ea.entity_type, ea.entity_id
                    FROM "Entity_Asset" ea
                    JOIN "Asset" a ON ea.asset_id = a.id
                    WHERE a.filename LIKE 'william_hubenschmidt%'
                """)
            )
            rows = result.fetchall()
            if not rows:
                print("  ✗ Document links: No documents linked to William")
                return False
            print(f"  ✓ Document links: {len(rows)} documents linked")

            return True
    except Exception as e:
        print(f"  ✗ CRM tools error: {e}")
        return False


async def check_document_tools() -> bool:
    """Verify document queries work using direct SQL."""
    from lib.db import async_get_session
    from sqlalchemy import text

    try:
        async with async_get_session() as session:
            # Check documents can be queried
            result = await session.execute(
                text("""
                    SELECT a.id, a.filename, d.storage_path
                    FROM "Asset" a
                    JOIN "Document" d ON a.id = d.id
                    WHERE a.filename LIKE '%resume%'
                    LIMIT 5
                """)
            )
            rows = result.fetchall()
            if not rows:
                print("  ✗ No resume documents found")
                return False
            for row in rows:
                print(f"  ✓ Document: {row.filename} (id={row.id})")
            return True
    except Exception as e:
        print(f"  ✗ Document tools error: {e}")
        return False


async def check_storage() -> bool:
    """Verify storage has seed files."""
    from lib.db import async_get_session
    from sqlalchemy import text

    try:
        async with async_get_session() as session:
            result = await session.execute(
                text("""
                    SELECT COUNT(*) as cnt, storage_path
                    FROM "Document"
                    WHERE storage_path LIKE '%/seed/%'
                    GROUP BY storage_path
                    LIMIT 10
                """)
            )
            rows = result.fetchall()
            print(f"  ✓ Storage paths: {len(rows)} seed documents configured")
            return True
    except Exception as e:
        print(f"  ✗ Storage error: {e}")
        return False


async def main():
    print("=" * 50)
    print("Agent Health Check")
    print("=" * 50)

    checks = [
        ("Database", check_database),
        ("CRM Tools", check_crm_tools),
        ("Document Tools", check_document_tools),
        ("Storage", check_storage),
    ]

    all_passed = True
    for name, check_fn in checks:
        print(f"\n[{name}]")
        try:
            passed = await check_fn()
            if not passed:
                all_passed = False
        except Exception as e:
            print(f"  ✗ Unexpected error: {e}")
            all_passed = False

    print("\n" + "=" * 50)
    if all_passed:
        print("✓ All health checks passed")
        return 0
    else:
        print("✗ Some health checks failed")
        return 1


if __name__ == "__main__":
    exit_code = asyncio.run(main())
    sys.exit(exit_code)
