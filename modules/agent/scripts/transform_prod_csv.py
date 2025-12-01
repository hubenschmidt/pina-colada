#!/usr/bin/env python3
"""Transform CSV exports from old schema to new schema.

Changes applied:
- Organization: Remove 'industry', add 'account_id'
- Deal: Remove 'owner_user_id', keep 'tenant_id'
- Lead: Remove 'owner_user_id', add 'tenant_id', 'account_id'
- Job: Remove 'organization_id', rename 'notes' to 'description'
- Creates Account.csv from Organizations
"""

import csv
import os
from pathlib import Path


EXPORTS_DIR = Path(__file__).parent.parent / "exports"
OUTPUT_DIR = Path(__file__).parent.parent / "exports_transformed"


def read_csv(filename: str) -> list[dict]:
    """Read CSV file and return list of dicts."""
    filepath = EXPORTS_DIR / filename
    if not filepath.exists():
        print(f"  Warning: {filename} not found")
        return []
    with open(filepath, "r", newline="", encoding="utf-8") as f:
        return list(csv.DictReader(f))


def write_csv(filename: str, data: list[dict], fieldnames: list[str]):
    """Write list of dicts to CSV file."""
    if not data:
        print(f"  Skipping {filename} - no data")
        return
    filepath = OUTPUT_DIR / filename
    with open(filepath, "w", newline="", encoding="utf-8") as f:
        writer = csv.DictWriter(f, fieldnames=fieldnames)
        writer.writeheader()
        writer.writerows(data)
    print(f"  Wrote {filename}: {len(data)} rows")


