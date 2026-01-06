import { searchOrganizations, searchIndividuals, getSalaryRanges } from "../../../api";
import { useState, useEffect, useCallback, useId, useContext } from "react";
import { useDebounce, DEBOUNCE_MS } from "../../../hooks/useDebounce";
import { ProjectContext } from "../../../context/projectContext";
import { useLookupsContext } from "../../../context/lookupsContext";
import { fetchOnce } from "../../../lib/lookup-cache";
import IndustrySelector from "../../AccountForm/IndustrySelector";

// Type guard for Organization vs Individual
const isOrganization = (item) => "name" in item && !("first_name" in item);

// Account Search/Autocomplete Component for Organizations and Individuals
const AccountSelector = ({ value, onChange, accountType = "Organization", readOnly = false }) => {
  const [query, setQuery] = useState(value || "");
  const [results, setResults] = useState([]);
  const [isOpen, setIsOpen] = useState(false);
  const [loading, setLoading] = useState(false);
  const [showNewIndividual, setShowNewIndividual] = useState(false);
  const [newFirstName, setNewFirstName] = useState("");
  const [newLastName, setNewLastName] = useState("");
  const dropdownId = useId();

  const searchFn = accountType === "Organization" ? searchOrganizations : searchIndividuals;

  const performSearch = useCallback(
    async (searchQuery) => {
      if (!searchQuery || searchQuery.length < 2) {
        setResults([]);
        return;
      }
      setLoading(true);
      try {
        const data = await searchFn(searchQuery);
        setResults(data);
        setIsOpen(true);
      } catch (error) {
        console.error("Search failed:", error);
        setResults([]);
      } finally {
        setLoading(false);
      }
    },
    [searchFn]
  );

  const searchDebounced = useDebounce(performSearch, DEBOUNCE_MS.SEARCH);

  useEffect(() => {
    setQuery(value || "");
  }, [value]);

  // Reset new individual form when account type changes
  useEffect(() => {
    setShowNewIndividual(false);
    setNewFirstName("");
    setNewLastName("");
  }, [accountType]);

  useEffect(() => {
    const handleClickOutside = (event) => {
      const dropdown = document.getElementById(dropdownId);
      if (dropdown && !dropdown.contains(event.target)) {
        setIsOpen(false);
      }
    };
    document.addEventListener("mousedown", handleClickOutside);
    return () => document.removeEventListener("mousedown", handleClickOutside);
  }, [dropdownId]);

  const handleInputChange = (e) => {
    const val = e.target.value;
    setQuery(val);
    onChange(val);
    searchDebounced(val);
  };

  const handleSelect = (item) => {
    const displayName = isOrganization(item) ? item.name : `${item.first_name} ${item.last_name}`;
    setQuery(displayName);
    onChange(displayName);
    setIsOpen(false);
    setResults([]);
    setShowNewIndividual(false);
  };

  const handleAddNewIndividual = () => {
    if (!newFirstName.trim() || !newLastName.trim()) return;
    const fullName = `${newFirstName.trim()} ${newLastName.trim()}`;
    setQuery(fullName);
    onChange(fullName);
    setIsOpen(false);
    setShowNewIndividual(false);
    setNewFirstName("");
    setNewLastName("");
  };

  const getDisplayName = (item) => {
    if (isOrganization(item)) return item.name;
    return `${item.first_name} ${item.last_name}`;
  };

  const getSubtext = (item) => {
    if (isOrganization(item)) {
      return item.industries?.length > 0 ? item.industries.join(", ") : null;
    }
    return item.email || item.title || null;
  };

  return (
    <div className="relative" id={dropdownId}>
      <input
        type="text"
        value={query}
        onChange={readOnly ? undefined : handleInputChange}
        onFocus={() => !readOnly && results.length > 0 && setIsOpen(true)}
        placeholder={
          accountType === "Organization"
            ? "Search or enter organization name..."
            : "Search or enter name..."
        }
        readOnly={readOnly}
        className={`w-full px-3 py-2 border border-zinc-300 dark:border-zinc-700 rounded focus:outline-none focus:ring-2 focus:ring-lime-500 bg-white dark:bg-zinc-800 text-zinc-900 dark:text-zinc-100 ${readOnly ? "bg-zinc-100 dark:bg-zinc-700 cursor-not-allowed opacity-75" : ""}`}
      />

      {loading && (
        <div className="absolute right-3 top-2.5">
          <svg
            className="animate-spin h-5 w-5 text-zinc-400"
            xmlns="http://www.w3.org/2000/svg"
            fill="none"
            viewBox="0 0 24 24">
            <circle
              className="opacity-25"
              cx="12"
              cy="12"
              r="10"
              stroke="currentColor"
              strokeWidth="4"></circle>
            <path
              className="opacity-75"
              fill="currentColor"
              d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
          </svg>
        </div>
      )}
      {isOpen && results.length > 0 && (
        <div className="absolute z-50 w-full mt-1 bg-white dark:bg-zinc-800 border border-zinc-300 dark:border-zinc-700 rounded shadow-lg max-h-60 overflow-y-auto">
          {results.map((item) => (
            <button
              key={item.id}
              type="button"
              onClick={() => handleSelect(item)}
              className="w-full px-3 py-2 text-left hover:bg-zinc-100 dark:hover:bg-zinc-700">
              <div className="text-zinc-900 dark:text-zinc-100">{getDisplayName(item)}</div>
              {getSubtext(item) && (
                <div className="text-xs text-zinc-500 dark:text-zinc-400">{getSubtext(item)}</div>
              )}
            </button>
          ))}
          {accountType === "Individual" && (
            <button
              type="button"
              onClick={() => setShowNewIndividual(true)}
              className="w-full px-3 py-2 text-left hover:bg-zinc-100 dark:hover:bg-zinc-700 border-t border-zinc-200 dark:border-zinc-600 text-lime-600 dark:text-lime-400 font-medium">
              + Add new individual...
            </button>
          )}
        </div>
      )}
      {isOpen &&
        query.length >= 2 &&
        results.length === 0 &&
        !loading &&
        accountType === "Organization" && (
          <div className="absolute z-50 w-full mt-1 bg-white dark:bg-zinc-800 border border-zinc-300 dark:border-zinc-700 rounded shadow-lg p-3 text-sm text-zinc-500 dark:text-zinc-400">
            No existing organizations found. A new one will be created.
          </div>
        )}
      {isOpen &&
        query.length >= 2 &&
        results.length === 0 &&
        !loading &&
        accountType === "Individual" && (
          <div className="absolute z-50 w-full mt-1 bg-white dark:bg-zinc-800 border border-zinc-300 dark:border-zinc-700 rounded shadow-lg p-3">
            <div className="text-sm text-zinc-500 dark:text-zinc-400 mb-2">
              No existing individuals found.
            </div>
            <button
              type="button"
              onClick={() => setShowNewIndividual(true)}
              className="text-lime-600 dark:text-lime-400 font-medium text-sm hover:underline">
              + Add new individual...
            </button>
          </div>
        )}
      {showNewIndividual && (
        <div className="absolute z-50 w-full mt-1 bg-white dark:bg-zinc-800 border border-zinc-300 dark:border-zinc-700 rounded shadow-lg p-3">
          <div className="text-sm font-medium text-zinc-700 dark:text-zinc-300 mb-2">
            Add New Individual
          </div>
          <div className="flex gap-2 mb-2">
            <input
              type="text"
              value={newFirstName}
              onChange={(e) => setNewFirstName(e.target.value)}
              placeholder="First name"
              className="flex-1 px-2 py-1 border border-zinc-300 dark:border-zinc-600 rounded text-sm bg-white dark:bg-zinc-700 text-zinc-900 dark:text-zinc-100"
            />
            <input
              type="text"
              value={newLastName}
              onChange={(e) => setNewLastName(e.target.value)}
              placeholder="Last name"
              className="flex-1 px-2 py-1 border border-zinc-300 dark:border-zinc-600 rounded text-sm bg-white dark:bg-zinc-700 text-zinc-900 dark:text-zinc-100"
            />
          </div>
          <div className="flex gap-2">
            <button
              type="button"
              onClick={handleAddNewIndividual}
              disabled={!newFirstName.trim() || !newLastName.trim()}
              className="px-3 py-1 bg-lime-500 text-white rounded text-sm hover:bg-lime-600 disabled:opacity-50 disabled:cursor-not-allowed">
              Add
            </button>
            <button
              type="button"
              onClick={() => {
                setShowNewIndividual(false);
                setNewFirstName("");
                setNewLastName("");
              }}
              className="px-3 py-1 text-zinc-600 dark:text-zinc-400 text-sm hover:underline">
              Cancel
            </button>
          </div>
        </div>
      )}
    </div>
  );
};

