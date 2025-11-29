import { formatPhoneNumber } from "../../../lib/phone";
import { AccountType, AccountFormConfig, FormFieldConfig } from "../types/AccountFormTypes";

const organizationFields: FormFieldConfig[] = [
  {
    name: "name",
    label: "Name",
    type: "text",
    required: true,
    gridColumn: "md:col-span-2",
  },
  {
    name: "website",
    label: "Website",
    type: "url",
    placeholder: "https://example.com",
    gridColumn: "md:col-span-1",
  },
  {
    name: "phone",
    label: "Phone",
    type: "tel",
    placeholder: "+1-555-123-4567",
    onChange: formatPhoneNumber,
    gridColumn: "md:col-span-1",
  },
  {
    name: "industry",
    label: "Industry",
    type: "custom",
    gridColumn: "md:col-span-1",
  },
  {
    name: "employee_count",
    label: "Employee Count",
    type: "number",
    min: 0,
    gridColumn: "md:col-span-1",
  },
  {
    name: "description",
    label: "Description",
    type: "textarea",
    rows: 3,
    gridColumn: "md:col-span-2",
  },
];

const individualFields: FormFieldConfig[] = [
  {
    name: "first_name",
    label: "First Name",
    type: "text",
    required: true,
    gridColumn: "md:col-span-1",
  },
  {
    name: "last_name",
    label: "Last Name",
    type: "text",
    required: true,
    gridColumn: "md:col-span-1",
  },
  {
    name: "email",
    label: "Email",
    type: "email",
    placeholder: "john@example.com",
    gridColumn: "md:col-span-1",
  },
  {
    name: "phone",
    label: "Phone",
    type: "tel",
    placeholder: "+1-555-123-4567",
    onChange: formatPhoneNumber,
    gridColumn: "md:col-span-1",
  },
  {
    name: "industry",
    label: "Industry",
    type: "custom",
    gridColumn: "md:col-span-1",
  },
  {
    name: "title",
    label: "Title",
    type: "text",
    placeholder: "Software Engineer",
    gridColumn: "md:col-span-1",
  },
  {
    name: "linkedin_url",
    label: "LinkedIn URL",
    type: "text",
    placeholder: "linkedin.com/in/username",
    gridColumn: "md:col-span-2",
  },
  {
    name: "notes",
    label: "Notes",
    type: "textarea",
    rows: 3,
    gridColumn: "md:col-span-2",
  },
];

const getOrganizationConfig = (): AccountFormConfig => ({
  title: "New Organization",
  editTitle: "Edit Organization",
  fields: organizationFields,
  sections: [
    { name: "Details", fieldNames: ["name", "website", "phone", "employee_count", "description"] },
  ],
  onValidate: (formData) => {
    if (!formData.name || String(formData.name).trim() === "") {
      return { name: "Name is required" };
    }
    return null;
  },
  onBeforeSubmit: (formData) => ({
    name: String(formData.name || "").trim(),
    website: formData.website ? String(formData.website).trim() : null,
    phone: formData.phone ? String(formData.phone).trim() : null,
    employee_count: formData.employee_count ? Number(formData.employee_count) : null,
    description: formData.description ? String(formData.description).trim() : null,
  }),
});

const getIndividualConfig = (): AccountFormConfig => ({
  title: "New Individual",
  editTitle: "Edit Individual",
  fields: individualFields,
  sections: [
    { name: "Details", fieldNames: ["first_name", "last_name", "email", "phone", "linkedin_url", "title", "notes"] },
  ],
  onValidate: (formData) => {
    const errors: Record<string, string> = {};
    if (!formData.first_name || String(formData.first_name).trim() === "") {
      errors.first_name = "First name is required";
    }
    if (!formData.last_name || String(formData.last_name).trim() === "") {
      errors.last_name = "Last name is required";
    }
    return Object.keys(errors).length > 0 ? errors : null;
  },
  onBeforeSubmit: (formData) => ({
    first_name: String(formData.first_name || "").trim(),
    last_name: String(formData.last_name || "").trim(),
    email: formData.email ? String(formData.email).trim() : null,
    phone: formData.phone ? String(formData.phone).trim() : null,
    linkedin_url: formData.linkedin_url ? String(formData.linkedin_url).trim() : null,
    title: formData.title ? String(formData.title).trim() : null,
    notes: formData.notes ? String(formData.notes).trim() : null,
  }),
});

export const useAccountFormConfig = (type: AccountType): AccountFormConfig => {
  if (type === "organization") return getOrganizationConfig();
  return getIndividualConfig();
};
