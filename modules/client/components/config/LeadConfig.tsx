import { useMemo } from "react";
import { ExternalLink } from "lucide-react";
import { CreatedJob } from "../../types/types";
import { getJobs, createJob, updateJob, deleteJob } from "../../api";
import { Column } from "../DataTable";
import { LeadTrackerConfig, BaseLead, LeadFormProps, LeadEditModalProps } from "../LeadTracker/LeadTrackerConfig";
import LeadForm from "../LeadTracker/LeadForm";
import LeadEditModal from "../LeadTracker/LeadEditModal";
import { useFormConfig } from "./FormConfig";

type LeadType = "job";

// Extend CreatedJob to match BaseLead interface
type JobLead = CreatedJob & BaseLead;

const getJobLeadConfig = (): LeadTrackerConfig<JobLead, Partial<CreatedJob>, Partial<CreatedJob>> => {
  const formConfig = useFormConfig("job");

  // Wrappers to inject config into generic components
  const JobFormAdapter = (props: LeadFormProps<JobLead>) => (
    <LeadForm {...props} config={formConfig} />
  );

  const JobEditModalAdapter = (props: LeadEditModalProps<JobLead>) => (
    <LeadEditModal {...props} config={formConfig} />
  );

  const columns: Column<JobLead>[] = [
    {
        header: "Company",
        accessor: "company",
        sortable: true,
        sortKey: "company",
        width: "12%",
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
        render: (job) => {
          const statusColors = {
            Lead: "bg-blue-100 text-blue-800 border-blue-300",
            Applied: "bg-green-100 text-green-800 border-green-300",
            Interviewing: "bg-yellow-100 text-yellow-800 border-yellow-300",
            Offer: "bg-purple-100 text-purple-800 border-purple-300",
            Accepted: "bg-teal-100 text-teal-800 border-teal-300",
            Rejected: "bg-gray-100 text-gray-800 border-gray-300",
            "Do Not Apply": "bg-red-100 text-red-800 border-red-300",
          };
          const colorClass =
            statusColors[job.status as keyof typeof statusColors] ||
            "bg-gray-100 text-gray-800";
          return (
            <span
              className={`inline-block px-2 py-1 text-xs font-medium rounded border ${colorClass}`}
            >
              {job.status}
            </span>
          );
        },
      },
      {
        header: "Date",
        accessor: "date",
        sortable: true,
        sortKey: "date",
        width: "10%",
        render: (job) =>
          job.date ? (
            new Date(job.date).toLocaleDateString()
          ) : (
            <span className="text-zinc-400">—</span>
          ),
      },
      {
        header: "Notes",
        accessor: "notes",
        width: "20%",
        render: (job) => {
          if (!job.notes) return <span className="text-zinc-400">—</span>;
          const truncated =
            job.notes.length > 50
              ? `${job.notes.substring(0, 50)}...`
              : job.notes;
          return (
            <span
              title={job.notes.length > 50 ? job.notes : undefined}
              className="text-sm"
            >
              {truncated}
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
        render: (job) =>
          job.resume ? (
            new Date(job.resume).toLocaleDateString()
          ) : (
            <span className="text-zinc-400">—</span>
          ),
      },
      {
        header: "URL",
        accessor: "job_url",
        width: "8%",
        render: (job) => (
          <div className="flex items-center gap-2">
            {job.job_url && (
              <a
                href={job.job_url}
                target="_blank"
                rel="noopener noreferrer"
                className="p-1 text-blue-600 hover:bg-blue-50 rounded"
                title="View job posting"
              >
                <ExternalLink size={16} />
              </a>
            )}
            {!job.job_url && <span className="text-zinc-400">—</span>}
          </div>
        ),
      },
  ];

  return {
    name: "job",
    entityName: "Job",
    entityNamePlural: "Jobs",
    columns,
    FormComponent: JobFormAdapter,
    EditModalComponent: JobEditModalAdapter,
    api: {
      getLeads: async (page, limit, sortBy, sortDirection, search) => {
        return getJobs(page, limit, sortBy, sortDirection, search);
      },
      createLead: async (job) => {
        await createJob(job);
      },
      updateLead: async (id, updates) => {
        await updateJob(id, updates);
      },
      deleteLead: async (id) => {
        await deleteJob(id);
      },
    },
    defaultSortBy: "date",
    defaultSortDirection: "DESC",
    defaultPageSize: 25,
    searchPlaceholder: "Search by company or job title...",
    emptyMessage: "No job applications yet. Add your first one above!",
    enableSearch: true,
  };
};

export const useLeadConfig = (type: LeadType): LeadTrackerConfig<any, any, any> => {
  if (type === "job") return getJobLeadConfig();
  throw new Error(`Unknown lead type: ${type}`);
};
