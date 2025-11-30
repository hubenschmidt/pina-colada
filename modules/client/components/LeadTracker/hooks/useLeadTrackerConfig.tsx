import { useContext } from "react";
import { ExternalLink } from "lucide-react";
import { CreatedJob } from "../../../types/types";
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
  CreatedOpportunity,
  CreatedPartnership,
} from "../../../api";
import { Column } from "../../DataTable/DataTable";
import {
  LeadTrackerConfig,
  BaseLead,
  LeadFormProps,
} from "../types/LeadTrackerTypes";
import LeadForm from "../LeadForm";
import { useLeadFormConfig } from "./useLeadFormConfig";
import { ProjectContext } from "../../../context/projectContext";

type LeadType = "job" | "opportunity" | "partnership";

const EmptyCell = () => <span className="text-zinc-400">—</span>;

const renderDate = (value: string | null | undefined) =>
  value ? new Date(value).toLocaleDateString() : <EmptyCell />;

const JOB_STATUS_COLORS: Record<string, string> = {
  Lead: "bg-blue-100 text-blue-800 border-blue-300",
  Applied: "bg-green-100 text-green-800 border-green-300",
  Interviewing: "bg-yellow-100 text-yellow-800 border-yellow-300",
  Offer: "bg-purple-100 text-purple-800 border-purple-300",
  Accepted: "bg-teal-100 text-teal-800 border-teal-300",
  Rejected: "bg-gray-100 text-gray-800 border-gray-300",
  "Do Not Apply": "bg-red-100 text-red-800 border-red-300",
};

const OPPORTUNITY_STATUS_COLORS: Record<string, string> = {
  Qualifying: "bg-blue-100 text-blue-800 border-blue-300",
  Discovery: "bg-cyan-100 text-cyan-800 border-cyan-300",
  Proposal: "bg-yellow-100 text-yellow-800 border-yellow-300",
  Negotiation: "bg-orange-100 text-orange-800 border-orange-300",
  "Closed Won": "bg-green-100 text-green-800 border-green-300",
  "Closed Lost": "bg-red-100 text-red-800 border-red-300",
};

const PARTNERSHIP_STATUS_COLORS: Record<string, string> = {
  Exploring: "bg-blue-100 text-blue-800 border-blue-300",
  Negotiating: "bg-yellow-100 text-yellow-800 border-yellow-300",
  Active: "bg-green-100 text-green-800 border-green-300",
  "On Hold": "bg-orange-100 text-orange-800 border-orange-300",
  Ended: "bg-gray-100 text-gray-800 border-gray-300",
};

// Extend types to match BaseLead interface
type JobLead = CreatedJob & BaseLead;
type OpportunityLead = CreatedOpportunity & BaseLead;
type PartnershipLead = CreatedPartnership & BaseLead;

const getJobLeadConfig = (selectedProjectId: number | null): LeadTrackerConfig<
  JobLead,
  Partial<CreatedJob>,
  Partial<CreatedJob>