// Project Multi-Select Component
const ProjectSelector = ({ value, onChange, defaultProjectIds }) => {
  const { projectState } = useContext(ProjectContext);
  const { projects } = projectState;
  const [isOpen, setIsOpen] = useState(false);
  const [initialized, setInitialized] = useState(false);
  const dropdownId = useId();
  const loaded = projects.length > 0;

  useEffect(() => {
    if (!initialized && loaded && defaultProjectIds?.length && (!value || value.length === 0)) {
      onChange(defaultProjectIds);
      setInitialized(true);
    }
  }, [loaded, defaultProjectIds, value, onChange, initialized]);

  useEffect(() => {
    const handleClickOutside = (event) => {
      const dropdown = document.getElementById(dropdownId);
      if (dropdown && !dropdown.contains(event.target)) {
        setIsOpen(false);
      }
    };
    document.addEventListener("mousedown", handleClickOutside);
    return () => document.removeEventListener("mousedown", handleClickOutside);
  }, [dropdownId]);

  const handleToggleProject = (projectId) => {
    const updated = value.includes(projectId)
      ? value.filter((id) => id !== projectId)
      : [...value, projectId];
    onChange(updated);
  };

  if (!loaded) {
    return (
      <div className="w-full px-3 py-2 border border-zinc-300 dark:border-zinc-700 rounded bg-zinc-100 dark:bg-zinc-800 text-zinc-500 dark:text-zinc-400">
        Loading...
      </div>
    );
  }

  const selectedNames = projects.filter((p) => value.includes(p.id)).map((p) => p.name);
  const displayText = selectedNames.length > 0 ? selectedNames.join(", ") : "Select projects...";

  return (
    <div className="relative" id={dropdownId}>
      <button
        type="button"
        onClick={() => setIsOpen(!isOpen)}
        className="w-full px-3 py-2 border border-zinc-300 dark:border-zinc-700 rounded focus:outline-none focus:ring-2 focus:ring-lime-500 bg-white dark:bg-zinc-800 text-zinc-900 dark:text-zinc-100 text-left flex justify-between items-center">
        <span className={selectedNames.length === 0 ? "text-zinc-500 dark:text-zinc-400" : ""}>
          {displayText}
        </span>
        <svg
          className={`w-4 h-4 transition-transform ${isOpen ? "rotate-180" : ""}`}
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
        </svg>
      </button>
      {isOpen && (
        <div className="absolute z-50 w-full mt-1 bg-white dark:bg-zinc-800 border border-zinc-300 dark:border-zinc-700 rounded shadow-lg max-h-60 overflow-y-auto">
          {projects.map((project) => {
            const isSelected = value.includes(project.id);
            return (
              <label
                key={project.id}
                className="flex items-center px-3 py-2 hover:bg-zinc-100 dark:hover:bg-zinc-700 cursor-pointer">
                <input
                  type="checkbox"
                  checked={isSelected}
                  onChange={() => handleToggleProject(project.id)}
                  className="w-4 h-4 text-lime-500 border-zinc-300 dark:border-zinc-600 rounded focus:ring-lime-500 bg-white dark:bg-zinc-700"
                />

                <span className="ml-2 text-zinc-900 dark:text-zinc-100">{project.name}</span>
              </label>
            );
          })}
        </div>
      )}
    </div>
  );
};

