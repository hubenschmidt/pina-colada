-- Migration: Add phone format and founding_year validation constraints
-- These ensure data integrity when AI agent bypasses API validation

-- Contact table: phone format constraint
ALTER TABLE "Contact"
ADD CONSTRAINT contact_phone_format_check
CHECK (phone IS NULL OR phone ~ '^\+1-\d{3}-\d{3}-\d{4}$');

-- Organization table: phone format constraint
ALTER TABLE "Organization"
ADD CONSTRAINT organization_phone_format_check
CHECK (phone IS NULL OR phone ~ '^\+1-\d{3}-\d{3}-\d{4}$');

-- Organization table: founding_year range constraint
ALTER TABLE "Organization"
ADD CONSTRAINT organization_founding_year_check
CHECK (founding_year IS NULL OR (founding_year >= 1800 AND founding_year <= EXTRACT(YEAR FROM NOW())));

-- Individual table: phone format constraint
ALTER TABLE "Individual"
ADD CONSTRAINT individual_phone_format_check
CHECK (phone IS NULL OR phone ~ '^\+1-\d{3}-\d{3}-\d{4}$');
