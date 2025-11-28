# Contacts as Standalone Entity

## Overview

Refactor Contacts to be standalone entities that can optionally link to Individuals, rather than requiring an Individual record. Adds a dedicated Contacts page under Accounts.

## Current State

- `Contact.individual_id` is NOT NULL (required)
- Creating an Individual auto-creates a Contact
- ContactSection requires selecting an Individual via search
- No standalone Contacts page exists

## Target State

- `Contact.individual_id` is nullable
- Contact has its own `first_name`, `last_name` fields
- Contacts page under Accounts with full CRUD
- Individual creation does NOT auto-create Contact
- ContactSection allows manual entry without Individual selection

---

## Implementation Steps

### 1. Database Migration

**File**: `modules/agent/src/alembic/versions/xxx_add_contact_name_fields.py`

```sql
ALTER TABLE "Contact" ADD COLUMN first_name TEXT;
ALTER TABLE "Contact" ADD COLUMN last_name TEXT;
ALTER TABLE "Contact" ALTER COLUMN individual_id DROP NOT NULL;

-- Migrate existing: copy names from linked Individual
UPDATE "Contact" c
SET first_name = i.first_name, last_name = i.last_name
FROM "Individual" i
WHERE c.individual_id = i.id;
```

### 2. Update Contact Model

**File**: `modules/agent/src/models/Contact.py`

- Add `first_name = Column(Text, nullable=True)`
- Add `last_name = Column(Text, nullable=True)`
- Change `individual_id` to `nullable=True`

### 3. Backend: New Contacts Routes

**File**: `modules/agent/src/api/routes/contacts.py`

Expand existing file with full CRUD:

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/contacts` | List all contacts |
| GET | `/contacts/{id}` | Get single contact |
| POST | `/contacts` | Create contact |
| PUT | `/contacts/{id}` | Update contact |
| DELETE | `/contacts/{id}` | Delete contact |

### 4. Update `_contact_to_dict` Functions

**Files**:
- `modules/agent/src/api/routes/individuals.py`
- `modules/agent/src/api/routes/organizations.py`
- `modules/agent/src/api/routes/contacts.py`

Change name resolution:
```python
def _contact_to_dict(contact):
    return {
        ...
        "first_name": contact.first_name or (contact.individual.first_name if contact.individual else ""),
        "last_name": contact.last_name or (contact.individual.last_name if contact.individual else ""),
        ...
    }
```

### 5. Remove Auto-Create Contact Logic

**File**: `modules/agent/src/api/routes/individuals.py`

Remove lines 171-178 in `create_individual_route`:
```python
# DELETE THIS BLOCK:
await create_contact({
    "individual_id": individual.id,
    "email": individual.email,
    ...
})
```

### 6. Frontend API

**File**: `modules/client/api/index.ts`

Add:
```typescript
export async function getContacts(): Promise<Contact[]>
export async function getContact(id: number): Promise<Contact>
export async function createContact(data: ContactInput): Promise<Contact>
export async function updateContact(id: number, data: Partial<ContactInput>): Promise<Contact>
export async function deleteContact(id: number): Promise<void>
```

### 7. Frontend: Contacts List Page

**File**: `modules/client/app/accounts/contacts/page.tsx`

Follow pattern from `accounts/individuals/page.tsx`:
- DataTable with columns: Name, Email, Phone, Organization, Individual (if linked)
- Search/filter
- Click row → navigate to detail

### 8. Frontend: Contact Detail Page

**File**: `modules/client/app/accounts/contacts/[id]/page.tsx`

Form fields:
- First Name, Last Name (editable)
- Email, Phone, Title, Department, Role, Notes
- Individual (optional dropdown/search)
- Organization (optional dropdown/search)

### 9. Frontend: New Contact Page

**File**: `modules/client/app/accounts/contacts/new/page.tsx`

Same form as detail, create mode.

### 10. Update Sidebar

**File**: `modules/client/components/Sidebar/Sidebar.tsx`

Add under Accounts section:
```tsx
<Link href="/accounts/contacts">Contacts</Link>
```

Order: Individuals, Organizations, Contacts

### 11. Update ContactSection Component

**File**: `modules/client/components/ContactSection.tsx`

- Make individual search optional (not required to add contact)
- Allow manual first_name/last_name entry when no Individual selected
- Keep search as optional convenience feature

---

## Files to Modify

| File | Changes |
|------|---------|
| `models/Contact.py` | Add first_name, last_name; make individual_id nullable |
| `api/routes/contacts.py` | Full CRUD endpoints |
| `api/routes/individuals.py` | Remove auto-create, update _contact_to_dict |
| `api/routes/organizations.py` | Update _contact_to_dict |
| `repositories/contact_repository.py` | Add find_all_contacts, ensure eager loading |
| `client/api/index.ts` | Add contact CRUD functions |
| `client/app/accounts/contacts/page.tsx` | New: list page |
| `client/app/accounts/contacts/[id]/page.tsx` | New: detail page |
| `client/app/accounts/contacts/new/page.tsx` | New: create page |
| `client/components/Sidebar/Sidebar.tsx` | Add Contacts link |
| `client/components/ContactSection.tsx` | Make individual optional |
| `alembic/versions/xxx_*.py` | New: migration |

---

## Migration Strategy

1. Run migration to add columns and copy data
2. Deploy backend changes
3. Deploy frontend changes
4. Existing data preserved with names copied from Individuals
