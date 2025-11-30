"use client";

import { useState, useEffect, useContext } from "react";
import { ExternalLink, FolderKanban } from "lucide-react";
import { Badge, Group } from "@mantine/core";
import { BaseLead } from "./types/LeadTrackerTypes";
import { LeadFormConfig, FormFieldConfig } from "./types/LeadFormTypes";
import { ContactInput } from "../../types/types";
import ContactSection, {
  ContactFieldConfig,
  SearchResult,
} from "../ContactSection/ContactSection";
import { searchContacts, updateJob, createNote, createTask, TaskInput, linkDocumentToEntity } from "../../api";
import FormActions from "../FormActions/FormActions";
import NotesSection from "../NotesSection/NotesSection";
import TasksSection from "../TasksSection/TasksSection";
import CommentsSection from "../CommentsSection/CommentsSection";
import DocumentsSection from "../DocumentsSection/DocumentsSection";
import Timestamps from "../Timestamps/Timestamps";
import { usePendingChanges } from "../../hooks/usePendingChanges";
import { ProjectContext } from "../../context/projectContext";

const emptyContact = (): ContactInput => ({
  first_name: "",
  last_name: "",
  email: "",
  phone: "",
});

interface LeadFormProps<T extends BaseLead> {
  onClose: () => void;
  onAdd?: (lead: Omit<T, "id" | "created_at" | "updated_at">) => Promise<T | void>;
  config: LeadFormConfig<T>;
  // Edit mode props
  lead?: T | null;
  onUpdate?: (id: string, updates: Partial<T>) => Promise<void>;
  onDelete?: (id: string) => Promise<void>;
}

