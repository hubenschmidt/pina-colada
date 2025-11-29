"use client";

import { useState, useEffect, useId } from "react";
import { ExternalLink } from "lucide-react";
import ContactSection, { ContactFieldConfig } from "../ContactSection";
import RelationshipsSection, { Relationship } from "../RelationshipsSection";
import NotesSection from "../NotesSection";
import FormActions from "../FormActions";
import Timestamps from "../Timestamps";
import {
  Contact,
  createIndividualContact,
  deleteIndividualContact,
  updateIndividualContact,
  createOrganizationContact,
  deleteOrganizationContact,
  updateOrganizationContact,
  searchContacts,
  searchAccounts,
  createNote,
  getIndustries,
  createIndustry,
  getEmployeeCountRanges,
  getFundingStages,
  getRevenueRanges,
  Industry,
  EmployeeCountRange,
  FundingStage,
  RevenueRange,
} from "../../api";
import { useProjectContext } from "../../context/projectContext";
import { SearchResult } from "../ContactSection";
import { AccountType, FormFieldConfig } from "./types/AccountFormTypes";
import { useAccountFormConfig } from "./hooks/useAccountFormConfig";
import { usePendingChanges } from "../../hooks/usePendingChanges";

interface OrganizationData {
  id?: number;
  name: string;
  website?: string | null;
  phone?: string | null;
  employee_count?: number | null;  // Legacy field
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

interface IndividualData {
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

interface AccountFormProps {
  type: AccountType;
  onClose: () => void;
  onAdd?: (data: OrganizationData | IndividualData) => Promise<{ id: number }>;
  account?: OrganizationData | IndividualData | null;
  onUpdate?: (id: number, data: Partial<OrganizationData | IndividualData>) => Promise<void>;
  onDelete?: (id: number) => Promise<void>;
}

interface PendingContact {
  individual_id?: number;
  first_name: string;
  last_name: string;
  email: string;
  phone: string;
  title?: string | null;
  notes?: string | null;
  is_primary?: boolean;
}

const emptyPendingContact = (): PendingContact => ({
  first_name: "",
  last_name: "",
  email: "",
  phone: "",
});

const inputClasses =
  "w-full px-3 py-2 border border-zinc-300 dark:border-zinc-700 rounded focus:outline-none focus:ring-2 focus:ring-lime-500 bg-white dark:bg-zinc-800 text-zinc-900 dark:text-zinc-100";

// Industry Multi-Select Dropdown Component
const IndustrySelector = ({ value, onChange }: { value: string[]; onChange: (value: string[]) => void }) => {
  const [industries, setIndustries] = useState<Industry[]>([]);
  const [loading, setLoading] = useState(true);
  const [selectedIndustries, setSelectedIndustries] = useState<string[]>(value || []);
  const [isOpen, setIsOpen] = useState(false);
  const [showNewInput, setShowNewInput] = useState(false);
  const [newIndustry, setNewIndustry] = useState("");
  const dropdownId = useId();

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
      const dropdown = document.getElementById(dropdownId);
      if (dropdown && !dropdown.contains(event.target as Node)) {
        setIsOpen(false);
        setShowNewInput(false);
      }
    };
    document.addEventListener("mousedown", handleClickOutside);
    return () => document.removeEventListener("mousedown", handleClickOutside);
  }, [dropdownId]);

  const handleToggleIndustry = (industryName: string) => {
    const updated = selectedIndustries.includes(industryName)
      ? selectedIndustries.filter((i) => i !== industryName)
      : [...selectedIndustries, industryName];
    setSelectedIndustries(updated);
    onChange(updated);
  };

  const handleAddNew = async () => {
    if (!newIndustry.trim()) return;
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
    <div className="relative" id={dropdownId}>
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

// Employee Count Range Dropdown Component
const EmployeeCountRangeSelector = ({
  value,
  onChange,
}: {
  value: number | null;
  onChange: (value: number | null) => void;
}) => {
  const [ranges, setRanges] = useState<EmployeeCountRange[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const loadRanges = async () => {
      try {
        const data = await getEmployeeCountRanges();
        setRanges(data);
      } catch (error) {
        console.error("Failed to fetch employee count ranges:", error);
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
      className={`w-full px-3 py-2 border border-zinc-300 dark:border-zinc-700 rounded focus:outline-none focus:ring-2 focus:ring-lime-500 bg-white dark:bg-zinc-800 ${value ? "text-zinc-900 dark:text-zinc-100" : "text-zinc-500 dark:text-zinc-400"}`}
    >
      <option value="">Select employee count...</option>
      {ranges.map((range) => (
        <option key={range.id} value={range.id} className="text-zinc-900 dark:text-zinc-100">
          {range.label}
        </option>
      ))}
    </select>
  );
};

// Funding Stage Dropdown Component
const FundingStageSelector = ({
  value,
  onChange,
}: {
  value: number | null;
  onChange: (value: number | null) => void;
}) => {
  const [stages, setStages] = useState<FundingStage[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const loadStages = async () => {
      try {
        const data = await getFundingStages();
        setStages(data);
      } catch (error) {
        console.error("Failed to fetch funding stages:", error);
      } finally {
        setLoading(false);
      }
    };
    loadStages();
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
      className={`w-full px-3 py-2 border border-zinc-300 dark:border-zinc-700 rounded focus:outline-none focus:ring-2 focus:ring-lime-500 bg-white dark:bg-zinc-800 ${value ? "text-zinc-900 dark:text-zinc-100" : "text-zinc-500 dark:text-zinc-400"}`}
    >
      <option value="">Select funding stage...</option>
      {stages.map((stage) => (
        <option key={stage.id} value={stage.id} className="text-zinc-900 dark:text-zinc-100">
          {stage.label}
        </option>
      ))}
    </select>
  );
};

// Revenue Range Dropdown Component
const RevenueRangeSelector = ({
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
        const data = await getRevenueRanges();
        setRanges(data);
      } catch (error) {
        console.error("Failed to fetch revenue ranges:", error);
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
      className={`w-full px-3 py-2 border border-zinc-300 dark:border-zinc-700 rounded focus:outline-none focus:ring-2 focus:ring-lime-500 bg-white dark:bg-zinc-800 ${value ? "text-zinc-900 dark:text-zinc-100" : "text-zinc-500 dark:text-zinc-400"}`}
    >
      <option value="">Select revenue range...</option>
      {ranges.map((range) => (
        <option key={range.id} value={range.id} className="text-zinc-900 dark:text-zinc-100">
          {range.label}
        </option>
      ))}
    </select>
  );
};

// Project Multi-Select Dropdown Component
const ProjectSelector = ({
  value,
  onChange,
  projects,
}: {
  value: number[];
  onChange: (value: number[]) => void;
  projects: { id: number; name: string }[];
}) => {
  const [isOpen, setIsOpen] = useState(false);
  const dropdownId = useId();

  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      const dropdown = document.getElementById(dropdownId);
      if (dropdown && !dropdown.contains(event.target as Node)) {
        setIsOpen(false);
      }
    };
    document.addEventListener("mousedown", handleClickOutside);
    return () => document.removeEventListener("mousedown", handleClickOutside);
  }, [dropdownId]);

  const handleToggleProject = (projectId: number) => {
    const updated = value.includes(projectId)
      ? value.filter((id) => id !== projectId)
      : [...value, projectId];
    onChange(updated);
  };

  const displayText = value.length > 0
    ? projects.filter((p) => value.includes(p.id)).map((p) => p.name).join(", ")
    : "Select projects...";

  return (
    <div className="relative" id={dropdownId}>
      <button
        type="button"
        onClick={() => setIsOpen(!isOpen)}
        className="w-full px-3 py-2 border border-zinc-300 dark:border-zinc-700 rounded focus:outline-none focus:ring-2 focus:ring-lime-500 bg-white dark:bg-zinc-800 text-zinc-900 dark:text-zinc-100 text-left flex justify-between items-center"
      >
        <span className={value.length === 0 ? "text-zinc-500 dark:text-zinc-400" : ""}>
          {displayText}
        </span>
        <svg className={`w-4 h-4 transition-transform ${isOpen ? "rotate-180" : ""}`} fill="none" stroke="currentColor" viewBox="0 0 24 24">
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
                className="flex items-center px-3 py-2 hover:bg-zinc-100 dark:hover:bg-zinc-700 cursor-pointer"
              >
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
          {projects.length === 0 && (
            <div className="px-3 py-2 text-zinc-500 dark:text-zinc-400">No projects available</div>
          )}
        </div>
      )}
    </div>
  );
};

