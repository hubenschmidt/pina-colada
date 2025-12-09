"use client";

import { useState, useEffect } from "react";
import { X, Search, Plus } from "lucide-react";
import {
  getTechnologies,
  getOrganizationTechnologies,
  addOrganizationTechnology,
  removeOrganizationTechnology,
} from "../../api";

const TechnologiesSection = ({ organizationId }) => {
  const [technologies, setTechnologies] = useState([]);
  const [allTechnologies, setAllTechnologies] = useState([]);
  const [isLoading, setIsLoading] = useState(false);
  const [showAddForm, setShowAddForm] = useState(false);
  const [searchQuery, setSearchQuery] = useState("");
  const [error, setError] = useState(null);

  useEffect(() => {
    if (organizationId) {
      fetchOrgTechnologies();
      fetchAllTechnologies();
    }
  }, [organizationId]);

  const fetchOrgTechnologies = async () => {
    setIsLoading(true);
    try {
      const data = await getOrganizationTechnologies(organizationId);
      setTechnologies(data.technologies || []);
    } catch (err) {
      console.error("Failed to fetch organization technologies:", err);
    } finally {
      setIsLoading(false);
    }
  };

  const fetchAllTechnologies = async () => {
    try {
      const data = await getTechnologies();
      setAllTechnologies(data || []);
    } catch (err) {
      console.error("Failed to fetch all technologies:", err);
    }
  };

  const handleToggleTechnology = async (tech) => {
    setError(null);
    const isSelected = technologies.some((t) => (t.technology_id || t.id) === tech.id);

    if (isSelected) {
      // Remove
      try {
        await removeOrganizationTechnology(organizationId, tech.id);
        setTechnologies(technologies.filter((t) => (t.technology_id || t.id) !== tech.id));
      } catch (err) {
        setError(err?.message || "Failed to remove technology");
      }
    } else {
      // Add
      try {
        await addOrganizationTechnology(organizationId, { technology_id: tech.id });
        setTechnologies([...technologies, { ...tech, technology_id: tech.id, technology: tech }]);
      } catch (err) {
        setError(err?.message || "Failed to add technology");
      }
    }
  };

  const handleRemoveTechnology = async (techId) => {
    setError(null);
    try {
      await removeOrganizationTechnology(organizationId, techId);
      setTechnologies(technologies.filter((t) => (t.technology_id || t.id) !== techId));
    } catch (err) {
      setError(err?.message || "Failed to remove technology");
    }
  };

  const handleDone = () => {
    setShowAddForm(false);
    setSearchQuery("");
  };

  // Get the display name for a technology (handles both direct and nested structures)
  const getTechName = (tech) => {
    return tech.technology?.name || tech.name || "Unknown";
  };

  const selectedTechIds = new Set(technologies.map((t) => t.technology_id || t.id));

  // Filter and sort technologies alphabetically
  const filteredTechnologies = allTechnologies
    .filter((tech) => {
      if (!searchQuery) return true;
      const query = searchQuery.toLowerCase();
      return (
        tech.name.toLowerCase().includes(query) ||
        (tech.category && tech.category.toLowerCase().includes(query))
      );
    })
    .sort((a, b) => a.name.localeCompare(b.name));

  const inputClasses =
    "w-full px-3 py-2 border border-zinc-300 dark:border-zinc-700 rounded focus:outline-none focus:ring-2 focus:ring-lime-500 bg-white dark:bg-zinc-800 text-zinc-900 dark:text-zinc-100";

  return (
    <div className="space-y-3">
      {/* Header */}
      <div className="flex items-center justify-between">
        <span className="text-sm font-medium text-zinc-700 dark:text-zinc-300">Technologies</span>
        {!showAddForm && (
          <button
            type="button"
            onClick={() => setShowAddForm(true)}
            className="flex items-center gap-1 text-sm text-lime-600 dark:text-lime-400 hover:text-lime-700 dark:hover:text-lime-300">
            <Plus size={16} />
            Add Technology
          </button>
        )}
      </div>

      {/* Error */}
      {error && (
        <div className="p-2 text-sm text-red-600 dark:text-red-400 bg-red-100 dark:bg-red-900/30 rounded">
          {error}
        </div>
      )}

      {/* Selected technologies as tags */}
      {technologies.length > 0 && (
        <div className="flex flex-wrap gap-2">
          {technologies.map((tech) => (
            <span
              key={tech.technology_id || tech.id}
              className="inline-flex items-center gap-1 px-2 py-1 text-sm bg-zinc-100 dark:bg-zinc-700 text-zinc-700 dark:text-zinc-300 rounded">
              {getTechName(tech)}
              <button
                type="button"
                onClick={() => handleRemoveTechnology(tech.technology_id || tech.id)}
                className="p-0.5 hover:text-red-500"
                title="Remove">
                <X size={14} />
              </button>
            </span>
          ))}
        </div>
      )}

      {/* Add form */}
      {showAddForm && (
        <div className="p-3 border border-zinc-200 dark:border-zinc-700 rounded-lg bg-zinc-50 dark:bg-zinc-800/50 space-y-3">
          {/* Search box */}
          <div className="relative">
            <Search
              size={16}
              className="absolute left-3 top-1/2 -translate-y-1/2 text-zinc-400"
            />
            <input
              type="text"
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className={`${inputClasses} pl-9`}
              placeholder="Search technologies..."
              autoFocus
            />
          </div>

          {/* Loading */}
          {isLoading && <p className="text-sm text-zinc-500">Loading technologies...</p>}

          {/* Multi-select picklist */}
          {!isLoading && (
            <div className="border border-zinc-200 dark:border-zinc-700 rounded max-h-48 overflow-y-auto bg-white dark:bg-zinc-800">
              {filteredTechnologies.length === 0 ? (
                <div className="px-3 py-2 text-sm text-zinc-500 dark:text-zinc-400">
                  No technologies found
                </div>
              ) : (
                filteredTechnologies.map((tech) => {
                  const isSelected = selectedTechIds.has(tech.id);
                  return (
                    <label
                      key={tech.id}
                      className="flex items-center px-3 py-2 hover:bg-zinc-50 dark:hover:bg-zinc-700 cursor-pointer">
                      <input
                        type="checkbox"
                        checked={isSelected}
                        onChange={() => handleToggleTechnology(tech)}
                        className="w-4 h-4 text-lime-500 border-zinc-300 dark:border-zinc-600 rounded focus:ring-lime-500 bg-white dark:bg-zinc-700"
                      />
                      <span className="ml-2 text-sm text-zinc-900 dark:text-zinc-100">
                        {tech.name}
                      </span>
                      {tech.category && (
                        <span className="ml-3 text-xs text-zinc-500 dark:text-zinc-400">
                          {tech.category}
                        </span>
                      )}
                    </label>
                  );
                })
              )}
            </div>
          )}

          {/* Done button */}
          <button
            type="button"
            onClick={handleDone}
            className="px-3 py-1 text-sm bg-lime-600 text-white rounded hover:bg-lime-700">
            Done
          </button>
        </div>
      )}

      {/* Empty state */}
      {!showAddForm && technologies.length === 0 && (
        <p className="text-sm text-zinc-500 dark:text-zinc-400">No technologies added yet.</p>
      )}
    </div>
  );
};

export default TechnologiesSection;
