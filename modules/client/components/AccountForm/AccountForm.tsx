"use client";

import { useState, useEffect } from "react";
import ContactSection, { ContactFieldConfig } from "../ContactSection";
import NotesSection from "../NotesSection";
import FormActions from "../FormActions";
import {
  Contact,
  createIndividualContact,
  deleteIndividualContact,
  createOrganizationContact,
  deleteOrganizationContact,
  searchContacts,
  createNote,
} from "../../api";
import { SearchResult } from "../ContactSection";
import { AccountType, FormFieldConfig } from "./types/AccountFormTypes";
import { useAccountFormConfig } from "./hooks/useAccountFormConfig";

interface OrganizationData {
  id?: number;
  name: string;
  website?: string | null;
  phone?: string | null;
  employee_count?: number | null;
  description?: string | null;
  contacts?: Contact[];
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
}

const emptyPendingContact = (): PendingContact => ({
  first_name: "",
  last_name: "",
  email: "",
  phone: "",
});

const inputClasses =
  "w-full px-3 py-2 border border-zinc-300 dark:border-zinc-700 rounded focus:outline-none focus:ring-2 focus:ring-lime-500 bg-white dark:bg-zinc-800 text-zinc-900 dark:text-zinc-100";

const renderField = (
  field: FormFieldConfig,
  value: string | number,
  onChange: (name: string, value: string) => void
) => {
  const handleChange = (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement>) => {
    const newValue = field.onChange ? field.onChange(e.target.value) : e.target.value;
    onChange(field.name, newValue);
  };

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
  const isEditMode = !!account;
  const isOrganization = type === "organization";

  const [formData, setFormData] = useState<Record<string, unknown>>({});
  const [contacts, setContacts] = useState<Contact[]>([]);
  const [pendingContacts, setPendingContacts] = useState<PendingContact[]>([]);
  const [pendingNotes, setPendingNotes] = useState<string[]>([]);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [isDeleting, setIsDeleting] = useState(false);
  const [errors, setErrors] = useState<Record<string, string>>({});

  useEffect(() => {
    if (isEditMode && account) {
      setFormData({ ...account });
      setContacts(account.contacts || []);
      setPendingContacts([]);
      setPendingNotes([]);
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
    setPendingContacts([]);
    setPendingNotes([]);
    setErrors({});
    setIsDeleting(false);
  }, [isEditMode, account, config.fields]);

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

      if (isEditMode && account && onUpdate) {
        await onUpdate(account.id!, submitData as Partial<OrganizationData | IndividualData>);
        onClose();
        return;
      }

      if (!onAdd) {
        onClose();
        return;
      }

      const created = await onAdd(submitData as unknown as OrganizationData | IndividualData);

      for (const pending of pendingContacts) {
        try {
          if (isOrganization && pending.individual_id) {
            await createOrganizationContact(created.id, {
              individual_id: pending.individual_id,
              email: pending.email?.trim() || undefined,
              phone: pending.phone?.trim() || undefined,
              title: pending.title?.trim() || undefined,
              notes: pending.notes?.trim() || undefined,
              is_primary: false,
            });
          }
          if (!isOrganization) {
            await createIndividualContact(created.id, {
              email: pending.email?.trim() || undefined,
              phone: pending.phone?.trim() || undefined,
              title: pending.title?.trim() || undefined,
              notes: pending.notes?.trim() || undefined,
              is_primary: false,
            });
          }
        } catch (err) {
          console.error("Failed to create contact:", err);
        }
      }

      const entityType = isOrganization ? "organization" : "individual";
      for (const noteContent of pendingNotes) {
        try {
          await createNote(entityType, created.id, noteContent);
        } catch (err) {
          console.error("Failed to create note:", err);
        }
      }

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

    try {
      if (isOrganization && !contact.individual_id) {
        setErrors({ _form: "Please select an individual from the search to add as a contact" });
        return;
      }

      if (isOrganization) {
        const newContactData = await createOrganizationContact(account.id, {
          individual_id: contact.individual_id!,
          email: contact.email?.trim() || undefined,
          phone: contact.phone?.trim() || undefined,
          is_primary: contacts.length === 0,
        });
        setContacts([...contacts, newContactData]);
        return;
      }

      const newContactData = await createIndividualContact(account.id, {
        email: contact.email?.trim() || undefined,
        phone: contact.phone?.trim() || undefined,
        is_primary: contacts.length === 0,
      });
      setContacts([...contacts, newContactData]);
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : "Failed to add contact";
      setErrors({ _form: message });
    }
  };

  const handleRemoveContact = async (index: number) => {
    if (!isEditMode) {
      setPendingContacts(pendingContacts.filter((_, i) => i !== index));
      return;
    }

    const contact = contacts[index];
    if (!account?.id || !contact) return;

    try {
      if (isOrganization) {
        await deleteOrganizationContact(account.id, contact.id);
        setContacts(contacts.filter((_, i) => i !== index));
        return;
      }

      await deleteIndividualContact(account.id, contact.id);
      setContacts(contacts.filter((_, i) => i !== index));
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : "Failed to remove contact";
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
    }));
  };

  const contactFields: ContactFieldConfig[] = [
    { name: "first_name", label: "First Name", placeholder: "e.g., John" },
    { name: "last_name", label: "Last Name", placeholder: "e.g., Doe" },
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
      }))
    : pendingContacts;

  // Check if we have paired fields (for 2-column grid layout)
  const hasPairedFields = config.fields.some((f) => f.gridColumn === "md:col-span-1");

  return (
    <div className="p-6 max-w-2xl">
      <h1 className="text-2xl font-bold text-zinc-900 dark:text-zinc-100 mb-6">
        {title}
      </h1>

      {errors._form && (
        <div className="mb-4 p-3 bg-red-100 dark:bg-red-900/30 text-red-600 dark:text-red-400 rounded">
          {errors._form}
        </div>
      )}

      <form onSubmit={handleSubmit} className="space-y-4">
        {hasPairedFields ? (
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            {config.fields.map((field) => (
              <div key={field.name} className={field.gridColumn === "md:col-span-1" ? "" : "md:col-span-2"}>
                <label className="block text-sm font-medium text-zinc-700 dark:text-zinc-300 mb-1">
                  {field.label}
                  {field.required && <span className="text-red-500"> *</span>}
                </label>
                {renderField(field, formData[field.name] as string | number, handleChange)}
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
              {renderField(field, formData[field.name] as string | number, handleChange)}
              {errors[field.name] && (
                <p className="mt-1 text-sm text-red-500">{errors[field.name]}</p>
              )}
            </div>
          ))
        )}

        <div className="border-t border-zinc-300 dark:border-zinc-700 pt-4 mt-4">
          <ContactSection<PendingContact>
            contacts={displayContacts}
            onAdd={handleAddContact}
            onRemove={handleRemoveContact}
            fields={contactFields}
            emptyContact={emptyPendingContact}
            display={{ primaryLabel: "Primary" }}
            showFormByDefault={false}
            isContactLocked={(_, index) => isEditMode && contacts[index]?.is_primary}
            individualSearch={{
              enabled: true,
              onSearch: handleSearchContacts,
            }}
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

        <FormActions
          isEditMode={isEditMode}
          isSubmitting={isSubmitting}
          isDeleting={isDeleting}
          onClose={onClose}
          onDelete={onDelete ? handleDelete : undefined}
          variant="compact"
        />
      </form>
    </div>
  );
};

export default AccountForm;