// Salary Range Dropdown Component
const SalaryRangeSelector = ({ value, onChange }) => {
  const { lookupsState, dispatchLookups } = useLookupsContext();
  const { salaryRanges, loaded } = lookupsState;

  useEffect(() => {
    if (loaded.salaryRanges) return;
    fetchOnce("salaryRanges", getSalaryRanges).then((data) => {
      dispatchLookups({ type: "SET_SALARY_RANGES", payload: data });
    });
  }, [loaded.salaryRanges, dispatchLookups]);

  if (!loaded.salaryRanges) {
    return (
      <div className="w-full px-3 py-2 border border-zinc-300 dark:border-zinc-700 rounded bg-zinc-100 dark:bg-zinc-800 text-zinc-500 dark:text-zinc-400">
        Loading...
      </div>
    );
  }

  return (
    <select
      value={value ?? ""}
      onChange={(e) => onChange(e.target.value ? Number(e.target.value) : null)}
      className="w-full px-3 py-2 border border-zinc-300 dark:border-zinc-700 rounded focus:outline-none focus:ring-2 focus:ring-lime-500 bg-white dark:bg-zinc-800 text-zinc-900 dark:text-zinc-100">
      <option value="">Select salary range...</option>
      {salaryRanges.map((range) => (
        <option key={range.id} value={range.id}>
          {range.label}
        </option>
      ))}
    </select>
  );
};