def transform():
    """Transform all CSV files to new schema."""
    print("=" * 60)
    print("CSV Schema Transformation")
    print("=" * 60)
    print(f"Source: {EXPORTS_DIR}")
    print(f"Output: {OUTPUT_DIR}")
    print()

    # Create output directory
    OUTPUT_DIR.mkdir(exist_ok=True)

    # Read source files
    print("Reading source files...")
    tenants = read_csv("Tenant.csv")
    users = read_csv("User.csv")
    organizations = read_csv("Organization.csv")
    deals = read_csv("Deal.csv")
    leads = read_csv("Lead.csv")
    jobs = read_csv("Job.csv")

    # Get default tenant_id (assume single tenant for now)
    default_tenant_id = tenants[0]["id"] if tenants else "1"
    print(f"  Default tenant_id: {default_tenant_id}")

    # Get max org ID to offset Account IDs for user accounts
    max_org_id = max(int(org["id"]) for org in organizations) if organizations else 0
    print(f"  Max organization ID: {max_org_id}")

    # Build org_id -> job mapping for lead.account_id
    job_org_map = {job["id"]: job.get("organization_id") for job in jobs}

    # =========================================
    # 1. Create Account records from Organizations
    # =========================================
    print("\nCreating Account records...")
    accounts = []
    org_to_account = {}  # org_id -> account_id mapping

    for org in organizations:
        org_id = org["id"]
        account = {
            "id": org_id,  # Use same ID for simplicity
            "tenant_id": default_tenant_id,
            "name": org["name"],
            "created_at": org["created_at"],
            "updated_at": org["updated_at"],
        }
        accounts.append(account)
        org_to_account[org_id] = org_id

    write_csv("Account.csv", accounts, ["id", "tenant_id", "name", "created_at", "updated_at"])

    # =========================================
    # 2. Transform Organizations
    # =========================================
    print("\nTransforming Organization...")
    transformed_orgs = []
    for org in organizations:
        new_org = {
            "id": org["id"],
            "name": org["name"],
            "website": org.get("website", ""),
            "phone": org.get("phone", ""),
            "employee_count": org.get("employee_count", ""),
            "description": org.get("description", ""),
            "notes": org.get("notes", ""),
            "created_at": org["created_at"],
            "updated_at": org["updated_at"],
            "account_id": org_to_account.get(org["id"], ""),
        }
        transformed_orgs.append(new_org)

    write_csv(
        "Organization.csv",
        transformed_orgs,
        ["id", "name", "website", "phone", "employee_count", "description", "notes", "created_at", "updated_at", "account_id"],
    )

    # =========================================
    # 3. Transform Deals
    # =========================================
    print("\nTransforming Deal...")
    transformed_deals = []
    for deal in deals:
        new_deal = {
            "id": deal["id"],
            "name": deal["name"],
            "description": deal.get("description", ""),
            "current_status_id": deal.get("current_status_id", ""),
            "value_amount": deal.get("value_amount", ""),
            "value_currency": deal.get("value_currency", "USD"),
            "probability": deal.get("probability", ""),
            "expected_close_date": deal.get("expected_close_date", ""),
            "close_date": deal.get("close_date", ""),
            "created_at": deal["created_at"],
            "updated_at": deal["updated_at"],
            "tenant_id": deal.get("tenant_id") or default_tenant_id,
            "owner_individual_id": "",  # Was owner_user_id, now nullable
            "project_id": "",  # New field, nullable
        }
        transformed_deals.append(new_deal)

    write_csv(
        "Deal.csv",
        transformed_deals,
        ["id", "name", "description", "current_status_id", "value_amount", "value_currency", "probability", "expected_close_date", "close_date", "created_at", "updated_at", "tenant_id", "owner_individual_id", "project_id"],
    )

    # =========================================
    # 4. Transform Leads
    # =========================================
    print("\nTransforming Lead...")
    transformed_leads = []
    for lead in leads:
        lead_id = lead["id"]
        # Get account_id from the Job's organization_id
        org_id = job_org_map.get(lead_id)
        account_id = org_to_account.get(org_id, "") if org_id else ""

        new_lead = {
            "id": lead_id,
            "deal_id": lead["deal_id"],
            "type": lead["type"],
            "title": lead["title"],
            "description": lead.get("description", ""),
            "source": lead.get("source", ""),
            "current_status_id": lead.get("current_status_id", ""),
            "created_at": lead["created_at"],
            "updated_at": lead["updated_at"],
            "owner_individual_id": "",  # Was owner_user_id, now nullable
            "tenant_id": default_tenant_id,
            "account_id": account_id,
        }
        transformed_leads.append(new_lead)

    write_csv(
        "Lead.csv",
        transformed_leads,
        ["id", "deal_id", "type", "title", "description", "source", "current_status_id", "created_at", "updated_at", "owner_individual_id", "tenant_id", "account_id"],
    )

    # =========================================
    # 5. Transform Jobs
    # =========================================
    print("\nTransforming Job...")
    transformed_jobs = []
    for job in jobs:
        new_job = {
            "id": job["id"],
            "job_title": job["job_title"],
            "job_url": job.get("job_url", ""),
            "description": job.get("notes", ""),  # Renamed from notes
            "resume_date": job.get("resume_date", ""),
            "salary_range": job.get("salary_range", ""),
            "created_at": job["created_at"],
            "updated_at": job["updated_at"],
            "salary_range_id": "",  # New field, nullable
        }
        transformed_jobs.append(new_job)

    write_csv(
        "Job.csv",
        transformed_jobs,
        ["id", "job_title", "job_url", "description", "resume_date", "salary_range", "created_at", "updated_at", "salary_range_id"],
    )

    # =========================================
    # 6. Create Individual records for Users
    # =========================================
    print("\nCreating Individual records for Users...")
    individuals = []
    user_to_individual = {}  # user_id -> individual_id mapping

    for user in users:
        user_id = user["id"]
        # Account ID for user = max_org_id + user_id (to avoid collision with org accounts)
        user_account_id = max_org_id + int(user_id)
        individual_id = user_id  # Use same ID as user for simplicity

        # Create account for the user's individual
        user_account = {
            "id": str(user_account_id),
            "tenant_id": default_tenant_id,
            "name": user.get("email", f"User {user_id}"),
            "created_at": user["created_at"],
            "updated_at": user["updated_at"],
        }
        accounts.append(user_account)

        # Create individual - extract name from email if not provided
        email = user.get("email", "")
        first_name = user.get("first_name", "").strip()
        last_name = user.get("last_name", "").strip()

        # If no name, derive from email (e.g., whubenschmidt@gmail.com -> "whubenschmidt")
        if not first_name and email:
            email_prefix = email.split("@")[0]
            first_name = email_prefix
        if not first_name:
            first_name = "Unknown"
        if not last_name:
            last_name = "User"  # Both first_name and last_name are NOT NULL

        individual = {
            "id": individual_id,
            "account_id": str(user_account_id),
            "first_name": first_name,
            "last_name": last_name,
            "email": email,
            "phone": "",
            "linkedin_url": "",
            "title": "",
            "description": "",
            "created_at": user["created_at"],
            "updated_at": user["updated_at"],
        }
        individuals.append(individual)
        user_to_individual[user_id] = individual_id

    # Re-write Account.csv with user accounts added
    write_csv("Account.csv", accounts, ["id", "tenant_id", "name", "created_at", "updated_at"])

    write_csv(
        "Individual.csv",
        individuals,
        ["id", "account_id", "first_name", "last_name", "email", "phone", "linkedin_url", "title", "description", "created_at", "updated_at"],
    )

    # =========================================
    # 7. Transform Users (add individual_id)
    # =========================================
    print("\nTransforming User...")
    transformed_users = []
    for user in users:
        new_user = {
            "id": user["id"],
            "tenant_id": user.get("tenant_id") or default_tenant_id,
            "email": user.get("email", ""),
            "first_name": user.get("first_name", ""),
            "last_name": user.get("last_name", ""),
            "avatar_url": user.get("avatar_url", ""),
            "status": user.get("status", "active"),
            "last_login_at": user.get("last_login_at", ""),
            "created_at": user["created_at"],
            "updated_at": user["updated_at"],
            "auth0_sub": user.get("auth0_sub", ""),
            "individual_id": user_to_individual.get(user["id"], ""),
        }
        transformed_users.append(new_user)

    write_csv(
        "User.csv",
        transformed_users,
        ["id", "tenant_id", "email", "first_name", "last_name", "avatar_url", "status", "last_login_at", "created_at", "updated_at", "auth0_sub", "individual_id"],
    )

    # =========================================
    # 8. Create Role and UserRole records
    # =========================================
    print("\nCreating Role and UserRole records...")

    # Get current timestamp for created_at
    from datetime import datetime, timezone
    now = datetime.now(timezone.utc).strftime("%Y-%m-%dT%H:%M:%SZ")

    # Create tenant-specific roles (Owner role for each tenant)
    roles = []
    user_roles = []
    role_id = 1000  # Start high to avoid conflicts with global roles

    for tenant in tenants:
        tenant_id = tenant["id"]
        # Create Owner role for this tenant
        role = {
            "id": str(role_id),
            "name": "Owner",
            "description": "Tenant owner with full access",
            "created_at": now,
            "tenant_id": tenant_id,
        }
        roles.append(role)

        # Assign all users with this tenant_id to the Owner role
        for user in users:
            if user.get("tenant_id") == tenant_id:
                user_role = {
                    "user_id": user["id"],
                    "role_id": str(role_id),
                    "created_at": now,
                }
                user_roles.append(user_role)

        role_id += 1

    write_csv("Role.csv", roles, ["id", "name", "description", "created_at", "tenant_id"])
    write_csv("UserRole.csv", user_roles, ["user_id", "role_id", "created_at"])

    # =========================================
    # 9. Copy Tenant as-is
    # =========================================
    print("\nCopying unchanged files...")

    # Tenant - copy as-is
    write_csv(
        "Tenant.csv",
        tenants,
        ["id", "name", "slug", "status", "plan", "created_at", "updated_at"],
    )

    print()
    print("=" * 60)
    print("Transformation complete!")
    print(f"Output files in: {OUTPUT_DIR}")
    print("=" * 60)


if __name__ == "__main__":
    transform()
