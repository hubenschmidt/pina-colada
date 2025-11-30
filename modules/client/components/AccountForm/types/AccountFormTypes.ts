import { Contact } from "../../../api";
import { Relationship } from "../../RelationshipsSection";

export type AccountType = "organization" | "individual";

export type FormFieldType =
  | "text"
  | "email"
  | "url"
  | "number"
  | "tel"
  | "textarea"
  | "select"
  | "custom";

export interface FormFieldOption {
  label: string;
  value: string | number;
}

export interface FormFieldConfig {
  name: string;
  label: string;
  type: FormFieldType;
  required?: boolean;
  placeholder?: string;
  options?: FormFieldOption[];
  defaultValue?: string | number | null;
  gridColumn?: string;
  rows?: number;
  min?: number;
  max?: number;
  onChange?: (value: string) => string;
}

export interface FormSection {
  name: string;
  fieldNames: string[];
}

export interface AccountFormConfig {
  title: string;
  editTitle: string;
  fields: FormFieldConfig[];
  sections?: FormSection[];
  submitButtonText?: string;
  onValidate?: (formData: Record<string, unknown>) => Record<string, string> | null;
  onBeforeSubmit?: (formData: Record<string, unknown>) => Record<string, unknown>;
}

export interface OrganizationData {
  id?: number;
  name: string;
  website?: string | null;
  phone?: string | null;
  employee_count?: number | null;
  employee_count_range_id?: number | null;
  employee_count_range?: string | null;
  funding_stage_id?: number | null;
  funding_stage?: string | null;
  description?: string | null;
  contacts?: Contact[];
  relationships?: Relationship[];
  industries?: string[];
  project_ids?: number[];
  projects?: { id: number; name: string }[];
}

export interface IndividualData {
  id?: number;
  first_name: string;
  last_name: string;
  email?: string | null;
  phone?: string | null;
  linkedin_url?: string | null;
  title?: string | null;
  notes?: string | null;
  contacts?: Contact[];
  relationships?: Relationship[];
  industries?: string[];
  project_ids?: number[];
  projects?: { id: number; name: string }[];
}

export interface AccountFormProps {
  type: AccountType;
  onClose: () => void;
  onAdd?: (data: OrganizationData | IndividualData) => Promise<{ id: number }>;
  account?: OrganizationData | IndividualData | null;
  onUpdate?: (id: number, data: Partial<OrganizationData | IndividualData>) => Promise<void>;
  onDelete?: (id: number) => Promise<void>;
}

export interface PendingContact {
  individual_id?: number;
  first_name: string;
  last_name: string;
  email: string;
  phone: string;
  title?: string | null;
  notes?: string | null;
  is_primary?: boolean;
}

export const emptyPendingContact = (): PendingContact => ({
  first_name: "",
  last_name: "",
  email: "",
  phone: "",
});
