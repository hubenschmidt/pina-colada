# Validation Integrity Spec: API vs Database Layer

## Purpose

Ensure all data validation that occurs at the API level is also enforced at the database level, since the AI agent will call repositories directly (bypassing API validation).

---

## Current Validation Audit

### ✅ Covered (Both API + DB)

| Field                     | API Validation                    | DB Constraint                                                         | Location                     |
| ------------------------- | --------------------------------- | --------------------------------------------------------------------- | ---------------------------- |
| `Opportunity.probability` | `Field(ge=0, le=100)` in Pydantic | `SMALLINT CHECK (probability >= 0 AND probability <= 100)`            | `001_initial_schema.sql:295` |
| `Deal.probability`        | N/A (no API model found)          | `SMALLINT CHECK (probability >= 0 AND probability <= 100)`            | `001_initial_schema.sql:63`  |
| `Individual.email`        | Unique check in controller        | `UNIQUE INDEX idx_individual_email_lower ON Individual(LOWER(email))` | `001_initial_schema.sql:102` |

### ⚠️ Gaps (API Only - No DB Constraint)

| Field                | API Validation                                       | Missing DB Constraint | Risk Level                     |
| -------------------- | ---------------------------------------------------- | --------------------- | ------------------------------ |
| `Contact.phone`      | `validate_phone()` - regex `^\+1-\d{3}-\d{3}-\d{4}$` | No CHECK constraint   | **HIGH** - AI agent can bypass |
| `Organization.phone` | `validate_phone()`                                   | No CHECK constraint   | **HIGH** - AI agent can bypass |
| `Individual.phone`   | `validate_phone()`                                   | No CHECK constraint   | **HIGH** - AI agent can bypass |

### ✅ Query Parameter Validations (Not Data Integrity Concerns)

These are API-level pagination/search validations, not data integrity constraints:

- `page: ge=1` - Pagination parameter
- `limit: ge=1, le=100` - Pagination parameter
- `order: regex="^(ASC|DESC)$"` - Sort direction parameter
- `q: min_length=1` - Search query parameter

These don't need database constraints as they're not persisted to the database.

### ✅ Required Field Validations (Covered by NOT NULL)

Fields marked as required in Pydantic models (e.g., `first_name: str`, `name: str`) are already enforced at the database level via `NOT NULL` constraints in the schema.

---

## Recommended Database Migrations

### 1. Phone Format Constraint

Add CHECK constraint to enforce `+1-XXX-XXX-XXXX` format:

```sql
-- Contact table
ALTER TABLE "Contact"
ADD CONSTRAINT contact_phone_format_check
CHECK (phone IS NULL OR phone ~ '^\+1-\d{3}-\d{3}-\d{4}$');

-- Organization table
ALTER TABLE "Organization"
ADD CONSTRAINT organization_phone_format_check
CHECK (phone IS NULL OR phone ~ '^\+1-\d{3}-\d{3}-\d{4}$');

-- Individual table
ALTER TABLE "Individual"
ADD CONSTRAINT individual_phone_format_check
CHECK (phone IS NULL OR phone ~ '^\+1-\d{3}-\d{3}-\d{4}$');
```

---

## Files to Modify

1. **New migration file**: `modules/agent/migrations/XXX_add_phone_constraints.sql`

   - Add phone format CHECK constraints to Contact, Organization, Individual tables

2. **SQLAlchemy models** (optional but recommended for documentation):
   - `modules/agent/src/models/Contact.py` - Add CheckConstraint to `__table_args__`
   - `modules/agent/src/models/Organization.py` - Add CheckConstraint
   - `modules/agent/src/models/Individual.py` - Add CheckConstraint

---

## Summary

| Validation Type     | Count    | Status                                |
| ------------------- | -------- | ------------------------------------- |
| Probability (0-100) | 2        | ✅ Covered                            |
| Email uniqueness    | 1        | ✅ Covered                            |
| Phone format        | 3        | ⚠️ **API only - needs DB constraint** |
| Required fields     | Multiple | ✅ Covered (NOT NULL)                 |
| Query parameters    | Multiple | ✅ N/A (not persisted)                |

