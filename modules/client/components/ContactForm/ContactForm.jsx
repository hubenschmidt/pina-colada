"use client";

import { useState, useEffect } from "react";
import { createNote, searchAccounts } from "../../api";
import { formatPhoneNumber } from "../../lib/phone";
import FormActions from "../FormActions/FormActions";
import NotesSection from "../NotesSection/NotesSection";
import Timestamps from "../Timestamps/Timestamps";
import { usePendingChanges } from "../../hooks/usePendingChanges";
import { Building2, User, Plus, X } from "lucide-react";

const ContactForm = ({ contact, onSave, onDelete, onClose }) => {
  const isEditMode = !!contact;

  // Build initial linked accounts from contact's accounts
  const getInitialLinkedAccounts = () => {
    if (!contact?.accounts) return [];
    return contact.accounts.map((acc) => ({
      id: acc.id,
      name: acc.name,
      type: acc.type || "organization",
    }));
  };

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

  const [linkedAccounts, setLinkedAccounts] = useState(getInitialLinkedAccounts);

  const [isSubmitting, setIsSubmitting] = useState(false);
  const [isDeleting, setIsDeleting] = useState(false);
  const [error, setError] = useState(null);
  const [pendingNotes, setPendingNotes] = useState([]);
  const [showAccountSearch, setShowAccountSearch] = useState(false);
  const [accountSearchQuery, setAccountSearchQuery] = useState("");
  const [accountSearchResults, setAccountSearchResults] = useState([]);
  const [isSearchingAccounts, setIsSearchingAccounts] = useState(false);

  // Debounced account search
  useEffect(() => {
    if (accountSearchQuery.length < 2) {
      setAccountSearchResults([]);
      return;
    }

    setIsSearchingAccounts(true);
    const timer = setTimeout(async () => {
      try {
        const results = await searchAccounts(accountSearchQuery);
        // Filter out already linked accounts by account_id
        const linkedIds = new Set(linkedAccounts.map((a) => a.id));
        setAccountSearchResults(results.filter((r) => !linkedIds.has(r.account_id)));
      } catch {
        setAccountSearchResults([]);
      } finally {
        setIsSearchingAccounts(false);
      }
    }, 300);

    return () => clearTimeout(timer);
  }, [accountSearchQuery, linkedAccounts]);

  const hasPendingChanges = usePendingChanges({
    original: contact,
    current: formData,
  });

  const inputClasses =
    "w-full px-3 py-2 border border-zinc-300 dark:border-zinc-700 rounded focus:outline-none focus:ring-2 focus:ring-lime-500 bg-white dark:bg-zinc-800 text-zinc-900 dark:text-zinc-100";

  const handleChange = (field, value) => {
    if (field === "phone") {
      value = formatPhoneNumber(value);
    }
    setFormData((prev) => ({ ...prev, [field]: value }));
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    setIsSubmitting(true);
    setError(null);
    try {
      // Build save data with linked account IDs
      const saveData = {
        ...formData,
        account_ids: linkedAccounts.map((a) => a.id),
      };
      const savedContact = await onSave(saveData);
      // Create pending notes after contact is created
      if (!isEditMode && savedContact && pendingNotes.length > 0) {
        for (const noteContent of pendingNotes) {
          await createNote("contact", savedContact.id, noteContent);
        }
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to save contact");
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleAddAccount = (account) => {
    setLinkedAccounts([...linkedAccounts, {
      id: account.account_id,
      name: account.name,
      type: account.type,
    }]);
    setAccountSearchQuery("");
    setAccountSearchResults([]);
    setShowAccountSearch(false);
  };

  const handleRemoveAccount = (index) => {
    setLinkedAccounts(linkedAccounts.filter((_, i) => i !== index));
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

        {/* Linked Accounts Section */}
        <div>
          <div className="flex items-center justify-between mb-2">
            <label className="block text-sm font-medium text-zinc-700 dark:text-zinc-300">
              Linked Accounts
            </label>
            {!showAccountSearch && (
              <button
                type="button"
                onClick={() => setShowAccountSearch(true)}
                className="flex items-center gap-1 text-sm text-lime-600 dark:text-lime-400 hover:text-lime-700 dark:hover:text-lime-300">
                <Plus size={16} />
                Link Account
              </button>
            )}
          </div>

          {showAccountSearch && (
            <div className="mb-3 p-3 border border-zinc-200 dark:border-zinc-700 rounded-lg bg-zinc-50 dark:bg-zinc-800/50">
              <div className="relative">
                <input
                  type="text"
                  value={accountSearchQuery}
                  onChange={(e) => setAccountSearchQuery(e.target.value)}
                  className={inputClasses}
                  placeholder="Search accounts by name..."
                  autoFocus
                />
                {accountSearchResults.length > 0 && (
                  <div className="absolute z-10 w-full mt-1 bg-white dark:bg-zinc-800 border border-zinc-300 dark:border-zinc-700 rounded shadow-lg max-h-48 overflow-auto">
                    {accountSearchResults.map((result) => (
                      <button
                        key={`${result.type}-${result.account_id}`}
                        type="button"
                        onClick={() => handleAddAccount(result)}
                        className="w-full text-left px-3 py-2 hover:bg-zinc-100 dark:hover:bg-zinc-700 text-zinc-900 dark:text-zinc-100 flex items-center gap-2">
                        {result.type === "organization" ? (
                          <Building2 size={14} className="text-zinc-400 flex-shrink-0" />
                        ) : (
                          <User size={14} className="text-zinc-400 flex-shrink-0" />
                        )}
                        <span className="flex-1">{result.name}</span>
                        <span className="text-xs text-zinc-400 capitalize">{result.type}</span>
                      </button>
                    ))}
                  </div>
                )}
                {isSearchingAccounts && (
                  <div className="absolute z-10 w-full mt-1 bg-white dark:bg-zinc-800 border border-zinc-300 dark:border-zinc-700 rounded p-2 text-zinc-500">
                    Searching...
                  </div>
                )}
                {!isSearchingAccounts && accountSearchQuery.length >= 2 && accountSearchResults.length === 0 && (
                  <div className="absolute z-10 w-full mt-1 bg-white dark:bg-zinc-800 border border-zinc-300 dark:border-zinc-700 rounded p-2 text-zinc-500">
                    No accounts found
                  </div>
                )}
              </div>
              <button
                type="button"
                onClick={() => {
                  setShowAccountSearch(false);
                  setAccountSearchQuery("");
                  setAccountSearchResults([]);
                }}
                className="mt-2 px-3 py-1 text-sm bg-zinc-200 dark:bg-zinc-700 text-zinc-700 dark:text-zinc-300 rounded hover:bg-zinc-300 dark:hover:bg-zinc-600">
                Cancel
              </button>
            </div>
          )}

          {linkedAccounts.length > 0 ? (
            <div className="space-y-2">
              {linkedAccounts.map((account, index) => (
                <div
                  key={`${account.type}-${account.id}`}
                  className="flex items-center gap-2 p-2 border border-zinc-200 dark:border-zinc-700 rounded">
                  {account.type === "organization" ? (
                    <Building2 size={14} className="text-zinc-500" />
                  ) : (
                    <User size={14} className="text-zinc-500" />
                  )}
                  <span className="text-sm text-zinc-900 dark:text-zinc-100 flex-1">{account.name}</span>
                  <span className="text-xs text-zinc-400 capitalize">{account.type}</span>
                  <button
                    type="button"
                    onClick={() => handleRemoveAccount(index)}
                    className="text-zinc-400 hover:text-red-500"
                    title="Remove">
                    <X size={16} />
                  </button>
                </div>
              ))}
            </div>
          ) : (
            <p className="text-sm text-zinc-500 dark:text-zinc-400">No accounts linked yet.</p>
          )}
        </div>

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
            Description
          </label>
          <textarea
            value={formData.notes}
            onChange={(e) => handleChange("notes", e.target.value)}
            className={`${inputClasses} min-h-[100px]`}
            placeholder="Description..."
          />
        </div>

        <div className="border-t border-zinc-300 dark:border-zinc-700 pt-4 mt-4">
          <NotesSection
            entityType="contact"
            entityId={isEditMode ? (contact?.id ?? null) : null}
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

        {isEditMode && contact && (
          <Timestamps createdAt={contact.created_at} updatedAt={contact.updated_at} />
        )}

        {error && (
          <div className="mt-4 p-3 bg-red-100 dark:bg-red-900/30 text-red-600 dark:text-red-400 rounded">
            {error}
          </div>
        )}
      </form>
    </div>
  );
};

export default ContactForm;
