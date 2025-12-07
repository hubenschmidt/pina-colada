/**
 * API client
 */

import axios from "axios";

import { env } from "next-runtime-env";
import { fetchBearerToken } from "../lib/fetch-bearer-token";

const getClient = () => axios.create();

const makeRequest = async (client, method, url, data, config) => {
  if (method === "get") return client.get(url, config);
  if (method === "delete") return client.delete(url, config);
  if (method === "post") return client.post(url, data, config);
  if (method === "patch") return client.patch(url, data, config);
  return client.put(url, data, config);
};

const apiRequest = async (method, path, data, config) => {
  const client = getClient();
  const authHeaders = await fetchBearerToken();
  const mergedConfig = { ...authHeaders, ...config };
  const apiUrl = env("NEXT_PUBLIC_API_URL");
  const url = `${apiUrl}${path}`;

  return makeRequest(client, method, url, data, mergedConfig)
    .then((response) => response.data)
    .catch((error) => {
      const errorData = error.response?.data || {};
      // Handle Pydantic validation errors (422) which return detail as array
      const message = Array.isArray(errorData.detail)
        ? errorData.detail
            .map((e) => `${e.loc?.slice(1).join(".") || "field"}: ${e.msg || "invalid"}`)
            .join("; ")
        : errorData.detail || errorData.error || error.message || "Request failed";
      console.error("API Error:", {
        status: error.response?.status,
        message,
        data: errorData,
      });
      const errorObj = new Error(message);
      errorObj.response = error.response;
      errorObj.errorData = errorData;
      throw errorObj;
    });
};

const apiGet = (path, config) => apiRequest("get", path, undefined, config);

const apiPost = (path, data, config) => apiRequest("post", path, data, config);

const apiPut = (path, data, config) => apiRequest("put", path, data, config);

const apiPatch = (path, data, config) => apiRequest("patch", path, data, config);

const apiDelete = (path, config) => apiRequest("delete", path, undefined, config);

export const getJobs = async (
  page = 1,
  limit = 50,
  orderBy = "date",
  order = "DESC",
  search,
  projectId
) => {
  const params = new URLSearchParams({
    page: page.toString(),
    limit: limit.toString(),
    orderBy,
    order,
  });
  if (search && search.trim()) {
    params.append("search", search.trim());
  }
  if (projectId) {
    params.append("projectId", projectId.toString());
  }
  return apiGet(`/jobs?${params}`);
};

const sanitize = (data) => {
  const sanitized = { ...data };
  if (sanitized.date === "") sanitized.date = null;
  if (sanitized.resume === "") sanitized.resume = null;
  return sanitized;
};

export const createJob = async (job) => {
  return apiPost("/jobs", sanitize(job));
};

export const getJob = async (id) => {
  return apiGet(`/jobs/${id}`);
};

export const updateJob = async (id, job) => {
  return apiPut(`/jobs/${id}`, sanitize(job));
};

export const deleteJob = async (id) => {
  await apiDelete(`/jobs/${id}`);
};

export const getRecentResumeDate = async () => {
  try {
    const data = await apiGet("/jobs/recent-resume-date");
    return data.resume_date || null;
  } catch {
    return null;
  }
};

// Lead Status Types

/**
 * Fetch all lead statuses from the database
 */
export const getStatuses = async () => {
  return apiGet("/leads/statuses");
};

/**
 * Fetch all job leads (jobs with non-null lead_status_id)
 * Optionally filter by lead status names
 */
export const getLeads = async (statusNames) => {
  const params = new URLSearchParams();
  if (statusNames && statusNames.length > 0) {
    params.append("statuses", statusNames.join(","));
  }
  return apiGet(`/leads?${params}`);
};

/**
 * Mark a lead as applied
 * Sets status to 'applied', clears lead_status_id, and sets date to now if not set
 */
export const markLeadAsApplied = async (jobId) => {
  return apiPost(`/leads/${jobId}/apply`, {});
};

/**
 * Mark a lead as "do not apply"
 * Sets status to 'do_not_apply' and clears lead_status_id
 */
export const markLeadAsDoNotApply = async (jobId) => {
  return apiPost(`/leads/${jobId}/do-not-apply`, {});
};

// Tenant Types

// User Types

