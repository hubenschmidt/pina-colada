import { Column, PageData } from "../DataTable";

export type LeadStatus = string;

export interface BaseLead {
  id: string;
  created_at: string;
  updated_at: string;
  status: LeadStatus;
}

export interface LeadFormProps<T extends BaseLead> {
  isOpen: boolean;
  onClose: () => void;
  // Add mode
  onAdd?: (lead: Omit<T, "id" | "created_at" | "updated_at">) => Promise<void>;
  // Edit mode
  lead?: T | null;
  onUpdate?: (id: string, updates: Partial<T>) => Promise<void>;
  onDelete?: (id: string) => Promise<void>;
}

export interface LeadAPI<T extends BaseLead, TInsert = Omit<T, "id" | "created_at" | "updated_at">, TUpdate = Partial<T>> {
  getLeads: (
    page: number,
    limit: number,
    sortBy: string,
    sortDirection: "ASC" | "DESC",
    search?: string
  ) => Promise<PageData<T>>;
  createLead: (lead: TInsert) => Promise<void>;
  updateLead: (id: string, updates: TUpdate) => Promise<void>;
  deleteLead: (id: string) => Promise<void>;
}

export interface LeadTrackerConfig<
  T extends BaseLead,
  TInsert = Omit<T, "id" | "created_at" | "updated_at">,
  TUpdate = Partial<T>
> {
  name: string;
  entityName: string;
  entityNamePlural: string;
  columns: Column<T>[];
  FormComponent: React.ComponentType<LeadFormProps<T>>;
  api: LeadAPI<T, TInsert, TUpdate>;
  defaultSortBy?: string;
  defaultSortDirection?: "ASC" | "DESC";
  defaultPageSize?: number;
  emptyMessage?: string;
  searchPlaceholder?: string;
  searchFields?: (keyof T)[];
  enableSearch?: boolean;
  enableExport?: boolean;
  detailPagePath?: string; // If provided, navigates to detail page instead of modal
  newPagePath?: string; // If provided, navigates to new page instead of modal
}
