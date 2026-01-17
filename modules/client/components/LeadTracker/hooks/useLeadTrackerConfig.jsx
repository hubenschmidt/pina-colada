import { useContext } from "react";
import { ExternalLink } from "lucide-react";
import { StatusSelect } from "../../StatusSelect/StatusSelect";

import {
  getJobs,
  createJob,
  updateJob,
  deleteJob,
  getOpportunities,
  createOpportunity,
  updateOpportunity,
  deleteOpportunity,
  getPartnerships,
  createPartnership,
  updatePartnership,
  deletePartnership,
} from "../../../api";

import LeadForm from "../LeadForm";
import { useLeadFormConfig } from "./useLeadFormConfig";
import { ProjectContext } from "../../../context/projectContext";

const EmptyCell = () => <span className="text-zinc-400">—</span>;

const JOB_STATUS_COLORS = {
  Lead: "bg-blue-100 text-blue-800 border-blue-300",
  Applied: "bg-green-100 text-green-800 border-green-300",
  Interviewing: "bg-yellow-100 text-yellow-800 border-yellow-300",
  Offer: "bg-purple-100 text-purple-800 border-purple-300",
  Accepted: "bg-teal-100 text-teal-800 border-teal-300",
  Rejected: "bg-gray-100 text-gray-800 border-gray-300",
  "Do Not Apply": "bg-red-100 text-red-800 border-red-300",
};

const JOB_STATUS_OPTIONS = ["Lead", "Applied", "Interviewing", "Rejected", "Offer", "Accepted", "Do Not Apply"];

const OPPORTUNITY_STATUS_COLORS = {
  Qualifying: "bg-blue-100 text-blue-800 border-blue-300",
  Discovery: "bg-cyan-100 text-cyan-800 border-cyan-300",
  Proposal: "bg-yellow-100 text-yellow-800 border-yellow-300",
  Negotiation: "bg-orange-100 text-orange-800 border-orange-300",
  "Closed Won": "bg-green-100 text-green-800 border-green-300",
  "Closed Lost": "bg-red-100 text-red-800 border-red-300",
};

const OPPORTUNITY_STATUS_OPTIONS = ["Qualifying", "Discovery", "Proposal", "Negotiation", "Closed Won", "Closed Lost"];

const PARTNERSHIP_STATUS_COLORS = {
  Exploring: "bg-blue-100 text-blue-800 border-blue-300",
  Negotiating: "bg-yellow-100 text-yellow-800 border-yellow-300",
  Active: "bg-green-100 text-green-800 border-green-300",
  "On Hold": "bg-orange-100 text-orange-800 border-orange-300",
  Ended: "bg-gray-100 text-gray-800 border-gray-300",
};

const PARTNERSHIP_STATUS_OPTIONS = ["Exploring", "Negotiating", "Active", "On Hold", "Ended"];

// Extend types to match BaseLead interface

