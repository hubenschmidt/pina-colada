# Document Versioning Implementation

## Overview

Add versioning support to documents so users can upload new versions of files with the same filename. Versions share entity links and are identified by version number and timestamp.

## Database Schema Changes

### Migration: Add versioning fields to Asset table

- Add `parent_id BIGINT REFERENCES "Asset"(id) ON DELETE CASCADE` - links versions together (NULL for original)
- Add `version_number INTEGER NOT NULL DEFAULT 1` - sequential version number
- Add `is_current_version BOOLEAN DEFAULT TRUE` - marks the active version
- Add index on `parent_id` for version queries
- Update existing documents: set `version_number=1`, `is_current_version=TRUE`, `parent_id=NULL`

## Backend Changes

### 1. Update Document Model (`modules/agent/src/models/Asset.py`)

- Add `parent_id`, `version_number`, `is_current_version` columns
- Add relationship to parent/children for version navigation

### 2. Update Document Repository (`modules/agent/src/repositories/document_repository.py`)

- Add `find_document_versions(document_id)` - get all versions of a document
- Add `get_current_version(parent_id)` - get the current version
- Add `create_new_version()` - create new version, increment version_number, mark old as not current
- Update `find_documents_by_tenant()` to optionally filter by `is_current_version=TRUE` for default list view

### 3. Update Documents API Route (`modules/agent/src/api/routes/documents.py`)

- Add `GET /assets/documents/{id}/versions` endpoint - return all versions
- Add `POST /assets/documents/{id}/versions` endpoint - create new version
- Add `PATCH /assets/documents/{id}/set-current` endpoint - set which version is current
- Update upload route to check for existing filename on entity before upload
- Return `has_existing_version: boolean` in upload response if duplicate filename found

### 4. Add Version Check Logic

- Before upload: Check if filename exists on the same entity (via EntityAsset)
- If exists: Return flag indicating version should be created
- On version creation: Link new version to same entities as parent

## Frontend Changes

### 1. Update API Client (`modules/client/api/index.ts`)

- Add `getDocumentVersions(id: number)` - fetch all versions
- Add `createDocumentVersion(id: number, file: File)` - upload new version
- Add `setCurrentVersion(id: number)` - mark version as current
- Update `uploadDocument()` to handle version creation

### 2. Update DocumentUpload Component (`modules/client/components/Documents/DocumentUpload.tsx`)

- Add check for existing filename on entity before upload
- Show modal/dialog when duplicate filename detected:
- Options: "Create New Version" or "Cancel"
- If "Create New Version": proceed with version creation
- Pass `parentDocumentId` when creating version

### 3. Update Document Detail Page (`modules/client/app/assets/documents/[id]/page.tsx`)

- Add version selector dropdown in header (shows current version, allows switching)
- Add "Version History" section below document info:
- List all versions with version number, timestamp, file size
- Show which is current version
- Allow setting different version as current
- Allow downloading any version
- Update document display to show version number and timestamp

### 4. Update DocumentList Component (`modules/client/components/Documents/DocumentList.tsx`)

- Add version badge/indicator if document has multiple versions
- Show version count (e.g., "v3 (2 other versions)")
- Filter to show only current versions by default (add toggle for "Show all versions")

### 5. Update DocumentsSection Component (`modules/client/components/DocumentsSection/DocumentsSection.tsx`)

- Pass entity context to DocumentUpload for duplicate filename checking
- Handle version creation in upload completion handler

## Implementation Details

### Version Numbering

- Original document: `version_number = 1`
- Each new version increments: `version_number = 2, 3, 4...`
- Version numbers are sequential and never reused

### Entity Linking

- When creating a new version:

1. Copy all EntityAsset links from parent document
2. New version inherits all entity relationships
3. All versions remain linked to same entities

### Current Version Logic

- Only one version per document group can be `is_current_version = TRUE`
- Default view shows only current versions
- Users can switch which version is "current"
- Download defaults to current version but any version can be downloaded

### Filename Matching

- Check for duplicate filename only on the specific entity being edited
- Example: If uploading "resume.pdf" to Lead #9, only check if "resume.pdf" exists on Lead #9
- Don't check across different entities (same filename on different leads is OK)

## Testing Considerations

- Test version creation from AccountForm and LeadForm
- Test version selector and history display
- Test entity link inheritance across versions
- Test current version switching
- Test duplicate filename detection on same entity
