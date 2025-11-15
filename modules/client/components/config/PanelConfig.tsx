import { CheckCircle, Trash2, XCircle } from "lucide-react";
import { CreatedJob } from "../../types/types";
import {
  LeadPanelConfig,
  LeadWithStatus,
} from "../LeadTracker/LeadPanelConfig";
import {
  get_leads,
  get_statuses,
  updateJob,
  mark_lead_as_applied,
  mark_lead_as_do_not_apply,
  deleteJob,
  type JobWithLeadStatus,
} from "../../api";
import styles from "../LeadTracker/LeadPanel.module.css";

type LeadType = "job";

const getJobPanelConfig = (): LeadPanelConfig<CreatedJob> => ({
  title: "Job Leads",
  entityName: "job",
  entityNamePlural: "jobs",

  getCompanyName: (job) => job.company,
  getTitle: (job) => job.job_title,
  getUrl: (job) => job.job_url,

  actions: [
    {
      label: "Applied",
      icon: CheckCircle,
      variant: "success",
      onClick: async (lead) => {
        await mark_lead_as_applied(lead.id);
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
        await mark_lead_as_do_not_apply(lead.id);
      },
    },
  ],

  api: {
    getLeads: async (statusNames) => {
      const leads = await get_leads(
        statusNames as ("Qualifying" | "Cold" | "Warm" | "Hot")[] | undefined
      );
      return leads.map((lead) => ({
        ...lead,
        lead_status: lead.lead_status || null,
      })) as LeadWithStatus<CreatedJob>[];
    },
    getStatuses: async () => {
      return get_statuses();
    },
    updateLeadStatus: async (id, statusId) => {
      await updateJob(id, { lead_status_id: statusId });
    },
    deleteLead: async (id) => {
      await deleteJob(id);
    },
  },

  getStatusBadgeClass: (statusName: string) => {
    switch (statusName) {
      case "Qualifying":
        return styles.badgeQualifying;
      case "Cold":
        return styles.badgeCold;
      case "Warm":
        return styles.badgeWarm;
      case "Hot":
        return styles.badgeHot;
      default:
        return styles.badgeQualifying;
    }
  },

  defaultStatusFilter: ["Qualifying"],
  emptyMessage: "No job leads found. Try adjusting your filters.",
});

export const usePanelConfig = (type: LeadType): LeadPanelConfig<any> => {
  if (type === "job") return getJobPanelConfig();
  throw new Error(`Unknown lead type: ${type}`);
};
