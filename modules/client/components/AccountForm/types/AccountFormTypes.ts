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
