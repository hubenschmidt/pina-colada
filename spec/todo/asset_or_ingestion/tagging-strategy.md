# Tagging Strategy - Technical Specification

## Status: Future Enhancement

This spec documents the tagging intelligence layer as a separate concern from the ingestion service.

**Related specs:**
- `spec/todo/asset_or_ingestion/asset-upload.md` - Phase 1: Client upload with tagging
- `spec/todo/asset_or_ingestion/asset-api-ingest.md` - Phase 2: API ingestion (tag-agnostic)

---

## Context

The ingestion service accepts tags but doesn't determine them. This spec covers the intelligence layer that decides *what* tags to apply.

**Goal:** Flexible tagging strategies that can evolve independently from storage/ingestion.

**Note:** Documents use the existing polymorphic `EntityTag` table for tagging (via `EntityTag(tag_id, 'Asset', asset_id)`) and `EntityAsset` for entity associations. A document can be tagged and linked to multiple entities (Projects, Leads, Accounts) independently.

---

## Tagging Approaches

### 1. User-Assigned
- Manual tagging in UI
- Full control, highest accuracy
- Doesn't scale for bulk imports

### 2. Rule-Based
- ETL pipeline applies tags based on source/patterns
- Examples: source URL domain, file type, folder structure
- Predictable, fast, but inflexible

### 3. LLM-Suggested
- Analyze content, propose relevant tags
- Works for unknown/varied content
- Requires LLM call overhead
- May need user confirmation

### 4. Hybrid
- LLM suggests, user confirms
- Balance of automation and accuracy
- Best for semi-automated workflows

---

## Future Implementation

### LLM Tagging Service

```python
async def suggest_tags(
    content: str,
    content_type: str,
    existing_tags: list[str],  # Tenant's tag vocabulary
) -> list[str]:
    """
    Analyze content and suggest relevant tags.

    - Uses tenant's existing tags as vocabulary
    - Returns 3-5 most relevant tags
    - Optionally creates new tags if needed
    """
```

### Integration Points

1. **Ingestion API** - `suggest_tags=true` flag triggers LLM analysis
2. **Background worker** - Process tagging queue async
3. **UI component** - Show suggestions for user confirmation
4. **Agent tool** - Agent can request tag suggestions

---

## Considerations

### Token Efficiency
- Batch similar content for single LLM call
- Cache tag suggestions for similar content types
- Use smaller model for tagging (Haiku)

### Tag Consistency
- Prefer existing tags over creating new ones
- Normalize tag names (lowercase, hyphenated)
- Merge similar tags periodically

### User Feedback Loop
- Track which suggestions are accepted/rejected
- Fine-tune prompts based on feedback
- Per-tenant tag preferences

---

## Dependencies

- Asset/Document tables from Phase 1 (joined table inheritance)
- EntityTag table (existing polymorphic tagging)
- EntityAsset table (polymorphic entity linking)
- Ingestion API from Phase 2
- LLM integration (existing infrastructure)

---

## Not In Scope (for now)

- Automatic tag hierarchy/taxonomy
- Cross-tenant tag sharing
- Tag embeddings for semantic search
- Auto-tagging on content change