const JOB_STATUS_OPTIONS = [
  "Lead",
  "Applied",
  "Interviewing",
  "Rejected",
  "Offer",
  "Accepted",
  "Do Not Apply",
];

const getJobFormConfig = (selectedProjectId) => ({
  title: "Add New",
  submitButtonText: "Add",
  sections: [
    {
      name: "Job Info",
      fieldNames: [
        "job_title",
        "project_ids",
        "date",
        "resume",
        "salary_range_id",
        "job_url",
        "status",
        "description",
      ],
    },
    {
      name: "Account Info",
      fieldNames: ["account", "account_type", "industry"],
    },
    {
      name: "Contact",
      fieldNames: [], // Contacts are now handled dynamically in LeadForm
    },
  ],

  fields: [
    // Job Info Section
    {
      name: "job_title",
      label: "Job Title",
      type: "text",
      required: true,
      placeholder: "e.g., Software Engineer",
    },
    {
      name: "project_ids",
      label: "Projects",
      type: "custom",
      defaultValue: selectedProjectId ? [selectedProjectId] : [],
      renderCustom: ({ value, onChange }) => (
        <ProjectSelector
          value={value || []}
          onChange={onChange}
          defaultProjectIds={selectedProjectId ? [selectedProjectId] : []}
        />
      ),
    },
    {
      name: "date",
      label: "Date",
      type: "date",
      required: true,
      defaultValue: new Date().toISOString().split("T")[0],
    },
    {
      name: "resume",
      label: "Resume Date (Legacy)",
      type: "custom",
      // Only show in edit mode when there's existing data
      showCondition: ({ isEditMode, formData }) => isEditMode && !!formData?.resume,
      renderCustom: ({ value }) => (
        <input
          type="date"
          value={value ? value.split(/[T ]/)[0] : ""}
          readOnly
          disabled
          className="w-full px-3 py-2 border border-zinc-300 dark:border-zinc-700 rounded bg-zinc-100 dark:bg-zinc-700 text-zinc-500 dark:text-zinc-400 cursor-not-allowed opacity-75"
        />
      ),
    },
    {
      name: "salary_range_id",
      label: "Salary Range",
      type: "custom",
      gridColumn: "md:col-span-1",
      renderCustom: ({ value, onChange }) => (
        <SalaryRangeSelector value={value} onChange={onChange} />
      ),
    },
    {
      name: "job_url",
      label: "Job URL",
      type: "text",
      placeholder: "linkedin.com/jobs/...",
      gridColumn: "md:col-span-1",
    },
    {
      name: "status",
      label: "Status",
      type: "select",
      defaultValue: "Applied",
      options: JOB_STATUS_OPTIONS.map((status) => ({
        label: status,
        value: status,
      })),
    },
    {
      name: "description",
      label: "Description",
      type: "textarea",
      placeholder: "Description of this job opportunity...",
      gridColumn: "md:col-span-2",
      rows: 3,
    },
    // Account Info Section
    {
      name: "account",
      label: "Organization",
      type: "custom",
      gridColumn: "md:col-span-1",
      renderCustom: ({ value, onChange, formData }) => (
        <AccountSelector
          value={value}
          onChange={onChange}
          accountType={formData?.account_type || "Organization"}
        />
      ),
    },
    {
      name: "account_type",
      label: "Account Type",
      type: "select",
      required: false,
      defaultValue: "Organization",
      gridColumn: "md:col-span-1",
      options: [
        { label: "Organization", value: "Organization" },
        { label: "Individual", value: "Individual" },
      ],
    },
    {
      name: "industry",
      label: "Industry",
      type: "custom",
      defaultValue: [],
      gridColumn: "md:col-span-1",
      renderCustom: ({ value, onChange }) => (
        <IndustrySelector value={value || []} onChange={onChange} />
      ),
    },
    // Contact Section fields removed - contacts are now handled dynamically in LeadForm
  ],
  onBeforeSubmit: (formData, isEditMode = false) => {
    // Clean up and normalize data before submission
    const cleaned = { ...formData };

    // In edit mode, only include fields allowed by JobUpdate model
    if (isEditMode) {
      const allowedFields = [
        "account",
        "contacts",
        "job_title",
        "date",
        "job_url",
        "salary_range",
        "salary_range_id",
        "description",
        "resume",
        "status",
        "source",
        "lead_status_id",
        "project_ids",
      ];

      const filtered = {};
      allowedFields.forEach((field) => {
        const key = field;
        if (cleaned[key] !== undefined) {
          filtered[key] = cleaned[key];
        }
      });
      // Ensure salary_range_id is a number or null (select elements coerce to string)
      const rawSalaryId = filtered.salary_range_id;
      filtered.salary_range_id =
        rawSalaryId === "" || rawSalaryId === undefined || rawSalaryId === null
          ? null
          : Number(rawSalaryId);
      return filtered;
    }

    // Create mode: clean up and normalize data
    // Ensure industry is an array or null
    if (!cleaned.industry || (Array.isArray(cleaned.industry) && cleaned.industry.length === 0)) {
      cleaned.industry = null;
    }

    // Ensure contacts is an array or remove if empty
    if (!cleaned.contacts || (Array.isArray(cleaned.contacts) && cleaned.contacts.length === 0)) {
      delete cleaned.contacts;
    }

    // Add default source
    cleaned.source = "manual";

    // Ensure salary_range_id is a number or null (select elements coerce to string)
    const rawSalaryId = cleaned.salary_range_id;
    cleaned.salary_range_id =
      rawSalaryId === "" || rawSalaryId === undefined || rawSalaryId === null
        ? null
        : Number(rawSalaryId);

    return cleaned;
  },
  onValidate: (formData) => {
    const errors = {};

    if (!formData.job_title) {
      errors.job_title = "Job Title is required";
    }

    return Object.keys(errors).length > 0 ? errors : null;
  },
});