const LeadForm = <T extends BaseLead>({
  onClose,
  onAdd,
  config,
  lead,
  onUpdate,
  onDelete,
}: LeadFormProps<T>) => {
  const { projectState } = useContext(ProjectContext);
  const selectedProject = projectState.selectedProject;
  const isEditMode = !!lead;
  const [formData, setFormData] = useState<any>({});
  const [contacts, setContacts] = useState<ContactInput[]>([]);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [isDeleting, setIsDeleting] = useState(false);
  const [errors, setErrors] = useState<{ [key: string]: string }>({});
  const [pendingNotes, setPendingNotes] = useState<string[]>([]);
  const [pendingTasks, setPendingTasks] = useState<TaskInput[]>([]);
  const [pendingComments, setPendingComments] = useState<string[]>([]);
  const [pendingDocumentIds, setPendingDocumentIds] = useState<number[]>([]);

  // Parse field value based on field type
  const parseFieldValue = (field: FormFieldConfig<T>, value: any): any => {
    if (field.type === "date" && value) {
      return new Date(value).toISOString().split("T")[0];
    }
    // Preserve arrays as-is (e.g., project_ids)
    if (Array.isArray(value)) {
      return value;
    }
    return value ?? "";
  };

  // Parse Individual account name into first/last name
  const parseIndividualAccount = (data: any) => {
    if (data.account_type !== "Individual" || !data.account) {
      return;
    }

    const parts = data.account.split(", ");
    if (parts.length === 2) {
      data.individual_last_name = parts[0];
      data.individual_first_name = parts[1];
      return;
    }

    data.individual_first_name = data.account;
  };

  // Update contact with individual field value
  const updateContactWithIndividualField = (
    prevContacts: ContactInput[],
    fieldName: string,
    value: string
  ): ContactInput[] => {
    const newContacts =
      prevContacts.length > 0 ? [...prevContacts] : [emptyContact()];
    if (fieldName === "individual_first_name") {
      newContacts[0] = { ...newContacts[0], first_name: value };
      return newContacts;
    }
    if (fieldName === "individual_last_name") {
      newContacts[0] = { ...newContacts[0], last_name: value };
    }
    return newContacts;
  };

  // Validate a single field
  const validateField = (
    field: FormFieldConfig<T>,
    value: any,
    errors: { [key: string]: string }
  ) => {
    if (field.required && (!value || value === "")) {
      errors[String(field.name)] = `${field.label} is required`;
      return;
    }

    if (!field.validation || !value) {
      return;
    }

    const error = field.validation(value);
    if (error) {
      errors[String(field.name)] = error;
    }
  };

  // Validate form-level rules
  const validateFormLevel = (
    formData: any,
    errors: { [key: string]: string }
  ) => {
    if (!config.onValidate) {
      return;
    }

    const formErrors = config.onValidate(formData);
    if (formErrors) {
      Object.assign(errors, formErrors);
    }
  };

  // Contact fields config for LeadForm
  const contactFields: ContactFieldConfig[] = [
    { name: "first_name", label: "First Name", placeholder: "e.g., John" },
    { name: "last_name", label: "Last Name", placeholder: "e.g., Doe" },
    { name: "title", label: "Title", placeholder: "e.g., VP of Sales" },
    {
      name: "email",
      label: "Email",
      type: "email",
      placeholder: "email@example.com",
    },
    {
      name: "phone",
      label: "Phone",
      type: "tel",
      placeholder: "+1-555-123-4567",
    },
  ];

  // Initialize form data
  useEffect(() => {
    if (isEditMode && lead) {
      // Edit mode: populate from existing lead
      const data: any = {};
      config.fields.forEach((field) => {
        const value = (lead as any)[field.name];
        data[field.name] = parseFieldValue(field, value);
      });
      // Handle Individual account type - parse account into first/last name
      parseIndividualAccount(data);
      setFormData(data);
      // Load contacts from lead if available
      const leadContacts = (lead as any).contacts;
      setContacts(Array.isArray(leadContacts) ? leadContacts : []);
      setErrors({});
      setIsDeleting(false);
      return;
    }

    // Add mode: use defaults
    const initialData: any = {};
    config.fields.forEach((field) => {
      if (field.defaultValue !== undefined) {
        initialData[field.name] = field.defaultValue;
        return;
      }
      if (field.type === "date") {
        initialData[field.name] = new Date().toISOString().split("T")[0];
        return;
      }
      if (field.type === "checkbox") {
        initialData[field.name] = false;
        return;
      }
      initialData[field.name] = "";
    });
    setFormData(initialData);
    setContacts([]);
    setErrors({});
    setIsDeleting(false);

    // Run onInit for fields that have it (only in add mode)
    config.fields.forEach(async (field) => {
      if (field.onInit) {
        try {
          const value = await field.onInit();
          setFormData((prev: any) => ({ ...prev, [field.name]: value }));
        } catch (error) {
          console.error(
            `Failed to initialize field ${String(field.name)}:`,
            error
          );
        }
      }
    });
  }, [config.fields, isEditMode, lead]);

  const hasPendingChanges = usePendingChanges({
    original: lead as Record<string, unknown> | null,
    current: formData,
  });

  const handleFieldChange = (fieldName: string, value: any) => {
    const field = config.fields.find((f) => f.name === fieldName);

    let processedValue = value;
    if (field?.onChange) {
      processedValue = field.onChange(value, formData);
    }

    setFormData((prev: any) => {
      const newData = { ...prev, [fieldName]: processedValue };

      // When account_type changes, clear account and individual fields
      if (fieldName === "account_type") {
        newData.account = "";
        newData.individual_first_name = "";
        newData.individual_last_name = "";
        // Reset contacts when switching account type
        setContacts([]);
      }

      // When individual fields change and Account Type is Individual, auto-populate first contact
      if (
        (fieldName === "individual_first_name" ||
          fieldName === "individual_last_name") &&
        newData.account_type === "Individual"
      ) {
        setContacts((prevContacts) =>
          updateContactWithIndividualField(
            prevContacts,
            fieldName,
            processedValue
          )
        );
      }

      return newData;
    });

    // Clear error for this field
    if (errors[fieldName]) {
      setErrors((prev) => {
        const newErrors = { ...prev };
        delete newErrors[fieldName];
        return newErrors;
      });
    }
  };

  const validate = (): boolean => {
    const newErrors: { [key: string]: string } = {};

    // Field-level validation
    config.fields.forEach((field) => {
      const value = formData[field.name];
      validateField(field, value, newErrors);
    });

    // Form-level validation
    validateFormLevel(formData, newErrors);

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const getErrorMessage = (err: any): string => {
    const extractors = [
      () => typeof err?.message === "string" ? err.message : null,
      () => typeof err?.error === "string" ? err.error : null,
      () => typeof err?.errorData?.detail === "string" ? err.errorData.detail : null,
      () => typeof err?.errorData?.error === "string" ? err.errorData.error : null,
      () => err?.errorData && typeof err.errorData === "object"
        ? Object.entries(err.errorData).map(([k, v]) => `${k}: ${v}`).join(", ")
        : null,
      () => typeof err === "string" ? err : null,
    ];
    for (const extractor of extractors) {
      const result = extractor();
      if (result) return result;
    }
    return `Failed to ${isEditMode ? "update" : "add"}. Please try again.`;
  };

  const buildIndividualAccountName = (data: any): string | undefined => {
    if (data.account_type !== "Individual" || isEditMode) return undefined;
    const lastName = data.individual_last_name || "";
    const firstName = data.individual_first_name || "";
    if (!lastName && !firstName) return undefined;
    return `${lastName}, ${firstName}`.trim().replace(/^,\s*|,\s*$/g, "");
  };

  const prepareSubmitData = (data: any): any => {
    const submitData = { ...data };

    const accountName = buildIndividualAccountName(submitData);
    if (accountName) {
      submitData.account = accountName;
    }

    const validContacts = contacts.filter((c) => c.first_name.trim() && c.last_name.trim());
    if (validContacts.length > 0) {
      submitData.contacts = validContacts;
    }

    delete submitData.individual_first_name;
    delete submitData.individual_last_name;

    return config.onBeforeSubmit ? config.onBeforeSubmit(submitData, isEditMode) : submitData;
  };

  const createPendingNotes = async (leadId: number) => {
    for (const noteContent of pendingNotes) {
      await createNote("job", leadId, noteContent);
    }
  };

  const createPendingTasks = async (leadId: number) => {
    for (const taskData of pendingTasks) {
      await createTask({
        ...taskData,
        taskable_type: "Lead",
        taskable_id: leadId,
      });
    }
  };

  const getFieldDefaultValue = (field: FormFieldConfig<T>): any => {
    if (field.defaultValue !== undefined) return field.defaultValue;
    if (field.type === "date") return new Date().toISOString().split("T")[0];
    if (field.type === "checkbox") return false;
    return "";
  };

  const resetForm = () => {
    const resetData: any = {};
    config.fields.forEach((field) => {
      resetData[field.name] = getFieldDefaultValue(field);
    });
    setFormData(resetData);
    setContacts([]);
    setErrors({});
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!validate()) return;

    setIsSubmitting(true);
    try {
      const submitData = prepareSubmitData(formData);

      if (isEditMode && lead && onUpdate) {
        await onUpdate(lead.id, submitData);
        onClose();
        return;
      }

      if (!onAdd) {
        onClose();
        return;
      }

      const createdLead = await onAdd(submitData);
      if (createdLead) {
        const leadId = parseInt(createdLead.id, 10);
        if (pendingNotes.length > 0) {
          await createPendingNotes(leadId);
        }
        if (pendingTasks.length > 0) {
          await createPendingTasks(leadId);
        }
        if (pendingDocumentIds.length > 0) {
          await linkPendingDocuments(leadId);
        }
      }

      resetForm();
      onClose();
    } catch (error: any) {
      console.error(isEditMode ? "Failed to update lead:" : "Failed to add lead:", error);
      setErrors({ _form: getErrorMessage(error) });
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleDelete = async () => {
    if (!lead || !onDelete) return;

    if (!isDeleting) {
      setIsDeleting(true);
      return;
    }

    try {
      await onDelete(lead.id);
      onClose();
    } catch (error: any) {
      console.error("Failed to delete:", error);
      setErrors({ _form: error?.message || "Failed to delete. Please try again." });
      setIsDeleting(false);
    }
  };

  const renderField = (field: FormFieldConfig<T>) => {
    if (field.hidden) return null;

    const accountType = formData["account_type"] || "Organization";
    const isAccountField = field.name === "account";
    const isAccountTypeField = field.name === "account_type";

    // In edit mode, account-related fields are read-only
    const isAccountReadOnly =
      isEditMode && (isAccountField || isAccountTypeField);
    const readOnlyClasses =
      "bg-zinc-100 dark:bg-zinc-700 cursor-not-allowed opacity-75";

    // If Individual is selected, render First Name instead of account field
    if (isAccountField && accountType === "Individual") {
      return (
        <div className={field.gridColumn || ""} key="individual_first_name">
          <label className="block text-sm font-medium text-zinc-700 dark:text-zinc-300 mb-1">
            First Name{" "}
            {field.required && !isEditMode && (
              <span className="text-red-500">*</span>
            )}
          </label>
          <input
            type="text"
            value={formData.individual_first_name || ""}
            onChange={(e) =>
              !isEditMode &&
              handleFieldChange("individual_first_name", e.target.value)
            }
            readOnly={isEditMode}
            className={`w-full px-3 py-2 border ${
              errors["individual_first_name"]
                ? "border-red-500"
                : "border-zinc-300 dark:border-zinc-700"
            } rounded focus:outline-none focus:ring-2 focus:ring-lime-500 bg-white dark:bg-zinc-800 text-zinc-900 dark:text-zinc-100 ${
              isEditMode ? readOnlyClasses : ""
            }`}
            placeholder="e.g., John"
            required={field.required && !isEditMode}
          />
          {errors["individual_first_name"] && (
            <p className="text-red-500 text-xs mt-1">
              {errors["individual_first_name"]}
            </p>
          )}
        </div>
      );
    }

    // If Individual is selected, render Account Type followed by Last Name
    if (isAccountTypeField && accountType === "Individual") {
      return [
        <div className={field.gridColumn || ""} key="account_type">
          <label className="block text-sm font-medium text-zinc-700 dark:text-zinc-300 mb-1">
            {field.label}
          </label>
          <select
            value={formData.account_type || "Organization"}
            onChange={(e) =>
              !isEditMode && handleFieldChange("account_type", e.target.value)
            }
            disabled={isEditMode}
            className={`w-full px-3 py-2 border border-zinc-300 dark:border-zinc-700 rounded focus:outline-none focus:ring-2 focus:ring-lime-500 bg-white dark:bg-zinc-800 text-zinc-900 dark:text-zinc-100 ${
              isEditMode ? readOnlyClasses : ""
            }`}
          >
            {field.options?.map((option) => (
              <option key={option.value} value={option.value}>
                {option.label}
              </option>
            ))}
          </select>
        </div>,
        <div className={field.gridColumn || ""} key="individual_last_name">
          <label className="block text-sm font-medium text-zinc-700 dark:text-zinc-300 mb-1">
            Last Name {!isEditMode && <span className="text-red-500">*</span>}
          </label>
          <input
            type="text"
            value={formData.individual_last_name || ""}
            onChange={(e) =>
              !isEditMode &&
              handleFieldChange("individual_last_name", e.target.value)
            }
            readOnly={isEditMode}
            className={`w-full px-3 py-2 border ${
              errors["individual_last_name"]
                ? "border-red-500"
                : "border-zinc-300 dark:border-zinc-700"
            } rounded focus:outline-none focus:ring-2 focus:ring-lime-500 bg-white dark:bg-zinc-800 text-zinc-900 dark:text-zinc-100 ${
              isEditMode ? readOnlyClasses : ""
            }`}
            placeholder="e.g., Doe"
            required={!isEditMode}
          />
          {errors["individual_last_name"] && (
            <p className="text-red-500 text-xs mt-1">
              {errors["individual_last_name"]}
            </p>
          )}
        </div>,
      ];
    }

    const value = formData[field.name];
    const error = errors[String(field.name)];

    const inputClasses = `w-full px-3 py-2 border ${
      error ? "border-red-500" : "border-zinc-300 dark:border-zinc-700"
    } rounded focus:outline-none focus:ring-2 focus:ring-lime-500 bg-white dark:bg-zinc-800 text-zinc-900 dark:text-zinc-100`;

    const fieldWrapper = (content: React.ReactNode) => (
      <div className={field.gridColumn || ""} key={String(field.name)}>
        <label className="block text-sm font-medium text-zinc-700 dark:text-zinc-300 mb-1">
          {field.label}{" "}
          {field.required && <span className="text-red-500">*</span>}
        </label>
        {content}
        {error && <p className="text-red-500 text-xs mt-1">{error}</p>}
      </div>
    );

    if (field.type === "custom" && field.renderCustom) {
      return fieldWrapper(
        field.renderCustom({
          value,
          onChange: (v) =>
            !isAccountReadOnly && handleFieldChange(String(field.name), v),
          field,
          formData,
          isEditMode,
        })
      );
    }

    if (field.type === "textarea") {
      return fieldWrapper(
        <textarea
          value={value || ""}
          onChange={(e) =>
            handleFieldChange(String(field.name), e.target.value)
          }
          className={inputClasses}
          placeholder={field.placeholder}
          rows={field.rows || 3}
          required={field.required}
          disabled={field.disabled}
        />
      );
    }

    if (field.type === "select") {
      const isDisabled = field.disabled || isAccountReadOnly;
      return fieldWrapper(
        <select
          value={value || ""}
          onChange={(e) =>
            !isDisabled && handleFieldChange(String(field.name), e.target.value)
          }
          className={`${inputClasses} ${isDisabled ? readOnlyClasses : ""}`}
          required={field.required}
          disabled={isDisabled}
        >
          {field.options?.map((option) => (
            <option key={option.value} value={option.value}>
              {option.label}
            </option>
          ))}
        </select>
      );
    }

    if (field.type === "checkbox") {
      return fieldWrapper(
        <label className="flex items-center gap-2 cursor-pointer">
          <input
            type="checkbox"
            checked={value || false}
            onChange={(e) =>
              handleFieldChange(String(field.name), e.target.checked)
            }
            className="w-4 h-4 text-lime-500 border-zinc-300 dark:border-zinc-700 bg-white dark:bg-zinc-800 rounded focus:ring-lime-500"
            disabled={field.disabled}
          />
          <span className="text-sm text-zinc-600 dark:text-zinc-400">
            {field.placeholder}
          </span>
        </label>
      );
    }

    // Handle URL fields with clickable link
    const isUrlField = String(field.name).endsWith("_url");
    const hasUrl = isUrlField && value && String(value).trim();

    if (hasUrl) {
      const url = String(value).startsWith("http") ? String(value) : `https://${value}`;
      return fieldWrapper(
        <div>
          <input
            type="text"
            value={value || ""}
            onChange={(e) => handleFieldChange(String(field.name), e.target.value)}
            className={inputClasses}
            placeholder={field.placeholder}
            required={field.required}
            disabled={field.disabled}
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

    return fieldWrapper(
      <input
        type={field.type}
        value={value || ""}
        onChange={(e) => handleFieldChange(String(field.name), e.target.value)}
        className={inputClasses}
        placeholder={field.placeholder}
        required={field.required}
        min={field.min}
        max={field.max}
        step={field.step}
        pattern={field.pattern}
        disabled={field.disabled}
      />
    );
  };

  // Contact handlers - auto-save in edit mode (bypasses onUpdate to avoid navigation)
  const saveContacts = async (newContacts: ContactInput[]) => {
    if (isEditMode && lead) {
      try {
        await updateJob((lead as any).id, { contacts: newContacts });
      } catch (err) {
        console.error("Failed to save contacts:", err);
      }
    }
  };

  const handleAddContact = (contact: ContactInput) => {
    const newContacts = [...contacts, contact];
    setContacts(newContacts);
    saveContacts(newContacts);
  };

  const handleUpdateContact = (index: number, updatedContact: ContactInput) => {
    const newContacts = contacts.map((c, i) =>
      i === index ? updatedContact : c
    );
    setContacts(newContacts);
    saveContacts(newContacts);
  };

  const handleSetPrimaryContact = (index: number) => {
    const newContacts = contacts.map((c, i) => ({
      ...c,
      is_primary: i === index,
    }));
    setContacts(newContacts);
    saveContacts(newContacts);
  };

  const handleSearchContacts = async (
    query: string
  ): Promise<SearchResult[]> => {
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

  const isContactLocked = (_: ContactInput, index: number) => {
    const accountType = formData["account_type"] || "Organization";
    return accountType === "Individual" && index === 0;
  };

  const formContent = (
    <form onSubmit={handleSubmit}>
      <div className="space-y-6">
        {config.sections && config.sections.length > 0 ? (
          // Render sections if configured
          config.sections.map((section, sectionIndex) => {
            // Special handling for Contact section - use dynamic contacts UI
            const isContactSection = section.name === "Contact";

            if (isContactSection) {
              return (
                <div key={section.name}>
                  {sectionIndex > 0 && (
                    <div className="border-t border-zinc-300 dark:border-zinc-700 mb-6"></div>
                  )}
                  <ContactSection<ContactInput>
                    contacts={contacts}
                    onAdd={handleAddContact}
                    onUpdate={handleUpdateContact}
                    onSetPrimary={handleSetPrimaryContact}
                    fields={contactFields}
                    emptyContact={emptyContact}
                    display={{ nameFields: ["first_name", "last_name"] }}
                    isContactLocked={isContactLocked}
                    individualSearch={{
                      enabled: true,
                      onSearch: handleSearchContacts,
                    }}
                    disabled={!isEditMode && !formData.account}
                    disabledMessage="Select an account first to add contacts..."
                  />
                </div>
              );
            }

            const sectionFields = config.fields.filter((field) =>
              section.fieldNames.includes(String(field.name))
            );

            return (
              <div key={section.name}>
                {sectionIndex > 0 && (
                  <div className="border-t border-zinc-300 dark:border-zinc-700 mb-6"></div>
                )}
                <div>
                  <h4 className="text-sm font-semibold text-zinc-900 dark:text-zinc-100 mb-4">
                    {section.name}
                  </h4>
                  <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                    {sectionFields.flatMap((field) => {
                      const rendered = renderField(field);
                      // Handle case where renderField returns an array (for Individual fields)
                      if (Array.isArray(rendered)) {
                        return rendered;
                      }
                      return rendered ? [rendered] : [];
                    })}
                  </div>
                </div>
              </div>
            );
          })
        ) : (
          // Fallback: render all fields in a single grid if no sections configured
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            {config.fields.map((field) => renderField(field))}
          </div>
        )}
      </div>

      <div className="border-t border-zinc-300 dark:border-zinc-700 pt-4 mt-4">
        <NotesSection
          entityType="job"
          entityId={isEditMode && lead ? parseInt(lead.id, 10) : null}
          pendingNotes={!isEditMode ? pendingNotes : undefined}
          onPendingNotesChange={!isEditMode ? setPendingNotes : undefined}
        />
      </div>

      <div className="border-t border-zinc-300 dark:border-zinc-700 pt-4 mt-4">
        <TasksSection
          entityType="Lead"
          entityId={isEditMode && lead ? parseInt(lead.id, 10) : null}
          pendingTasks={!isEditMode ? pendingTasks : undefined}
          onPendingTasksChange={!isEditMode ? setPendingTasks : undefined}
        />
      </div>

      <div className="border-t border-zinc-300 dark:border-zinc-700 pt-4 mt-4">
        <DocumentsSection
          entityType="Lead"
          entityId={isEditMode && lead ? parseInt(lead.id, 10) : null}
          pendingDocumentIds={!isEditMode ? pendingDocumentIds : undefined}
          onPendingDocumentIdsChange={!isEditMode ? setPendingDocumentIds : undefined}
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
          cancelButtonText={config.cancelButtonText}
        />
      </div>

      {isEditMode && lead && (
        <Timestamps
          createdAt={(lead as any).created_at}
          updatedAt={(lead as any).updated_at}
        />
      )}
    </form>
  );

  return (
    <div>
      <Group justify="space-between" mb="md">
        <h1 className="text-2xl font-bold text-zinc-900 dark:text-zinc-100">
          {isEditMode ? config.title.replace("Add New", "Edit") : config.title}
        </h1>
        {selectedProject ? (
          <Badge variant="light" color="lime" leftSection={<FolderKanban className="h-3 w-3" />}>
            {selectedProject.name}
          </Badge>
        ) : (
          <Badge variant="light" color="gray">
            Global
          </Badge>
        )}
      </Group>
      {formContent}
      {errors._form && (
        <div className="mt-4 p-3 bg-red-100 dark:bg-red-900/30 text-red-600 dark:text-red-400 rounded">
          {errors._form}
        </div>
      )}

      {/* Comments Section - below the form */}
      <div className="border-t border-zinc-200 dark:border-zinc-700 pt-6 mt-6">
        <CommentsSection
          entityType="Lead"
          entityId={isEditMode && lead ? parseInt(lead.id, 10) : null}
          pendingComments={!isEditMode ? pendingComments : undefined}
          onPendingCommentsChange={!isEditMode ? setPendingComments : undefined}
        />
      </div>
    </div>
  );
};

export default LeadForm;