const renderField = (
  field: FormFieldConfig,
  value: string | number,
  onChange: (name: string, value: string) => void,
  customRender?: React.ReactNode
) => {
  const handleChange = (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement>) => {
    const newValue = field.onChange ? field.onChange(e.target.value) : e.target.value;
    onChange(field.name, newValue);
  };

  if (field.type === "custom" && customRender) {
    return customRender;
  }

  if (field.type === "textarea") {
    return (
      <textarea
        value={value || ""}
        onChange={handleChange}
        className={inputClasses}
        rows={field.rows || 3}
        placeholder={field.placeholder}
      />
    );
  }

  if (field.type === "select" && field.options) {
    return (
      <select value={value || ""} onChange={handleChange} className={inputClasses}>
        <option value="">Select...</option>
        {field.options.map((opt) => (
          <option key={opt.value} value={opt.value}>
            {opt.label}
          </option>
        ))}
      </select>
    );
  }

  const isUrlField = field.name.endsWith("_url") || field.name === "website";
  const hasUrl = isUrlField && value && String(value).trim();

  if (hasUrl) {
    const url = String(value).startsWith("http") ? String(value) : `https://${value}`;
    return (
      <div>
        <input
          type={field.type === "tel" ? "tel" : field.type}
          value={value || ""}
          onChange={handleChange}
          className={inputClasses}
          placeholder={field.placeholder}
          required={field.required}
          min={field.min}
          max={field.max}
        />
        <a
          href={url}
          target="_blank"
          rel="noopener noreferrer"
          className="inline-flex items-center gap-1 mt-1 text-sm text-blue-600 dark:text-blue-400 hover:underline"
        >
          {String(value)}
          <ExternalLink size={14} />
        </a>
      </div>
    );
  }

  return (
    <input
      type={field.type === "tel" ? "tel" : field.type}
      value={value || ""}
      onChange={handleChange}
      className={inputClasses}
      placeholder={field.placeholder}
      required={field.required}
      min={field.min}
      max={field.max}
    />
  );
};

