"use client";

import { useState, useEffect } from "react";
import { formatPhoneNumber } from "../utils/phone";
import ContactSection, { ContactFieldConfig } from "./ContactSection";
import NotesSection from "./NotesSection";
import FormActions from "./FormActions";
import {
  Contact,
  createIndividualContact,
  deleteIndividualContact,
  createOrganizationContact,
  deleteOrganizationContact,
  searchContacts,
  createNote,
} from "../api";
import { SearchResult } from "./ContactSection";

type AccountType = "organization" | "individual";

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
  // onAdd now returns the created entity with ID so we can create contacts
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

const AccountForm = ({
  type,
  onClose,
  onAdd,
  account,
  onUpdate,
  onDelete,
}: AccountFormProps) => {
  const isEditMode = !!account;
  const isOrganization = type === "organization";

  const [formData, setFormData] = useState<any>({});
  const [contacts, setContacts] = useState<Contact[]>([]);
  const [pendingContacts, setPendingContacts] = useState<PendingContact[]>([]);
  const [pendingNotes, setPendingNotes] = useState<string[]>([]);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [isDeleting, setIsDeleting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (isEditMode && account) {
      setFormData({ ...account });
      setContacts(account.contacts || []);
      setPendingContacts([]);
      setPendingNotes([]);
      setError(null);
      setIsDeleting(false);
      return;
    }
    
    setFormData(
      isOrganization
        ? { name: "", website: "", phone: "", employee_count: "", description: "" }
        : { first_name: "", last_name: "", email: "", phone: "", linkedin_url: "", title: "", notes: "" }
    );
    setContacts([]);
    setPendingContacts([]);
    setPendingNotes([]);
    setError(null);
    setIsDeleting(false);
  }, [isEditMode, account, isOrganization]);

  const handleChange = (field: string, value: string) => {
    setFormData((prev: any) => ({ ...prev, [field]: value }));
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);

    if (isOrganization && !formData.name?.trim()) {
      setError("Name is required");
      return;
    }
    
    if (!isOrganization && (!formData.first_name?.trim() || !formData.last_name?.trim())) {
      setError("First name and last name are required");
      return;
    }

    setIsSubmitting(true);
    try {
      const submitData = isOrganization
        ? {
            name: formData.name.trim(),
            website: formData.website?.trim() || null,
            phone: formData.phone?.trim() || null,
            employee_count: formData.employee_count ? Number(formData.employee_count) : null,
            description: formData.description?.trim() || null,
          }
        : {
            first_name: formData.first_name.trim(),
            last_name: formData.last_name.trim(),
            email: formData.email?.trim() || null,
            phone: formData.phone?.trim() || null,
            linkedin_url: formData.linkedin_url?.trim() || null,
            title: formData.title?.trim() || null,
            notes: formData.notes?.trim() || null,
          };

      if (isEditMode && account && onUpdate) {
        await onUpdate(account.id!, submitData);
        onClose();
        return;
      }
      
      if (!onAdd) {
        onClose();
        return;
      }
      
      const created = await onAdd(submitData as any);

      // Create pending contacts after entity is created
      const createContact = async (pending: PendingContact) => {
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
            return;
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
      };
      
      for (const pending of pendingContacts) {
        await createContact(pending);
      }

      // Create pending notes after entity is created
      const entityType = isOrganization ? "organization" : "individual";
      for (const noteContent of pendingNotes) {
        try {
          await createNote(entityType, created.id, noteContent);
        } catch (err) {
          console.error("Failed to create note:", err);
        }
      }
      
      onClose();
    } catch (err: any) {
      setError(err?.message || `Failed to ${isEditMode ? "update" : "create"}`);
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
    } catch (err: any) {
      setError(err?.message || "Failed to delete");
      setIsDeleting(false);
    }
  };

  // Contact management
  const handleAddContact = async (contact: PendingContact) => {
    if (!isEditMode) {
      // Create mode: add to pending
      setPendingContacts([...pendingContacts, contact]);
      return;
    }

    // Edit mode: create via API
    if (!account?.id) return;
    setError(null);

    try {
      if (isOrganization && !contact.individual_id) {
        setError("Please select an individual from the search to add as a contact");
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
    } catch (err: any) {
      setError(err?.message || "Failed to add contact");
    }
  };

  const handleRemoveContact = async (index: number) => {
    if (!isEditMode) {
      // Create mode: remove from pending
      setPendingContacts(pendingContacts.filter((_, i) => i !== index));
      return;
    }

    // Edit mode: delete via API
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
    } catch (err: any) {
      setError(err?.message || "Failed to remove contact");
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

  // Contact fields config - consistent across all entity types
  const contactFields: ContactFieldConfig[] = [
    { name: "first_name", label: "First Name", placeholder: "e.g., John" },
    { name: "last_name", label: "Last Name", placeholder: "e.g., Doe" },
    { name: "email", label: "Email", type: "email", placeholder: "email@example.com" },
    { name: "phone", label: "Phone", type: "tel", placeholder: "+1-555-123-4567" },
  ];

  const inputClasses =
    "w-full px-3 py-2 border border-zinc-300 dark:border-zinc-700 rounded focus:outline-none focus:ring-2 focus:ring-lime-500 bg-white dark:bg-zinc-800 text-zinc-900 dark:text-zinc-100";

  const title = isEditMode
    ? `Edit ${isOrganization ? "Organization" : "Individual"}`
    : `New ${isOrganization ? "Organization" : "Individual"}`;

  // Get contacts to display (existing in edit mode, pending in create mode)
  const displayContacts: PendingContact[] = isEditMode
    ? contacts.map((c) => ({
        first_name: (c as any).first_name || "",
        last_name: (c as any).last_name || "",
        email: c.email || "",
        phone: c.phone || "",
      }))
    : pendingContacts;

  return (
    <div className="p-6 max-w-2xl">
      <h1 className="text-2xl font-bold text-zinc-900 dark:text-zinc-100 mb-6">
        {title}
      </h1>

      {error && (
        <div className="mb-4 p-3 bg-red-100 dark:bg-red-900/30 text-red-600 dark:text-red-400 rounded">
          {error}
        </div>
      )}

      <form onSubmit={handleSubmit} className="space-y-4">
        {isOrganization ? (
          <>
            <div>
              <label className="block text-sm font-medium text-zinc-700 dark:text-zinc-300 mb-1">
                Name <span className="text-red-500">*</span>
              </label>
              <input
                type="text"
                value={formData.name || ""}
                onChange={(e) => handleChange("name", e.target.value)}
                className={inputClasses}
                required
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-zinc-700 dark:text-zinc-300 mb-1">
                Website
              </label>
              <input
                type="url"
                value={formData.website || ""}
                onChange={(e) => handleChange("website", e.target.value)}
                className={inputClasses}
                placeholder="https://example.com"
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-zinc-700 dark:text-zinc-300 mb-1">
                Phone
              </label>
              <input
                type="tel"
                value={formData.phone || ""}
                onChange={(e) => handleChange("phone", formatPhoneNumber(e.target.value))}
                className={inputClasses}
                placeholder="+1-555-123-4567"
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-zinc-700 dark:text-zinc-300 mb-1">
                Employee Count
              </label>
              <input
                type="number"
                value={formData.employee_count || ""}
                onChange={(e) => handleChange("employee_count", e.target.value)}
                className={inputClasses}
                min="0"
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-zinc-700 dark:text-zinc-300 mb-1">
                Description
              </label>
              <textarea
                value={formData.description || ""}
                onChange={(e) => handleChange("description", e.target.value)}
                className={inputClasses}
                rows={3}
              />
            </div>
          </>
        ) : (
          <>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div>
                <label className="block text-sm font-medium text-zinc-700 dark:text-zinc-300 mb-1">
                  First Name <span className="text-red-500">*</span>
                </label>
                <input
                  type="text"
                  value={formData.first_name || ""}
                  onChange={(e) => handleChange("first_name", e.target.value)}
                  className={inputClasses}
                  required
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-zinc-700 dark:text-zinc-300 mb-1">
                  Last Name <span className="text-red-500">*</span>
                </label>
                <input
                  type="text"
                  value={formData.last_name || ""}
                  onChange={(e) => handleChange("last_name", e.target.value)}
                  className={inputClasses}
                  required
                />
              </div>
            </div>

            <div>
              <label className="block text-sm font-medium text-zinc-700 dark:text-zinc-300 mb-1">
                Email
              </label>
              <input
                type="email"
                value={formData.email || ""}
                onChange={(e) => handleChange("email", e.target.value)}
                className={inputClasses}
                placeholder="john@example.com"
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-zinc-700 dark:text-zinc-300 mb-1">
                Phone
              </label>
              <input
                type="tel"
                value={formData.phone || ""}
                onChange={(e) => handleChange("phone", formatPhoneNumber(e.target.value))}
                className={inputClasses}
                placeholder="+1-555-123-4567"
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-zinc-700 dark:text-zinc-300 mb-1">
                LinkedIn URL
              </label>
              <input
                type="url"
                value={formData.linkedin_url || ""}
                onChange={(e) => handleChange("linkedin_url", e.target.value)}
                className={inputClasses}
                placeholder="https://linkedin.com/in/username"
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-zinc-700 dark:text-zinc-300 mb-1">
                Title
              </label>
              <input
                type="text"
                value={formData.title || ""}
                onChange={(e) => handleChange("title", e.target.value)}
                className={inputClasses}
                placeholder="Software Engineer"
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-zinc-700 dark:text-zinc-300 mb-1">
                Notes
              </label>
              <textarea
                value={formData.notes || ""}
                onChange={(e) => handleChange("notes", e.target.value)}
                className={inputClasses}
                rows={3}
              />
            </div>
          </>
        )}

        {/* Contacts Section */}
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

        {/* Notes Section */}
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
