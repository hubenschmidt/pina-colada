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

export const fetchJobs = async (
  page: number = 1,
  limit: number = 25,
  orderBy: string = "date",
  order: "ASC" | "DESC" = "DESC",
  search?: string
): Promise<PageData<AppliedJob>> => {
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
  
  const getOrderColumn = (orderBy: string): string => {
    if (orderBy === "job_title") return "job_title"
    if (orderBy === "company") return "company"
    if (orderBy === "status") return "status"
    if (orderBy === "date") return "date"
    if (orderBy === "resume") return "resume"
    return "date"
  }
  const orderColumn = getOrderColumn(orderBy)
  
  let countQuery = supabase.from('Job').select('*', { count: 'exact', head: true });
  let dataQuery = supabase.from('Job').select('*');
  
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
  
  if (USE_API_ROUTES) {
    const response = await fetch('/api/jobs', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(sanitized),
    });
    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}));
      throw new Error(errorData.error || 'Failed to create job');
    }
    return response.json();
  }
  
  const { supabase } = await import('./supabase');
  if (!supabase) {
    throw new Error('Supabase client not available');
  }
  const { data, error } = await supabase
    .from('Job')
    .insert(sanitized)
    .select()
    .single();
  
  if (error) throw error;
  return data;
}

export const updateJob = async (id: string, job: AppliedJobUpdate): Promise<AppliedJob> => {
  const sanitized = sanitizeJobUpdate(job);
  
  if (USE_API_ROUTES) {
    const response = await fetch(`/api/jobs/${id}`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(sanitized),
    });
    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}));
      throw new Error(errorData.error || 'Failed to update job');
    }
    return response.json();
  }
  
  const { supabase } = await import('./supabase');
  if (!supabase) {
    throw new Error('Supabase client not available');
  }
  const { data, error } = await supabase
    .from('Job')
    .update(sanitized)
    .eq('id', id)
    .select()
    .single();
  
  if (error) throw error;
  return data;
}

export const deleteJob = async (id: string): Promise<void> => {
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
    .from('Job')
    .delete()
    .eq('id', id);
  
  if (error) throw error;
}

export const getMostRecentResumeDate = async (): Promise<string | null> => {
  if (USE_API_ROUTES) {
    const response = await fetch('/api/jobs/recent-resume-date');
    if (!response.ok) {
      return null;
    }
    const data = await response.json();
    return data.resume_date || null;
  }
  
  const { supabase } = await import('./supabase');
  if (!supabase) {
    return null;
  }
  
  // Get the most recent job with a resume date, ordered by date DESC
  const { data, error } = await supabase
    .from('Job')
    .select('resume')
    .not('resume', 'is', null)
    .order('date', { ascending: false })
    .limit(1)
    .maybeSingle();
  
  if (error || !data || !data.resume) {
    return null;
  }
  
  // Convert timestamp to date string (YYYY-MM-DD format for date input)
  const date = new Date(data.resume);
  return date.toISOString().split('T')[0];
}