const AccountForm = ({
  type,
  onClose,
  onAdd,
  account,
  onUpdate,
  onDelete,
}: AccountFormProps) => {
  const config = useAccountFormConfig(type);
  const { projectState } = useProjectContext();
  const isEditMode = !!account;
  const isOrganization = type === "organization";

  const [formData, setFormData] = useState<Record<string, unknown>>({});
  const [contacts, setContacts] = useState<Contact[]>([]);
  const [pendingContacts, setPendingContacts] = useState<PendingContact[]>([]);
  const [pendingDeletions, setPendingDeletions] = useState<Contact[]>([]);
  const [pendingNotes, setPendingNotes] = useState<string[]>([]);
  const [relationships, setRelationships] = useState<Relationship[]>([]);
  const [pendingRelationships, setPendingRelationships] = useState<Relationship[]>([]);
  const [selectedIndustries, setSelectedIndustries] = useState<string[]>([]);
  const [selectedProjectIds, setSelectedProjectIds] = useState<number[]>([]);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [isDeleting, setIsDeleting] = useState(false);
  const [errors, setErrors] = useState<Record<string, string>>({});

  useEffect(() => {
    if (isEditMode && account) {
      setFormData({ ...account });
      setContacts(account.contacts || []);
      setSelectedIndustries(account.industries || []);
      // Load project IDs from projects array or project_ids
      const projectIds = account.projects?.map((p) => p.id) || account.project_ids || [];
      setSelectedProjectIds(projectIds);
      setRelationships(account.relationships || []);
      setPendingContacts([]);
      setPendingDeletions([]);
      setPendingNotes([]);
      setPendingRelationships([]);
      setErrors({});
      setIsDeleting(false);
      return;
    }

    const initialData: Record<string, unknown> = {};
    config.fields.forEach((field) => {
      initialData[field.name] = field.defaultValue ?? "";
    });
    setFormData(initialData);
    setContacts([]);
    setRelationships([]);
    setSelectedIndustries([]);
    // Default to currently selected project for new accounts
    setSelectedProjectIds(projectState.selectedProject ? [projectState.selectedProject.id] : []);
    setPendingContacts([]);
    setPendingDeletions([]);
    setPendingNotes([]);
    setPendingRelationships([]);
    setErrors({});
    setIsDeleting(false);
  }, [isEditMode, account, config.fields, projectState.selectedProject?.id]);

  const formDataHasChanges = usePendingChanges({
    original: account as unknown as Record<string, unknown> | null,
    current: formData,
    pendingDeletions,
  });

  // Check if projects have changed
  const originalProjectIds = account?.projects?.map((p) => p.id).sort() || [];
  const currentProjectIds = [...selectedProjectIds].sort();
  const projectsChanged = JSON.stringify(originalProjectIds) !== JSON.stringify(currentProjectIds);

  const hasPendingChanges = formDataHasChanges || projectsChanged;

  const handleChange = (field: string, value: string) => {
    setFormData((prev) => ({ ...prev, [field]: value }));
    if (errors[field]) {
      setErrors((prev) => {
        const next = { ...prev };
        delete next[field];
        return next;
      });
    }
  };

  const resolveIndustryIds = async (): Promise<number[]> => {
    if (selectedIndustries.length === 0) return [];
    const allIndustries = await getIndustries();
    return selectedIndustries
      .map((name) => allIndustries.find((ind) => ind.name === name)?.id)
      .filter((id): id is number => id !== undefined);
  };

  const processPendingContactDeletions = async (accountId: number) => {
    const deleteContact = isOrganization ? deleteOrganizationContact : deleteIndividualContact;
    for (const contact of pendingDeletions) {
      try {
        await deleteContact(accountId, contact.id);
      } catch (err) {
        console.error("Failed to delete contact:", err);
      }
    }
  };

  const buildContactData = (pending: PendingContact) => ({
    individual_id: pending.individual_id || undefined,
    first_name: pending.first_name?.trim() || undefined,
    last_name: pending.last_name?.trim() || undefined,
    email: pending.email?.trim() || undefined,
    phone: pending.phone?.trim() || undefined,
    title: pending.title?.trim() || undefined,
    notes: pending.notes?.trim() || undefined,
    is_primary: false,
  });

  const createPendingContacts = async (entityId: number) => {
    const createContact = isOrganization ? createOrganizationContact : createIndividualContact;
    for (const pending of pendingContacts) {
      try {
        await createContact(entityId, buildContactData(pending));
      } catch (err) {
        console.error("Failed to create contact:", err);
      }
    }
  };

  const createRelationshipContacts = async (entityId: number) => {
    for (const rel of pendingRelationships) {
      try {
        if (isOrganization && rel.type === "individual") {
          await createOrganizationContact(entityId, {
            individual_id: rel.id,
            first_name: rel.name.split(" ")[0] || "",
            last_name: rel.name.split(" ").slice(1).join(" ") || "",
            is_primary: false,
          });
        } else if (!isOrganization && rel.type === "organization") {
          await createIndividualContact(entityId, {
            organization_id: rel.id,
            is_primary: false,
          });
        }
      } catch (err) {
        console.error("Failed to create relationship:", err);
      }
    }
  };

  const createPendingNotes = async (entityId: number) => {
    const entityType = isOrganization ? "organization" : "individual";
    for (const noteContent of pendingNotes) {
      try {
        await createNote(entityType, entityId, noteContent);
      } catch (err) {
        console.error("Failed to create note:", err);
      }
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setErrors({});

    const validationErrors = config.onValidate?.(formData);
    if (validationErrors) {
      setErrors(validationErrors);
      return;
    }

    setIsSubmitting(true);
    try {
      const submitData = config.onBeforeSubmit?.(formData) ?? formData;
      const industryIds = await resolveIndustryIds();
      const dataWithIndustriesAndProjects = {
        ...submitData,
        industry_ids: industryIds.length > 0 ? industryIds : undefined,
        project_ids: selectedProjectIds.length > 0 ? selectedProjectIds : [],
      };

      if (isEditMode && account && onUpdate) {
        await onUpdate(account.id!, dataWithIndustriesAndProjects as Partial<OrganizationData | IndividualData>);
        await processPendingContactDeletions(account.id!);
        onClose();
        return;
      }

      if (!onAdd) {
        onClose();
        return;
      }

      const created = await onAdd(dataWithIndustriesAndProjects as unknown as OrganizationData | IndividualData);
      await createPendingContacts(created.id);
      await createRelationshipContacts(created.id);
      await createPendingNotes(created.id);
      onClose();
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : `Failed to ${isEditMode ? "update" : "create"}`;
      setErrors({ _form: message });
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleDelete = async () => {
    if (!account || !onDelete) return;

    if (!isDeleting) {
      setIsDeleting(true);
      return;
    }

    try {
      await onDelete(account.id!);
      onClose();
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : "Failed to delete";
      setErrors({ _form: message });
      setIsDeleting(false);
    }
  };

  const handleAddContact = async (contact: PendingContact) => {
    if (!isEditMode) {
      setPendingContacts([...pendingContacts, contact]);
      return;
    }

    if (!account?.id) return;
    setErrors({});

    const createContact = isOrganization ? createOrganizationContact : createIndividualContact;
    const contactData = {
      individual_id: isOrganization ? (contact.individual_id || undefined) : undefined,
      first_name: contact.first_name?.trim() || undefined,
      last_name: contact.last_name?.trim() || undefined,
      email: contact.email?.trim() || undefined,
      phone: contact.phone?.trim() || undefined,
      is_primary: contacts.length === 0,
    };

    try {
      const newContactData = await createContact(account.id, contactData);
      setContacts([...contacts, newContactData]);
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : "Failed to add contact";
      setErrors({ _form: message });
    }
  };

  const handleRemoveContact = (index: number) => {
    if (!isEditMode) {
      setPendingContacts(pendingContacts.filter((_, i) => i !== index));
      return;
    }

    const contact = contacts[index];
    if (!contact) return;

    // Defer deletion until save
    setPendingDeletions([...pendingDeletions, contact]);
    setContacts(contacts.filter((_, i) => i !== index));
  };

  const handleUpdateContact = async (index: number, updatedContact: PendingContact) => {
    if (!isEditMode) {
      setPendingContacts(pendingContacts.map((c, i) => (i === index ? updatedContact : c)));
      return;
    }

    const contact = contacts[index];
    if (!account?.id || !contact) return;

    const updateContact = isOrganization ? updateOrganizationContact : updateIndividualContact;
    const updateData = {
      email: updatedContact.email?.trim() || undefined,
      phone: updatedContact.phone?.trim() || undefined,
    };

    try {
      const updated = await updateContact(account.id, contact.id, updateData);
      setContacts(contacts.map((c, i) => (i === index ? updated : c)));
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : "Failed to update contact";
      setErrors({ _form: message });
    }
  };

  const handleSetPrimaryContact = async (index: number) => {
    if (!isEditMode || !account?.id) return;

    const contact = contacts[index];
    if (!contact) return;

    try {
      // Update the selected contact to be primary
      const updateContact = isOrganization ? updateOrganizationContact : updateIndividualContact;
      await updateContact(account.id, contact.id, { is_primary: true });

      // Update local state - set selected as primary, others as not
      setContacts(contacts.map((c, i) => ({
        ...c,
        is_primary: i === index,
      })));
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : "Failed to set primary contact";
      setErrors({ _form: message });
    }
  };

  const handleSearchContacts = async (query: string): Promise<SearchResult[]> => {
    const results = await searchContacts(query);
    return results.map((result) => ({
      individual_id: result.individual_id,
      first_name: result.first_name,
      last_name: result.last_name,
      email: result.email,
      phone: result.phone,
      account_name: result.account_name,
    }));
  };

  const handleSearchRelationships = async (query: string) => {
    const results = await searchAccounts(query);
    return results
      .filter((r) => r.type !== "unknown")
      .map((r) => ({
        id: r.id,
        name: r.name,
        type: r.type as "organization" | "individual",
      }));
  };

  const handleAddRelationship = (relationship: Relationship) => {
    if (isEditMode) {
      setRelationships([...relationships, relationship]);
      return;
    }
    setPendingRelationships([...pendingRelationships, relationship]);
  };

  const handleRemoveRelationship = (index: number) => {
    if (isEditMode) {
      setRelationships(relationships.filter((_, i) => i !== index));
      return;
    }
    setPendingRelationships(pendingRelationships.filter((_, i) => i !== index));
  };

  const displayRelationships = isEditMode ? relationships : pendingRelationships;
  const relationshipSearchType = isOrganization ? "individual" : "organization";

  const contactFields: ContactFieldConfig[] = [
    { name: "first_name", label: "First Name", placeholder: "e.g., John" },
    { name: "last_name", label: "Last Name", placeholder: "e.g., Doe" },
    { name: "title", label: "Title", placeholder: "e.g., VP of Sales" },
    { name: "email", label: "Email", type: "email", placeholder: "email@example.com" },
    { name: "phone", label: "Phone", type: "tel", placeholder: "+1-555-123-4567" },
  ];

  const title = isEditMode ? config.editTitle : config.title;

  const displayContacts: PendingContact[] = isEditMode
    ? contacts.map((c) => ({
        first_name: (c as unknown as Record<string, string>).first_name || "",
        last_name: (c as unknown as Record<string, string>).last_name || "",
        email: c.email || "",
        phone: c.phone || "",
        title: c.title,
        is_primary: c.is_primary,
      }))
    : pendingContacts;

  // Check if we have paired fields (for 2-column grid layout)
  const hasPairedFields = config.fields.some((f) => f.gridColumn === "md:col-span-1");

  return (
    <div>
      <h1 className="text-2xl font-bold text-zinc-900 dark:text-zinc-100 mb-6">
        {title}
      </h1>

      <form onSubmit={handleSubmit} className="space-y-4">
        {hasPairedFields ? (
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            {config.fields.map((field) => (
              <div key={field.name} className={field.gridColumn === "md:col-span-1" ? "" : "md:col-span-2"}>
                <label className="block text-sm font-medium text-zinc-700 dark:text-zinc-300 mb-1">
                  {field.label}
                  {field.required && <span className="text-red-500"> *</span>}
                </label>
                {renderField(
                  field,
                  formData[field.name] as string | number,
                  handleChange,
                  field.name === "industry" ? (
                    <IndustrySelector value={selectedIndustries} onChange={setSelectedIndustries} />
                  ) : field.name === "employee_count_range_id" ? (
                    <EmployeeCountRangeSelector
                      value={formData.employee_count_range_id as number | null}
                      onChange={(val) => handleChange("employee_count_range_id", val?.toString() ?? "")}
                    />
                  ) : field.name === "funding_stage_id" ? (
                    <FundingStageSelector
                      value={formData.funding_stage_id as number | null}
                      onChange={(val) => handleChange("funding_stage_id", val?.toString() ?? "")}
                    />
                  ) : field.name === "revenue_range_id" ? (
                    <RevenueRangeSelector
                      value={formData.revenue_range_id as number | null}
                      onChange={(val) => handleChange("revenue_range_id", val?.toString() ?? "")}
                    />
                  ) : undefined
                )}
                {errors[field.name] && (
                  <p className="mt-1 text-sm text-red-500">{errors[field.name]}</p>
                )}
              </div>
            ))}
          </div>
        ) : (
          config.fields.map((field) => (
            <div key={field.name}>
              <label className="block text-sm font-medium text-zinc-700 dark:text-zinc-300 mb-1">
                {field.label}
                {field.required && <span className="text-red-500"> *</span>}
              </label>
              {renderField(
                field,
                formData[field.name] as string | number,
                handleChange,
                field.name === "industry" ? (
                  <IndustrySelector value={selectedIndustries} onChange={setSelectedIndustries} />
                ) : field.name === "employee_count_range_id" ? (
                  <EmployeeCountRangeSelector
                    value={formData.employee_count_range_id as number | null}
                    onChange={(val) => handleChange("employee_count_range_id", val?.toString() ?? "")}
                  />
                ) : field.name === "funding_stage_id" ? (
                  <FundingStageSelector
                    value={formData.funding_stage_id as number | null}
                    onChange={(val) => handleChange("funding_stage_id", val?.toString() ?? "")}
                  />
                ) : field.name === "revenue_range_id" ? (
                  <RevenueRangeSelector
                    value={formData.revenue_range_id as number | null}
                    onChange={(val) => handleChange("revenue_range_id", val?.toString() ?? "")}
                  />
                ) : undefined
              )}
              {errors[field.name] && (
                <p className="mt-1 text-sm text-red-500">{errors[field.name]}</p>
              )}
            </div>
          ))
        )}

        {/* Project Assignment */}
        <div className="mt-4">
          <label className="block text-sm font-medium text-zinc-700 dark:text-zinc-300 mb-1">
            Projects
          </label>
          <ProjectSelector
            value={selectedProjectIds}
            onChange={setSelectedProjectIds}
            projects={projectState.projects}
          />
        </div>

        <div className="border-t border-zinc-300 dark:border-zinc-700 pt-4 mt-4">
          <ContactSection<PendingContact>
            contacts={displayContacts}
            onAdd={handleAddContact}
            onRemove={handleRemoveContact}
            onUpdate={handleUpdateContact}
            onSetPrimary={handleSetPrimaryContact}
            fields={contactFields}
            emptyContact={emptyPendingContact}
            display={{ primaryLabel: "Primary" }}
            showFormByDefault={false}
            individualSearch={{
              enabled: true,
              onSearch: handleSearchContacts,
            }}
          />
        </div>

        <div className="border-t border-zinc-300 dark:border-zinc-700 pt-4 mt-4">
          <RelationshipsSection
            relationships={displayRelationships}
            onAdd={handleAddRelationship}
            onRemove={handleRemoveRelationship}
            searchType={relationshipSearchType}
            onSearch={handleSearchRelationships}
          />
        </div>

        <div className="border-t border-zinc-300 dark:border-zinc-700 pt-4 mt-4">
          <NotesSection
            entityType={isOrganization ? "organization" : "individual"}
            entityId={isEditMode ? account?.id ?? null : null}
            pendingNotes={!isEditMode ? pendingNotes : undefined}
            onPendingNotesChange={!isEditMode ? setPendingNotes : undefined}
          />
        </div>

        <div className="border-t border-zinc-200 dark:border-zinc-700 pt-4 mt-4">
          <FormActions
            isEditMode={isEditMode}
            isSubmitting={isSubmitting}
            isDeleting={isDeleting}
            hasPendingChanges={hasPendingChanges}
            onClose={onClose}
            onDelete={onDelete ? handleDelete : undefined}
            variant="compact"
          />
        </div>

        {isEditMode && account && (
          <Timestamps
            createdAt={(account as any).created_at}
            updatedAt={(account as any).updated_at}
          />
        )}

        {errors._form && (
          <div className="mt-4 p-3 bg-red-100 dark:bg-red-900/30 text-red-600 dark:text-red-400 rounded">
            {errors._form}
          </div>
        )}
      </form>
    </div>
  );
};

export default AccountForm;
