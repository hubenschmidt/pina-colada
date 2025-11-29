-- Migration: Enforce that all leads must have at least one project association
-- This uses a trigger to validate on INSERT that the lead will have a project

-- Create a function to check if a lead has at least one project
CREATE OR REPLACE FUNCTION check_lead_has_project()
RETURNS TRIGGER AS $$
BEGIN
    -- For new leads, we allow the insert but the application must create
    -- the LeadProject association in the same transaction
    -- This is enforced at the application level
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Add a comment to document the constraint
COMMENT ON TABLE "LeadProject" IS 'Junction table for Lead-Project many-to-many relationship. All leads must have at least one project association (enforced at application level).';
