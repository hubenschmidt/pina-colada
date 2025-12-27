-- Create system tenant for global resources (e.g., URL shortener cache)
-- Using id=0 to distinguish from user tenants which start at 1
INSERT INTO "Tenant" (id, name, slug, created_at)
VALUES (0, 'System', 'system', NOW())
ON CONFLICT (id) DO NOTHING;
