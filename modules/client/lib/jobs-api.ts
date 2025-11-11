/**
 * Jobs API client - abstracts Supabase vs local Postgres
 * In dev: uses API routes (which connect to local Postgres)
 * In prod: uses Supabase directly
 */

import { AppliedJob, AppliedJobInsert, AppliedJobUpdate } from './supabase'

// In development, use local Postgres via API routes
// In production, use Supabase directly
const USE_API_ROUTES = process.env.NODE_ENV === "development"

export async function fetchJobs(): Promise<AppliedJob[]> {
  if (USE_API_ROUTES) {
    const response = await fetch('/api/jobs');
    if (!response.ok) {
      throw new Error('Failed to fetch jobs');
    }
    return response.json();
  }
  
  // Production: Use Supabase directly
  const { supabase } = await import('./supabase');
  if (!supabase) {
    throw new Error('Supabase client not available');
  }
  const { data, error } = await supabase
    .from('applied_jobs')
    .select('*')
    .order('application_date', { ascending: false });
  
  if (error) throw error;
  return data || [];
}

export async function createJob(job: AppliedJobInsert): Promise<AppliedJob> {
  if (USE_API_ROUTES) {
    const response = await fetch('/api/jobs', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(job),
    });
    if (!response.ok) {
      throw new Error('Failed to create job');
    }
    return response.json();
  }
  
  const { supabase } = await import('./supabase');
  if (!supabase) {
    throw new Error('Supabase client not available');
  }
  const { data, error } = await supabase
    .from('applied_jobs')
    .insert(job)
    .select()
    .single();
  
  if (error) throw error;
  return data;
}

export async function updateJob(id: string, job: AppliedJobUpdate): Promise<AppliedJob> {
  if (USE_API_ROUTES) {
    const response = await fetch(`/api/jobs/${id}`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(job),
    });
    if (!response.ok) {
      throw new Error('Failed to update job');
    }
    return response.json();
  }
  
  const { supabase } = await import('./supabase');
  if (!supabase) {
    throw new Error('Supabase client not available');
  }
  const { data, error } = await supabase
    .from('applied_jobs')
    .update(job)
    .eq('id', id)
    .select()
    .single();
  
  if (error) throw error;
  return data;
}

export async function deleteJob(id: string): Promise<void> {
  if (USE_API_ROUTES) {
    const response = await fetch(`/api/jobs/${id}`, {
      method: 'DELETE',
    });
    if (!response.ok) {
      throw new Error('Failed to delete job');
    }
    return;
  }
  
  const { supabase } = await import('./supabase');
  if (!supabase) {
    throw new Error('Supabase client not available');
  }
  const { error } = await supabase
    .from('applied_jobs')
    .delete()
    .eq('id', id);
  
  if (error) throw error;
}