/**
 * Get current authenticated user
 */
export const getMe = async () => {
  return apiGet("/auth/me");
};

/**
 * Check if user has an associated tenant
 * Calls backend directly with user email
 * Returns tenant object if found, throws 404 error if not
 */
export const checkUserTenant = async (user) => {
  if (!user.email) {
    throw new Error("User email is required");
  }
  return apiGet(`/users/${encodeURIComponent(user.email)}/tenant`);
};

/**
 * Create a new tenant/organization
 */
export const createTenant = async (name, plan = "free") => {
  const data = await apiPost("/auth/tenant/create", {
    name,
    plan,
  });
  return data.tenant;
};

// Preferences Types

/**
 * Get list of common timezones for dropdown
 */
export const getTimezones = async () => {
  return apiGet("/preferences/timezones");
};

/**
 * Get current user's preferences
 */
export const getUserPreferences = async () => {
  return apiGet("/preferences/user");
};

/**
 * Update current user's preferences (theme and/or timezone)
 */
export const updateUserPreferences = async (updates) => {
  return apiPatch("/preferences/user", updates);
};

/**
 * Get tenant preferences (Admin/SuperAdmin only)
 */
export const getTenantPreferences = async () => {
  return apiGet("/preferences/tenant");
};

/**
 * Update tenant theme preference (Admin/SuperAdmin only)
 */
export const updateTenantPreferences = async (theme) => {
  return apiPatch("/preferences/tenant", { theme });
};

// Contact Types

// Individual Types

/**
 * Get individuals for the current tenant with pagination
 */
export const getIndividuals = async (
  page = 1,
  limit = 50,
  orderBy = "updated_at",
  order = "DESC",
  search
) => {
  const params = new URLSearchParams({
    page: page.toString(),
    limit: limit.toString(),
    orderBy,
    order,
  });
  if (search && search.trim()) {
    params.append("search", search.trim());
  }
  return apiGet(`/individuals?${params}`);
};

/**
 * Get a single individual by ID
 */
export const getIndividual = async (id) => {
  return apiGet(`/individuals/${id}`);
};

/**
 * Create a new individual
 */
export const createIndividual = async (data) => {
  return apiPost("/individuals", data);
};

/**
 * Update an individual
 */
export const updateIndividual = async (id, data) => {
  return apiPut(`/individuals/${id}`, data);
};

/**
 * Delete an individual
 */
export const deleteIndividual = async (id) => {
  await apiDelete(`/individuals/${id}`);
};

// Organization Types

/**
 * Get all organizations for the current tenant with pagination, sorting, and search
 */
export const getOrganizations = async (
  page = 1,
  limit = 50,
  orderBy = "updated_at",
  order = "DESC",
  search
) => {
  const params = new URLSearchParams({
    page: page.toString(),
    limit: limit.toString(),
    orderBy,
    order,
  });
  if (search && search.trim()) params.append("search", search.trim());
  const query = params.toString() ? `?${params}` : "";
  return apiGet(`/organizations${query}`);
};

/**
 * Get a single organization by ID
 */
export const getOrganization = async (id) => {
  return apiGet(`/organizations/${id}`);
};

/**
 * Create a new organization
 */
export const createOrganization = async (data) => {
  return apiPost("/organizations", data);
};

/**
 * Update an organization
 */
export const updateOrganization = async (id, data) => {
  return apiPut(`/organizations/${id}`, data);
};

/**
 * Delete an organization
 */
export const deleteOrganization = async (id) => {
  await apiDelete(`/organizations/${id}`);
};

/**
 * Search organizations by name
 */
export const searchOrganizations = async (query) => {
  return apiGet(`/organizations/search?q=${encodeURIComponent(query)}`);
};

/**
 * Search individuals by name or email
 */
export const searchIndividuals = async (query) => {
  return apiGet(`/individuals/search?q=${encodeURIComponent(query)}`);
};

// Account Search Types

/**
 * Search accounts by name
 */
export const searchAccounts = async (query) => {
  return apiGet(`/accounts/search?q=${encodeURIComponent(query)}`);
};

// Contact Search Types

/**
 * Search contacts and individuals by name or email.
 * Returns results that can be linked as contacts via individual_id.
 */
