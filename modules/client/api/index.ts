/**
 * API client
 */

import axios, { AxiosRequestConfig, AxiosResponse } from "axios";
import { CreatedJob } from "../types/types";
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
  const apiUrl = env("NEXT_PUBLIC_API_URL");
  const url = `${apiUrl}${path}`;

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
): Promise<PageData<CreatedJob>> => {
  const params = new URLSearchParams({
    page: page.toString(),
    limit: limit.toString(),
    orderBy,
    order,
  });
  if (search && search.trim()) {
    params.append("search", search.trim());
  }
  return apiGet<PageData<CreatedJob>>(`/jobs?${params}`);
};

const sanitize = (data: any): any => {
  const sanitized = { ...data };
  if (sanitized.date === "") sanitized.date = null;
  if (sanitized.resume === "") sanitized.resume = null;
  return sanitized;
};

export const createJob = async (
  job: Partial<CreatedJob>
): Promise<CreatedJob> => {
  return apiPost<CreatedJob>("/jobs", sanitize(job));
};

export const getJob = async (id: string): Promise<CreatedJob> => {
  return apiGet<CreatedJob>(`/jobs/${id}`);
};

export const updateJob = async (
  id: string,
  job: Partial<CreatedJob>
): Promise<CreatedJob> => {
  return apiPut<CreatedJob>(`/jobs/${id}`, sanitize(job));
};

export const deleteJob = async (id: string): Promise<void> => {
  await apiDelete(`/jobs/${id}`);
};

export const getRecentResumeDate = async (): Promise<string | null> => {
  try {
    const data = await apiGet<{ resume_date?: string }>(
      "/jobs/recent-resume-date"
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

export type JobWithLeadStatus = CreatedJob & {
  lead_status?: LeadStatus;
};

/**
 * Fetch all lead statuses from the database
 */
export const getStatuses = async (): Promise<LeadStatus[]> => {
  return apiGet<LeadStatus[]>("/lead-statuses");
};

/**
 * Fetch all job leads (jobs with non-null lead_status_id)
 * Optionally filter by lead status names
 */
export const getLeads = async (
  statusNames?: ("Qualifying" | "Cold" | "Warm" | "Hot")[]
): Promise<JobWithLeadStatus[]> => {
  const params = new URLSearchParams();
  if (statusNames && statusNames.length > 0) {
    params.append("statuses", statusNames.join(","));
  }
  return apiGet<JobWithLeadStatus[]>(`/leads?${params}`);
};

/**
 * Mark a lead as applied
 * Sets status to 'applied', clears lead_status_id, and sets date to now if not set
 */
export const markLeadAsApplied = async (jobId: string): Promise<CreatedJob> => {
  return apiPost<CreatedJob>(`/leads/${jobId}/apply`, {});
};

/**
 * Mark a lead as "do not apply"
 * Sets status to 'do_not_apply' and clears lead_status_id
 */
export const markLeadAsDoNotApply = async (
  jobId: string
): Promise<CreatedJob> => {
  return apiPost<CreatedJob>(`/leads/${jobId}/do-not-apply`, {});
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

export type TenantResponse = {
  id: number;
  name: string;
  created_at: string;
};

/**
 * Get current authenticated user
 */
export const getMe = async (): Promise<UserMeResponse> => {
  return apiGet<UserMeResponse>("/auth/me");
};

/**
 * Check if user has an associated tenant
 * Calls backend directly with user email
 * Returns tenant object if found, throws 404 error if not
 */
export const checkUserTenant = async (user: {
  email?: string | null;
}): Promise<TenantResponse> => {
  if (!user.email) {
    throw new Error("User email is required");
  }
  return apiGet<TenantResponse>(
    `/users/${encodeURIComponent(user.email)}/tenant`
  );
};

/**
 * Create a new tenant/organization
 */
export const createTenant = async (
  name: string,
  plan: string = "free"
): Promise<Tenant> => {
  const data = await apiPost<{ tenant: Tenant }>("/auth/tenant/create", {
    name,
    plan,
  });
  return data.tenant;
};
