/**
 * Jobs API client - always calls agent's REST API
 * The agent's repository layer handles Supabase vs local Postgres based on environment
 */

import { AppliedJob, AppliedJobInsert, AppliedJobUpdate } from '../lib/supabase'
import { PageData } from '../components/DataTable'

const getAgentApiUrl = (): string => {
  if (typeof window === "undefined") return "http://localhost:8000"
  if (window.location.hostname === "pinacolada.co" || window.location.hostname === "www.pinacolada.co") {
    return "https://api.pinacolada.co"
  }
  return "http://localhost:8000"
}

export const fetchJobs = async (
  page: number = 1,
  limit: number = 25,
  orderBy: string = "date",
  order: "ASC" | "DESC" = "DESC",
  search?: string
): Promise<PageData<AppliedJob>> => {
  const baseUrl = getAgentApiUrl();
  const params = new URLSearchParams({
    page: page.toString(),
    limit: limit.toString(),
    orderBy,
    order,
  });
  if (search && search.trim()) {
    params.append('search', search.trim());
  }
  const response = await fetch(`${baseUrl}/api/jobs?${params}`);
  if (!response.ok) {
    throw new Error('Failed to fetch jobs');
  }
  return response.json();
}

export const createJob = async (job: AppliedJobInsert): Promise<AppliedJob> => {
  const baseUrl = getAgentApiUrl();
  const response = await fetch(`${baseUrl}/api/jobs`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(job),
  });
  if (!response.ok) {
    throw new Error('Failed to create job');
  }
  return response.json();
}

export const updateJob = async (id: string, job: AppliedJobUpdate): Promise<AppliedJob> => {
  const baseUrl = getAgentApiUrl();
  const response = await fetch(`${baseUrl}/api/jobs/${id}`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(job),
  });
  if (!response.ok) {
    throw new Error('Failed to update job');
  }
  return response.json();
}

export const deleteJob = async (id: string): Promise<void> => {
  const baseUrl = getAgentApiUrl();
  const response = await fetch(`${baseUrl}/api/jobs/${id}`, {
    method: 'DELETE',
  });
  if (!response.ok) {
    throw new Error('Failed to delete job');
  }
}

// Lead Status Types
export type LeadStatus = {
  id: string
  name: 'Qualifying' | 'Cold' | 'Warm' | 'Hot'
  description: string | null
  created_at: string
}

export type JobWithLeadStatus = AppliedJob & {
  lead_status?: LeadStatus
}

/**
 * Fetch all lead statuses from the database
 */
export const fetchLeadStatuses = async (): Promise<LeadStatus[]> => {
  const baseUrl = getAgentApiUrl();
  const response = await fetch(`${baseUrl}/api/lead-statuses`);
  if (!response.ok) {
    throw new Error('Failed to fetch lead statuses');
  }
  return response.json();
}

/**
 * Fetch all job leads (jobs with non-null lead_status_id)
 * Optionally filter by lead status names
 */
export const fetchLeads = async (
  statusNames?: ('Qualifying' | 'Cold' | 'Warm' | 'Hot')[]
): Promise<JobWithLeadStatus[]> => {
  const baseUrl = getAgentApiUrl();
  const params = new URLSearchParams();
  if (statusNames && statusNames.length > 0) {
    params.append('statuses', statusNames.join(','));
  }
  const response = await fetch(`${baseUrl}/api/leads?${params}`);
  if (!response.ok) {
    throw new Error('Failed to fetch leads');
  }
  return response.json();
}

/**
 * Update a job's lead status
 */
export const updateJobLeadStatus = async (
  jobId: string,
  leadStatusId: string | null
): Promise<AppliedJob> => {
  const baseUrl = getAgentApiUrl();
  const response = await fetch(`${baseUrl}/api/jobs/${jobId}`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ lead_status_id: leadStatusId }),
  });
  if (!response.ok) {
    throw new Error('Failed to update job lead status');
  }
  return response.json();
}

/**
 * Mark a lead as applied
 * Sets status to 'applied', clears lead_status_id, and sets date to now if not set
 */
export const markLeadAsApplied = async (jobId: string): Promise<AppliedJob> => {
  const baseUrl = getAgentApiUrl();
  const response = await fetch(`${baseUrl}/api/leads/${jobId}/apply`, {
    method: 'POST',
  });
  if (!response.ok) {
    throw new Error('Failed to mark lead as applied');
  }
  return response.json();
}

/**
 * Mark a lead as "do not apply"
 * Sets status to 'do_not_apply' and clears lead_status_id
 */
export const markLeadAsDoNotApply = async (jobId: string): Promise<AppliedJob> => {
  const baseUrl = getAgentApiUrl();
  const response = await fetch(`${baseUrl}/api/leads/${jobId}/do-not-apply`, {
    method: 'POST',
  });
  if (!response.ok) {
    throw new Error('Failed to mark lead as do not apply');
  }
  return response.json();
}