export const searchContacts = async (query) => {
  return apiGet(`/contacts/search?q=${encodeURIComponent(query)}`);
};

// Industry Types

/**
 * Get all industries
 */
export const getIndustries = async () => {
  return apiGet("/industries");
};

/**
 * Create a new industry
 */
export const createIndustry = async (name) => {
  return apiPost("/industries", { name });
};

// Salary Range Types

/**
 * Get all salary ranges
 */
export const getSalaryRanges = async () => {
  return apiGet("/salary-ranges");
};

// Employee Count Range Types

/**
 * Get all employee count ranges
 */
export const getEmployeeCountRanges = async () => {
  return apiGet("/employee-count-ranges");
};

// Funding Stage Types

/**
 * Get all funding stages
 */
export const getFundingStages = async () => {
  return apiGet("/funding-stages");
};

// Standalone Contact CRUD

/**
 * Get contacts with pagination
 */
export const getContacts = async (
  page = 1,
  limit = 50,
  orderBy = "updated_at",
  order = "DESC",
  search
) => {
  const params = new URLSearchParams({
    page: page.toString(),
    limit: limit.toString(),
    orderBy,
    order,
  });
  if (search && search.trim()) {
    params.append("search", search.trim());
  }
  return apiGet(`/contacts?${params}`);
};

/**
 * Get a single contact by ID
 */
export const getContact = async (id) => {
  return apiGet(`/contacts/${id}`);
};

/**
 * Create a standalone contact
 */
export const createContact = async (data) => {
  return apiPost("/contacts", data);
};

/**
 * Update a contact
 */
export const updateContact = async (id, data) => {
  return apiPut(`/contacts/${id}`, data);
};

/**
 * Delete a contact
 */
export const deleteContact = async (id) => {
  await apiDelete(`/contacts/${id}`);
};

// Contact Management for Individuals

/**
 * Get contacts for an individual
 */
export const getIndividualContacts = async (individualId) => {
  return apiGet(`/individuals/${individualId}/contacts`);
};

/**
 * Create a contact for an individual
 */
export const createIndividualContact = async (individualId, data) => {
  return apiPost(`/individuals/${individualId}/contacts`, data);
};

/**
 * Update a contact for an individual
 */
export const updateIndividualContact = async (individualId, contactId, data) => {
  return apiPut(`/individuals/${individualId}/contacts/${contactId}`, data);
};

/**
 * Delete a contact for an individual
 */
export const deleteIndividualContact = async (individualId, contactId) => {
  await apiDelete(`/individuals/${individualId}/contacts/${contactId}`);
};

// Contact Management for Organizations

/**
 * Get contacts for an organization
 */
export const getOrganizationContacts = async (orgId) => {
  return apiGet(`/organizations/${orgId}/contacts`);
};

/**
 * Input type for creating an organization contact
 */

/**
 * Add a contact (existing individual) to an organization
 */
export const createOrganizationContact = async (orgId, data) => {
  return apiPost(`/organizations/${orgId}/contacts`, data);
};

/**
 * Update a contact for an organization
 */
export const updateOrganizationContact = async (orgId, contactId, data) => {
  return apiPut(`/organizations/${orgId}/contacts/${contactId}`, data);
};

/**
 * Delete a contact for an organization
 */
export const deleteOrganizationContact = async (orgId, contactId) => {
  await apiDelete(`/organizations/${orgId}/contacts/${contactId}`);
};

// Account Relationship Management

/**
 * Create a relationship between two accounts
 */
export const createAccountRelationship = async (accountId, data) => {
  return apiPost(`/accounts/${accountId}/relationships`, data);
};

/**
 * Delete a relationship from an account
 */
export const deleteAccountRelationship = async (accountId, relationshipId) => {
  await apiDelete(`/accounts/${accountId}/relationships/${relationshipId}`);
};

// Note Types

/**
 * Get notes for an entity
 */
export const getNotes = async (entityType, entityId) => {
  return apiGet(`/notes?entity_type=${encodeURIComponent(entityType)}&entity_id=${entityId}`);
};

/**
 * Create a new note
 */
export const createNote = async (entityType, entityId, content) => {
  return apiPost("/notes", {
    entity_type: entityType,
    entity_id: entityId,
    content,
  });
};

