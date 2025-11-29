/**
 * API client
 */

import axios, { AxiosRequestConfig, AxiosResponse } from "axios";
import { CreatedJob } from "../types/types";
export type { CreatedJob } from "../types/types";
import { PageData } from "../components/DataTable/DataTable";
import { env } from "next-runtime-env";
import { fetchBearerToken } from "../lib/fetch-bearer-token";

const getClient = () => axios.create();

const makeRequest = async <T>(
  client: ReturnType<typeof getClient>,
  method: "get" | "post" | "put" | "delete" | "patch",
  url: string,
  data: any,
  config: AxiosRequestConfig
): Promise<AxiosResponse<T>> => {
  if (method === "get") return client.get(url, config);
  if (method === "delete") return client.delete(url, config);
  if (method === "post") return client.post(url, data, config);
  if (method === "patch") return client.patch(url, data, config);
  return client.put(url, data, config);
};

const apiRequest = async <T>(
  method: "get" | "post" | "put" | "delete" | "patch",
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
      // Handle Pydantic validation errors (422) which return detail as array
      const message = Array.isArray(errorData.detail)
        ? errorData.detail
            .map((e: { loc?: string[]; msg?: string }) =>
              `${e.loc?.slice(1).join(".") || "field"}: ${e.msg || "invalid"}`
            )
            .join("; ")
        : errorData.detail || errorData.error || error.message || "Request failed";
      console.error("API Error:", { status: error.response?.status, message, data: errorData });
      const errorObj = new Error(message);
      (errorObj as any).response = error.response;
      (errorObj as any).errorData = errorData;
      throw errorObj;
    });
};

const apiGet = <T>(path: string, config?: AxiosRequestConfig) =>
  apiRequest<T>("get", path, undefined, config);

const apiPost = <T>(path: string, data?: any, config?: AxiosRequestConfig) =>
  apiRequest<T>("post", path, data, config);

const apiPut = <T>(path: string, data?: any, config?: AxiosRequestConfig) =>
  apiRequest<T>("put", path, data, config);

const apiPatch = <T>(path: string, data?: any, config?: AxiosRequestConfig) =>
  apiRequest<T>("patch", path, data, config);

const apiDelete = <T>(path: string, config?: AxiosRequestConfig) =>
  apiRequest<T>("delete", path, undefined, config);

