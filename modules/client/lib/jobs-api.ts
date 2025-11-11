/**
 * Jobs API client - abstracts Supabase vs local Postgres
 * In dev: uses API routes (which connect to local Postgres)
 * In prod: uses Supabase directly
 */

import { AppliedJob, AppliedJobInsert, AppliedJobUpdate } from './supabase'
import { PageData } from '../components/DataTable'

// In development, use local Postgres via API routes
// In production, use Supabase directly
const USE_API_ROUTES = process.env.NODE_ENV === "development"

export async function fetchJobs(
  page: number = 1,
  limit: number = 25,
  orderBy: string = "application_date",
  order: "ASC" | "DESC" = "DESC",
  search?: string
): Promise<PageData<AppliedJob>> {
  if (USE_API_ROUTES) {
    const params = new URLSearchParams({
      page: page.toString(),
      limit: limit.toString(),
      orderBy,
      order,
    });
    if (search && search.trim()) {
      params.append('search', search.trim());
    }
    const response = await fetch(`/api/jobs?${params}`);
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
  
  const offset = (page - 1) * limit;
  const ascending = order === "ASC";
  const orderColumn = orderBy === "job_title" ? "job_title" :
                     orderBy === "company" ? "company" :
                     orderBy === "status" ? "status" :
                     "application_date";
  
  let countQuery = supabase.from('applied_jobs').select('*', { count: 'exact', head: true });
  let dataQuery = supabase.from('applied_jobs').select('*');
  
  // Apply search filter
  if (search && search.trim()) {
    const searchPattern = `%${search.trim()}%`;
    countQuery = countQuery.or(`company.ilike."${searchPattern}",job_title.ilike."${searchPattern}"`);
    dataQuery = dataQuery.or(`company.ilike."${searchPattern}",job_title.ilike."${searchPattern}"`);
  }
  
  // Get total count
  const { count } = await countQuery;
  
  // Get paginated data
  const { data, error } = await dataQuery
    .order(orderColumn, { ascending })
    .range(offset, offset + limit - 1);
  
  if (error) throw error;
  
  const total = count || 0;
  const totalPages = Math.max(1, Math.ceil(total / limit));
  
  return {
    items: data || [],
    currentPage: page,
    totalPages,
    total,
    pageSize: limit,
  };
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

