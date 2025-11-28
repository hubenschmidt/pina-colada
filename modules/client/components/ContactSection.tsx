"use client";

import { useState } from "react";
import { Plus, Trash2, Pencil, Star } from "lucide-react";
import { formatPhoneNumber } from "../lib/phone";

export interface ContactFieldConfig {
  name: string;
  label: string;
  type?: "text" | "email" | "tel";
  placeholder?: string;
  colSpan?: 1 | 2;
}

export interface ContactDisplayConfig {
  nameFields?: string[]; // Fields to combine for display name
  primaryLabel?: string;
}

export interface SearchResult {
  individual_id: number;
  first_name: string;
  last_name: string;
  email?: string | null;
  phone?: string | null;
}

interface ContactSectionProps<T extends Record<string, any>> {
  contacts: T[];
  onAdd: (contact: T) => void;
  onRemove: (index: number) => void;
  onUpdate?: (index: number, contact: T) => void;
  onSetPrimary?: (index: number) => void;
  fields: ContactFieldConfig[];
  emptyContact: () => T;
  display?: ContactDisplayConfig;
  // For locked contacts (e.g., primary contact auto-filled from form)
  isContactLocked?: (contact: T, index: number) => boolean;
  // Individual search - searches by first/last name and auto-populates fields
  individualSearch?: {
    enabled: boolean;
    onSearch: (query: string) => Promise<SearchResult[]>;
  };
  showFormByDefault?: boolean;
}

