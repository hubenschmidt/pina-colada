"use client";

import { useState, useEffect } from "react";
import { Plus, Trash2 } from "lucide-react";
import { Modal } from "@mantine/core";
import { BaseLead } from "./LeadTrackerConfig";
import { LeadFormConfig, FormFieldConfig } from "./LeadFormConfig";
import { ContactInput } from "../../types/types";

const emptyContact = (): ContactInput => ({
  first_name: "",
  last_name: "",
  email: "",
  phone: "",
});

interface LeadFormProps<T extends BaseLead> {
  isOpen: boolean;
  onClose: () => void;
  onAdd?: (lead: Omit<T, "id" | "created_at" | "updated_at">) => Promise<void>;
  config: LeadFormConfig<T>;
  // Edit mode props
  lead?: T | null;
  onUpdate?: (id: string, updates: Partial<T>) => Promise<void>;
  onDelete?: (id: string) => Promise<void>;
}

const LeadForm = <T extends BaseLead>({
  isOpen,
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

  // Initialize form data
  useEffect(() => {
    if (isOpen) {
      if (isEditMode && lead) {
        // Edit mode: populate from existing lead
        const data: any = {};
        config.fields.forEach((field) => {
          const value = (lead as any)[field.name];
          if (field.type === "date" && value) {
            data[field.name] = new Date(value).toISOString().split("T")[0];
            return;
          }
          data[field.name] = value ?? "";
        });
        // Handle Individual account type - parse account into first/last name
        if (data.account_type === "Individual" && data.account) {
          const parts = data.account.split(", ");
          if (parts.length === 2) {
            data.individual_last_name = parts[0];
            data.individual_first_name = parts[1];
          } else {
            data.individual_first_name = data.account;
          }
        }
        setFormData(data);
        // Load contacts from lead if available
        const leadContacts = (lead as any).contacts;
        setContacts(Array.isArray(leadContacts) ? leadContacts : []);
        setErrors({});
        setIsDeleting(false);
      } else {
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
              console.error(`Failed to initialize field ${String(field.name)}:`, error);
            }
          }
        });
      }
    }
  }, [isOpen, config.fields, isEditMode, lead]);

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
      if ((fieldName === "individual_first_name" || fieldName === "individual_last_name") &&
          newData.account_type === "Individual") {
        setContacts((prevContacts) => {
          const newContacts = prevContacts.length > 0 ? [...prevContacts] : [emptyContact()];
          if (fieldName === "individual_first_name") {
            newContacts[0] = { ...newContacts[0], first_name: processedValue };
          } else if (fieldName === "individual_last_name") {
            newContacts[0] = { ...newContacts[0], last_name: processedValue };
          }
          return newContacts;
        });
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

  // Contact management handlers
  const handleAddContact = () => {
    setContacts((prev) => [...prev, emptyContact()]);
  };

  const handleRemoveContact = (index: number) => {
    setContacts((prev) => prev.filter((_, i) => i !== index));
  };

  const handleContactChange = (index: number, field: keyof ContactInput, value: string) => {
    setContacts((prev) => {
      const newContacts = [...prev];
      newContacts[index] = { ...newContacts[index], [field]: value };
      return newContacts;
    });
  };

  const validate = (): boolean => {
    const newErrors: { [key: string]: string } = {};

    // Field-level validation
    config.fields.forEach((field) => {
      const value = formData[field.name];

      // Required check
      if (field.required && (!value || value === "")) {
        newErrors[String(field.name)] = `${field.label} is required`;
      }

      // Custom validation
      if (field.validation && value) {
        const error = field.validation(value);
        if (error) {
          newErrors[String(field.name)] = error;
        }
      }
    });

    // Form-level validation
    if (config.onValidate) {
      const formErrors = config.onValidate(formData);
      if (formErrors) {
        Object.assign(newErrors, formErrors);
      }
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
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
          submitData.account = `${lastName}, ${firstName}`.trim().replace(/^,\s*|,\s*$/g, "");
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
        submitData = config.onBeforeSubmit(submitData);
      }

      if (isEditMode && lead && onUpdate) {
        // Edit mode: update existing lead
        await onUpdate(lead.id, submitData);
      } else if (onAdd) {
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
      }

      onClose();
    } catch (error: any) {
      console.error(isEditMode ? "Failed to update lead:" : "Failed to add lead:", error);
      const errorMessage =
        error?.message || error?.error || `Failed to ${isEditMode ? "update" : "add"}. Please try again.`;
      alert(errorMessage);
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
    } catch (error) {
      console.error("Failed to delete:", error);
      alert("Failed to delete. Please try again.");
      setIsDeleting(false);
    }
  };

  const renderField = (field: FormFieldConfig<T>) => {
    if (field.hidden) return null;

    const accountType = formData["account_type"] || "Organization";
    const isAccountField = field.name === "account";
    const isAccountTypeField = field.name === "account_type";

    // In edit mode, account-related fields are read-only
    const isAccountReadOnly = isEditMode && (isAccountField || isAccountTypeField);
    const readOnlyClasses = "bg-zinc-100 dark:bg-zinc-700 cursor-not-allowed opacity-75";

    // If Individual is selected, render First Name instead of account field
    if (isAccountField && accountType === "Individual") {
      return (
        <div className={field.gridColumn || ""} key="individual_first_name">
          <label className="block text-sm font-medium text-zinc-700 dark:text-zinc-300 mb-1">
            First Name{" "}
            {field.required && !isEditMode && <span className="text-red-500">*</span>}
          </label>
          <input
            type="text"
            value={formData.individual_first_name || ""}
            onChange={(e) => !isEditMode && handleFieldChange("individual_first_name", e.target.value)}
            readOnly={isEditMode}
            className={`w-full px-3 py-2 border ${
              errors["individual_first_name"] ? "border-red-500" : "border-zinc-300 dark:border-zinc-700"
            } rounded focus:outline-none focus:ring-2 focus:ring-lime-500 bg-white dark:bg-zinc-800 text-zinc-900 dark:text-zinc-100 ${isEditMode ? readOnlyClasses : ""}`}
            placeholder="e.g., John"
            required={field.required && !isEditMode}
          />
          {errors["individual_first_name"] && <p className="text-red-500 text-xs mt-1">{errors["individual_first_name"]}</p>}
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
            onChange={(e) => !isEditMode && handleFieldChange("account_type", e.target.value)}
            disabled={isEditMode}
            className={`w-full px-3 py-2 border border-zinc-300 dark:border-zinc-700 rounded focus:outline-none focus:ring-2 focus:ring-lime-500 bg-white dark:bg-zinc-800 text-zinc-900 dark:text-zinc-100 ${isEditMode ? readOnlyClasses : ""}`}
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
            Last Name{" "}
            {!isEditMode && <span className="text-red-500">*</span>}
          </label>
          <input
            type="text"
            value={formData.individual_last_name || ""}
            onChange={(e) => !isEditMode && handleFieldChange("individual_last_name", e.target.value)}
            readOnly={isEditMode}
            className={`w-full px-3 py-2 border ${
              errors["individual_last_name"] ? "border-red-500" : "border-zinc-300 dark:border-zinc-700"
            } rounded focus:outline-none focus:ring-2 focus:ring-lime-500 bg-white dark:bg-zinc-800 text-zinc-900 dark:text-zinc-100 ${isEditMode ? readOnlyClasses : ""}`}
            placeholder="e.g., Doe"
            required={!isEditMode}
          />
          {errors["individual_last_name"] && <p className="text-red-500 text-xs mt-1">{errors["individual_last_name"]}</p>}
        </div>
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
          onChange: (v) => !isAccountReadOnly && handleFieldChange(String(field.name), v),
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
          onChange={(e) => handleFieldChange(String(field.name), e.target.value)}
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
          onChange={(e) => !isDisabled && handleFieldChange(String(field.name), e.target.value)}
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
          <span className="text-sm text-zinc-600 dark:text-zinc-400">{field.placeholder}</span>
        </label>
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

  const renderContactsSection = () => {
    const accountType = formData["account_type"] || "Organization";
    const isFirstContactLocked = accountType === "Individual" && contacts.length > 0;

    return (
      <div className="space-y-4">
        {contacts.map((contact, index) => (
          <div key={index} className="p-4 border border-zinc-200 dark:border-zinc-700 rounded-lg relative">
            {index === 0 && (
              <span className="absolute -top-2.5 left-3 px-2 bg-white dark:bg-zinc-900 text-xs text-lime-600 dark:text-lime-400 font-medium">
                Primary Contact
              </span>
            )}
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div>
                <label className="block text-sm font-medium text-zinc-700 dark:text-zinc-300 mb-1">
                  First Name
                </label>
                <input
                  type="text"
                  value={contact.first_name}
                  onChange={(e) => handleContactChange(index, "first_name", e.target.value)}
                  disabled={index === 0 && isFirstContactLocked}
                  className={`w-full px-3 py-2 border border-zinc-300 dark:border-zinc-700 rounded focus:outline-none focus:ring-2 focus:ring-lime-500 bg-white dark:bg-zinc-800 text-zinc-900 dark:text-zinc-100 ${
                    index === 0 && isFirstContactLocked ? "opacity-60 cursor-not-allowed" : ""
                  }`}
                  placeholder="e.g., John"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-zinc-700 dark:text-zinc-300 mb-1">
                  Last Name
                </label>
                <input
                  type="text"
                  value={contact.last_name}
                  onChange={(e) => handleContactChange(index, "last_name", e.target.value)}
                  disabled={index === 0 && isFirstContactLocked}
                  className={`w-full px-3 py-2 border border-zinc-300 dark:border-zinc-700 rounded focus:outline-none focus:ring-2 focus:ring-lime-500 bg-white dark:bg-zinc-800 text-zinc-900 dark:text-zinc-100 ${
                    index === 0 && isFirstContactLocked ? "opacity-60 cursor-not-allowed" : ""
                  }`}
                  placeholder="e.g., Doe"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-zinc-700 dark:text-zinc-300 mb-1">
                  Email
                </label>
                <input
                  type="email"
                  value={contact.email || ""}
                  onChange={(e) => handleContactChange(index, "email", e.target.value)}
                  className="w-full px-3 py-2 border border-zinc-300 dark:border-zinc-700 rounded focus:outline-none focus:ring-2 focus:ring-lime-500 bg-white dark:bg-zinc-800 text-zinc-900 dark:text-zinc-100"
                  placeholder="e.g., john@example.com"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-zinc-700 dark:text-zinc-300 mb-1">
                  Phone
                </label>
                <input
                  type="text"
                  value={contact.phone || ""}
                  onChange={(e) => handleContactChange(index, "phone", e.target.value)}
                  className="w-full px-3 py-2 border border-zinc-300 dark:border-zinc-700 rounded focus:outline-none focus:ring-2 focus:ring-lime-500 bg-white dark:bg-zinc-800 text-zinc-900 dark:text-zinc-100"
                  placeholder="e.g., +1 (555) 123-4567"
                />
              </div>
            </div>
            {!(index === 0 && isFirstContactLocked) && (
              <button
                type="button"
                onClick={() => handleRemoveContact(index)}
                className="absolute top-2 right-2 p-1 text-zinc-400 hover:text-red-500 transition-colors"
                title="Remove contact"
              >
                <Trash2 size={16} />
              </button>
            )}
          </div>
        ))}
        <button
          type="button"
          onClick={handleAddContact}
          className="flex items-center gap-2 px-4 py-2 text-sm text-zinc-600 dark:text-zinc-400 hover:text-zinc-800 dark:hover:text-zinc-200 border border-dashed border-zinc-300 dark:border-zinc-600 rounded-lg hover:border-zinc-400 dark:hover:border-zinc-500 transition-colors"
        >
          <Plus size={16} />
          Add Contact
        </button>
      </div>
    );
  };

  return (
    <Modal.Root opened={isOpen} onClose={onClose} fullScreen>
      <Modal.Overlay backgroundOpacity={0.5} />
      <Modal.Content>
        <Modal.Header>
          <Modal.Title>{isEditMode ? config.title.replace("Add New", "Edit") : config.title}</Modal.Title>
          <Modal.CloseButton />
        </Modal.Header>
        <Modal.Body>
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
                    <div>
                      <h4 className="text-sm font-semibold text-zinc-900 dark:text-zinc-100 mb-4">
                        Contacts
                      </h4>
                      {renderContactsSection()}
                    </div>
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

            <div className="flex gap-3 mt-6">
              <button
                type="submit"
                disabled={isSubmitting}
                className="flex items-center gap-2 px-6 py-3 bg-zinc-800 dark:bg-zinc-700 text-white dark:text-zinc-100 rounded-lg hover:bg-zinc-700 dark:hover:bg-zinc-600 font-semibold disabled:opacity-50 disabled:cursor-not-allowed"
              >
                {!isEditMode && <Plus size={18} />}
                {isSubmitting
                  ? (isEditMode ? "Saving..." : "Adding...")
                  : (isEditMode ? "Save Changes" : (config.submitButtonText || "Add"))}
              </button>
              <button
                type="button"
                onClick={onClose}
                className="px-6 py-3 bg-zinc-200 dark:bg-zinc-800 text-zinc-700 dark:text-zinc-300 rounded-lg hover:bg-zinc-300 dark:hover:bg-zinc-700 font-semibold"
              >
                {config.cancelButtonText || "Cancel"}
              </button>
              {isEditMode && onDelete && (
                <button
                  type="button"
                  onClick={handleDelete}
                  className={`px-6 py-3 rounded-lg font-semibold ml-auto ${
                    isDeleting
                      ? "bg-red-600 text-white hover:bg-red-700"
                      : "bg-red-100 dark:bg-red-900/30 text-red-600 dark:text-red-400 hover:bg-red-200 dark:hover:bg-red-900/50"
                  }`}
                >
                  <Trash2 size={18} className="inline mr-2" />
                  {isDeleting ? "Confirm Delete" : "Delete"}
                </button>
              )}
            </div>
          </form>
        </Modal.Body>
      </Modal.Content>
    </Modal.Root>
  );
};

export default LeadForm;
