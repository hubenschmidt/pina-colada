"use client";

import { useState, useEffect } from "react";
import { ExternalLink } from "lucide-react";
import { BaseLead } from "./types/LeadTrackerTypes";
import { LeadFormConfig, FormFieldConfig } from "./types/LeadFormTypes";
import { ContactInput } from "../../types/types";
import ContactSection, {
  ContactFieldConfig,
  SearchResult,
} from "../ContactSection";
import { searchContacts, updateJob } from "../../api";
import FormActions from "../FormActions";
import { usePendingChanges } from "../../hooks/usePendingChanges";

const emptyContact = (): ContactInput => ({
  first_name: "",
  last_name: "",
  email: "",
  phone: "",
});

interface LeadFormProps<T extends BaseLead> {
  onClose: () => void;
  onAdd?: (lead: Omit<T, "id" | "created_at" | "updated_at">) => Promise<void>;
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
  const isEditMode = !!lead;
  const [formData, setFormData] = useState<any>({});
  const [contacts, setContacts] = useState<ContactInput[]>([]);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [isDeleting, setIsDeleting] = useState(false);
  const [errors, setErrors] = useState<{ [key: string]: string }>({});

  // Parse field value based on field type
  const parseFieldValue = (field: FormFieldConfig<T>, value: any): string => {
    if (field.type === "date" && value) {
      return new Date(value).toISOString().split("T")[0];
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
    if (typeof err?.message === "string") return err.message;
    if (typeof err?.error === "string") return err.error;
    if (typeof err?.errorData?.detail === "string") return err.errorData.detail;
    if (typeof err?.errorData?.error === "string") return err.errorData.error;
    if (err?.errorData && typeof err.errorData === "object") {
      const validationErrors = Object.entries(err.errorData)
        .map(([key, value]) => `${key}: ${value}`)
        .join(", ");
      return validationErrors;
    }
    if (typeof err === "string") return err;
    return `Failed to ${isEditMode ? "update" : "add"}. Please try again.`;
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!validate()) {
      return;
    }

    setIsSubmitting(true);
    try {
      let submitData = { ...formData };

      // If Account Type is Individual, combine individual first/last name into account
      if (submitData.account_type === "Individual" && !isEditMode) {
        const lastName = submitData.individual_last_name || "";
        const firstName = submitData.individual_first_name || "";
        if (lastName || firstName) {
          submitData.account = `${lastName}, ${firstName}`
            .trim()
            .replace(/^,\s*|,\s*$/g, "");
        }
      }

      // Filter out empty contacts and add to submit data
      const validContacts = contacts.filter(
        (c) => c.first_name.trim() && c.last_name.trim()
      );
      if (validContacts.length > 0) {
        submitData.contacts = validContacts;
      }

      // Remove temporary fields (they're for form only)
      delete submitData.individual_first_name;
      delete submitData.individual_last_name;

      // Pre-process data if needed
      if (config.onBeforeSubmit) {
        submitData = config.onBeforeSubmit(submitData, isEditMode);
      }

      if (isEditMode && lead && onUpdate) {
        // Edit mode: update existing lead
        await onUpdate(lead.id, submitData);
        onClose();
        return;
      }

      if (!onAdd) {
        onClose();
        return;
      }

      // Add mode: create new lead
      await onAdd(submitData);

      // Reset form only in add mode
      const resetData: any = {};
      config.fields.forEach((field) => {
        if (field.defaultValue !== undefined) {
          resetData[field.name] = field.defaultValue;
          return;
        }
        if (field.type === "date") {
          resetData[field.name] = new Date().toISOString().split("T")[0];
          return;
        }
        if (field.type === "checkbox") {
          resetData[field.name] = false;
          return;
        }
        resetData[field.name] = "";
      });
      setFormData(resetData);
      setContacts([]);
      setErrors({});
      onClose();
    } catch (error: any) {
      console.error(
        isEditMode ? "Failed to update lead:" : "Failed to add lead:",
        error
      );
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

      <FormActions
        isEditMode={isEditMode}
        isSubmitting={isSubmitting}
        isDeleting={isDeleting}
        hasPendingChanges={hasPendingChanges}
        onClose={onClose}
        onDelete={onDelete ? handleDelete : undefined}
        cancelButtonText={config.cancelButtonText}
      />
    </form>
  );

  return (
    <div>
      <h1 className="text-2xl font-bold text-zinc-900 dark:text-zinc-100 mb-6">
        {isEditMode ? config.title.replace("Add New", "Edit") : config.title}
      </h1>
      {formContent}
      {errors._form && (
        <div className="mt-4 p-3 bg-red-100 dark:bg-red-900/30 text-red-600 dark:text-red-400 rounded">
          {errors._form}
        </div>
      )}
    </div>
  );
};

export default LeadForm;
