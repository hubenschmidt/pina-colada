-- Drop unused Lead junction tables
-- These were replaced by the Account-centric model (Lead -> Account -> Individual/Organization)

DROP TABLE IF EXISTS "Lead_Individual";
DROP TABLE IF EXISTS "Lead_Organization";