/**
 * Update a note
 */
export const updateNote = async (noteId, content) => {
  return apiPut(`/notes/${noteId}`, { content });
};

/**
 * Delete a note
 */
export const deleteNote = async (noteId) => {
  await apiDelete(`/notes/${noteId}`);
};

// ==============================================
// Revenue Range Types and API
// ==============================================

/**
 * Get all revenue ranges
 */
export const getRevenueRanges = async () => {
  return apiGet("/revenue-ranges");
};

// ==============================================
// Technology Types and API
// ==============================================

/**
 * Get all technologies
 */
export const getTechnologies = async (category) => {
  const params = category ? `?category=${encodeURIComponent(category)}` : "";
  return apiGet(`/technologies${params}`);
};

/**
 * Create a new technology
 */
export const createTechnology = async (data) => {
  return apiPost("/technologies", data);
};

/**
 * Get technologies for an organization
 */
export const getOrganizationTechnologies = async (orgId) => {
  return apiGet(`/organizations/${orgId}/technologies`);
};

/**
 * Add a technology to an organization
 */
export const addOrganizationTechnology = async (orgId, data) => {
  return apiPost(`/organizations/${orgId}/technologies`, data);
};

/**
 * Remove a technology from an organization
 */
export const removeOrganizationTechnology = async (orgId, technologyId) => {
  await apiDelete(`/organizations/${orgId}/technologies/${technologyId}`);
};

// ==============================================
// Funding Round Types and API
// ==============================================

/**
 * Get funding rounds for an organization
 */
export const getOrganizationFundingRounds = async (orgId) => {
  return apiGet(`/organizations/${orgId}/funding-rounds`);
};

/**
 * Create a funding round for an organization
 */
export const createOrganizationFundingRound = async (orgId, data) => {
  return apiPost(`/organizations/${orgId}/funding-rounds`, data);
};

/**
 * Delete a funding round
 */
export const deleteOrganizationFundingRound = async (orgId, roundId) => {
  await apiDelete(`/organizations/${orgId}/funding-rounds/${roundId}`);
};

// ==============================================
// Company Signal Types and API
// ==============================================

/**
 * Get signals for an organization
 */
export const getOrganizationSignals = async (orgId, options) => {
  const params = new URLSearchParams();
  if (options?.signal_type) params.append("signal_type", options.signal_type);
  if (options?.limit) params.append("limit", options.limit.toString());
  const query = params.toString() ? `?${params}` : "";
  return apiGet(`/organizations/${orgId}/signals${query}`);
};

/**
 * Create a signal for an organization
 */
export const createOrganizationSignal = async (orgId, data) => {
  return apiPost(`/organizations/${orgId}/signals`, data);
};

/**
 * Delete a signal
 */
export const deleteOrganizationSignal = async (orgId, signalId) => {
  await apiDelete(`/organizations/${orgId}/signals/${signalId}`);
};

// ==============================================
// Extended Organization Type (with research data)
// ==============================================

// ==============================================
// Extended Individual Type (with research data)
// ==============================================

// ==============================================
// Reports Types and API
// ==============================================

// Canned Reports

export const getLeadPipelineReport = async (dateFrom, dateTo, projectId) => {
  const params = new URLSearchParams();
  if (dateFrom) params.append("date_from", dateFrom);
  if (dateTo) params.append("date_to", dateTo);
  if (projectId !== undefined && projectId !== null) {
    params.append("project_id", projectId.toString());
  }
  const query = params.toString() ? `?${params}` : "";
  return apiGet(`/reports/canned/lead-pipeline${query}`);
};

export const getAccountOverviewReport = async () => {
  return apiGet("/reports/canned/account-overview");
};

export const getContactCoverageReport = async () => {
  return apiGet("/reports/canned/contact-coverage");
};

export const getNotesActivityReport = async (projectId) => {
  const params = new URLSearchParams();
  if (projectId !== undefined && projectId !== null) {
    params.append("project_id", projectId.toString());
  }
  const query = params.toString() ? `?${params}` : "";
  return apiGet(`/reports/canned/notes-activity${query}`);
};

// User Audit Report