const OPPORTUNITY_STATUS_OPTIONS = [
  "Qualifying",
  "Discovery",
  "Proposal",
  "Negotiation",
  "Closed Won",
  "Closed Lost",
];

const getOpportunityFormConfig = (selectedProjectId) => ({
  title: "Add New Opportunity",
  submitButtonText: "Add",
  sections: [
    {
      name: "Opportunity Info",
      fieldNames: [
        "opportunity_name",
        "project_ids",
        "estimated_value",
        "probability",
        "expected_close_date",
        "status",
        "description",
      ],
    },
    {
      name: "Account Info",
      fieldNames: ["account", "account_type", "industry"],
    },
    {
      name: "Contact",
      fieldNames: [],
    },
  ],

  fields: [
    {
      name: "opportunity_name",
      label: "Name",
      type: "text",
      required: true,
      placeholder: "e.g., Acme Corp - Website Redesign",
    },
    {
      name: "project_ids",
      label: "Projects",
      type: "custom",
      defaultValue: selectedProjectId ? [selectedProjectId] : [],
      renderCustom: ({ value, onChange }) => (
        <ProjectSelector
          value={value || []}
          onChange={onChange}
          defaultProjectIds={selectedProjectId ? [selectedProjectId] : []}
        />
      ),
    },
    {
      name: "estimated_value",
      label: "Estimated Value ($)",
      type: "number",
      placeholder: "e.g., 50000",
    },
    {
      name: "probability",
      label: "Probability (%)",
      type: "number",
      placeholder: "e.g., 75",
      min: 0,
      max: 100,
      step: 1,
    },
    {
      name: "expected_close_date",
      label: "Expected Close Date",
      type: "date",
    },
    {
      name: "status",
      label: "Status",
      type: "select",
      defaultValue: "Qualifying",
      options: OPPORTUNITY_STATUS_OPTIONS.map((status) => ({
        label: status,
        value: status,
      })),
    },
    {
      name: "description",
      label: "Description",
      type: "textarea",
      placeholder: "Description of this opportunity...",
      gridColumn: "md:col-span-2",
      rows: 3,
    },
    {
      name: "account",
      label: "Organization",
      type: "custom",
      gridColumn: "md:col-span-1",
      renderCustom: ({ value, onChange, formData }) => (
        <AccountSelector
          value={value}
          onChange={onChange}
          accountType={formData?.account_type || "Organization"}
        />
      ),
    },
    {
      name: "account_type",
      label: "Account Type",
      type: "select",
      required: false,
      defaultValue: "Organization",
      gridColumn: "md:col-span-1",
      options: [
        { label: "Organization", value: "Organization" },
        { label: "Individual", value: "Individual" },
      ],
    },
    {
      name: "industry",
      label: "Industry",
      type: "custom",
      defaultValue: [],
      gridColumn: "md:col-span-1",
      renderCustom: ({ value, onChange }) => (
        <IndustrySelector value={value || []} onChange={onChange} />
      ),
    },
  ],

  onBeforeSubmit: (formData) => {
    const cleaned = { ...formData };
    if (!cleaned.industry || (Array.isArray(cleaned.industry) && cleaned.industry.length === 0)) {
      cleaned.industry = null;
    }
    if (!cleaned.contacts || (Array.isArray(cleaned.contacts) && cleaned.contacts.length === 0)) {
      delete cleaned.contacts;
    }
    cleaned.source = "manual";
    return cleaned;
  },
  onValidate: (formData) => {
    const errors = {};

    if (!formData.opportunity_name) {
      errors.opportunity_name = "Name is required";
    }

    return Object.keys(errors).length > 0 ? errors : null;
  },
});