> => {
  const formConfig = useLeadFormConfig("job");

  // Wrapper to inject config into LeadForm
  const JobFormAdapter = (props: LeadFormProps<JobLead>) => (
    <LeadForm {...props} config={formConfig} />
  );

  const columns: Column<JobLead>[] = [
    {
      header: "Account",
      accessor: "account",
      sortable: true,
      sortKey: "account",
      width: "12%",
      render: (job: any) =>
        job.account?.trim() ? <span>{job.account}</span> : <EmptyCell />,
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
      render: (job) => (
        <span
          className={`inline-block px-2 py-1 text-xs font-medium rounded border ${JOB_STATUS_COLORS[job.status] || "bg-gray-100 text-gray-800"}`}
        >
          {job.status}
        </span>
      ),
    },
    {
      header: "Notes",
      accessor: "notes",
      width: "20%",
      render: (job) => {
        if (!job.notes) return <EmptyCell />;
        const isTruncated = job.notes.length > 50;
        return (
          <span title={isTruncated ? job.notes : undefined} className="text-sm">
            {isTruncated ? `${job.notes.substring(0, 50)}...` : job.notes}
          </span>
        );
      },
    },
    {
      header: "Resume",
      accessor: "resume",
      sortable: true,
      sortKey: "resume",
      width: "10%",
      render: (job) => renderDate(job.resume),
    },
    {
      header: "URL",
      accessor: "job_url",
      width: "8%",
      render: (job) =>
        job.job_url ? (
          <a
            href={job.job_url}
            target="_blank"
            rel="noopener noreferrer"
            className="p-1 text-blue-600 hover:bg-blue-50 rounded"
            title="View job posting"
          >
            <ExternalLink size={16} />
          </a>
        ) : (
          <EmptyCell />
        ),
    },
    {
      header: "Created",
      accessor: "date",
      sortable: true,
      sortKey: "date",
      width: "8%",
      render: (job) => renderDate(job.date),
    },
    {
      header: "Updated",
      accessor: "updated_at",
      sortable: true,
      sortKey: "updated_at",
      width: "8%",
      render: (job) => renderDate(job.updated_at),
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
    detailPagePath: "/leads/jobs",
    newPagePath: "/leads/jobs/new",
  };
};

const getOpportunityLeadConfig = (selectedProjectId: number | null): LeadTrackerConfig<
  OpportunityLead,
  Partial<CreatedOpportunity>,
  Partial<CreatedOpportunity>
> => {
  const formConfig = useLeadFormConfig("opportunity");

  const OpportunityFormAdapter = (props: LeadFormProps<OpportunityLead>) => (
    <LeadForm {...props} config={formConfig} />
  );

  const columns: Column<OpportunityLead>[] = [
    {
      header: "Account",
      accessor: "account",
      sortable: true,
      sortKey: "account",
      width: "15%",
      render: (opp: OpportunityLead) =>
        opp.account?.trim() ? <span>{opp.account}</span> : <EmptyCell />,
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
      render: (opp) => (
        <span
          className={`inline-block px-2 py-1 text-xs font-medium rounded border ${OPPORTUNITY_STATUS_COLORS[opp.status] || "bg-gray-100 text-gray-800"}`}
        >
          {opp.status}
        </span>
      ),
    },
    {
      header: "Value",
      accessor: "estimated_value",
      sortable: true,
      sortKey: "estimated_value",
      width: "10%",
      render: (opp) =>
        opp.estimated_value ? (
          <span>${opp.estimated_value.toLocaleString()}</span>
        ) : (
          <EmptyCell />
        ),
    },
    {
      header: "Probability",
      accessor: "probability",
      sortable: true,
      sortKey: "probability",
      width: "10%",
      render: (opp) =>
        opp.probability !== null ? <span>{opp.probability}%</span> : <EmptyCell />,
    },
    {
      header: "Close Date",
      accessor: "expected_close_date",
      sortable: true,
      sortKey: "expected_close_date",
      width: "10%",
      render: (opp) => renderDate(opp.expected_close_date),
    },
    {
      header: "Updated",
      accessor: "updated_at",
      sortable: true,
      sortKey: "updated_at",
      width: "10%",
      render: (opp) => renderDate(opp.updated_at),
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
    detailPagePath: "/leads/opportunities",
    newPagePath: "/leads/opportunities/new",
  };
};

const getPartnershipLeadConfig = (selectedProjectId: number | null): LeadTrackerConfig<
  PartnershipLead,
  Partial<CreatedPartnership>,
  Partial<CreatedPartnership>
> => {
  const formConfig = useLeadFormConfig("partnership");

  const PartnershipFormAdapter = (props: LeadFormProps<PartnershipLead>) => (
    <LeadForm {...props} config={formConfig} />
  );

  const columns: Column<PartnershipLead>[] = [
    {
      header: "Account",
      accessor: "account",
      sortable: true,
      sortKey: "account",
      width: "15%",
      render: (partner: PartnershipLead) =>
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
      render: (partner) => (
        <span
          className={`inline-block px-2 py-1 text-xs font-medium rounded border ${PARTNERSHIP_STATUS_COLORS[partner.status] || "bg-gray-100 text-gray-800"}`}
        >
          {partner.status}
        </span>
      ),
    },
    {
      header: "Start Date",
      accessor: "start_date",
      sortable: true,
      sortKey: "start_date",
      width: "10%",
      render: (partner) => renderDate(partner.start_date),
    },
    {
      header: "End Date",
      accessor: "end_date",
      sortable: true,
      sortKey: "end_date",
      width: "10%",
      render: (partner) => renderDate(partner.end_date),
    },
    {
      header: "Updated",
      accessor: "updated_at",
      sortable: true,
      sortKey: "updated_at",
      width: "10%",
      render: (partner) => renderDate(partner.updated_at),
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
    detailPagePath: "/leads/partnerships",
    newPagePath: "/leads/partnerships/new",
  };
};

export const useLeadTrackerConfig = (
  type: LeadType
): LeadTrackerConfig<any, any, any> => {
  const { projectState } = useContext(ProjectContext);
  const selectedProjectId = projectState.selectedProject?.id ?? null;

  if (type === "job") return getJobLeadConfig(selectedProjectId);
  if (type === "opportunity") return getOpportunityLeadConfig(selectedProjectId);
  if (type === "partnership") return getPartnershipLeadConfig(selectedProjectId);
  throw new Error(`Unknown lead type: ${type}`);
};