export const getUserAuditReport = async (userId) => {
  const params = new URLSearchParams();
  if (userId !== undefined && userId !== null) {
    params.append("user_id", userId.toString());
  }
  const query = params.toString() ? `?${params}` : "";
  return apiGet(`/reports/canned/user-audit${query}`);
};

// Custom Reports - Fields

export const getEntityFields = async (entity) => {
  return apiGet(`/reports/fields/${entity}`);
};

// Custom Reports - Execution

export const previewCustomReport = async (query) => {
  return apiPost("/reports/custom/preview", query);
};

export const runCustomReport = async (query) => {
  return apiPost("/reports/custom/run", query);
};

export const exportCustomReport = async (query) => {
  const client = axios.create();
  const authHeaders = await fetchBearerToken();
  const apiUrl = env("NEXT_PUBLIC_API_URL");
  const response = await client.post(`${apiUrl}/reports/custom/export`, query, {
    ...authHeaders,
    responseType: "blob",
  });
  return response.data;
};

// Saved Reports CRUD

export const getSavedReports = async (
  projectId,
  includeGlobal = true,
  page = 1,
  limit = 50,
  sortBy = "updated_at",
  order = "DESC",
  search
) => {
  const params = new URLSearchParams();
  if (projectId !== undefined && projectId !== null) {
    params.append("project_id", projectId.toString());
  }
  params.append("include_global", includeGlobal.toString());
  params.append("page", page.toString());
  params.append("limit", limit.toString());
  params.append("sort_by", sortBy);
  params.append("order", order);
  if (search) {
    params.append("q", search);
  }
  const query = params.toString() ? `?${params}` : "";
  return apiGet(`/reports/saved${query}`);
};

export const getSavedReport = async (id) => {
  return apiGet(`/reports/saved/${id}`);
};

export const createSavedReport = async (data) => {
  return apiPost("/reports/saved", data);
};

export const updateSavedReport = async (id, data) => {
  return apiPut(`/reports/saved/${id}`, data);
};

export const deleteSavedReport = async (id) => {
  await apiDelete(`/reports/saved/${id}`);
};

// ==============================================
// Project Types and API
// ==============================================

export const getProjects = async () => {
  return apiGet("/projects");
};

export const getProject = async (id) => {
  return apiGet(`/projects/${id}`);
};

export const createProject = async (data) => {
  return apiPost("/projects", data);
};

export const updateProject = async (id, data) => {
  return apiPut(`/projects/${id}`, data);
};

export const deleteProject = async (id) => {
  await apiDelete(`/projects/${id}`);
};

export const getProjectLeads = async (projectId) => {
  return apiGet(`/projects/${projectId}/leads`);
};

export const getProjectDeals = async (projectId) => {
  return apiGet(`/projects/${projectId}/deals`);
};

export const setSelectedProject = async (projectId) => {
  return apiPut("/users/me/selected-project", { project_id: projectId });
};

// Opportunity types and API functions

export const createOpportunity = async (data) => {
  return apiPost("/opportunities", data);
};

export const getOpportunity = async (id) => {
  return apiGet(`/opportunities/${id}`);
};

export const updateOpportunity = async (id, data) => {
  return apiPut(`/opportunities/${id}`, data);
};

export const deleteOpportunity = async (id) => {
  await apiDelete(`/opportunities/${id}`);
};

export const getOpportunities = async (
  page = 1,
  limit = 50,
  orderBy = "updated_at",
  order = "DESC",
  search,
  projectId
) => {
  const params = new URLSearchParams({
    page: page.toString(),
    limit: limit.toString(),
    orderBy,
    order,
  });
  if (search && search.trim()) {
    params.append("search", search.trim());
  }
  if (projectId) {
    params.append("projectId", projectId.toString());
  }
  return apiGet(`/opportunities?${params}`);
};

// Partnership types and API functions

export const createPartnership = async (data) => {
  return apiPost("/partnerships", data);
};

export const getPartnership = async (id) => {
  return apiGet(`/partnerships/${id}`);
};

export const updatePartnership = async (id, data) => {
  return apiPut(`/partnerships/${id}`, data);
};

export const deletePartnership = async (id) => {
  await apiDelete(`/partnerships/${id}`);
};

