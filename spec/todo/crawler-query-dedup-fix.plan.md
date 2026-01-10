# Plan: Improve Job Search Query & Fix Duplicate Check Bug

## Summary
Three issues to fix:
1. **Search quality**: Query returns wrong job types, evaluator rejects valid matches
2. **Duplicate check bug**: "skipping duplicate [Pending]" log appears AFTER agent approval (wastes LLM tokens)
3. **Self-healing prompts**: When <50% approval rate, LLM suggests system_prompt improvements saved to DB for user review

---

## Issue 1: Search Query & Evaluator Improvements

### Problem
- Query `"Typescript javascript or python or golang engineer remote salary"` returns management/QA roles
- Evaluator prompt is too vague, rejects valid matches (Staff roles, adjacent skills)
- 8 results found, 0 proposals created

### Fix A: Better Query Structure

Add title exclusions to queries. Update in UI or seeder:

```
(senior OR staff OR lead) software engineer -manager -director -QA -SDET -test -intern
```

**Query syntax tips:**
- `intitle:"senior engineer"` - title must contain
- `(typescript OR golang)` - match any skill
- `-manager -QA` - exclude unwanted roles

### Fix B: Better Evaluator Prompt

Replace vague `system_prompt` with explicit criteria:

```
Evaluate for William Hubenschmidt, Senior Software/AI Engineer (Go/TypeScript/Python/Node/React).

APPROVE if:
- IC role at senior/staff/lead/principal level
- Uses ANY of: Go, TypeScript, JavaScript, Node.js, Python, React
- Full-time OR contract, remote OR hybrid OR NYC-based
- "Staff" or "Lead" in title (these are IC roles, NOT management)

APPROVE WITH CAUTION:
- "Engineering Manager" with <50% people duties → APPROVE if IC-heavy
- "Staff Security Engineer" → APPROVE if involves software dev
- Roles with secondary skills (Kubernetes, etc.) → APPROVE if core skills match

REJECT if:
- People management role (>50% people responsibilities)
- QA/SDET/Test Automation as PRIMARY function
- Junior/intern/entry-level
- Requires skills not in resume that can't be learned quickly

Be LENIENT on adjacent skills. "Platform Engineer" or "Backend Engineer" likely matches.
```

### Files to Modify
- **Database/UI**: Update `system_prompt` field on Automation_Config
- **Optional code change**: Add `JobTitleExclusions` to `lib/ats.go`:
```go
var JobTitleExclusions = []string{
    "-manager", "-director", "-QA", "-SDET", "-test", "-intern", "-junior",
}
```

---

## Issue 2: Duplicate Check Bug

### Problem
Log shows:
```
agent approved 1/5 results
skipping duplicate [Pending]: Job Application for Software Engineer at Applied Intuition
```

The LLM evaluated a job that was already marked "Pending" - wasting tokens.

### Root Cause
In `automation_service.go`, `tryCreateProposal()` (line ~673) checks duplicates AFTER the LLM has already approved them. This is redundant because:
1. `filterDuplicates()` already ran at line 585 BEFORE LLM
2. Items reaching `tryCreateProposal()` are already filtered

### Fix
In `tryCreateProposal()`, remove the redundant duplicate check that logs "skipping duplicate":

**File**: `internal/services/automation_service.go`

Find in `tryCreateProposal()` (~line 671-677):
```go
if source := dedup.duplicateSource(job.URL); source != "" {
    log.Printf("Automation service: skipping duplicate [%s]: %s at %s", source, job.Title, job.Company)
    return nil
}
```

This check is redundant - items already passed `filterDuplicates()` before reaching the LLM. Keep the `dedup.markURL(job.URL, "Proposal")` call that happens after successful creation.

---

## Issue 3: Self-Healing System Prompt Suggestions

### Problem
When the evaluator rejects too many jobs, the system_prompt criteria may be too strict or misaligned. Users must manually figure out what's wrong.

### Solution
After evaluation completes, if approval rate < 50%, ask LLM to suggest a system_prompt improvement.

### Logic Flow
1. After `evaluateJobsWithAgent()` returns, calculate approval rate: `approved / total`
2. If rate < 0.5 AND `suggested_prompt` is NULL → generate suggestion
3. Save suggestion to `Automation_Config.suggested_prompt`
4. UI displays suggestion with "Accept" button
5. On accept: copy suggestion to `system_prompt`, clear `suggested_prompt`

### Database Change

