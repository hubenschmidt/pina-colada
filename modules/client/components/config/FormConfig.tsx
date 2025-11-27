import { CreatedJob } from "../../types/types";
import { LeadFormConfig } from "../LeadTracker/LeadFormConfig";
import { getRecentResumeDate, getIndustries, createIndustry, searchOrganizations, searchIndividuals, getRevenueRanges, type Industry, type Organization, type Individual, type RevenueRange } from "../../api";
import { useState, useEffect, useCallback } from "react";
import React from "react";

type LeadType = "job";

// Account Search/Autocomplete Component for Organizations
const AccountSelector = ({
  value,
  onChange,
  accountType = "Organization",
  readOnly = false
}: {
  value: string;
  onChange: (value: string) => void;
  accountType?: "Organization" | "Individual";
  readOnly?: boolean;
}) => {
  const [query, setQuery] = useState(value || "");
  const [results, setResults] = useState<(Organization | Individual)[]>([]);
  const [isOpen, setIsOpen] = useState(false);
  const [loading, setLoading] = useState(false);
  const dropdownRef = React.useRef<HTMLDivElement>(null);

  // Debounced search
  const searchDebounced = useCallback(
    debounce(async (searchQuery: string) => {
      if (!searchQuery || searchQuery.length < 2) {
        setResults([]);
        return;
      }
      setLoading(true);
      try {
        const data = accountType === "Organization"
          ? await searchOrganizations(searchQuery)
          : await searchIndividuals(searchQuery);
        setResults(data);
        setIsOpen(true);
      } catch (error) {
        console.error("Search failed:", error);
        setResults([]);
      } finally {
        setLoading(false);
      }
    }, 300),
    [accountType]
  );

  useEffect(() => {
    setQuery(value || "");
  }, [value]);

  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (dropdownRef.current && !dropdownRef.current.contains(event.target as Node)) {
        setIsOpen(false);
      }
    };
    document.addEventListener("mousedown", handleClickOutside);
    return () => document.removeEventListener("mousedown", handleClickOutside);
  }, []);

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const val = e.target.value;
    setQuery(val);
    onChange(val);
    searchDebounced(val);
  };

  const handleSelect = (item: Organization | Individual) => {
    const displayName = accountType === "Organization"
      ? (item as Organization).name
      : `${(item as Individual).first_name} ${(item as Individual).last_name}`;
    setQuery(displayName);
    onChange(displayName);
    setIsOpen(false);
    setResults([]);
  };

  const getDisplayName = (item: Organization | Individual) => {
    if (accountType === "Organization") {
      return (item as Organization).name;
    }
    const ind = item as Individual;
    return `${ind.first_name} ${ind.last_name}`;
  };

  const getSubtext = (item: Organization | Individual) => {
    if (accountType === "Organization") {
      const org = item as Organization;
      return org.industries?.length > 0 ? org.industries.join(", ") : null;
    }
    const ind = item as Individual;
    return ind.email || ind.title || null;
  };

  return (
    <div className="relative" ref={dropdownRef}>
      <input
        type="text"
        value={query}
        onChange={readOnly ? undefined : handleInputChange}
        onFocus={() => !readOnly && results.length > 0 && setIsOpen(true)}
        placeholder={accountType === "Organization" ? "Search or enter organization name..." : "Search or enter name..."}
        readOnly={readOnly}
        className={`w-full px-3 py-2 border border-zinc-300 dark:border-zinc-700 rounded focus:outline-none focus:ring-2 focus:ring-lime-500 bg-white dark:bg-zinc-800 text-zinc-900 dark:text-zinc-100 ${readOnly ? "bg-zinc-100 dark:bg-zinc-700 cursor-not-allowed opacity-75" : ""}`}
      />
      {loading && (
        <div className="absolute right-3 top-2.5">
          <svg className="animate-spin h-5 w-5 text-zinc-400" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
            <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
            <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
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
              className="w-full px-3 py-2 text-left hover:bg-zinc-100 dark:hover:bg-zinc-700"
            >
              <div className="text-zinc-900 dark:text-zinc-100">{getDisplayName(item)}</div>
              {getSubtext(item) && (
                <div className="text-xs text-zinc-500 dark:text-zinc-400">{getSubtext(item)}</div>
              )}
            </button>
          ))}
        </div>
      )}
      {isOpen && query.length >= 2 && results.length === 0 && !loading && (
        <div className="absolute z-50 w-full mt-1 bg-white dark:bg-zinc-800 border border-zinc-300 dark:border-zinc-700 rounded shadow-lg p-3 text-sm text-zinc-500 dark:text-zinc-400">
          No existing {accountType.toLowerCase()}s found. A new one will be created.
        </div>
      )}
    </div>
  );
};