export const getPartnerships = async (
  page = 1,
  limit = 50,
  orderBy = "updated_at",
  order = "DESC",
  search,
  projectId
) => {
  const params = new URLSearchParams({
    page: page.toString(),
    limit: limit.toString(),
    orderBy,
    order,
  });
  if (search && search.trim()) {
    params.append("search", search.trim());
  }
  if (projectId) {
    params.append("projectId", projectId.toString());
  }
  return apiGet(`/partnerships?${params}`);
};

// ==============================================
// Task Types and API
// ==============================================

export const getTasksByEntity = async (entityType, entityId) => {
  return apiGet(`/tasks/entity/${entityType}/${entityId}`);
};

export const getTasks = async (
  page = 1,
  limit = 20,
  orderBy = "created_at",
  order = "DESC",
  scope = "global",
  projectId,
  search
) => {
  const params = new URLSearchParams({
    page: page.toString(),
    limit: limit.toString(),
    orderBy,
    order,
    scope,
  });
  if (projectId) {
    params.append("projectId", projectId.toString());
  }
  if (search && search.trim()) {
    params.append("search", search.trim());
  }
  return apiGet(`/tasks?${params}`);
};

export const getTask = async (id) => {
  return apiGet(`/tasks/${id}`);
};

export const getTaskStatuses = async () => {
  return apiGet("/tasks/statuses");
};

export const getTaskPriorities = async () => {
  return apiGet("/tasks/priorities");
};

export const createTask = async (data) => {
  return apiPost("/tasks", data);
};

export const updateTask = async (id, data) => {
  return apiPut(`/tasks/${id}`, data);
};

export const deleteTask = async (id) => {
  await apiDelete(`/tasks/${id}`);
};

// ==============================================
// Comment Types and API
// ==============================================

export const getComments = async (commentableType, commentableId) => {
  return apiGet(
    `/comments?commentable_type=${encodeURIComponent(
      commentableType
    )}&commentable_id=${commentableId}`
  );
};

export const createComment = async (commentableType, commentableId, content, parentCommentId) => {
  return apiPost("/comments", {
    commentable_type: commentableType,
    commentable_id: commentableId,
    content,
    parent_comment_id: parentCommentId ?? null,
  });
};

export const updateComment = async (commentId, content) => {
  return apiPut(`/comments/${commentId}`, { content });
};

export const deleteComment = async (commentId) => {
  await apiDelete(`/comments/${commentId}`);
};

// ==============================================
// Notification Types and API
// ==============================================

export const getNotificationCount = async () => {
  return apiGet("/notifications/count");
};

export const getNotifications = async (limit = 20) => {
  return apiGet(`/notifications?limit=${limit}`);
};

export const markNotificationsRead = async (notificationIds) => {
  await apiPost("/notifications/mark-read", {
    notification_ids: notificationIds,
  });
};

export const markEntityNotificationsRead = async (entityType, entityId) => {
  await apiPost("/notifications/mark-entity-read", {
    entity_type: entityType,
    entity_id: entityId,
  });
};

// ==============================================
// Document Types and API
// ==============================================

export const getDocuments = async (
  page = 1,
  limit = 50,
  orderBy = "updated_at",
  order = "DESC",
  search,
  tags,
  entityType,
  entityId
) => {
  const params = new URLSearchParams({
    page: page.toString(),
    limit: limit.toString(),
    orderBy,
    order,
  });
  if (search && search.trim()) params.append("search", search.trim());
  if (tags && tags.length > 0) params.append("tags", tags.join(","));
  if (entityType) params.append("entity_type", entityType);
  if (entityId) params.append("entity_id", entityId.toString());
  return apiGet(`/assets/documents?${params}`);
};

export const getDocument = async (id) => {
  return apiGet(`/assets/documents/${id}`);
};

export const uploadDocument = async (file, tags, entityType, entityId, description) => {
  const formData = new FormData();
  formData.append("file", file);
  if (tags && tags.length > 0) formData.append("tags", tags.join(","));
  if (entityType) formData.append("entity_type", entityType);
  if (entityId) formData.append("entity_id", entityId.toString());
  if (description) formData.append("description", description);
  return apiPost("/assets/documents", formData);
};

export const deleteDocument = async (id) => {
  await apiDelete(`/assets/documents/${id}`);
};