**Migration**: Add column to `Automation_Config`:
```sql
ALTER TABLE "Automation_Config" ADD COLUMN suggested_prompt TEXT;
```

### Backend Changes

**File**: `internal/services/automation_service.go`

After evaluation batch completes:
```go
func (s *AutomationService) checkAndSuggestPromptImprovement(cfg *repositories.AutomationConfigDTO, approved, total int, rejectedJobs []jobResult) {
    if total == 0 {
        return
    }
    rate := float64(approved) / float64(total)
    if rate >= 0.5 {
        return
    }
    if cfg.SuggestedPrompt != nil && *cfg.SuggestedPrompt != "" {
        return // suggestion already pending
    }

    suggestion := s.generatePromptSuggestion(cfg, approved, total, rejectedJobs)
    if suggestion != "" {
        s.repo.UpdateSuggestedPrompt(cfg.ID, suggestion)
    }
}

func (s *AutomationService) generatePromptSuggestion(cfg *repositories.AutomationConfigDTO, approved, total int, rejectedJobs []jobResult) string {
    // Include sample of rejected job titles to help LLM understand what's being filtered out
    rejectedSample := formatRejectedSample(rejectedJobs, 5)

    prompt := fmt.Sprintf(`The evaluator system_prompt had a %d%% approval rate (%d/%d jobs approved).

Current system_prompt:
%s

Sample rejected jobs:
%s

Suggest an improved system_prompt that would be more lenient on valid matches while still filtering irrelevant roles. Consider:
- Are adjacent skills being rejected unnecessarily?
- Are valid title variations (Staff, Lead, Principal) being missed?
- Is the prompt too strict on specific technologies?

Return ONLY the improved system_prompt text, nothing else.`,
        approved*100/total, approved, total, cfg.SystemPrompt, rejectedSample)

    // Call LLM and return suggestion
}
```

**File**: `internal/repositories/automation_repository.go`

```go
func (r *AutomationRepository) UpdateSuggestedPrompt(configID int64, suggestion string) error {
    return r.db.Model(&models.AutomationConfig{}).
        Where("id = ?", configID).
        Update("suggested_prompt", suggestion).Error
}

func (r *AutomationRepository) AcceptSuggestedPrompt(configID int64) error {
    return r.db.Transaction(func(tx *gorm.DB) error {
        var cfg models.AutomationConfig
        if err := tx.First(&cfg, configID).Error; err != nil {
            return err
        }
        if cfg.SuggestedPrompt == nil || *cfg.SuggestedPrompt == "" {
            return errors.New("no suggestion to accept")
        }
        return tx.Model(&cfg).Updates(map[string]interface{}{
            "system_prompt":     *cfg.SuggestedPrompt,
            "suggested_prompt":  nil,
        }).Error
    })
}
```

**File**: `internal/controllers/automation_controller.go`

New endpoint:
```go
// POST /api/automation-configs/:id/accept-suggestion
func (c *AutomationController) AcceptSuggestedPrompt(ctx *gin.Context) {
    id, _ := strconv.ParseInt(ctx.Param("id"), 10, 64)
    if err := c.repo.AcceptSuggestedPrompt(id); err != nil {
        ctx.JSON(500, gin.H{"error": err.Error()})
        return
    }
    ctx.JSON(200, gin.H{"message": "suggestion accepted"})
}
```

### Frontend Changes

Display suggestion banner on crawler config page when `suggested_prompt` is present:
```jsx
{config.suggested_prompt && (
  <div className="suggestion-banner">
    <p>Suggested system prompt improvement (based on low approval rate):</p>
    <pre>{config.suggested_prompt}</pre>
    <button onClick={() => acceptSuggestion(config.id)}>Accept</button>
    <button onClick={() => dismissSuggestion(config.id)}>Dismiss</button>
  </div>
)}
```

---

## Verification

1. **Query improvements**:
   - Update system_prompt via UI
   - Run crawler, verify fewer rejections for valid roles

2. **Duplicate bug fix**:
   - Run crawler twice with overlapping results
   - Verify "skipping duplicate" logs appear BEFORE "agent approved" (or not at all)
   - Check no wasted LLM calls on already-pending jobs

3. **Self-healing prompt suggestions**:
   - Run crawler with strict system_prompt that produces <50% approval rate
   - Verify `suggested_prompt` is populated in DB
   - Verify suggestion appears in crawler config UI
   - Click "Accept", verify system_prompt updated and suggestion cleared
   - Run again with same low rate, verify no new suggestion (already pending logic works)
