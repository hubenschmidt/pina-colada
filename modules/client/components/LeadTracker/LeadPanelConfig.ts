import { BaseLead } from "./LeadTrackerConfig";

export interface LeadStatus {
  id: string;
  name: string;
  description: string | null;
  created_at: string;
}

export type LeadWithStatus<T extends BaseLead> = T & {
  lead_status: LeadStatus | null;
  lead_status_id: string | null;
};

export interface LeadAction<T extends BaseLead> {
  label: string;
  icon: React.ComponentType<{ size?: number }>;
  variant: "primary" | "secondary" | "danger" | "success";
  onClick: (lead: LeadWithStatus<T>) => Promise<void>;
  showLabel?: boolean; // Whether to show text label (default: true)
}

export interface LeadPanelAPI<T extends BaseLead> {
  getLeads: (statusNames?: string[]) => Promise<LeadWithStatus<T>[]>;
  getStatuses: () => Promise<LeadStatus[]>;
  updateLeadStatus: (id: string, statusId: string) => Promise<void>;
  deleteLead: (id: string) => Promise<void>;
}

export interface LeadPanelConfig<T extends BaseLead> {
  title: string; // e.g., "Job Leads", "Opportunity Leads"
  entityName: string; // e.g., "job", "opportunity"
  entityNamePlural: string; // e.g., "jobs", "opportunities"

  // Field accessors
  getCompanyName: (lead: LeadWithStatus<T>) => string;
  getTitle: (lead: LeadWithStatus<T>) => string;
  getUrl?: (lead: LeadWithStatus<T>) => string | null;

  // Actions available for each lead
  actions: LeadAction<T>[];

  // API functions
  api: LeadPanelAPI<T>;

  // Status badge styling
  getStatusBadgeClass?: (statusName: string) => string;

  // Default filter
  defaultStatusFilter?: string[]; // e.g., ["Qualifying"]

  // Empty state message
  emptyMessage?: string;
}