export const downloadDocument = async (id, filename) => {
  const apiUrl = env("NEXT_PUBLIC_API_URL");
  const { headers } = await fetchBearerToken();

  const response = await fetch(`${apiUrl}/assets/documents/${id}/download`, {
    headers,
  });

  if (!response.ok) {
    const text = await response.text();
    console.error("Download error:", response.status, text);
    throw new Error(`Download failed: ${response.status}`);
  }

  const blob = await response.blob();
  const url = window.URL.createObjectURL(blob);
  const a = document.createElement("a");
  a.href = url;
  a.download = filename;
  document.body.appendChild(a);
  a.click();
  document.body.removeChild(a);
  window.URL.revokeObjectURL(url);
};

export const getDocumentPreviewUrl = async (id) => {
  const apiUrl = env("NEXT_PUBLIC_API_URL");
  const { headers } = await fetchBearerToken();

  const response = await fetch(`${apiUrl}/assets/documents/${id}/download`, {
    headers,
  });

  if (!response.ok) throw new Error("Failed to load preview");

  const blob = await response.blob();
  return window.URL.createObjectURL(blob);
};

export const linkDocumentToEntity = async (documentId, entityType, entityId) => {
  await apiPost(`/assets/documents/${documentId}/link`, {
    entity_type: entityType,
    entity_id: entityId,
  });
};

export const unlinkDocumentFromEntity = async (documentId, entityType, entityId) => {
  await apiDelete(
    `/assets/documents/${documentId}/link?entity_type=${entityType}&entity_id=${entityId}`
  );
};

export const getTags = async () => {
  return apiGet("/tags");
};

// ============== Document Version API ==============

export const checkDocumentFilename = async (filename, entityType, entityId) => {
  const params = new URLSearchParams({
    filename,
    entity_type: entityType,
    entity_id: entityId.toString(),
  });
  return apiGet(`/assets/documents/check-filename?${params}`);
};

export const getDocumentVersions = async (documentId) => {
  return apiGet(`/assets/documents/${documentId}/versions`);
};

export const createDocumentVersion = async (documentId, file) => {
  const formData = new FormData();
  formData.append("file", file);
  return apiPost(`/assets/documents/${documentId}/versions`, formData);
};

export const setCurrentDocumentVersion = async (documentId) => {
  return apiPatch(`/assets/documents/${documentId}/set-current`, {});
};

// ============== Conversation API ==============

export const getConversations = async (limit = 50, includeArchived = false) => {
  const params = new URLSearchParams({
    limit: limit.toString(),
    include_archived: includeArchived.toString(),
  });
  return apiGet(`/conversations?${params}`);
};

export const getAllConversations = async ({ search, limit = 100, offset = 0, includeArchived = false } = {}) => {
  const params = new URLSearchParams({
    limit: limit.toString(),
    offset: offset.toString(),
    include_archived: includeArchived.toString(),
  });
  if (search) params.set("search", search);
  return apiGet(`/conversations/all?${params}`);
};

export const getConversation = async (threadId) => {
  return apiGet(`/conversations/${threadId}`);
};

export const updateConversationTitle = async (threadId, title) => {
  return apiPatch(`/conversations/${threadId}`, { title });
};

export const archiveConversation = async (threadId) => {
  return apiDelete(`/conversations/${threadId}`);
};

export const unarchiveConversation = async (threadId) => {
  return apiPost(`/conversations/${threadId}/unarchive`);
};

export const deleteConversationPermanent = async (threadId) => {
  return apiDelete(`/conversations/${threadId}/permanent`);
};

// ============== Usage Analytics API ==============

export const getUserUsage = async (period = "monthly") => {
  return apiGet(`/usage/user?period=${period}`);
};

export const getTenantUsage = async (period = "monthly") => {
  return apiGet(`/usage/tenant?period=${period}`);
};

export const getUsageTimeseries = async (period = "monthly", scope = "user") => {
  return apiGet(`/usage/timeseries?period=${period}&scope=${scope}`);
};

export const getDeveloperAnalytics = async (period = "monthly", groupBy = "node") => {
  return apiGet(`/usage/analytics?period=${period}&group_by=${groupBy}`);
};

export const checkDeveloperAccess = async () => {
  return apiGet("/usage/developer-access");
};
