import { CheckCircle, Trash2, XCircle } from "lucide-react";
import { CreatedJob } from "../../types/types";
import {
  LeadPanelConfig,
  LeadWithStatus,
} from "../LeadTracker/LeadPanelConfig";
import {
  getLeads,
  getStatuses,
  updateJob,
  markLeadAsApplied,
  markLeadAsDoNotApply,
  deleteJob,
  type JobWithLeadStatus,
} from "../../api";
import styles from "../LeadTracker/LeadPanel.module.css";

type LeadType = "job";

const getJobPanelConfig = (): LeadPanelConfig<CreatedJob> => ({
  title: "Job Leads",
  entityName: "job",
  entityNamePlural: "jobs",

  getAccountName: (job) => job.account,
  getTitle: (job) => job.job_title,
  getUrl: (job) => job.job_url,

  actions: [
    {
      label: "Applied",
      icon: CheckCircle,
      variant: "success",
      onClick: async (lead) => {
        await markLeadAsApplied(lead.id);
      },
    },
    {
      label: "",
      icon: Trash2,
      variant: "danger",
      showLabel: false,
      onClick: async (lead) => {
        await deleteJob(lead.id);
      },
    },
    {
      label: "Do Not Apply",
      icon: XCircle,
      variant: "secondary",
      onClick: async (lead) => {
        await markLeadAsDoNotApply(lead.id);
      },
    },
  ],

  api: {
    getLeads: async (statusNames) => {
      const leads = await getLeads(
        statusNames as ("Qualifying" | "Cold" | "Warm" | "Hot")[] | undefined
      );
      return leads.map((lead) => ({
        ...lead,
        lead_status: lead.lead_status || null,
      })) as LeadWithStatus<CreatedJob>[];
    },
    getStatuses: async () => {
      return getStatuses();
    },
    updateLeadStatus: async (id, statusId) => {
      await updateJob(id, { lead_status_id: statusId });
    },
    deleteLead: async (id) => {
      await deleteJob(id);
    },
  },

  getStatusBadgeClass: (statusName: string) => {
    const statusMap: Record<string, string> = {
      Lead: styles.badgeQualifying,
      Applied: styles.badgeWarm,
      Interviewing: styles.badgeHot,
      Offer: styles.badgeHot,
      Accepted: styles.badgeHot,
      Rejected: styles.badgeCold,
      "Do Not Apply": styles.badgeCold,
    };
    return statusMap[statusName] || styles.badgeQualifying;
  },

  defaultStatusFilter: ["Lead"],
  emptyMessage: "No job leads found. Try adjusting your filters.",
});

export const usePanelConfig = (type: LeadType): LeadPanelConfig<any> => {
  if (type === "job") return getJobPanelConfig();
  throw new Error(`Unknown lead type: ${type}`);
};
