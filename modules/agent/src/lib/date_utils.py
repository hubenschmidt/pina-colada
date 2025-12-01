"""Shared date utilities for parsing and formatting dates.

All datetimes are stored in UTC. Use user's timezone preference for display.
"""

from datetime import datetime, timezone
from zoneinfo import ZoneInfo, available_timezones


# Common timezones for frontend dropdown (curated list)
COMMON_TIMEZONES = [
    ("UTC", "UTC"),
    ("America/New_York", "Eastern Time (ET)"),
    ("America/Chicago", "Central Time (CT)"),
    ("America/Denver", "Mountain Time (MT)"),
    ("America/Los_Angeles", "Pacific Time (PT)"),
    ("America/Phoenix", "Arizona (no DST)"),
    ("America/Anchorage", "Alaska (AKT)"),
    ("Pacific/Honolulu", "Hawaii (HST)"),
    ("Europe/London", "London (GMT/BST)"),
    ("Europe/Paris", "Central European (CET)"),
    ("Europe/Berlin", "Berlin (CET)"),
    ("Asia/Tokyo", "Japan (JST)"),
    ("Asia/Shanghai", "China (CST)"),
    ("Asia/Singapore", "Singapore (SGT)"),
    ("Australia/Sydney", "Sydney (AEST)"),
]


def is_valid_timezone(tz: str) -> bool:
    """Check if timezone string is valid IANA timezone."""
    return tz in available_timezones()


def get_common_timezones() -> list[dict]:
    """Return list of common timezones for frontend dropdown."""
    return [{"value": tz, "label": label} for tz, label in COMMON_TIMEZONES]


def to_user_timezone(utc_dt: datetime, user_tz: str) -> datetime:
    """Convert UTC datetime to user's timezone for display."""
    if utc_dt is None:
        return None
    if not is_valid_timezone(user_tz):
        user_tz = "America/New_York"
    user_zone = ZoneInfo(user_tz)
    if utc_dt.tzinfo is None:
        utc_dt = utc_dt.replace(tzinfo=timezone.utc)
    return utc_dt.astimezone(user_zone)


def parse_date_input(date_str: str | None) -> datetime | None:
    """Parse date input from frontend and convert to UTC.

    For date-only strings like '2025-12-01', use the selected date with
    current UTC time. All returned datetimes are timezone-aware UTC.
    """
    if not date_str:
        return None

    # If already has time component, parse and ensure UTC
    if "T" in date_str or " " in date_str:
        try:
            dt = datetime.fromisoformat(date_str.replace("Z", "+00:00"))
            # Ensure timezone-aware UTC
            if dt.tzinfo is None:
                dt = dt.replace(tzinfo=timezone.utc)
            else:
                dt = dt.astimezone(timezone.utc)
            return dt
        except (ValueError, TypeError):
            return None

    # Date-only: use selected date with current UTC time
    try:
        now_utc = datetime.now(timezone.utc)
        parsed_date = datetime.strptime(date_str, "%Y-%m-%d")
        return parsed_date.replace(
            hour=now_utc.hour,
            minute=now_utc.minute,
            second=now_utc.second,
            tzinfo=timezone.utc
        )
    except (ValueError, TypeError):
        return None


def format_date(value) -> str:
    """Format datetime to ISO date string (YYYY-MM-DD)."""
    if not value:
        return ""
    date_str = value.isoformat() if isinstance(value, datetime) else str(value)
    return date_str[:10] if date_str else ""


def format_datetime(value) -> str:
    """Format datetime to ISO8601 string with Z suffix (UTC)."""
    if not value:
        return ""
    if isinstance(value, datetime):
        # Ensure UTC and format with Z suffix
        if value.tzinfo is not None:
            value = value.astimezone(timezone.utc)
        return value.strftime("%Y-%m-%dT%H:%M:%SZ")
    return str(value)


def format_display_date(value) -> str:
    """Format datetime to MM/DD/YYYY for frontend display."""
    if not value:
        return ""
    if isinstance(value, datetime):
        return value.strftime("%m/%d/%Y")
    date_str = str(value)[:10]
    if len(date_str) >= 10 and "-" in date_str:
        parts = date_str.split("-")
        if len(parts) == 3:
            return f"{int(parts[1])}/{int(parts[2])}/{parts[0]}"
    return date_str
