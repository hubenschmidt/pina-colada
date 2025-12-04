"use client";

import { useState } from "react";
import Link from "next/link";
import { Building2, User, Plus, X } from "lucide-react";

const RelationshipsSection = ({
  relationships,
  onAdd,
  onRemove,
  searchType,
  onSearch,
  readOnly = false,
}) => {
  const [showAddForm, setShowAddForm] = useState(false);
  const [searchQuery, setSearchQuery] = useState("");
  const [searchResults, setSearchResults] = useState([]);
  const [isSearching, setIsSearching] = useState(false);
  const [showResults, setShowResults] = useState(false);
  const [debounceTimer, setDebounceTimer] = useState(null);

  const inputClasses =
    "w-full px-3 py-2 border border-zinc-300 dark:border-zinc-700 rounded focus:outline-none focus:ring-2 focus:ring-lime-500 bg-white dark:bg-zinc-800 text-zinc-900 dark:text-zinc-100";

  const handleSearch = (query) => {
    setSearchQuery(query);

    if (debounceTimer) {
      clearTimeout(debounceTimer);
    }

    if (!onSearch || query.length < 2) {
      setSearchResults([]);
      setShowResults(false);
      return;
    }

    const timer = setTimeout(async () => {
      setIsSearching(true);
      setShowResults(true);
      try {
        const results = await onSearch(query);
        // Filter out already added relationships
        const existingIds = new Set(relationships.map((r) => r.id));
        setSearchResults(results.filter((r) => !existingIds.has(r.id)));
      } catch {
        setSearchResults([]);
      } finally {
        setIsSearching(false);
      }
    }, 300);

    setDebounceTimer(timer);
  };

  const handleSelect = (result) => {
    if (!onAdd) return;

    const name = result.name || `${result.first_name || ""} ${result.last_name || ""}`.trim();

    const type = result.type || searchType || "organization";

    onAdd({
      id: result.id,
      name,
      type,
    });

    setSearchQuery("");
    setSearchResults([]);
    setShowResults(false);
    setShowAddForm(false);
  };

  const canAdd = onAdd && onSearch && !readOnly;

  if (!relationships?.length && !canAdd) {
    return null;
  }

  const searchLabel = "Search Accounts";

  return (
    <div className="space-y-2">
      <div className="flex items-center justify-between">
        <span className="text-sm font-medium text-zinc-700 dark:text-zinc-300">Relationships</span>
        {canAdd && !showAddForm && (
          <button
            type="button"
            onClick={() => setShowAddForm(true)}
            className="flex items-center gap-1 text-sm text-lime-600 dark:text-lime-400 hover:text-lime-700 dark:hover:text-lime-300">
            <Plus size={16} />
            Add Relationship
          </button>
        )}
      </div>

      {showAddForm && canAdd && (
        <div className="p-3 border border-zinc-200 dark:border-zinc-700 rounded-lg bg-zinc-50 dark:bg-zinc-800/50">
          <div className="relative">
            <label className="block text-xs font-medium text-zinc-600 dark:text-zinc-400 mb-1">
              {searchLabel}
            </label>
            <input
              type="text"
              value={searchQuery}
              onChange={(e) => handleSearch(e.target.value)}
              onFocus={() => searchResults.length > 0 && setShowResults(true)}
              onBlur={() => setTimeout(() => setShowResults(false), 200)}
              className={inputClasses}
              placeholder={`Search by name...`}
            />

            {showResults && searchResults.length > 0 && (
              <div className="absolute z-10 w-full mt-1 bg-white dark:bg-zinc-800 border border-zinc-300 dark:border-zinc-700 rounded shadow-lg max-h-48 overflow-auto">
                {searchResults.map((result) => (
                  <button
                    key={`${result.type}-${result.id}`}
                    type="button"
                    onClick={() => handleSelect(result)}
                    className="w-full text-left px-3 py-2 hover:bg-zinc-100 dark:hover:bg-zinc-700 text-zinc-900 dark:text-zinc-100 flex items-center gap-2">
                    {result.type === "organization" ? (
                      <Building2 size={14} className="text-zinc-400 flex-shrink-0" />
                    ) : (
                      <User size={14} className="text-zinc-400 flex-shrink-0" />
                    )}
                    <span className="flex-1">
                      {result.name || `${result.first_name || ""} ${result.last_name || ""}`.trim()}
                    </span>
                    <span className="text-xs text-zinc-400 capitalize">{result.type}</span>
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
          <button
            type="button"
            onClick={() => {
              setShowAddForm(false);
              setSearchQuery("");
              setSearchResults([]);
            }}
            className="mt-2 px-3 py-1 text-sm bg-zinc-200 dark:bg-zinc-700 text-zinc-700 dark:text-zinc-300 rounded hover:bg-zinc-300 dark:hover:bg-zinc-600">
            Cancel
          </button>
        </div>
      )}

      <div className="space-y-2">
        {relationships.map((rel, index) => {
          const relUrl = `/accounts/${
            rel.type === "organization" ? "organizations" : "individuals"
          }/${rel.id}`;

          return (
            <div
              key={`${rel.type}-${rel.id}`}
              className="flex items-center gap-2 p-3 border border-zinc-200 dark:border-zinc-700 rounded">
              {rel.type === "organization" ? (
                <Building2 size={16} className="text-zinc-500" />
              ) : (
                <User size={16} className="text-zinc-500" />
              )}
              <Link
                href={relUrl}
                className="text-sm text-zinc-900 dark:text-zinc-100 hover:text-zinc-500 dark:hover:text-zinc-400 hover:underline">
                {rel.name}
              </Link>
              <span className="text-xs text-zinc-500 dark:text-zinc-400 ml-auto capitalize">
                {rel.type}
              </span>
              {onRemove && !readOnly && (
                <button
                  type="button"
                  onClick={(e) => {
                    e.preventDefault();
                    e.stopPropagation();
                    onRemove(index);
                  }}
                  className="text-zinc-400 hover:text-red-500 ml-2"
                  title="Remove">
                  <X size={16} />
                </button>
              )}
            </div>
          );
        })}
      </div>

      {relationships.length === 0 && !showAddForm && (
        <p className="text-sm text-zinc-500 dark:text-zinc-400">No relationships added yet.</p>
      )}
    </div>
  );
};

export default RelationshipsSection;
