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


def get_account_relationships(account, owner_account_id: int) -> list:
    """Get relationships for an account from Account_Relationship table.

    Returns relationships in a format suitable for the frontend:
    - id: the entity id (individual or organization)
    - account_id: the related account id
    - name: display name
    - type: 'individual' or 'organization'
    - relationship_id: the Account_Relationship.id for deletion
    """
    relationships = []
    seen = set()

    # Outgoing relationships (this account -> other account)
    for rel in (account.outgoing_relationships or []):
        related_account = rel.to_account
        account_type, entity_id = get_account_type_and_entity_id(related_account)
        key = (account_type, entity_id)
        if key not in seen:
            seen.add(key)
            relationships.append({
                "id": entity_id,
                "account_id": related_account.id,
                "name": related_account.name,
                "type": account_type,
                "relationship_id": rel.id,
                "relationship_type": rel.relationship_type,
            })

    # Incoming relationships (other account -> this account)
    for rel in (account.incoming_relationships or []):
        related_account = rel.from_account
        account_type, entity_id = get_account_type_and_entity_id(related_account)
        key = (account_type, entity_id)
        if key not in seen:
            seen.add(key)
            relationships.append({
                "id": entity_id,
                "account_id": related_account.id,
                "name": related_account.name,
                "type": account_type,
                "relationship_id": rel.id,
                "relationship_type": rel.relationship_type,
            })

    return relationships