const getJobLeadConfig = (selectedProjectId) => {
  const formConfig = useLeadFormConfig("job");

  // Wrapper to inject config into LeadForm
  const JobFormAdapter = (props) => <LeadForm {...props} config={formConfig} />;

  const columns = [
    {
      header: "Account",
      accessor: "account",
      sortable: true,
      sortKey: "account",
      width: "12%",
      render: (job) => (job.account?.trim() ? <span>{job.account}</span> : <EmptyCell />),
    },
    {
      header: "Job Title",
      accessor: "job_title",
      sortable: true,
      sortKey: "job_title",
      width: "15%",
    },
    {
      header: "Status",
      accessor: "status",
      sortable: true,
      sortKey: "status",
      width: "10%",
      render: (row, { onCellUpdate } = {}) => (
        <StatusSelect
          value={row.status}
          options={JOB_STATUS_OPTIONS}
          colors={JOB_STATUS_COLORS}
          onUpdate={onCellUpdate}
        />
      ),
    },
    {
      header: "Description",
      accessor: "description",
      width: "20%",
      render: (job) => {
        if (!job.description) return <EmptyCell />;
        const isTruncated = job.description.length > 50;
        return (
          <span title={isTruncated ? job.description : undefined} className="text-sm">
            {isTruncated ? `${job.description.substring(0, 50)}...` : job.description}
          </span>
        );
      },
    },
    {
      header: "URL",
      accessor: "job_url",
      width: "6%",
      render: (job) => {
        if (!job.job_url) return <EmptyCell />;
        const url = job.job_url.startsWith("http") ? job.job_url : `https://${job.job_url}`;
        return (
          <a
            href={url}
            target="_blank"
            rel="noopener noreferrer"
            className="p-1 text-blue-600 hover:bg-blue-50 rounded"
            title="View job posting">
            <ExternalLink size={16} />
          </a>
        );
      },
    },
    {
      header: "Date Posted",
      accessor: "date_posted",
      sortable: true,
      sortKey: "date_posted",
      width: "8%",
      render: (job) => job.formatted_date_posted || <EmptyCell />,
    },
    {
      header: "Confidence",
      accessor: "date_posted_confidence",
      width: "6%",
      render: (job) => {
        if (!job.date_posted_confidence) return <EmptyCell />;
        const colors = {
          high: "bg-green-100 text-green-800",
          medium: "bg-yellow-100 text-yellow-800",
          low: "bg-orange-100 text-orange-800",
          none: "bg-zinc-100 text-zinc-500",
        };
        const tooltips = {
          high: "High: Date from structured data (JSON-LD or meta tags)",
          medium: "Medium: Date from search API (absolute date like 'Jan 15, 2026')",
          low: "Low: Date from relative text ('2 days ago') or URL pattern",
          none: "None: No date source found",
        };
        return (
          <span
            className={`px-1.5 py-0.5 text-xs rounded cursor-help ${colors[job.date_posted_confidence] || colors.none}`}
            title={tooltips[job.date_posted_confidence] || tooltips.none}>
            {job.date_posted_confidence}
          </span>
        );
      },
    },
    {
      header: "Created",
      accessor: "created_at",
      sortable: true,
      sortKey: "created_at",
      width: "8%",
      render: (job) => job.formatted_created_at || <EmptyCell />,
    },
    {
      header: "Updated",
      accessor: "updated_at",
      sortable: true,
      sortKey: "updated_at",
      width: "8%",
      render: (job) => job.formatted_updated_at || <EmptyCell />,
    },
  ];

  return {
    name: "job",
    entityName: "Job",
    entityNamePlural: "Jobs",
    columns,
    FormComponent: JobFormAdapter,
    api: {
      getLeads: async (page, limit, sortBy, sortDirection, search, projectId) => {
        return getJobs(page, limit, sortBy, sortDirection, search, projectId);
      },
      createLead: async (job) => {
        return await createJob(job);
      },
      updateLead: async (id, updates) => {
        await updateJob(id, updates);
      },
      deleteLead: async (id) => {
        await deleteJob(id);
      },
    },
    defaultSortBy: "updated_at",
    defaultSortDirection: "DESC",
    defaultPageSize: 50,
    searchPlaceholder: "Search by account or job title...",
    emptyMessage: "No job applications yet. Add your first one above!",
    enableSearch: true,
    getSuggestionLabel: (job) => `${job.account || "Unknown"} - ${job.job_title || "Untitled"}`,
    getSuggestionValue: (job) => job.job_title || "",
    detailPagePath: "/leads/jobs",
    newPagePath: "/leads/jobs/new",
  };
};

const getOpportunityLeadConfig = (selectedProjectId) => {
  const formConfig = useLeadFormConfig("opportunity");

  const OpportunityFormAdapter = (props) => <LeadForm {...props} config={formConfig} />;

  const columns = [
    {
      header: "Account",
      accessor: "account",
      sortable: true,
      sortKey: "account",
      width: "15%",
      render: (opp) => (opp.account?.trim() ? <span>{opp.account}</span> : <EmptyCell />),
    },
    {
      header: "Opportunity",
      accessor: "opportunity_name",
      sortable: true,
      sortKey: "opportunity_name",
      width: "20%",
    },
    {
      header: "Status",
      accessor: "status",
      sortable: true,
      sortKey: "status",
      width: "12%",
      render: (row, { onCellUpdate } = {}) => (
        <StatusSelect
          value={row.status}
          options={OPPORTUNITY_STATUS_OPTIONS}
          colors={OPPORTUNITY_STATUS_COLORS}
          onUpdate={onCellUpdate}
        />
      ),
    },
    {
      header: "Value",
      accessor: "estimated_value",
      sortable: true,
      sortKey: "estimated_value",
      width: "10%",
      render: (opp) =>
        opp.estimated_value ? <span>${opp.estimated_value.toLocaleString()}</span> : <EmptyCell />,
    },
    {
      header: "Probability",
      accessor: "probability",
      sortable: true,
      sortKey: "probability",
      width: "10%",
      render: (opp) => (opp.probability !== null ? <span>{opp.probability}%</span> : <EmptyCell />),
    },
    {
      header: "Close Date",
      accessor: "expected_close_date",
      sortable: true,
      sortKey: "expected_close_date",
      width: "10%",
      render: (opp) => opp.formatted_expected_close_date || <EmptyCell />,
    },
    {
      header: "Updated",
      accessor: "updated_at",
      sortable: true,
      sortKey: "updated_at",
      width: "10%",
      render: (opp) => opp.formatted_updated_at || <EmptyCell />,
    },
  ];

  return {
    name: "opportunity",
    entityName: "Opportunity",
    entityNamePlural: "Opportunities",
    columns,
    FormComponent: OpportunityFormAdapter,
    api: {
      getLeads: async (page, limit, sortBy, sortDirection, search, projectId) => {
        return getOpportunities(page, limit, sortBy, sortDirection, search, projectId);
      },
      createLead: async (opp) => {
        return await createOpportunity(opp);
      },
      updateLead: async (id, updates) => {
        await updateOpportunity(id, updates);
      },
      deleteLead: async (id) => {
        await deleteOpportunity(id);
      },
    },
    defaultSortBy: "updated_at",
    defaultSortDirection: "DESC",
    defaultPageSize: 50,
    searchPlaceholder: "Search by account or opportunity name...",
    emptyMessage: "No opportunities yet. Add your first one above!",
    enableSearch: true,
    getSuggestionLabel: (opp) =>
      `${opp.account || "Unknown"} - ${opp.opportunity_name || "Untitled"}`,
    getSuggestionValue: (opp) => opp.opportunity_name || "",
    detailPagePath: "/leads/opportunities",
    newPagePath: "/leads/opportunities/new",
  };
};

