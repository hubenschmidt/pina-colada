"""Shared validation utilities."""

import re

# Phone number pattern: matches +1-XXX-XXX-XXXX
PHONE_PATTERN = re.compile(r'^\+1-\d{3}-\d{3}-\d{4}$')


def validate_phone(phone: str | None) -> str | None:
    """
    Validate phone number format.
    Expected format: +1-XXX-XXX-XXXX
    Returns the phone if valid, raises ValueError if invalid.
    """
    if phone is None or phone == "":
        return None

    if not PHONE_PATTERN.match(phone):
        raise ValueError(
            "Invalid phone format. Expected: +1-XXX-XXX-XXXX"
        )

    return phone