export const getJobs = async (
  page: number = 1,
  limit: number = 50,
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
  return apiGet<LeadStatus[]>("/leads/statuses");
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

// Preferences Types
export type UserPreferencesResponse = {
  theme: string | null;
  effective_theme: string;
  can_edit_tenant: boolean;
};

export type TenantPreferencesResponse = {
  theme: string;
};

/**
 * Get current user's preferences
 */
export const getUserPreferences = async (): Promise<UserPreferencesResponse> => {
  return apiGet<UserPreferencesResponse>("/preferences/user");
};

/**
 * Update current user's theme preference
 */
export const updateUserPreferences = async (
  theme: "light" | "dark" | null
): Promise<UserPreferencesResponse> => {
  return apiPatch<UserPreferencesResponse>("/preferences/user", { theme });
};

/**
 * Get tenant preferences (Admin/SuperAdmin only)
 */
export const getTenantPreferences = async (): Promise<TenantPreferencesResponse> => {
  return apiGet<TenantPreferencesResponse>("/preferences/tenant");
};

/**
 * Update tenant theme preference (Admin/SuperAdmin only)
 */
export const updateTenantPreferences = async (
  theme: "light" | "dark"
): Promise<TenantPreferencesResponse> => {
  return apiPatch<TenantPreferencesResponse>("/preferences/tenant", { theme });
};

// Contact Types
export type ContactIndividualRef = {
  id: number;
  first_name: string;
  last_name: string;
};

export type ContactOrganizationRef = {
  id: number;
  name: string;
};

export type Contact = {
  id: number;
  first_name: string;
  last_name: string;
  title: string | null;
  department: string | null;
  role: string | null;
  email: string | null;
  phone: string | null;
  is_primary: boolean;
  notes: string | null;
  individuals: ContactIndividualRef[];
  organizations: ContactOrganizationRef[];
  created_at: string | null;
  updated_at: string | null;
};

// Individual Types
export type Individual = {
  id: number;
  first_name: string;
  last_name: string;
  email: string | null;
  phone: string | null;
  linkedin_url: string | null;
  title: string | null;
  notes: string | null;
  industries: string[];
  contacts?: Contact[];
  created_at: string | null;
  updated_at: string | null;
};

/**
 * Get all individuals for the current tenant
 */
export const getIndividuals = async (): Promise<Individual[]> => {
  return apiGet<Individual[]>("/individuals");
};

/**
 * Get a single individual by ID
 */
export const getIndividual = async (id: number): Promise<Individual> => {
  return apiGet<Individual>(`/individuals/${id}`);
};

/**
 * Create a new individual
 */
export const createIndividual = async (
  data: Partial<Omit<Individual, "id" | "created_at" | "updated_at">>
): Promise<Individual> => {
  return apiPost<Individual>("/individuals", data);
};

/**
 * Update an individual
 */
export const updateIndividual = async (
  id: number,
  data: Partial<Omit<Individual, "id" | "created_at" | "updated_at">>
): Promise<Individual> => {
  return apiPut<Individual>(`/individuals/${id}`, data);
};

/**
 * Delete an individual
 */
export const deleteIndividual = async (id: number): Promise<void> => {
  await apiDelete(`/individuals/${id}`);
};

// Organization Types
export type Organization = {
  id: number;
  name: string;
  website: string | null;
  phone: string | null;
  industries: string[];
  employee_count: number | null;  // Legacy field
  employee_count_range_id: number | null;
  employee_count_range: string | null;  // Label from lookup
  funding_stage_id: number | null;
  funding_stage: string | null;  // Label from lookup
  description: string | null;
  contacts?: Contact[];
  created_at: string | null;
  updated_at: string | null;
};

/**
 * Get all organizations for the current tenant
 */
export const getOrganizations = async (): Promise<Organization[]> => {
  return apiGet<Organization[]>("/organizations");
};

/**
 * Get a single organization by ID
 */
export const getOrganization = async (id: number): Promise<Organization> => {
  return apiGet<Organization>(`/organizations/${id}`);
};

/**
 * Create a new organization
 */
export const createOrganization = async (
  data: Partial<Omit<Organization, "id" | "created_at" | "updated_at" | "industries">>
): Promise<Organization> => {
  return apiPost<Organization>("/organizations", data);
};

/**
 * Update an organization
 */
export const updateOrganization = async (
  id: number,
  data: Partial<Omit<Organization, "id" | "created_at" | "updated_at" | "industries">>
): Promise<Organization> => {
  return apiPut<Organization>(`/organizations/${id}`, data);
};

/**
 * Delete an organization
 */
export const deleteOrganization = async (id: number): Promise<void> => {
  await apiDelete(`/organizations/${id}`);
};

/**
 * Search organizations by name
 */
export const searchOrganizations = async (query: string): Promise<Organization[]> => {
  return apiGet<Organization[]>(`/organizations/search?q=${encodeURIComponent(query)}`);
};

/**
 * Search individuals by name or email
 */
export const searchIndividuals = async (query: string): Promise<Individual[]> => {
  return apiGet<Individual[]>(`/individuals/search?q=${encodeURIComponent(query)}`);
};

// Account Search Types
export type AccountSearchResult = {
  id: number;
  account_id: number;
  name: string;
  type: "organization" | "individual" | "unknown";
};

/**
 * Search accounts by name
 */
export const searchAccounts = async (query: string): Promise<AccountSearchResult[]> => {
  return apiGet<AccountSearchResult[]>(`/accounts/search?q=${encodeURIComponent(query)}`);
};

// Contact Search Types
export type ContactSearchResult = {
  individual_id: number | null;
  contact_id: number | null;
  first_name: string;
  last_name: string;
  email: string | null;
  phone: string | null;
  source: "individual" | "contact";
};

/**
 * Search contacts and individuals by name or email.
 * Returns results that can be linked as contacts via individual_id.
 */
export const searchContacts = async (query: string): Promise<ContactSearchResult[]> => {
  return apiGet<ContactSearchResult[]>(`/contacts/search?q=${encodeURIComponent(query)}`);
};

// Industry Types
export type Industry = {
  id: number;
  name: string;
  created_at: string | null;
};

/**
 * Get all industries
 */
export const getIndustries = async (): Promise<Industry[]> => {
  return apiGet<Industry[]>("/industries");
};

/**
 * Create a new industry
 */
export const createIndustry = async (name: string): Promise<Industry> => {
  return apiPost<Industry>("/industries", { name });
};

// Salary Range Types
export type SalaryRange = {
  id: number;
  label: string;
  min_value: number | null;
  max_value: number | null;
  display_order: number;
};

/**
 * Get all salary ranges
 */
export const getSalaryRanges = async (): Promise<SalaryRange[]> => {
  return apiGet<SalaryRange[]>("/salary-ranges");
};

// Employee Count Range Types
export type EmployeeCountRange = {
  id: number;
  label: string;
  min_value: number | null;
  max_value: number | null;
  display_order: number;
};

/**
 * Get all employee count ranges
 */
export const getEmployeeCountRanges = async (): Promise<EmployeeCountRange[]> => {
  return apiGet<EmployeeCountRange[]>("/employee-count-ranges");
};

// Funding Stage Types
export type FundingStage = {
  id: number;
  label: string;
  display_order: number;
};

/**
 * Get all funding stages
 */
export const getFundingStages = async (): Promise<FundingStage[]> => {
  return apiGet<FundingStage[]>("/funding-stages");
};

// Standalone Contact CRUD

/**
 * Get all contacts
 */
export const getContacts = async (): Promise<Contact[]> => {
  return apiGet<Contact[]>("/contacts");
};

/**
 * Get a single contact by ID
 */
export const getContact = async (id: number): Promise<Contact> => {
  return apiGet<Contact>(`/contacts/${id}`);
};

export type ContactInput = {
  first_name?: string;
  last_name?: string;
  title?: string;
  department?: string;
  role?: string;
  email?: string;
  phone?: string;
  is_primary?: boolean;
  notes?: string;
  individual_ids?: number[];
  organization_ids?: number[];
};

/**
 * Create a standalone contact
 */
export const createContact = async (data: ContactInput): Promise<Contact> => {
  return apiPost<Contact>("/contacts", data);
};

/**
 * Update a contact
 */
export const updateContact = async (
  id: number,
  data: ContactInput
): Promise<Contact> => {
  return apiPut<Contact>(`/contacts/${id}`, data);
};

/**
 * Delete a contact
 */
export const deleteContact = async (id: number): Promise<void> => {
  await apiDelete(`/contacts/${id}`);
};

// Contact Management for Individuals

/**
 * Get contacts for an individual
 */
export const getIndividualContacts = async (individualId: number): Promise<Contact[]> => {
  return apiGet<Contact[]>(`/individuals/${individualId}/contacts`);
};

/**
 * Create a contact for an individual
 */
export const createIndividualContact = async (
  individualId: number,
  data: Partial<Omit<Contact, "id" | "individual_id" | "created_at" | "updated_at">> & { organization_id?: number }
): Promise<Contact> => {
  return apiPost<Contact>(`/individuals/${individualId}/contacts`, data);
};

/**
 * Update a contact for an individual
 */
export const updateIndividualContact = async (
  individualId: number,
  contactId: number,
  data: Partial<Omit<Contact, "id" | "individual_id" | "created_at" | "updated_at">>
): Promise<Contact> => {
  return apiPut<Contact>(`/individuals/${individualId}/contacts/${contactId}`, data);
};

/**
 * Delete a contact for an individual
 */
export const deleteIndividualContact = async (
  individualId: number,
  contactId: number
): Promise<void> => {
  await apiDelete(`/individuals/${individualId}/contacts/${contactId}`);
};

// Contact Management for Organizations

/**
 * Get contacts for an organization
 */
export const getOrganizationContacts = async (orgId: number): Promise<Contact[]> => {
  return apiGet<Contact[]>(`/organizations/${orgId}/contacts`);
};

/**
 * Input type for creating an organization contact
 */
export type OrgContactInput = {
  individual_id?: number;
  first_name?: string;
  last_name?: string;
  title?: string;
  department?: string;
  role?: string;
  email?: string;
  phone?: string;
  is_primary?: boolean;
  notes?: string;
};

/**
 * Add a contact (existing individual) to an organization
 */
export const createOrganizationContact = async (
  orgId: number,
  data: OrgContactInput
): Promise<Contact> => {
  return apiPost<Contact>(`/organizations/${orgId}/contacts`, data);
};

/**
 * Update a contact for an organization
 */
export const updateOrganizationContact = async (
  orgId: number,
  contactId: number,
  data: Partial<Omit<Contact, "id" | "individual_id" | "organization_id" | "created_at" | "updated_at">>
): Promise<Contact> => {
  return apiPut<Contact>(`/organizations/${orgId}/contacts/${contactId}`, data);
};

/**
 * Delete a contact for an organization
 */
export const deleteOrganizationContact = async (
  orgId: number,
  contactId: number
): Promise<void> => {
  await apiDelete(`/organizations/${orgId}/contacts/${contactId}`);
};

// Note Types
export type Note = {
  id: number;
  tenant_id: number;
  entity_type: string;
  entity_id: number;
  content: string;
  created_by: number | null;
  created_at: string | null;
  updated_at: string | null;
};

/**
 * Get notes for an entity
 */
export const getNotes = async (
  entityType: string,
  entityId: number
): Promise<Note[]> => {
  return apiGet<Note[]>(`/notes?entity_type=${encodeURIComponent(entityType)}&entity_id=${entityId}`);
};

/**
 * Create a new note
 */
export const createNote = async (
  entityType: string,
  entityId: number,
  content: string
): Promise<Note> => {
  return apiPost<Note>("/notes", { entity_type: entityType, entity_id: entityId, content });
};

/**
 * Update a note
 */
export const updateNote = async (
  noteId: number,
  content: string
): Promise<Note> => {
  return apiPut<Note>(`/notes/${noteId}`, { content });
};

/**
 * Delete a note
 */
export const deleteNote = async (noteId: number): Promise<void> => {
  await apiDelete(`/notes/${noteId}`);
};

// ==============================================
// Revenue Range Types and API
// ==============================================

export type RevenueRange = {
  id: number;
  label: string;
  min_value: number | null;
  max_value: number | null;
  display_order: number;
};

/**
 * Get all revenue ranges
 */
export const getRevenueRanges = async (): Promise<RevenueRange[]> => {
  return apiGet<RevenueRange[]>("/revenue-ranges");
};

// ==============================================
// Technology Types and API
// ==============================================

export type Technology = {
  id: number;
  name: string;
  category: string;
  vendor: string | null;
};

export type OrganizationTechnology = {
  organization_id: number;
  technology_id: number;
  technology: Technology | null;
  detected_at: string | null;
  source: string | null;
  confidence: number | null;
};

/**
 * Get all technologies
 */
export const getTechnologies = async (category?: string): Promise<Technology[]> => {
  const params = category ? `?category=${encodeURIComponent(category)}` : "";
  return apiGet<Technology[]>(`/technologies${params}`);
};

/**
 * Create a new technology
 */
export const createTechnology = async (data: { name: string; category: string; vendor?: string }): Promise<Technology> => {
  return apiPost<Technology>("/technologies", data);
};

/**
 * Get technologies for an organization
 */
export const getOrganizationTechnologies = async (orgId: number): Promise<{ technologies: OrganizationTechnology[] }> => {
  return apiGet<{ technologies: OrganizationTechnology[] }>(`/organizations/${orgId}/technologies`);
};

/**
 * Add a technology to an organization
 */
export const addOrganizationTechnology = async (
  orgId: number,
  data: { technology_id: number; source?: string; confidence?: number }
): Promise<{ organization_technology: OrganizationTechnology }> => {
  return apiPost<{ organization_technology: OrganizationTechnology }>(`/organizations/${orgId}/technologies`, data);
};

/**
 * Remove a technology from an organization
 */
export const removeOrganizationTechnology = async (orgId: number, technologyId: number): Promise<void> => {
  await apiDelete(`/organizations/${orgId}/technologies/${technologyId}`);
};

// ==============================================
// Funding Round Types and API
// ==============================================

export type FundingRound = {
  id: number;
  organization_id: number;
  round_type: string;
  amount: number | null;
  announced_date: string | null;
  lead_investor: string | null;
  source_url: string | null;
  created_at: string | null;
};

/**
 * Get funding rounds for an organization
 */
export const getOrganizationFundingRounds = async (orgId: number): Promise<{ funding_rounds: FundingRound[] }> => {
  return apiGet<{ funding_rounds: FundingRound[] }>(`/organizations/${orgId}/funding-rounds`);
};

/**
 * Create a funding round for an organization
 */
export const createOrganizationFundingRound = async (
  orgId: number,
  data: {
    round_type: string;
    amount?: number;
    announced_date?: string;
    lead_investor?: string;
    source_url?: string;
  }
): Promise<{ funding_round: FundingRound }> => {
  return apiPost<{ funding_round: FundingRound }>(`/organizations/${orgId}/funding-rounds`, data);
};

/**
 * Delete a funding round
 */
export const deleteOrganizationFundingRound = async (orgId: number, roundId: number): Promise<void> => {
  await apiDelete(`/organizations/${orgId}/funding-rounds/${roundId}`);
};

// ==============================================
// Company Signal Types and API
// ==============================================

export type CompanySignal = {
  id: number;
  organization_id: number;
  signal_type: string;
  headline: string;
  description: string | null;
  signal_date: string | null;
  source: string | null;
  source_url: string | null;
  sentiment: string | null;
  relevance_score: number | null;
  created_at: string | null;
};

/**
 * Get signals for an organization
 */
export const getOrganizationSignals = async (
  orgId: number,
  options?: { signal_type?: string; limit?: number }
): Promise<{ signals: CompanySignal[] }> => {
  const params = new URLSearchParams();
  if (options?.signal_type) params.append("signal_type", options.signal_type);
  if (options?.limit) params.append("limit", options.limit.toString());
  const query = params.toString() ? `?${params}` : "";
  return apiGet<{ signals: CompanySignal[] }>(`/organizations/${orgId}/signals${query}`);
};

/**
 * Create a signal for an organization
 */
export const createOrganizationSignal = async (
  orgId: number,
  data: {
    signal_type: string;
    headline: string;
    description?: string;
    signal_date?: string;
    source?: string;
    source_url?: string;
    sentiment?: string;
    relevance_score?: number;
  }
): Promise<{ signal: CompanySignal }> => {
  return apiPost<{ signal: CompanySignal }>(`/organizations/${orgId}/signals`, data);
};

/**
 * Delete a signal
 */
export const deleteOrganizationSignal = async (orgId: number, signalId: number): Promise<void> => {
  await apiDelete(`/organizations/${orgId}/signals/${signalId}`);
};

// ==============================================
// Extended Organization Type (with research data)
// ==============================================

export type OrganizationWithResearch = Organization & {
  revenue_range_id: number | null;
  revenue_range: string | null;
  founding_year: number | null;
  headquarters_city: string | null;
  headquarters_state: string | null;
  headquarters_country: string | null;
  company_type: string | null;
  linkedin_url: string | null;
  crunchbase_url: string | null;
  technologies?: OrganizationTechnology[];
  funding_rounds?: FundingRound[];
  signals?: CompanySignal[];
};

// ==============================================
// Extended Individual Type (with research data)
// ==============================================

export type IndividualWithResearch = Individual & {
  twitter_url: string | null;
  github_url: string | null;
  bio: string | null;
  seniority_level: string | null;
  department: string | null;
  is_decision_maker: boolean | null;
  reports_to_id: number | null;
  reports_to?: { id: number; first_name: string; last_name: string };
  direct_reports?: { id: number; first_name: string; last_name: string; title: string | null }[];
};