const getPartnershipLeadConfig = (selectedProjectId) => {
  const formConfig = useLeadFormConfig("partnership");

  const PartnershipFormAdapter = (props) => <LeadForm {...props} config={formConfig} />;

  const columns = [
    {
      header: "Account",
      accessor: "account",
      sortable: true,
      sortKey: "account",
      width: "15%",
      render: (partner) =>
        partner.account?.trim() ? <span>{partner.account}</span> : <EmptyCell />,
    },
    {
      header: "Partnership",
      accessor: "partnership_name",
      sortable: true,
      sortKey: "partnership_name",
      width: "20%",
    },
    {
      header: "Type",
      accessor: "partnership_type",
      sortable: true,
      sortKey: "partnership_type",
      width: "12%",
      render: (partner) =>
        partner.partnership_type ? <span>{partner.partnership_type}</span> : <EmptyCell />,
    },
    {
      header: "Status",
      accessor: "status",
      sortable: true,
      sortKey: "status",
      width: "12%",
      render: (row, { onCellUpdate } = {}) => (
        <StatusSelect
          value={row.status}
          options={PARTNERSHIP_STATUS_OPTIONS}
          colors={PARTNERSHIP_STATUS_COLORS}
          onUpdate={onCellUpdate}
        />
      ),
    },
    {
      header: "Start Date",
      accessor: "start_date",
      sortable: true,
      sortKey: "start_date",
      width: "10%",
      render: (partner) => partner.formatted_start_date || <EmptyCell />,
    },
    {
      header: "End Date",
      accessor: "end_date",
      sortable: true,
      sortKey: "end_date",
      width: "10%",
      render: (partner) => partner.formatted_end_date || <EmptyCell />,
    },
    {
      header: "Updated",
      accessor: "updated_at",
      sortable: true,
      sortKey: "updated_at",
      width: "10%",
      render: (partner) => partner.formatted_updated_at || <EmptyCell />,
    },
  ];

  return {
    name: "partnership",
    entityName: "Partnership",
    entityNamePlural: "Partnerships",
    columns,
    FormComponent: PartnershipFormAdapter,
    api: {
      getLeads: async (page, limit, sortBy, sortDirection, search, projectId) => {
        return getPartnerships(page, limit, sortBy, sortDirection, search, projectId);
      },
      createLead: async (partner) => {
        return await createPartnership(partner);
      },
      updateLead: async (id, updates) => {
        await updatePartnership(id, updates);
      },
      deleteLead: async (id) => {
        await deletePartnership(id);
      },
    },
    defaultSortBy: "updated_at",
    defaultSortDirection: "DESC",
    defaultPageSize: 50,
    searchPlaceholder: "Search by account or partnership name...",
    emptyMessage: "No partnerships yet. Add your first one above!",
    enableSearch: true,
    getSuggestionLabel: (partner) =>
      `${partner.account || "Unknown"} - ${partner.partnership_name || "Untitled"}`,
    getSuggestionValue: (partner) => partner.partnership_name || "",
    detailPagePath: "/leads/partnerships",
    newPagePath: "/leads/partnerships/new",
  };
};

export const useLeadTrackerConfig = (type) => {
  const { projectState } = useContext(ProjectContext);
  const selectedProjectId = projectState.selectedProject?.id ?? null;

  if (type === "job") return getJobLeadConfig(selectedProjectId);
  if (type === "opportunity") return getOpportunityLeadConfig(selectedProjectId);
  if (type === "partnership") return getPartnershipLeadConfig(selectedProjectId);
  throw new Error(`Unknown lead type: ${type}`);
};