const ContactSection = <T extends Record<string, any>>({
  contacts,
  onAdd,
  onRemove,
  onUpdate,
  onSetPrimary,
  fields,
  emptyContact,
  display = {},
  isContactLocked,
  individualSearch,
  showFormByDefault = false,
}: ContactSectionProps<T>) => {
  const [showAddForm, setShowAddForm] = useState(showFormByDefault);
  const [newContact, setNewContact] = useState<T>(emptyContact());
  const [searchQuery, setSearchQuery] = useState("");
  const [searchResults, setSearchResults] = useState<SearchResult[]>([]);
  const [isSearching, setIsSearching] = useState(false);
  const [showResults, setShowResults] = useState(false);
  const [debounceTimer, setDebounceTimer] = useState<NodeJS.Timeout | null>(null);
  const [editingIndex, setEditingIndex] = useState<number | null>(null);
  const [editingContact, setEditingContact] = useState<T | null>(null);

  const inputClasses =
    "w-full px-3 py-2 border border-zinc-300 dark:border-zinc-700 rounded focus:outline-none focus:ring-2 focus:ring-lime-500 bg-white dark:bg-zinc-800 text-zinc-900 dark:text-zinc-100";

  const handleSearch = (query: string) => {
    setSearchQuery(query);

    // Clear previous debounce timer
    if (debounceTimer) {
      clearTimeout(debounceTimer);
    }

    if (!individualSearch || query.length < 2) {
      setSearchResults([]);
      setShowResults(false);
      return;
    }

    // Debounce the API call by 300ms
    const timer = setTimeout(async () => {
      setIsSearching(true);
      setShowResults(true);
      try {
        const results = await individualSearch.onSearch(query);
        setSearchResults(results);
      } catch {
        setSearchResults([]);
      } finally {
        setIsSearching(false);
      }
    }, 300);

    setDebounceTimer(timer);
  };

  const handleSelectIndividual = (individual: SearchResult) => {
    // Auto-populate the contact fields from the selected individual, including individual_id for linking
    setNewContact({
      ...newContact,
      individual_id: individual.individual_id,
      first_name: individual.first_name,
      last_name: individual.last_name,
      email: individual.email || "",
      phone: individual.phone || "",
    } as T);
    setSearchQuery("");
    setSearchResults([]);
    setShowResults(false);
  };

  const handleAdd = () => {
    onAdd({ ...newContact });
    setNewContact(emptyContact());
    setShowAddForm(false);
    setSearchQuery("");
    setSearchResults([]);
  };

  const handleCancel = () => {
    setShowAddForm(false);
    setNewContact(emptyContact());
    setSearchQuery("");
    setSearchResults([]);
    setShowResults(false);
  };

  const handleFieldChange = (fieldName: string, value: string, fieldType?: string) => {
    const processedValue = fieldType === "tel" ? formatPhoneNumber(value) : value;
    setNewContact({ ...newContact, [fieldName]: processedValue });
  };

  const getContactDisplayName = (contact: T): string => {
    if (display.nameFields && display.nameFields.length > 0) {
      return display.nameFields
        .map((f) => contact[f])
        .filter(Boolean)
        .join(" ");
    }
    // Fallback: try common name fields
    if (contact.first_name || contact.last_name) {
      return `${contact.first_name || ""} ${contact.last_name || ""}`.trim();
    }
    return "";
  };

  const handleStartEdit = (contact: T, index: number) => {
    setEditingIndex(index);
    setEditingContact({ ...contact });
  };

  const handleCancelEdit = () => {
    setEditingIndex(null);
    setEditingContact(null);
  };

  const handleSaveEdit = () => {
    if (editingIndex !== null && editingContact && onUpdate) {
      onUpdate(editingIndex, editingContact);
    }
    setEditingIndex(null);
    setEditingContact(null);
  };

  const handleEditFieldChange = (fieldName: string, value: string, fieldType?: string) => {
    if (!editingContact) return;
    const processedValue = fieldType === "tel" ? formatPhoneNumber(value) : value;
    setEditingContact({ ...editingContact, [fieldName]: processedValue });
  };

  const renderContactCard = (contact: T, index: number) => {
    const isPrimary = contact.is_primary === true || (index === 0 && !contacts.some(c => c.is_primary));
    const isLocked = isContactLocked?.(contact, index) ?? false;
    const displayName = getContactDisplayName(contact);
    const isEditing = editingIndex === index;

    if (isEditing && editingContact) {
      return (
        <div
          key={index}
          className="p-3 border border-lime-300 dark:border-lime-700 rounded-lg bg-zinc-50 dark:bg-zinc-800/50"
        >
          <div className="grid grid-cols-2 gap-3">
            {fields.map((field) => (
              <div key={field.name} className={field.colSpan === 2 ? "col-span-2" : ""}>
                <label className="block text-xs font-medium text-zinc-600 dark:text-zinc-400 mb-1">
                  {field.label}
                </label>
                <input
                  type={field.type || "text"}
                  value={editingContact[field.name] || ""}
                  onChange={(e) => handleEditFieldChange(field.name, e.target.value, field.type)}
                  className={inputClasses}
                  placeholder={field.placeholder}
                />
              </div>
            ))}
          </div>
          <div className="flex gap-2 mt-3">
            <button
              type="button"
              onClick={handleSaveEdit}
              className="px-3 py-1 text-sm bg-lime-600 text-white rounded hover:bg-lime-700"
            >
              Save
            </button>
            <button
              type="button"
              onClick={handleCancelEdit}
              className="px-3 py-1 text-sm bg-zinc-200 dark:bg-zinc-700 text-zinc-700 dark:text-zinc-300 rounded hover:bg-zinc-300 dark:hover:bg-zinc-600"
            >
              Cancel
            </button>
          </div>
        </div>
      );
    }

    return (
      <div
        key={index}
        className="flex items-center justify-between p-3 border border-zinc-200 dark:border-zinc-700 rounded"
      >
        <div>
          <div className="flex items-center gap-2">
            {displayName && (
              <span className="text-sm font-medium text-zinc-900 dark:text-zinc-100">
                {displayName}
              </span>
            )}
            {contact.title && (
              <span className="text-sm text-zinc-500 dark:text-zinc-400">
                — {contact.title}
              </span>
            )}
            {isPrimary && (
              <span className="text-xs bg-lime-100 dark:bg-lime-900/30 text-lime-700 dark:text-lime-400 px-2 py-0.5 rounded">
                {display.primaryLabel || "Primary"}
              </span>
            )}
          </div>
          <div className="text-sm text-zinc-500 dark:text-zinc-500 mt-1">
            {contact.email && <span className="mr-3">{contact.email}</span>}
            {contact.phone && <span>{contact.phone}</span>}
          </div>
        </div>
        <div className="flex items-center gap-2">
          {!isPrimary && onSetPrimary && !isLocked && (
            <button
              type="button"
              onClick={() => onSetPrimary(index)}
              className="text-zinc-400 hover:text-yellow-500"
              title="Set as primary"
            >
              <Star size={16} />
            </button>
          )}
          {onUpdate && !isLocked && (
            <button
              type="button"
              onClick={() => handleStartEdit(contact, index)}
              className="text-zinc-400 hover:text-lime-500"
              title="Edit contact"
            >
              <Pencil size={16} />
            </button>
          )}
          {!isLocked && (
            <button
              type="button"
              onClick={() => onRemove(index)}
              className="text-zinc-400 hover:text-red-500"
              title="Remove contact"
            >
              <Trash2 size={16} />
            </button>
          )}
        </div>
      </div>
    );
  };

  const renderAddForm = () => (
    <div className="mb-4 p-4 border border-zinc-200 dark:border-zinc-700 rounded-lg bg-zinc-50 dark:bg-zinc-800/50">
      {/* Individual search */}
      {individualSearch?.enabled && (
        <div className="mb-3 relative">
          <label className="block text-xs font-medium text-zinc-600 dark:text-zinc-400 mb-1">
            Search Existing Contacts
          </label>
          <input
            type="text"
            value={searchQuery}
            onChange={(e) => handleSearch(e.target.value)}
            onFocus={() => searchResults.length > 0 && setShowResults(true)}
            onBlur={() => setTimeout(() => setShowResults(false), 200)}
            className={inputClasses}
            placeholder="Search by first or last name..."
          />
          {showResults && searchResults.length > 0 && (
            <div className="absolute z-10 w-full mt-1 bg-white dark:bg-zinc-800 border border-zinc-300 dark:border-zinc-700 rounded shadow-lg max-h-48 overflow-auto">
              {searchResults.map((ind) => (
                <button
                  key={ind.individual_id}
                  type="button"
                  onClick={() => handleSelectIndividual(ind)}
                  className="w-full text-left px-3 py-2 hover:bg-zinc-100 dark:hover:bg-zinc-700 text-zinc-900 dark:text-zinc-100"
                >
                  {ind.first_name} {ind.last_name}
                  {ind.email && <span className="text-zinc-500 text-sm ml-2">{ind.email}</span>}
                </button>
              ))}
            </div>
          )}
          {isSearching && (
            <div className="absolute z-10 w-full mt-1 bg-white dark:bg-zinc-800 border border-zinc-300 dark:border-zinc-700 rounded p-2 text-zinc-500">
              Searching...
            </div>
          )}
        </div>
      )}

      {/* Contact fields - always visible */}
      <div className="grid grid-cols-2 gap-3">
        {fields.map((field) => (
          <div key={field.name} className={field.colSpan === 2 ? "col-span-2" : ""}>
            <label className="block text-xs font-medium text-zinc-600 dark:text-zinc-400 mb-1">
              {field.label}
            </label>
            <input
              type={field.type || "text"}
              value={newContact[field.name] || ""}
              onChange={(e) => handleFieldChange(field.name, e.target.value, field.type)}
              className={inputClasses}
              placeholder={field.placeholder}
            />
          </div>
        ))}
      </div>

      <div className="flex gap-2 mt-3">
        <button
          type="button"
          onClick={handleAdd}
          className="px-3 py-1 text-sm bg-lime-600 text-white rounded hover:bg-lime-700"
        >
          Add
        </button>
        <button
          type="button"
          onClick={handleCancel}
          className="px-3 py-1 text-sm bg-zinc-200 dark:bg-zinc-700 text-zinc-700 dark:text-zinc-300 rounded hover:bg-zinc-300 dark:hover:bg-zinc-600"
        >
          Cancel
        </button>
      </div>
    </div>
  );

  return (
    <div className="space-y-2">
      {/* Header */}
      <div className="flex items-center justify-between mb-3">
        <span className="text-sm font-medium text-zinc-700 dark:text-zinc-300">Contacts</span>
        {!showAddForm && (
          <button
            type="button"
            onClick={() => setShowAddForm(true)}
            className="flex items-center gap-1 text-sm text-lime-600 dark:text-lime-400 hover:text-lime-700 dark:hover:text-lime-300"
          >
            <Plus size={16} />
            Add Contact
          </button>
        )}
      </div>

      {/* Add Form */}
      {showAddForm && renderAddForm()}

      {/* Contacts List */}
      {contacts.map((contact, index) => renderContactCard(contact, index))}

      {contacts.length === 0 && !showAddForm && (
        <p className="text-sm text-zinc-500 dark:text-zinc-400">No contacts added yet.</p>
      )}
    </div>
  );
};

export default ContactSection;
