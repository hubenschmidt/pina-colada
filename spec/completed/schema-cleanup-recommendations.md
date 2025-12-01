# Schema Cleanup Recommendations

## Summary
Initial schema includes 6 models that aren't currently being used. This document provides recommendations for which to keep vs remove.

## Currently Unused Models

### ✅ KEEP - High Future Value

1. **Task** (`Task`)
   - **Why**: Core CRM functionality - task management is essential
   - **Usage**: Taskable polymorphic model (can attach to Deal, Lead, Job, Project, Organization, Individual, Contact)
   - **Recommendation**: Keep - will likely be needed soon

2. **Tag** (`Tag` + `EntityTag` join table)
   - **Why**: Flexible categorization system for any entity
   - **Usage**: Polymorphic tagging via EntityTag join table
   - **Recommendation**: Keep - useful for organization/filtering

3. **Project** (`Project`)
   - **Why**: Project management is common CRM feature
   - **Usage**: Tenant-scoped projects with dates/status
   - **Recommendation**: Keep - likely needed for deal/project tracking

4. **Opportunity** (`Opportunity`)
   - **Why**: Sales pipeline management - extends Lead via joined table inheritance
   - **Usage**: Business opportunities with value/probability tracking
   - **Recommendation**: Keep - core sales CRM functionality

5. **Partnership** (`Partnership`)
   - **Why**: Partnership tracking - extends Lead via joined table inheritance
   - **Usage**: Partnership opportunities with dates/type
   - **Recommendation**: Keep - useful for B2B relationship management

### ⚠️ CONSIDER REMOVING - Premature/Unclear Value

6. **Asset** (`Asset`)
   - **Why**: Generic content storage model - very abstract
   - **Usage**: Polymorphic content storage for AI agent context
   - **Recommendation**: **Remove** - too generic, can be added back if needed with clearer requirements
   - **Note**: If you need file/content storage later, design it for specific use cases

## Action Plan

### Option 1: Conservative (Recommended)
- Keep all 6 models
- They're already created, minimal overhead
- Allows flexibility for future features

### Option 2: Aggressive Cleanup
- Remove `Asset` model
- Keep the other 5 (Task, Tag, Project, Opportunity, Partnership)
- These align with standard CRM functionality

### Option 3: Full Cleanup (Not Recommended)
- Remove all unused models
- Risk: May need to recreate them later, losing data if any exists

## Models Currently In Use

These models ARE being actively used and should NOT be removed:

- **Core CRM**: Tenant, User, Role, UserRole, Status, Organization, Individual, Contact, Lead, Deal, Job, Account, Industry, Note
- **Lookup Tables**: SalaryRange, EmployeeCountRange, FundingStage, RevenueRange
- **Data Enrichment**: Technology, OrganizationTechnology, FundingRound, CompanySignal, DataProvenance
- **Reports**: SavedReport

## Migration Notes

If removing models:
1. Check for any foreign key relationships
2. Create migration to drop tables
3. Remove model files from `models/` directory
4. Remove imports from `models/__init__.py`
5. Update `Base.metadata.create_all()` will automatically exclude removed models

## Recommendation

**Keep Task, Tag, Project, Opportunity, Partnership** - these align with standard CRM features and will likely be needed.

**Remove Asset** - too generic, can be redesigned later if needed.