**Risk**: AI agent calling `contact_repository.create()`, `organization_repository.create()`, or `individual_repository.create()` directly could insert malformed phone numbers, breaking application assumptions and potentially causing issues with phone number parsing/display logic.

---

## Audit Methodology

This audit reviewed:

- ✅ All Pydantic `Field()` constraints (`ge`, `le`, `min_length`, `max_length`, `pattern`)
- ✅ All `@field_validator` decorators and custom validation functions
- ✅ All `Query()` parameter validations (excluded as not data integrity concerns)
- ✅ Database schema (`001_initial_schema.sql`) for CHECK constraints, UNIQUE indexes, NOT NULL constraints
- ✅ SQLAlchemy model definitions for constraint documentation

**Conclusion**: Phone format validation is the ONLY API validation that lacks a corresponding database constraint.

---

## Frontend Validation Audit

### ✅ Covered (Frontend + API + DB)

| Field                                                          | Frontend Validation             | API Validation           | DB Constraint                                     | Status        |
| -------------------------------------------------------------- | ------------------------------- | ------------------------ | ------------------------------------------------- | ------------- |
| `Opportunity.probability`                                      | `min: 0, max: 100` (HTML5)      | `Field(ge=0, le=100)`    | `CHECK (probability >= 0 AND probability <= 100)` | ✅ All layers |
| Required fields (name, first_name, last_name, job_title, etc.) | `required: true` in form config | Pydantic required fields | `NOT NULL` constraints                            | ✅ All layers |

### ⚠️ Frontend Only - No API/DB Validation

| Field                        | Frontend Validation                                | Missing API/DB Validation                       | Risk Level                           |
| ---------------------------- | -------------------------------------------------- | ----------------------------------------------- | ------------------------------------ |
| `Organization.founding_year` | `min: 1800, max: current year` (HTML5)             | No API validation, no DB constraint             | **MEDIUM** - Invalid years possible  |
| Phone formatting             | `formatPhoneNumber()` formats to `+1-XXX-XXX-XXXX` | API validates format, but DB constraint missing | **HIGH** - Format not enforced at DB |

### ✅ Frontend Formatting Only (No Validation Needed)

| Field            | Frontend Behavior                                | Notes                                                                      |
| ---------------- | ------------------------------------------------ | -------------------------------------------------------------------------- |
| Phone formatting | `formatPhoneNumber()` auto-formats as user types | Formatting only, not validation. API validates format.                     |
| Email type       | `type="email"` HTML5                             | Browser validation only, not enforced. API/DB don't validate email format. |
| Phone type       | `type="tel"` HTML5                               | Browser validation only, not enforced. API validates format.               |

### ⚠️ Gaps Identified

1. **`Organization.founding_year`** - Frontend has min/max constraints but:
   - No API validation (Pydantic model doesn't validate range)
   - No database CHECK constraint
   - **Recommendation**: Add API validation and DB constraint

---

## Recommended Additions

### 1. Organization Founding Year Validation

**API Layer** (`modules/agent/src/api/routes/organizations.py`):

```python
founding_year: Optional[int] = Field(default=None, ge=1800, le=func.now().year)
```

**Database Layer** (`modules/agent/migrations/XXX_add_founding_year_constraint.sql`):

```sql
ALTER TABLE "Organization"
ADD CONSTRAINT organization_founding_year_check
CHECK (founding_year IS NULL OR (founding_year >= 1800 AND founding_year <= EXTRACT(YEAR FROM NOW())));
```

---

## Updated Summary

| Validation Type     | Count    | Frontend        | API | DB  | Status                  |
| ------------------- | -------- | --------------- | --- | --- | ----------------------- |
| Probability (0-100) | 2        | ✅              | ✅  | ✅  | ✅ All layers           |
| Email uniqueness    | 1        | N/A             | ✅  | ✅  | ✅ API + DB             |
| Phone format        | 3        | Formatting only | ✅  | ❌  | ⚠️ **Missing DB**       |
| Required fields     | Multiple | ✅              | ✅  | ✅  | ✅ All layers           |
| Founding year range | 1        | ✅              | ❌  | ❌  | ⚠️ **Missing API + DB** |

**Final Conclusion**:

- **Phone format**: Missing DB constraint (already identified)
- **Founding year**: Missing both API validation and DB constraint (new finding)
