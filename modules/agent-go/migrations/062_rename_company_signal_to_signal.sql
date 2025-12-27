-- Rename Company_Signal to Signal and change from organization_id to account_id
-- This allows signals to be associated with any Account (Organization or Individual)

-- Step 1: Add account_id column
ALTER TABLE "Company_Signal" ADD COLUMN IF NOT EXISTS account_id BIGINT;

-- Step 2: Populate account_id from organization's account_id
UPDATE "Company_Signal" cs
SET account_id = o.account_id
FROM "Organization" o
WHERE cs.organization_id = o.id;

-- Step 3: Drop the old foreign key constraint and organization_id column
ALTER TABLE "Company_Signal" DROP CONSTRAINT IF EXISTS "Company_Signal_organization_id_fkey";
ALTER TABLE "Company_Signal" DROP COLUMN IF EXISTS organization_id;

-- Step 4: Make account_id NOT NULL and add foreign key
ALTER TABLE "Company_Signal" ALTER COLUMN account_id SET NOT NULL;
ALTER TABLE "Company_Signal" ADD CONSTRAINT "Signal_account_id_fkey"
    FOREIGN KEY (account_id) REFERENCES "Account"(id) ON DELETE CASCADE;

-- Step 5: Rename the table
ALTER TABLE "Company_Signal" RENAME TO "Signal";

-- Step 6: Rename indexes
ALTER INDEX IF EXISTS idx_company_signal_org RENAME TO idx_signal_account;
ALTER INDEX IF EXISTS idx_company_signal_date RENAME TO idx_signal_date;
ALTER INDEX IF EXISTS idx_company_signal_type RENAME TO idx_signal_type;

-- Step 7: Create new index on account_id (old one was on organization_id)
DROP INDEX IF EXISTS idx_signal_account;
CREATE INDEX IF NOT EXISTS idx_signal_account ON "Signal"(account_id);