const PARTNERSHIP_TYPE_OPTIONS = [
  "Strategic",
  "Referral",
  "Technology",
  "Channel",
  "Joint Venture",
  "Other",
];

const PARTNERSHIP_STATUS_OPTIONS = ["Exploring", "Negotiating", "Active", "On Hold", "Ended"];

const getPartnershipFormConfig = (selectedProjectId) => ({
  title: "Add New Partnership",
  submitButtonText: "Add",
  sections: [
    {
      name: "Partnership Info",
      fieldNames: [
        "partnership_name",
        "project_ids",
        "partnership_type",
        "start_date",
        "end_date",
        "status",
        "description",
      ],
    },
    {
      name: "Account Info",
      fieldNames: ["account", "account_type", "industry"],
    },
    {
      name: "Contact",
      fieldNames: [],
    },
  ],

  fields: [
    {
      name: "partnership_name",
      label: "Name",
      type: "text",
      required: true,
      placeholder: "e.g., Acme Corp - Technology Partnership",
    },
    {
      name: "project_ids",
      label: "Projects",
      type: "custom",
      defaultValue: selectedProjectId ? [selectedProjectId] : [],
      renderCustom: ({ value, onChange }) => (
        <ProjectSelector
          value={value || []}
          onChange={onChange}
          defaultProjectIds={selectedProjectId ? [selectedProjectId] : []}
        />
      ),
    },
    {
      name: "partnership_type",
      label: "Partnership Type",
      type: "select",
      defaultValue: "",
      options: [
        { label: "Select type...", value: "" },
        ...PARTNERSHIP_TYPE_OPTIONS.map((type) => ({
          label: type,
          value: type,
        })),
      ],
    },
    {
      name: "start_date",
      label: "Start Date",
      type: "date",
    },
    {
      name: "end_date",
      label: "End Date",
      type: "date",
    },
    {
      name: "status",
      label: "Status",
      type: "select",
      defaultValue: "Exploring",
      options: PARTNERSHIP_STATUS_OPTIONS.map((status) => ({
        label: status,
        value: status,
      })),
    },
    {
      name: "description",
      label: "Description",
      type: "textarea",
      placeholder: "Description of this partnership...",
      gridColumn: "md:col-span-2",
      rows: 3,
    },
    {
      name: "account",
      label: "Organization",
      type: "custom",
      gridColumn: "md:col-span-1",
      renderCustom: ({ value, onChange, formData }) => (
        <AccountSelector
          value={value}
          onChange={onChange}
          accountType={formData?.account_type || "Organization"}
        />
      ),
    },
    {
      name: "account_type",
      label: "Account Type",
      type: "select",
      required: false,
      defaultValue: "Organization",
      gridColumn: "md:col-span-1",
      options: [
        { label: "Organization", value: "Organization" },
        { label: "Individual", value: "Individual" },
      ],
    },
    {
      name: "industry",
      label: "Industry",
      type: "custom",
      defaultValue: [],
      gridColumn: "md:col-span-1",
      renderCustom: ({ value, onChange }) => (
        <IndustrySelector value={value || []} onChange={onChange} />
      ),
    },
  ],

  onBeforeSubmit: (formData) => {
    const cleaned = { ...formData };
    if (!cleaned.industry || (Array.isArray(cleaned.industry) && cleaned.industry.length === 0)) {
      cleaned.industry = null;
    }
    if (!cleaned.contacts || (Array.isArray(cleaned.contacts) && cleaned.contacts.length === 0)) {
      delete cleaned.contacts;
    }
    cleaned.source = "manual";
    return cleaned;
  },
  onValidate: (formData) => {
    const errors = {};

    if (!formData.partnership_name) {
      errors.partnership_name = "Name is required";
    }

    return Object.keys(errors).length > 0 ? errors : null;
  },
});

export const useLeadFormConfig = (type) => {
  const { projectState } = useContext(ProjectContext);
  const selectedProjectId = projectState.selectedProject?.id || null;

  if (type === "job") return getJobFormConfig(selectedProjectId);
  if (type === "opportunity") return getOpportunityFormConfig(selectedProjectId);
  if (type === "partnership") return getPartnershipFormConfig(selectedProjectId);
  throw new Error(`Unknown lead type: ${type}`);
};
