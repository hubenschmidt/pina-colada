import { ExternalLink } from "lucide-react";
import { CreatedJob } from "../../../types/types";
import { getJobs, createJob, updateJob, deleteJob } from "../../../api";
import { Column } from "../../DataTable/DataTable";
import {
  LeadTrackerConfig,
  BaseLead,
  LeadFormProps,
} from "../types/LeadTrackerTypes";
import LeadForm from "../LeadForm";
import { useLeadFormConfig } from "./useLeadFormConfig";

type LeadType = "job";

const EmptyCell = () => <span className="text-zinc-400">—</span>;

const renderDate = (value: string | null | undefined) =>
  value ? new Date(value).toLocaleDateString() : <EmptyCell />;

const STATUS_COLORS: Record<string, string> = {
  Lead: "bg-blue-100 text-blue-800 border-blue-300",
  Applied: "bg-green-100 text-green-800 border-green-300",
  Interviewing: "bg-yellow-100 text-yellow-800 border-yellow-300",
  Offer: "bg-purple-100 text-purple-800 border-purple-300",
  Accepted: "bg-teal-100 text-teal-800 border-teal-300",
  Rejected: "bg-gray-100 text-gray-800 border-gray-300",
  "Do Not Apply": "bg-red-100 text-red-800 border-red-300",
};

// Extend CreatedJob to match BaseLead interface
type JobLead = CreatedJob & BaseLead;

const getJobLeadConfig = (): LeadTrackerConfig<
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
          className={`inline-block px-2 py-1 text-xs font-medium rounded border ${STATUS_COLORS[job.status] || "bg-gray-100 text-gray-800"}`}
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
      getLeads: async (page, limit, sortBy, sortDirection, search) => {
        return getJobs(page, limit, sortBy, sortDirection, search);
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

export const useLeadTrackerConfig = (
  type: LeadType
): LeadTrackerConfig<JobLead, Partial<CreatedJob>, Partial<CreatedJob>> => {
  if (type === "job") return getJobLeadConfig();
  throw new Error(`Unknown lead type: ${type}`);
};
