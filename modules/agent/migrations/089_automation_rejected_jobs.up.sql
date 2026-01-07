-- Track jobs rejected by agent review (not user rejections)
CREATE TABLE "Automation_Rejected_Job" (
    id BIGSERIAL PRIMARY KEY,
    automation_config_id BIGINT NOT NULL REFERENCES "Automation_Config"(id) ON DELETE CASCADE,
    tenant_id BIGINT NOT NULL REFERENCES "Tenant"(id),
    url TEXT NOT NULL,
    job_title TEXT,
    company TEXT,
    reason TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_automation_rejected_job_url ON "Automation_Rejected_Job"(automation_config_id, url);
CREATE INDEX idx_automation_rejected_job_tenant ON "Automation_Rejected_Job"(tenant_id);
