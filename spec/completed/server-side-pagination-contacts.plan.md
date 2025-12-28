# Server-Side Pagination: Contacts Endpoint

## Summary

The `/accounts/contacts` datatable currently uses client-side pagination, meaning all contacts are loaded at once and filtered/sorted/paginated in the browser. This should be converted to server-side pagination for consistency and performance.

## Current State

| Page | API Function | Pagination | Backend Support |
|------|--------------|------------|-----------------|
| `/accounts/contacts` | `getContacts()` | Client-side (useMemo) | No pagination params |

All other datatable endpoints already use server-side pagination:
- `/accounts/organizations` ✅
- `/accounts/individuals` ✅
- `/leads/jobs` ✅
- `/leads/opportunities` ✅
- `/leads/partnerships` ✅
- `/tasks` ✅
- `/assets/documents` ✅

## Required Changes

### Backend

1. **Repository** (`src/repositories/contact_repository.py`)
   - Add `find_all_contacts_paginated()` function with params:
     - `page`, `page_size`, `order_by`, `order`, `search`, `tenant_id`
   - Search should filter by: `first_name`, `last_name`, `email`, `title`
   - Only load relationships needed for list view (not full detail)

2. **Routes** (`src/api/routes/contacts.py`)
   - Update `GET /contacts` to accept query params: `page`, `limit`, `orderBy`, `order`, `search`
   - Add `_contact_to_list_dict()` for minimal list response
   - Return paginated response format:
     ```json
     {
       "items": [...],
       "currentPage": 1,
       "totalPages": 5,
       "total": 250,
       "pageSize": 50
     }
     ```

### Frontend

3. **API** (`modules/client/api/index.ts`)
   - Update `getContacts()` to accept pagination params and return `PageData<Contact>`

4. **Page** (`modules/client/app/accounts/contacts/page.tsx`)
   - Remove client-side `useMemo` filtering/sorting/pagination
   - Add state for `page`, `limit`, `sortBy`, `sortDirection`, `searchQuery`
   - Pass params to `getContacts()` in `useCallback`
   - Re-fetch on param changes

## List View Fields (Minimal Response)

Based on the contacts datatable columns:
- `id`
- `first_name`
- `last_name`
- `title`
- `email`
- `phone`
- `is_primary`
- `updated_at`

## Files to Modify

- `modules/agent/src/repositories/contact_repository.py`
- `modules/agent/src/api/routes/contacts.py`
- `modules/client/api/index.ts`
- `modules/client/app/accounts/contacts/page.tsx`
