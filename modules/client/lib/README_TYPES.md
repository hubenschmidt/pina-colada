# Supabase Type Generation

This project uses TypeScript types generated from the Supabase database schema.

## Generating Types

### From Production Database

To generate types from your production Supabase project:

1. **Set your project reference** (found in your Supabase project URL):
   ```bash
   export SUPABASE_PROJECT_REF="your-project-ref"
   ```

2. **Login to Supabase** (if not already logged in):
   ```bash
   npx supabase login
   ```

3. **Generate types**:
   ```bash
   npm run generate-types
   ```

   Or manually:
   ```bash
   npx supabase gen types typescript --project-id "$SUPABASE_PROJECT_REF" --schema public > lib/database.types.ts
   ```

### From Local Database

If you're running Supabase locally:

1. **Start local Supabase** (if using Supabase CLI):
   ```bash
   npx supabase start
   ```

2. **Generate types**:
   ```bash
   npm run generate-types:local
   ```

   Or manually:
   ```bash
   npx supabase gen types typescript --local --schema public > lib/database.types.ts
   ```

## Using Types

The generated types are automatically used in `lib/supabase.ts`:

```typescript
import { supabase, AppliedJob, AppliedJobInsert, AppliedJobUpdate } from './lib/supabase'

// Type-safe queries
const { data } = await supabase
  .from('applied_jobs')
  .select('*')
  // TypeScript will autocomplete and validate fields

// Type-safe inserts
const newJob: AppliedJobInsert = {
  company: 'Acme Corp',
  job_title: 'Software Engineer',
  // TypeScript ensures required fields are present
}
```

## Updating Types

**Important**: Regenerate types whenever you:
- Add/modify tables or columns
- Change column types or constraints
- Add new enums or functions

The types in `lib/database.types.ts` should match your actual database schema.

## References

- [Supabase Type Generation Docs](https://supabase.com/docs/guides/api/rest/generating-types#generating-types-using-supabase-cli)
- [Supabase CLI Docs](https://supabase.com/docs/reference/cli)

