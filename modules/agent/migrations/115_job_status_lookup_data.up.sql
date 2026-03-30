-- Insert required job status lookup data
INSERT INTO "Status" (name, description, category, is_terminal, created_at, updated_at)
VALUES ('Lead', 'Job lead not yet applied to', 'job', FALSE, NOW(), NOW())
ON CONFLICT (name) DO NOTHING;

INSERT INTO "Status" (name, description, category, is_terminal, created_at, updated_at)
VALUES ('Do Not Apply', 'Job decided not to apply for', 'job', TRUE, NOW(), NOW())
ON CONFLICT (name) DO NOTHING;
