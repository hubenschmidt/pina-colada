# Configurable Crawler Behavior on Compilation Target

## Coding Rules
- **Go**: KISS, YAGNI, guard clauses, early returns, no nested conditionals. CSR pattern.
- **CSR**: Routes→Controllers→Services→Repositories. Services never import models directly.

## Goal
Give users two options for what happens when compilation target is reached:

1. **Pause** (default) - Crawler stays `enabled=true`, skips runs while at/above target, auto-resumes when proposals drop below target
2. **Disable** - Set `enabled=false`, requires manual re-enable via UI

## Config Field
`disable_on_compiled` (boolean, default: false)
- `false` = Pause behavior (current "skipping" logic)
- `true` = Disable behavior (set enabled=false)

## Implementation

### Model: `internal/models/AutomationConfig.go`
Add field:
```go
DisableOnCompiled bool `gorm:"column:disable_on_compiled;not null;default:false"`
```

### Migration: `087_automation_disable_on_compiled.up.sql`
```sql
ALTER TABLE "Automation_Config"
ADD COLUMN disable_on_compiled BOOLEAN NOT NULL DEFAULT FALSE;
```

### Repository: `internal/repositories/automation_repository.go`
- Add `DisableOnCompiled` to DTO and Input structs
- Add to `modelToDTO` and `applyInput`
- Keep existing `DisableConfig` method

### Serializer: `internal/serializers/automation_serializers.go`
Add `DisableOnCompiled` to response struct

### Service: `internal/services/automation_service.go`
Update run completion logic:
```go
if compiled && cfg.DisableOnCompiled {
    s.automationRepo.DisableConfig(cfg.ID)
    // Send config_enabled: false in SSE
}
// If !DisableOnCompiled, current "skipping" behavior continues
```

### Frontend: `app/automation/page.jsx`
- Add checkbox/toggle in config form: "Disable when target reached"
- Handle `config_enabled` in SSE events to update toggle state

## Critical Files
- `modules/agent/internal/models/AutomationConfig.go`
- `modules/agent/internal/repositories/automation_repository.go`
- `modules/agent/internal/services/automation_service.go`
- `modules/agent/internal/serializers/automation_serializers.go`
- `modules/agent/migrations/087_automation_disable_on_compiled.up.sql`
- `modules/client/app/automation/page.jsx`
