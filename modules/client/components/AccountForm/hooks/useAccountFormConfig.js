import { formatPhoneNumber } from "../../../lib/phone";

const organizationFields = [
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
    name: "employee_count_range_id",
    label: "Employee Count",
    type: "custom",
    gridColumn: "md:col-span-1",
  },
  {
    name: "revenue_range_id",
    label: "Revenue",
    type: "custom",
    gridColumn: "md:col-span-1",
  },
  {
    name: "funding_stage_id",
    label: "Funding Stage",
    type: "custom",
    gridColumn: "md:col-span-1",
  },
  {
    name: "founding_year",
    label: "Founding Year",
    type: "number",
    placeholder: "2020",
    min: 1800,
    max: new Date().getFullYear(),
    gridColumn: "md:col-span-1",
  },
  {
    name: "company_type",
    label: "Company Type",
    type: "select",
    options: [
      { value: "private", label: "Private" },
      { value: "public", label: "Public" },
      { value: "nonprofit", label: "Nonprofit" },
      { value: "government", label: "Government" },
    ],
    gridColumn: "md:col-span-1",
  },
  {
    name: "headquarters_city",
    label: "Headquarters City",
    type: "text",
    placeholder: "San Francisco",
    gridColumn: "md:col-span-1",
  },
  {
    name: "headquarters_state",
    label: "State",
    type: "text",
    placeholder: "CA",
    gridColumn: "md:col-span-1",
  },
  {
    name: "linkedin_url",
    label: "LinkedIn URL",
    type: "text",
    placeholder: "linkedin.com/company/name",
    gridColumn: "md:col-span-1",
  },
  {
    name: "crunchbase_url",
    label: "Crunchbase URL",
    type: "text",
    placeholder: "crunchbase.com/organization/name",
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

const individualFields = [
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
    name: "seniority_level",
    label: "Seniority Level",
    type: "select",
    options: [
      { value: "C-Level", label: "C-Level" },
      { value: "VP", label: "VP" },
      { value: "Director", label: "Director" },
      { value: "Manager", label: "Manager" },
      { value: "IC", label: "Individual Contributor" },
    ],
    gridColumn: "md:col-span-1",
  },
  {
    name: "department",
    label: "Department",
    type: "text",
    placeholder: "Engineering",
    gridColumn: "md:col-span-1",
  },
  {
    name: "is_decision_maker",
    label: "Decision Maker",
    type: "select",
    options: [
      { value: "true", label: "Yes" },
      { value: "false", label: "No" },
    ],
    gridColumn: "md:col-span-1",
  },
  {
    name: "linkedin_url",
    label: "LinkedIn URL",
    type: "text",
    placeholder: "linkedin.com/in/username",
    gridColumn: "md:col-span-1",
  },
  {
    name: "twitter_url",
    label: "Twitter URL",
    type: "text",
    placeholder: "twitter.com/username",
    gridColumn: "md:col-span-1",
  },
  {
    name: "github_url",
    label: "GitHub URL",
    type: "text",
    placeholder: "github.com/username",
    gridColumn: "md:col-span-1",
  },
  {
    name: "bio",
    label: "Bio",
    type: "textarea",
    rows: 2,
    gridColumn: "md:col-span-2",
  },
  {
    name: "description",
    label: "Description",
    type: "textarea",
    rows: 3,
    gridColumn: "md:col-span-2",
  },
];

const getOrganizationConfig = () => ({
  title: "New Organization",
  editTitle: "Edit Organization",
  fields: organizationFields,
  sections: [
    {
      name: "Details",
      fieldNames: [
        "name",
        "website",
        "phone",
        "employee_count_range_id",
        "revenue_range_id",
        "funding_stage_id",
        "founding_year",
        "company_type",
        "headquarters_city",
        "headquarters_state",
        "linkedin_url",
        "crunchbase_url",
        "description",
      ],
    },
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
    employee_count_range_id: formData.employee_count_range_id
      ? Number(formData.employee_count_range_id)
      : null,
    revenue_range_id: formData.revenue_range_id
      ? Number(formData.revenue_range_id)
      : null,
    funding_stage_id: formData.funding_stage_id
      ? Number(formData.funding_stage_id)
      : null,
    founding_year: formData.founding_year
      ? Number(formData.founding_year)
      : null,
    company_type: formData.company_type
      ? String(formData.company_type).trim()
      : null,
    headquarters_city: formData.headquarters_city
      ? String(formData.headquarters_city).trim()
      : null,
    headquarters_state: formData.headquarters_state
      ? String(formData.headquarters_state).trim()
      : null,
    linkedin_url: formData.linkedin_url
      ? String(formData.linkedin_url).trim()
      : null,
    crunchbase_url: formData.crunchbase_url
      ? String(formData.crunchbase_url).trim()
      : null,
    description: formData.description
      ? String(formData.description).trim()
      : null,
  }),
});

const getIndividualConfig = () => ({
  title: "New Individual",
  editTitle: "Edit Individual",
  fields: individualFields,
  sections: [
    {
      name: "Details",
      fieldNames: [
        "first_name",
        "last_name",
        "email",
        "phone",
        "title",
        "seniority_level",
        "department",
        "is_decision_maker",
        "linkedin_url",
        "twitter_url",
        "github_url",
        "bio",
        "description",
      ],
    },
  ],
  onValidate: (formData) => {
    const errors = {};
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
    linkedin_url: formData.linkedin_url
      ? String(formData.linkedin_url).trim()
      : null,
    twitter_url: formData.twitter_url
      ? String(formData.twitter_url).trim()
      : null,
    github_url: formData.github_url ? String(formData.github_url).trim() : null,
    title: formData.title ? String(formData.title).trim() : null,
    seniority_level: formData.seniority_level
      ? String(formData.seniority_level).trim()
      : null,
    department: formData.department ? String(formData.department).trim() : null,
    is_decision_maker:
      formData.is_decision_maker === "true"
        ? true
        : formData.is_decision_maker === "false"
          ? false
          : null,
    bio: formData.bio ? String(formData.bio).trim() : null,
    description: formData.description
      ? String(formData.description).trim()
      : null,
  }),
});

export const useAccountFormConfig = (type) => {
  if (type === "organization") return getOrganizationConfig();
  return getIndividualConfig();
};