// Simple debounce helper
function debounce<T extends (...args: any[]) => any>(func: T, wait: number): T {
  let timeout: NodeJS.Timeout;
  return ((...args: any[]) => {
    clearTimeout(timeout);
    timeout = setTimeout(() => func(...args), wait);
  }) as T;
}

// Industry Multi-Select Dropdown Component
const IndustrySelector = ({ value, onChange }: { value: string[]; onChange: (value: string[]) => void }) => {
  const [industries, setIndustries] = useState<Industry[]>([]);
  const [loading, setLoading] = useState(true);
  const [selectedIndustries, setSelectedIndustries] = useState<string[]>(value || []);
  const [isOpen, setIsOpen] = useState(false);
  const [showNewInput, setShowNewInput] = useState(false);
  const [newIndustry, setNewIndustry] = useState("");
  const dropdownRef = React.useRef<HTMLDivElement>(null);

  useEffect(() => {
    const loadIndustries = async () => {
      try {
        const data = await getIndustries();
        setIndustries(data.sort((a, b) => a.name.localeCompare(b.name)));
      } catch (error) {
        console.error("Failed to fetch industries:", error);
      } finally {
        setLoading(false);
      }
    };
    loadIndustries();
  }, []);

  useEffect(() => {
    setSelectedIndustries(Array.isArray(value) ? value : value ? [value] : []);
  }, [value]);

  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (dropdownRef.current && !dropdownRef.current.contains(event.target as Node)) {
        setIsOpen(false);
        setShowNewInput(false);
      }
    };
    document.addEventListener("mousedown", handleClickOutside);
    return () => document.removeEventListener("mousedown", handleClickOutside);
  }, []);

  const handleToggleIndustry = (industryName: string) => {
    const updated = selectedIndustries.includes(industryName)
      ? selectedIndustries.filter((i) => i !== industryName)
      : [...selectedIndustries, industryName];
    setSelectedIndustries(updated);
    onChange(updated);
  };

  const handleAddNew = async () => {
    if (newIndustry.trim()) {
      try {
        const created = await createIndustry(newIndustry.trim());
        setIndustries((prev) => [...prev, created].sort((a, b) => a.name.localeCompare(b.name)));
        const updated = [...selectedIndustries, created.name];
        setSelectedIndustries(updated);
        onChange(updated);
        setShowNewInput(false);
        setNewIndustry("");
      } catch (error) {
        console.error("Failed to create industry:", error);
      }
    }
  };

  if (loading) {
    return (
      <div className="w-full px-3 py-2 border border-zinc-300 dark:border-zinc-700 rounded bg-zinc-100 dark:bg-zinc-800 text-zinc-500 dark:text-zinc-400">
        Loading industries...
      </div>
    );
  }

  const displayText = selectedIndustries.length > 0
    ? selectedIndustries.join(", ")
    : "Select industries...";

  return (
    <div className="relative" ref={dropdownRef}>
      <button
        type="button"
        onClick={() => setIsOpen(!isOpen)}
        className="w-full px-3 py-2 border border-zinc-300 dark:border-zinc-700 rounded focus:outline-none focus:ring-2 focus:ring-lime-500 bg-white dark:bg-zinc-800 text-zinc-900 dark:text-zinc-100 text-left flex justify-between items-center"
      >
        <span className={selectedIndustries.length === 0 ? "text-zinc-500 dark:text-zinc-400" : ""}>
          {displayText}
        </span>
        <svg className={`w-4 h-4 transition-transform ${isOpen ? "rotate-180" : ""}`} fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
        </svg>
      </button>
      {isOpen && (
        <div className="absolute z-50 w-full mt-1 bg-white dark:bg-zinc-800 border border-zinc-300 dark:border-zinc-700 rounded shadow-lg max-h-60 overflow-y-auto">
          {industries.map((industry) => {
            const isSelected = selectedIndustries.includes(industry.name);
            return (
              <label
                key={industry.id}
                className="flex items-center px-3 py-2 hover:bg-zinc-100 dark:hover:bg-zinc-700 cursor-pointer"
              >
                <input
                  type="checkbox"
                  checked={isSelected}
                  onChange={() => handleToggleIndustry(industry.name)}
                  className="w-4 h-4 text-lime-500 border-zinc-300 dark:border-zinc-600 rounded focus:ring-lime-500 bg-white dark:bg-zinc-700"
                />
                <span className="ml-2 text-zinc-900 dark:text-zinc-100">{industry.name}</span>
              </label>
            );
          })}
          <div className="border-t border-zinc-200 dark:border-zinc-700">
            {!showNewInput ? (
              <button
                type="button"
                onClick={() => setShowNewInput(true)}
                className="w-full px-3 py-2 text-left text-zinc-600 dark:text-zinc-400 hover:bg-zinc-100 dark:hover:bg-zinc-700"
              >
                + Add New Industry
              </button>
            ) : (
              <div className="p-2 flex gap-2">
                <input
                  type="text"
                  value={newIndustry}
                  onChange={(e) => setNewIndustry(e.target.value)}
                  onKeyDown={(e) => {
                    if (e.key === "Enter") {
                      e.preventDefault();
                      handleAddNew();
                      return;
                    }
                    if (e.key === "Escape") {
                      setShowNewInput(false);
                      setNewIndustry("");
                    }
                  }}
                  placeholder="New industry..."
                  className="flex-1 px-2 py-1 text-sm border border-zinc-300 dark:border-zinc-600 rounded bg-white dark:bg-zinc-700 text-zinc-900 dark:text-zinc-100"
                  autoFocus
                />
                <button
                  type="button"
                  onClick={handleAddNew}
                  className="px-2 py-1 text-sm bg-lime-500 text-white rounded hover:bg-lime-600"
                >
                  Add
                </button>
              </div>
            )}
          </div>
        </div>
      )}
    </div>
  );
};

