import { BaseLead } from "./LeadTrackerConfig";

export type FormFieldType =
  | "text"
  | "email"
  | "url"
  | "number"
  | "date"
  | "datetime"
  | "textarea"
  | "select"
  | "checkbox"
  | "custom";

export interface FormFieldOption {
  label: string;
  value: string | number;
}

export interface FormFieldConfig<T extends BaseLead = BaseLead> {
  name: keyof T | string;
  label: string;
  type: FormFieldType;
  required?: boolean;
  placeholder?: string;
  options?: FormFieldOption[];
  defaultValue?: any;
  gridColumn?: string; // e.g., "md:col-span-2"
  rows?: number; // For textarea
  min?: number;
  max?: number;
  step?: number;
  pattern?: string;
  validation?: (value: any) => string | null; // Custom validation, returns error message
  renderCustom?: (props: {
    value: any;
    onChange: (value: any) => void;
    field: FormFieldConfig<T>;
  }) => React.ReactNode;
  onInit?: () => Promise<any>; // Called when form opens
  onChange?: (value: any, formData: any) => any; // Transform value or update other fields
  disabled?: boolean;
  hidden?: boolean;
}

export interface LeadFormConfig<T extends BaseLead = BaseLead> {
  title: string;
  fields: FormFieldConfig<T>[];
  submitButtonText?: string;
  cancelButtonText?: string;
  onValidate?: (formData: any) => { [key: string]: string } | null; // Form-level validation
  onBeforeSubmit?: (formData: any) => any; // Transform data before submission
}
