"use client";

import { useState, useEffect, useId } from "react";

interface Project {
  id: number;
  name: string;
}

interface ProjectSelectorProps {
  value: number[];
  onChange: (value: number[]) => void;
  projects: Project[];
}

const ProjectSelector = ({ value, onChange, projects }: ProjectSelectorProps) => {
  const [isOpen, setIsOpen] = useState(false);
  const dropdownId = useId();

  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      const dropdown = document.getElementById(dropdownId);
      if (dropdown && !dropdown.contains(event.target as Node)) {
        setIsOpen(false);
      }
    };
    document.addEventListener("mousedown", handleClickOutside);
    return () => document.removeEventListener("mousedown", handleClickOutside);
  }, [dropdownId]);

  const handleToggleProject = (projectId: number) => {
    const updated = value.includes(projectId)
      ? value.filter((id) => id !== projectId)
      : [...value, projectId];
    onChange(updated);
  };

  const displayText = value.length > 0
    ? projects.filter((p) => value.includes(p.id)).map((p) => p.name).join(", ")
    : "Select projects...";

  return (
    <div className="relative" id={dropdownId}>
      <button
        type="button"
        onClick={() => setIsOpen(!isOpen)}
        className="w-full px-3 py-2 border border-zinc-300 dark:border-zinc-700 rounded focus:outline-none focus:ring-2 focus:ring-lime-500 bg-white dark:bg-zinc-800 text-zinc-900 dark:text-zinc-100 text-left flex justify-between items-center"
      >
        <span className={value.length === 0 ? "text-zinc-500 dark:text-zinc-400" : ""}>
          {displayText}
        </span>
        <svg className={`w-4 h-4 transition-transform ${isOpen ? "rotate-180" : ""}`} fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
        </svg>
      </button>
      {isOpen && (
        <div className="absolute z-50 w-full mt-1 bg-white dark:bg-zinc-800 border border-zinc-300 dark:border-zinc-700 rounded shadow-lg max-h-60 overflow-y-auto">
          {projects.map((project) => {
            const isSelected = value.includes(project.id);
            return (
              <label
                key={project.id}
                className="flex items-center px-3 py-2 hover:bg-zinc-100 dark:hover:bg-zinc-700 cursor-pointer"
              >
                <input
                  type="checkbox"
                  checked={isSelected}
                  onChange={() => handleToggleProject(project.id)}
                  className="w-4 h-4 text-lime-500 border-zinc-300 dark:border-zinc-600 rounded focus:ring-lime-500 bg-white dark:bg-zinc-700"
                />
                <span className="ml-2 text-zinc-900 dark:text-zinc-100">{project.name}</span>
              </label>
            );
          })}
          {projects.length === 0 && (
            <div className="px-3 py-2 text-zinc-500 dark:text-zinc-400">No projects available</div>
          )}
        </div>
      )}
    </div>
  );
};

export default ProjectSelector;
