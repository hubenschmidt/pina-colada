CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_agent_proposal_tenant_status_entity
  ON "Agent_Proposal" (tenant_id, status, entity_type);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_automation_run_log_config_status
  ON "Automation_Run_Log" (automation_config_id, status);
