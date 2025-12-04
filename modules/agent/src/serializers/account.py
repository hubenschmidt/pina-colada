"""Serializers for account-related models."""


def get_account_type_and_entity_id(account) -> tuple:
    """Determine account type and entity ID from linked records."""
    if account.organizations:
        return "organization", account.organizations[0].id
    if account.individuals:
        return "individual", account.individuals[0].id
    return "unknown", account.id


def account_to_response(account) -> dict:
    """Convert account to dict with type info."""
    account_type, entity_id = get_account_type_and_entity_id(account)
    return {
        "id": entity_id,
        "account_id": account.id,
        "name": account.name,
        "type": account_type,
    }