// Salary Range Dropdown Component
const SalaryRangeSelector = ({
  value,
  onChange,
}: {
  value: number | null;
  onChange: (value: number | null) => void;
}) => {
  const [ranges, setRanges] = useState<RevenueRange[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const loadRanges = async () => {
      try {
        const data = await getRevenueRanges("salary");
        setRanges(data);
      } catch (error) {
        console.error("Failed to fetch salary ranges:", error);
      } finally {
        setLoading(false);
      }
    };
    loadRanges();
  }, []);

  if (loading) {
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
      className="w-full px-3 py-2 border border-zinc-300 dark:border-zinc-700 rounded focus:outline-none focus:ring-2 focus:ring-lime-500 bg-white dark:bg-zinc-800 text-zinc-900 dark:text-zinc-100"
    >
      <option value="">Select salary range...</option>
      {ranges.map((range) => (
        <option key={range.id} value={range.id}>
          {range.label}
        </option>
      ))}
    </select>
  );
};

const JOB_STATUS_OPTIONS: CreatedJob["status"][] = [
  "Lead",
  "Applied",
  "Interviewing",
  "Rejected",
  "Offer",
  "Accepted",
  "Do Not Apply",
];

const getJobFormConfig = (): LeadFormConfig<CreatedJob> => ({
  title: "Add New",
  submitButtonText: "Add",
  sections: [
    {
      name: "Job Info",
      fieldNames: ["job_title", "date", "resume", "revenue_range_id", "job_url", "industry", "status", "notes"],
    },
    {
      name: "Account Info",
      fieldNames: ["account", "account_type"],
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
      name: "date",
      label: "Date",
      type: "date",
      required: true,
      defaultValue: new Date().toISOString().split("T")[0],
    },
    {
      name: "resume",
      label: "Resume Date",
      type: "custom",
      onInit: async () => {
        try {
          const resumeDate = await getRecentResumeDate();
          return resumeDate || "";
        } catch (error) {
          console.error("Failed to fetch recent resume date:", error);
          return "";
        }
      },
      renderCustom: ({ value, onChange }) => (
        <div>
          <input
            type="date"
            value={value ? value.split("T")[0] : ""}
            onChange={(e) => onChange(e.target.value)}
            className="w-full px-3 py-2 border border-zinc-300 dark:border-zinc-700 rounded focus:outline-none focus:ring-2 focus:ring-lime-500 bg-white dark:bg-zinc-800 text-zinc-900 dark:text-zinc-100"
          />
          <label className="flex items-center gap-2 mt-2 text-sm text-zinc-600 dark:text-zinc-400 cursor-pointer">
            <input
              type="checkbox"
              checked={!!value}
              onChange={async (e) => {
                if (!e.target.checked) {
                  onChange("");
                  return;
                }
                try {
                  const resumeDate = await getRecentResumeDate();
                  onChange(resumeDate || "");
                } catch (error) {
                  console.error("Failed to fetch recent resume date:", error);
                }
              }}
              className="w-4 h-4 border-zinc-300 dark:border-zinc-600 bg-white dark:bg-zinc-700 rounded focus:ring-zinc-500 accent-zinc-600 dark:accent-zinc-400"
            />
            <span>Use latest resume on file?</span>
          </label>
        </div>
      ),
    },
    {
      name: "revenue_range_id",
      label: "Salary Range",
      type: "custom",
      renderCustom: ({ value, onChange }) => (
        <SalaryRangeSelector value={value} onChange={onChange} />
      ),
    },
    {
      name: "job_url",
      label: "Job URL",
      type: "url",
      placeholder: "https://...",
      gridColumn: "md:col-span-2",
    },
    {
      name: "industry",
      label: "Industry",
      type: "custom",
      defaultValue: [],
      renderCustom: ({ value, onChange }) => (
        <IndustrySelector value={value || []} onChange={onChange} />
      ),
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
      name: "notes",
      label: "Notes",
      type: "textarea",
      placeholder: "Additional notes about this application...",
      gridColumn: "md:col-span-2",
      rows: 3,
    },
    // Account Info Section
    {
      name: "account",
      label: "Organization",
      type: "custom",
      required: true,
      renderCustom: ({ value, onChange, formData, isEditMode }) => (
        <AccountSelector
          value={value}
          onChange={onChange}
          accountType={formData?.account_type || "Organization"}
          readOnly={isEditMode}
        />
      ),
    },
    {
      name: "account_type",
      label: "Account Type",
      type: "select",
      required: false,
      defaultValue: "Organization",
      options: [
        { label: "Organization", value: "Organization" },
        { label: "Individual", value: "Individual" },
      ],
    },
    // Contact Section fields removed - contacts are now handled dynamically in LeadForm
  ],
  onBeforeSubmit: (formData) => {
    // Clean up and normalize data before submission
    const cleaned = { ...formData };

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

    return cleaned;
  },
  onValidate: (formData) => {
    const errors: { [key: string]: string } = {};
    const accountType = formData.account_type || "Organization";

    // Account (Organization or Individual) is required
    if (accountType === "Individual") {
      // For Individual, require first and last name
      if (!formData.individual_first_name || formData.individual_first_name.trim() === "") {
        errors.individual_first_name = "First Name is required";
      }
      if (!formData.individual_last_name || formData.individual_last_name.trim() === "") {
        errors.individual_last_name = "Last Name is required";
      }
      return errors;
    }
    
    // For Organization, require account field
    if (!formData.account || formData.account.trim() === "") {
      errors.account = "Organization is required";
    }

    if (!formData.job_title) {
      errors.job_title = "Job Title is required";
    }

    return Object.keys(errors).length > 0 ? errors : null;
  },
});

export const useFormConfig = (type: LeadType): LeadFormConfig<any> => {
  if (type === "job") return getJobFormConfig();
  throw new Error(`Unknown lead type: ${type}`);
};

