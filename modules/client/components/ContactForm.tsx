"use client";

import { useState } from "react";
import { Contact } from "../api";
import { formatPhoneNumber } from "../lib/phone";
import FormActions from "./FormActions";

interface ContactFormProps {
  contact?: Contact;
  onSave: (data: Partial<Contact>) => Promise<void>;
  onDelete?: () => Promise<void>;
  onClose: () => void;
}

const ContactForm = ({ contact, onSave, onDelete, onClose }: ContactFormProps) => {
  const isEditMode = !!contact;

  const [formData, setFormData] = useState({
    first_name: contact?.first_name || "",
    last_name: contact?.last_name || "",
    email: contact?.email || "",
    phone: contact?.phone || "",
    title: contact?.title || "",
    department: contact?.department || "",
    role: contact?.role || "",
    notes: contact?.notes || "",
  });

  const [isSubmitting, setIsSubmitting] = useState(false);
  const [isDeleting, setIsDeleting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const inputClasses =
    "w-full px-3 py-2 border border-zinc-300 dark:border-zinc-700 rounded focus:outline-none focus:ring-2 focus:ring-lime-500 bg-white dark:bg-zinc-800 text-zinc-900 dark:text-zinc-100";

  const handleChange = (field: string, value: string) => {
    if (field === "phone") {
      value = formatPhoneNumber(value);
    }
    setFormData((prev) => ({ ...prev, [field]: value }));
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setIsSubmitting(true);
    setError(null);
    try {
      await onSave(formData);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to save contact");
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleDelete = async () => {
    if (!onDelete) return;

    if (!isDeleting) {
      setIsDeleting(true);
      return;
    }

    try {
      await onDelete();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to delete contact");
      setIsDeleting(false);
    }
  };

  return (
    <div>
      <h1 className="text-xl font-bold text-zinc-900 dark:text-zinc-100 mb-6">
        {isEditMode ? "Edit Contact" : "New Contact"}
      </h1>

      {error && (
        <div className="mb-4 p-3 bg-red-100 dark:bg-red-900/30 text-red-700 dark:text-red-400 rounded">
          {error}
        </div>
      )}

      <form onSubmit={handleSubmit} className="space-y-4">
        <div className="grid grid-cols-2 gap-4">
          <div>
            <label className="block text-sm font-medium text-zinc-700 dark:text-zinc-300 mb-1">
              First Name
            </label>
            <input
              type="text"
              value={formData.first_name}
              onChange={(e) => handleChange("first_name", e.target.value)}
              className={inputClasses}
              placeholder="First name"
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-zinc-700 dark:text-zinc-300 mb-1">
              Last Name
            </label>
            <input
              type="text"
              value={formData.last_name}
              onChange={(e) => handleChange("last_name", e.target.value)}
              className={inputClasses}
              placeholder="Last name"
            />
          </div>
        </div>

        {isEditMode && contact?.organizations && contact.organizations.length > 0 && (
          <div>
            <label className="block text-sm font-medium text-zinc-700 dark:text-zinc-300 mb-1">
              Account
            </label>
            <div className="w-full px-3 py-2 border border-zinc-200 dark:border-zinc-700 rounded bg-zinc-100 dark:bg-zinc-800/50 text-zinc-700 dark:text-zinc-300">
              {contact.organizations.map((org) => org.name).join(", ")}
            </div>
          </div>
        )}

        <div className="grid grid-cols-2 gap-4">
          <div>
            <label className="block text-sm font-medium text-zinc-700 dark:text-zinc-300 mb-1">
              Email
            </label>
            <input
              type="email"
              value={formData.email}
              onChange={(e) => handleChange("email", e.target.value)}
              className={inputClasses}
              placeholder="email@example.com"
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-zinc-700 dark:text-zinc-300 mb-1">
              Phone
            </label>
            <input
              type="tel"
              value={formData.phone}
              onChange={(e) => handleChange("phone", e.target.value)}
              className={inputClasses}
              placeholder="(555) 555-5555"
            />
          </div>
        </div>

        <div className="grid grid-cols-3 gap-4">
          <div>
            <label className="block text-sm font-medium text-zinc-700 dark:text-zinc-300 mb-1">
              Title
            </label>
            <input
              type="text"
              value={formData.title}
              onChange={(e) => handleChange("title", e.target.value)}
              className={inputClasses}
              placeholder="Job title"
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-zinc-700 dark:text-zinc-300 mb-1">
              Department
            </label>
            <input
              type="text"
              value={formData.department}
              onChange={(e) => handleChange("department", e.target.value)}
              className={inputClasses}
              placeholder="Department"
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-zinc-700 dark:text-zinc-300 mb-1">
              Role
            </label>
            <input
              type="text"
              value={formData.role}
              onChange={(e) => handleChange("role", e.target.value)}
              className={inputClasses}
              placeholder="Role"
            />
          </div>
        </div>

        <div>
          <label className="block text-sm font-medium text-zinc-700 dark:text-zinc-300 mb-1">
            Notes
          </label>
          <textarea
            value={formData.notes}
            onChange={(e) => handleChange("notes", e.target.value)}
            className={`${inputClasses} min-h-[100px]`}
            placeholder="Additional notes..."
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

export default ContactForm;
