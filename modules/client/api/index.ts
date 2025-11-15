/**
 * API client
 */

import axios, { AxiosRequestConfig, AxiosResponse } from "axios";
import {
  AppliedJob,
  AppliedJobInsert,
  AppliedJobUpdate,
} from "../lib/supabase";
import { PageData } from "../components/DataTable";
import { env } from "next-runtime-env";
import { fetchBearerToken } from "../lib/fetch-bearer-token";

const getClient = () => axios.create();

const makeRequest = async <T>(
  client: ReturnType<typeof getClient>,
  method: "get" | "post" | "put" | "delete",
  url: string,
  data: any,
  config: AxiosRequestConfig
): Promise<AxiosResponse<T>> => {
  if (method === "get") return client.get(url, config);
  if (method === "delete") return client.delete(url, config);
  if (method === "post") return client.post(url, data, config);
  return client.put(url, data, config);
};

const apiRequest = async <T>(
  method: "get" | "post" | "put" | "delete",
  path: string,
  data?: any,
  config?: AxiosRequestConfig
): Promise<T> => {
  const client = getClient();
  const authHeaders = await fetchBearerToken();
  const mergedConfig = { ...authHeaders, ...config };
  const url = `${env("NEXT_PUBLIC_API_URL")}${path}`;

  return makeRequest<T>(client, method, url, data, mergedConfig)
    .then((response) => response.data)
    .catch((error) => {
      const errorData = error.response?.data || {};
      const message =
        errorData.detail ||
        errorData.error ||
        error.message ||
        "Request failed";
      throw new Error(message);
    });
};

const apiGet = <T>(path: string, config?: AxiosRequestConfig) =>
  apiRequest<T>("get", path, undefined, config);

const apiPost = <T>(path: string, data?: any, config?: AxiosRequestConfig) =>
  apiRequest<T>("post", path, data, config);

const apiPut = <T>(path: string, data?: any, config?: AxiosRequestConfig) =>
  apiRequest<T>("put", path, data, config);

const apiDelete = <T>(path: string, config?: AxiosRequestConfig) =>
  apiRequest<T>("delete", path, undefined, config);

export const getJobs = async (
  page: number = 1,
  limit: number = 25,
  orderBy: string = "date",
  order: "ASC" | "DESC" = "DESC",
  search?: string
): Promise<PageData<AppliedJob>> => {
  const params = new URLSearchParams({
    page: page.toString(),
    limit: limit.toString(),
    orderBy,
    order,
  });
  if (search && search.trim()) {
    params.append("search", search.trim());
  }
  return apiGet<PageData<AppliedJob>>(`/api/jobs?${params}`);
};

// Helper to sanitize date fields: convert empty strings to null
const sanitizeJobInsert = (job: AppliedJobInsert): AppliedJobInsert => {
  const sanitized = { ...job };
  // Convert empty strings to null for date/timestamp fields
  if (sanitized.date === "") {
    sanitized.date = null as any;
  }
  if (sanitized.resume === "") {
    sanitized.resume = null as any;
  }
  return sanitized;
};

const sanitizeJobUpdate = (job: AppliedJobUpdate): AppliedJobUpdate => {
  const sanitized = { ...job };
  // Convert empty strings to null for date/timestamp fields
  if ("date" in sanitized && sanitized.date === "") {
    sanitized.date = null as any;
  }
  if ("resume" in sanitized && sanitized.resume === "") {
    sanitized.resume = null as any;
  }
  return sanitized;
};

export const createJob = async (job: AppliedJobInsert): Promise<AppliedJob> => {
  const sanitized = sanitizeJobInsert(job);
  return apiPost<AppliedJob>("/api/jobs", sanitized);
};

export const get_job = async (id: string): Promise<AppliedJob> => {
  return apiGet<AppliedJob>(`/api/jobs/${id}`);
};

export const updateJob = async (
  id: string,
  job: AppliedJobUpdate
): Promise<AppliedJob> => {
  const sanitized = sanitizeJobUpdate(job);
  return apiPut<AppliedJob>(`/api/jobs/${id}`, sanitized);
};

export const deleteJob = async (id: string): Promise<void> => {
  await apiDelete(`/api/jobs/${id}`);
};

export const get_recent_resume_date = async (): Promise<string | null> => {
  try {
    const data = await apiGet<{ resume_date?: string }>(
      "/api/jobs/recent-resume-date"
    );
    return data.resume_date || null;
  } catch {
    return null;
  }
};

// Lead Status Types
export type LeadStatus = {
  id: string;
  name: "Qualifying" | "Cold" | "Warm" | "Hot";
  description: string | null;
  created_at: string;
};

export type JobWithLeadStatus = AppliedJob & {
  lead_status?: LeadStatus;
};

/**
 * Fetch all lead statuses from the database
 */
export const get_statuses = async (): Promise<LeadStatus[]> => {
  return apiGet<LeadStatus[]>("/api/lead-statuses");
};

/**
 * Fetch all job leads (jobs with non-null lead_status_id)
 * Optionally filter by lead status names
 */
export const get_leads = async (
  statusNames?: ("Qualifying" | "Cold" | "Warm" | "Hot")[]
): Promise<JobWithLeadStatus[]> => {
  const params = new URLSearchParams();
  if (statusNames && statusNames.length > 0) {
    params.append("statuses", statusNames.join(","));
  }
  return apiGet<JobWithLeadStatus[]>(`/api/leads?${params}`);
};

/**
 * Mark a lead as applied
 * Sets status to 'applied', clears lead_status_id, and sets date to now if not set
 */
export const mark_lead_as_applied = async (
  jobId: string
): Promise<AppliedJob> => {
  return apiPost<AppliedJob>(`/api/leads/${jobId}/apply`, {});
};

/**
 * Mark a lead as "do not apply"
 * Sets status to 'do_not_apply' and clears lead_status_id
 */
export const mark_lead_as_do_not_apply = async (
  jobId: string
): Promise<AppliedJob> => {
  return apiPost<AppliedJob>(`/api/leads/${jobId}/do-not-apply`, {});
};

// Tenant Types
export type Tenant = {
  id: number;
  name: string;
  slug: string;
  plan: string;
  role?: string;
};

// User Types
export type User = {
  id: number;
  auth0_sub: string;
  email: string | null;
  tenant_id: number | null;
  created_at: string;
  updated_at: string;
};

export type UserMeResponse = {
  user: User;
  tenants: Array<{ id: number; name: string; role: string }>;
  current_tenant_id: number | null;
};

/**
 * Get current authenticated user
 */
export const get_me = async (): Promise<UserMeResponse> => {
  return apiGet<UserMeResponse>("/api/auth/me");
};

/**
 * Create a new tenant/organization
 */
export const createTenant = async (
  name: string,
  plan: string = "free"
): Promise<Tenant> => {
  const data = await apiPost<{ tenant: Tenant }>("/api/auth/tenant/create", {
    name,
    plan,
  });
  return data.tenant;
};
