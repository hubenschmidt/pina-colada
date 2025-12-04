"use client";

import { useState, useEffect, useId } from "react";
import Link from "next/link";
import { X } from "lucide-react";












const ProjectSelector = ({ value, onChange, projects }) => {
  const [isOpen, setIsOpen] = useState(false);
  const dropdownId = useId();

  useEffect(() => {
    const handleClickOutside = (event) => {
      const dropdown = document.getElementById(dropdownId);
      if (dropdown && !dropdown.contains(event.target)) {
        setIsOpen(false);
      }
    };
    document.addEventListener("mousedown", handleClickOutside);
    return () => document.removeEventListener("mousedown", handleClickOutside);
  }, [dropdownId]);

  const handleToggleProject = (projectId) => {
    const updated = value.includes(projectId) ?
    value.filter((id) => id !== projectId) :
    [...value, projectId];
    onChange(updated);
  };

  const selectedProjects = projects.filter((p) => value.includes(p.id));

  return (
    <div className="space-y-2">
      <div className="relative" id={dropdownId}>
        <button
          type="button"
          onClick={() => setIsOpen(!isOpen)}
          className="w-full px-3 py-2 border border-zinc-300 dark:border-zinc-700 rounded focus:outline-none focus:ring-2 focus:ring-lime-500 bg-white dark:bg-zinc-800 text-zinc-900 dark:text-zinc-100 text-left flex justify-between items-center">

          <span className="text-zinc-500 dark:text-zinc-400">
            {value.length === 0 ? "Select projects..." : `${value.length} project${value.length > 1 ? "s" : ""} selected`}
          </span>
          <svg className={`w-4 h-4 transition-transform ${isOpen ? "rotate-180" : ""}`} fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
          </svg>
        </button>
        {isOpen &&
        <div className="absolute z-50 w-full mt-1 bg-white dark:bg-zinc-800 border border-zinc-300 dark:border-zinc-700 rounded shadow-lg max-h-60 overflow-y-auto">
            {projects.map((project) => {
            const isSelected = value.includes(project.id);
            return (
              <label
                key={project.id}
                className="flex items-center px-3 py-2 hover:bg-zinc-100 dark:hover:bg-zinc-700 cursor-pointer">

                  <input
                  type="checkbox"
                  checked={isSelected}
                  onChange={() => handleToggleProject(project.id)}
                  className="w-4 h-4 text-lime-500 border-zinc-300 dark:border-zinc-600 rounded focus:ring-lime-500 bg-white dark:bg-zinc-700" />

                  <span className="ml-2 text-zinc-900 dark:text-zinc-100">{project.name}</span>
                </label>);

          })}
            {projects.length === 0 &&
          <div className="px-3 py-2 text-zinc-500 dark:text-zinc-400">No projects available</div>
          }
          </div>
        }
      </div>
      {selectedProjects.length > 0 &&
      <div className="flex flex-wrap gap-2">
          {selectedProjects.map((project) =>
        <div
          key={project.id}
          className="flex items-center gap-1 px-2 py-1 bg-zinc-100 dark:bg-zinc-700 rounded text-sm">

              <Link
            href={`/projects/${project.id}`}
            className="text-zinc-900 dark:text-zinc-100 hover:text-zinc-500 dark:hover:text-zinc-400 hover:underline">

                {project.name}
              </Link>
              <button
            type="button"
            onClick={() => handleToggleProject(project.id)}
            className="text-zinc-400 hover:text-zinc-600 dark:hover:text-zinc-300"
            title="Remove">

                <X size={14} />
              </button>
            </div>
        )}
        </div>
      }
    </div>);

};

export default ProjectSelector;