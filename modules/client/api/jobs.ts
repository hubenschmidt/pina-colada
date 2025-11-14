/**
 * Jobs API client - always calls agent's REST API
 * The agent's repository layer handles Supabase vs local Postgres based on environment
 */

import { AppliedJob, AppliedJobInsert, AppliedJobUpdate } from '../lib/supabase'
import { PageData } from '../components/DataTable'
import { env } from 'next-runtime-env'

export const fetchJobs = async (
  page: number = 1,
  limit: number = 25,
  orderBy: string = "date",
  order: "ASC" | "DESC" = "DESC",
  search?: string
): Promise<PageData<AppliedJob>> => {
  const baseUrl = env('NEXT_PUBLIC_API_URL');
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

// Helper to sanitize date fields: convert empty strings to null
const sanitizeJobInsert = (job: AppliedJobInsert): AppliedJobInsert => {
  const sanitized = { ...job };
  // Convert empty strings to null for date/timestamp fields
  if (sanitized.date === '') {
    sanitized.date = null as any;
  }
  if (sanitized.resume === '') {
    sanitized.resume = null as any;
  }
  return sanitized;
};

const sanitizeJobUpdate = (job: AppliedJobUpdate): AppliedJobUpdate => {
  const sanitized = { ...job };
  // Convert empty strings to null for date/timestamp fields
  if ('date' in sanitized && sanitized.date === '') {
    sanitized.date = null as any;
  }
  if ('resume' in sanitized && sanitized.resume === '') {
    sanitized.resume = null as any;
  }
  return sanitized;
};

export const createJob = async (job: AppliedJobInsert): Promise<AppliedJob> => {
  const sanitized = sanitizeJobInsert(job);
  const baseUrl = env('NEXT_PUBLIC_API_URL');
  const response = await fetch(`${baseUrl}/api/jobs`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(sanitized),
  });
  if (!response.ok) {
    const errorData = await response.json().catch(() => ({}));
    throw new Error(errorData.detail || errorData.error || 'Failed to create job');
  }
  return response.json();
}

export const updateJob = async (id: string, job: AppliedJobUpdate): Promise<AppliedJob> => {
  const sanitized = sanitizeJobUpdate(job);
  const baseUrl = env('NEXT_PUBLIC_API_URL');
  const response = await fetch(`${baseUrl}/api/jobs/${id}`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(sanitized),
  });
  if (!response.ok) {
    const errorData = await response.json().catch(() => ({}));
    throw new Error(errorData.detail || errorData.error || 'Failed to update job');
  }
  return response.json();
}

export const deleteJob = async (id: string): Promise<void> => {
  const baseUrl = env('NEXT_PUBLIC_API_URL');
  const response = await fetch(`${baseUrl}/api/jobs/${id}`, {
    method: 'DELETE',
  });
  if (!response.ok) {
    const errorData = await response.json().catch(() => ({}));
    throw new Error(errorData.detail || errorData.error || 'Failed to delete job');
  }
}

export const getMostRecentResumeDate = async (): Promise<string | null> => {
  const baseUrl = env('NEXT_PUBLIC_API_URL');
  const response = await fetch(`${baseUrl}/api/jobs/recent-resume-date`);
  if (!response.ok) {
    return null;
  }
  const data = await response.json();
  return data.resume_date || null;
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
  const baseUrl = env('NEXT_PUBLIC_API_URL');
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
  const baseUrl = env('NEXT_PUBLIC_API_URL');
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
  const baseUrl = env('NEXT_PUBLIC_API_URL');
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
  const baseUrl = env('NEXT_PUBLIC_API_URL');
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
  const baseUrl = env('NEXT_PUBLIC_API_URL');
  const response = await fetch(`${baseUrl}/api/leads/${jobId}/do-not-apply`, {
    method: 'POST',
  });
  if (!response.ok) {
    throw new Error('Failed to mark lead as do not apply');
  }
  return response.json();
}

