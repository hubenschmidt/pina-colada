"use client";

import { useState, useEffect, useId } from "react";
import { useLookupsContext } from "../../context/lookupsContext";
import { getIndustries, createIndustry } from "../../api";
import { fetchOnce } from "../../lib/lookup-cache";

const IndustrySelector = ({ value, onChange }) => {
  const { lookupsState, dispatchLookups } = useLookupsContext();
  const { industries, loaded } = lookupsState;

  useEffect(() => {
    if (loaded.industries) return;
    fetchOnce("industries", getIndustries).then((data) => {
      dispatchLookups({ type: "SET_INDUSTRIES", payload: data });
    });
  }, [loaded.industries, dispatchLookups]);

  const [selectedIndustries, setSelectedIndustries] = useState(value || []);
  const [isOpen, setIsOpen] = useState(false);
  const [showNewInput, setShowNewInput] = useState(false);
  const [newIndustry, setNewIndustry] = useState("");
  const dropdownId = useId();

  useEffect(() => {
    setSelectedIndustries(Array.isArray(value) ? value : value ? [value] : []);
  }, [value]);

  useEffect(() => {
    const handleClickOutside = (event) => {
      const dropdown = document.getElementById(dropdownId);
      if (dropdown && !dropdown.contains(event.target)) {
        setIsOpen(false);
        setShowNewInput(false);
      }
    };
    document.addEventListener("mousedown", handleClickOutside);
    return () => document.removeEventListener("mousedown", handleClickOutside);
  }, [dropdownId]);

  const handleToggleIndustry = (industryName) => {
    const updated = selectedIndustries.includes(industryName)
      ? selectedIndustries.filter((i) => i !== industryName)
      : [...selectedIndustries, industryName];
    setSelectedIndustries(updated);
    onChange(updated);
  };

  const handleAddNew = async () => {
    if (!newIndustry.trim()) return;
    try {
      const created = await createIndustry(newIndustry.trim());
      dispatchLookups({ type: "ADD_INDUSTRY", payload: created });
      const updated = [...selectedIndustries, created.name];
      setSelectedIndustries(updated);
      onChange(updated);
      setShowNewInput(false);
      setNewIndustry("");
    } catch (error) {
      console.error("Failed to create industry:", error);
    }
  };

  if (!loaded.industries) {
    return (
      <div className="w-full px-3 py-2 border border-zinc-300 dark:border-zinc-700 rounded bg-zinc-100 dark:bg-zinc-800 text-zinc-500 dark:text-zinc-400">
        Loading industries...
      </div>
    );
  }

  const displayText =
    selectedIndustries.length > 0
      ? selectedIndustries.join(", ")
      : "Select industries...";

  return (
    <div className="relative" id={dropdownId}>
      <button
        type="button"
        onClick={() => setIsOpen(!isOpen)}
        className="w-full px-3 py-2 border border-zinc-300 dark:border-zinc-700 rounded focus:outline-none focus:ring-2 focus:ring-lime-500 bg-white dark:bg-zinc-800 text-zinc-900 dark:text-zinc-100 text-left flex justify-between items-center"
      >
        <span
          className={
            selectedIndustries.length === 0
              ? "text-zinc-500 dark:text-zinc-400"
              : ""
          }
        >
          {displayText}
        </span>
        <svg
          className={`w-4 h-4 transition-transform ${isOpen ? "rotate-180" : ""}`}
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth={2}
            d="M19 9l-7 7-7-7"
          />
        </svg>
      </button>
      {isOpen && (
        <div className="absolute z-50 w-full mt-1 bg-white dark:bg-zinc-800 border border-zinc-300 dark:border-zinc-700 rounded shadow-lg max-h-60 overflow-y-auto">
          {industries.map((industry) => {
            const isSelected = selectedIndustries.includes(industry.name);
            return (
              <label
                key={industry.id}
                className="flex items-center px-3 py-2 hover:bg-zinc-100 dark:hover:bg-zinc-700 cursor-pointer"
              >
                <input
                  type="checkbox"
                  checked={isSelected}
                  onChange={() => handleToggleIndustry(industry.name)}
                  className="w-4 h-4 text-lime-500 border-zinc-300 dark:border-zinc-600 rounded focus:ring-lime-500 bg-white dark:bg-zinc-700"
                />

                <span className="ml-2 text-zinc-900 dark:text-zinc-100">
                  {industry.name}
                </span>
              </label>
            );
          })}
          <div className="border-t border-zinc-200 dark:border-zinc-700">
            {!showNewInput ? (
              <button
                type="button"
                onClick={() => setShowNewInput(true)}
                className="w-full px-3 py-2 text-left text-zinc-600 dark:text-zinc-400 hover:bg-zinc-100 dark:hover:bg-zinc-700"
              >
                + Add New Industry
              </button>
            ) : (
              <div className="p-2 flex gap-2">
                <input
                  type="text"
                  value={newIndustry}
                  onChange={(e) => setNewIndustry(e.target.value)}
                  onKeyDown={(e) => {
                    if (e.key === "Enter") {
                      e.preventDefault();
                      handleAddNew();
                      return;
                    }
                    if (e.key === "Escape") {
                      setShowNewInput(false);
                      setNewIndustry("");
                    }
                  }}
                  placeholder="New industry..."
                  className="flex-1 px-2 py-1 text-sm border border-zinc-300 dark:border-zinc-600 rounded bg-white dark:bg-zinc-700 text-zinc-900 dark:text-zinc-100"
                  autoFocus
                />

                <button
                  type="button"
                  onClick={handleAddNew}
                  className="px-2 py-1 text-sm bg-lime-500 text-white rounded hover:bg-lime-600"
                >
                  Add
                </button>
              </div>
            )}
          </div>
        </div>
      )}
    </div>
  );
};

export default IndustrySelector;
