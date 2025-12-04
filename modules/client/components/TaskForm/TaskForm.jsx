"use client";

import { useState, useEffect } from "react";
import { Badge, Group } from "@mantine/core";
import { FolderKanban } from "lucide-react";

import FormActions from "../FormActions/FormActions";
import Timestamps from "../Timestamps/Timestamps";
import CommentsSection from "../CommentsSection/CommentsSection";
import { usePendingChanges } from "../../hooks/usePendingChanges";
import {
  getTaskStatuses,
  getTaskPriorities,
  searchAccounts,
  getJobs,
  getOpportunities,
  getPartnerships,
} from "../../api";

const COMPLEXITY_OPTIONS = [
  { value: "", label: "Select complexity..." },
  { value: "1", label: "1" },
  { value: "2", label: "2" },
  { value: "3", label: "3" },
  { value: "5", label: "5" },
  { value: "8", label: "8" },
  { value: "13", label: "13" },
  { value: "21", label: "21" },
];

const ENTITY_TYPES = [
  { value: "", label: "None (standalone task)" },
  { value: "Account", label: "Account" },
  { value: "Lead", label: "Lead" },
];

const TaskForm = ({ onClose, onAdd, task, onUpdate, onDelete, selectedProject }) => {
  const isEditMode = !!task;

  const [formData, setFormData] = useState({
    title: "",
    description: "",
    current_status_id: "",
    priority_id: "",
    start_date: "",
    due_date: "",
    estimated_hours: "",
    actual_hours: "",
    complexity: "",
  });

  const [entityType, setEntityType] = useState("");
  const [entityId, setEntityId] = useState("");
  const [entitySearch, setEntitySearch] = useState("");
  const [entityOptions, setEntityOptions] = useState([]);
  const [isSearching, setIsSearching] = useState(false);

  const [statuses, setStatuses] = useState([]);
  const [priorities, setPriorities] = useState([]);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [isDeleting, setIsDeleting] = useState(false);
  const [errors, setErrors] = useState({});

  // Load statuses and priorities
  useEffect(() => {
    const loadOptions = async () => {
      try {
        const [statusData, priorityData] = await Promise.all([
          getTaskStatuses(),
          getTaskPriorities(),
        ]);
        setStatuses(statusData);
        setPriorities(priorityData);
      } catch (err) {
        console.error("Failed to load task options:", err);
      }
    };
    loadOptions();
  }, []);

  // Initialize form in edit mode
  useEffect(() => {
    if (isEditMode && task) {
      setFormData({
        title: task.title || "",
        description: task.description || "",
        current_status_id: task.status?.id?.toString() || "",
        priority_id: task.priority?.id?.toString() || "",
        start_date: task.start_date || "",
        due_date: task.due_date || "",
        estimated_hours: task.estimated_hours ?? "",
        actual_hours: task.actual_hours ?? "",
        complexity: task.complexity?.toString() || "",
      });

      if (task.entity?.type && task.entity?.id) {
        setEntityType(task.entity.type);
        setEntityId(task.entity.id.toString());
        setEntitySearch(task.entity.display_name || "");
      }
    }
  }, [isEditMode, task]);

  const hasPendingChanges = usePendingChanges({
    original: task,
    current: formData,
  });

  // Search for entities when typing
  useEffect(() => {
    if (!entityType || !entitySearch.trim() || entitySearch.length < 2) {
      setEntityOptions([]);
      return;
    }

    const searchEntities = async () => {
      setIsSearching(true);
      try {
        let results = [];
        if (entityType === "Account") {
          const accounts = await searchAccounts(entitySearch);
          results = accounts.map((a) => ({
            id: a.id,
            name: a.name,
            type: a.type,
          }));
        } else if (entityType === "Lead") {
          // Fetch all lead types and filter client-side
          const [jobs, opportunities, partnerships] = await Promise.all([
            getJobs(1, 50, "title", "ASC", entitySearch).catch(() => ({ items: [] })),
            getOpportunities(1, 50, "title", "ASC", entitySearch).catch(() => ({ items: [] })),
            getPartnerships(1, 50, "title", "ASC", entitySearch).catch(() => ({ items: [] })),
          ]);
          results = [
            ...(jobs.items || []).map((j) => ({ id: j.id, name: j.title, type: "Job" })),
            ...(opportunities.items || []).map((o) => ({
              id: o.id,
              name: o.title,
              type: "Opportunity",
            })),
            ...(partnerships.items || []).map((p) => ({
              id: p.id,
              name: p.title,
              type: "Partnership",
            })),
          ];
        }
        setEntityOptions(results);
      } catch (err) {
        console.error("Failed to search entities:", err);
        setEntityOptions([]);
      } finally {
        setIsSearching(false);
      }
    };

    const debounce = setTimeout(searchEntities, 300);
    return () => clearTimeout(debounce);
  }, [entityType, entitySearch]);

  const handleFieldChange = (name, value) => {
    setFormData((prev) => ({ ...prev, [name]: value }));
    if (errors[name]) {
      setErrors((prev) => {
        const newErrors = { ...prev };
        delete newErrors[name];
        return newErrors;
      });
    }
  };

  const handleEntityTypeChange = (value) => {
    setEntityType(value);
    setEntityId("");
    setEntitySearch("");
    setEntityOptions([]);
  };

  const handleEntitySelect = (entity) => {
    setEntityId(entity.id.toString());
    setEntitySearch(entity.name);
    setEntityOptions([]);
  };

  const validate = () => {
    const newErrors = {};
    if (!formData.title.trim()) {
      newErrors.title = "Title is required";
    }
    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    if (!validate()) return;

    setIsSubmitting(true);
    try {
      const submitData = {
        title: formData.title.trim(),
        description: formData.description.trim() || null,
        current_status_id: formData.current_status_id ? Number(formData.current_status_id) : null,
        priority_id: formData.priority_id ? Number(formData.priority_id) : null,
        start_date: formData.start_date || null,
        due_date: formData.due_date || null,
        estimated_hours: formData.estimated_hours !== "" ? Number(formData.estimated_hours) : null,
        actual_hours: formData.actual_hours !== "" ? Number(formData.actual_hours) : null,
        complexity: formData.complexity ? Number(formData.complexity) : null,
      };

      // Add entity linking if selected
      if (entityType && entityId) {
        submitData.taskable_type = entityType;
        submitData.taskable_id = Number(entityId);
      }

      if (isEditMode && task && onUpdate) {
        await onUpdate(task.id, submitData);
      } else if (onAdd) {
        await onAdd(submitData);
      }

      onClose();
    } catch (error) {
      console.error("Failed to save task:", error);
      setErrors({ _form: error?.message || "Failed to save task. Please try again." });
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleDelete = async () => {
    if (!task || !onDelete) return;

    if (!isDeleting) {
      setIsDeleting(true);
      return;
    }

    try {
      await onDelete(task.id);
      onClose();
    } catch (error) {
      console.error("Failed to delete task:", error);
      setErrors({ _form: error?.message || "Failed to delete. Please try again." });
      setIsDeleting(false);
    }
  };

  const inputClasses =
    "w-full px-3 py-2 border border-zinc-300 dark:border-zinc-700 rounded focus:outline-none focus:ring-2 focus:ring-lime-500 bg-white dark:bg-zinc-800 text-zinc-900 dark:text-zinc-100";

  const selectClasses =
    "w-full px-3 py-2 border border-zinc-300 dark:border-zinc-700 rounded focus:outline-none focus:ring-2 focus:ring-lime-500 bg-white dark:bg-zinc-800 text-zinc-900 dark:text-zinc-100";

  return (
    <div>
      <Group justify="space-between" mb="md">
        <h1 className="text-2xl font-bold text-zinc-900 dark:text-zinc-100">
          {isEditMode ? "Edit Task" : "New Task"}
        </h1>
        {selectedProject ? (
          <Badge variant="light" color="lime" leftSection={<FolderKanban className="h-3 w-3" />}>
            {selectedProject.name}
          </Badge>
        ) : (
          <Badge variant="light" color="gray">
            Global
          </Badge>
        )}
      </Group>

      <form onSubmit={handleSubmit}>
        <div className="space-y-6">
          {/* Basic Info Section */}
          <div>
            <h4 className="text-sm font-semibold text-zinc-900 dark:text-zinc-100 mb-4">
              Task Details
            </h4>
            <div className="grid grid-cols-1 gap-4">
              <div>
                <label className="block text-sm font-medium text-zinc-700 dark:text-zinc-300 mb-1">
                  Title <span className="text-red-500">*</span>
                </label>
                <input
                  type="text"
                  value={formData.title}
                  onChange={(e) => handleFieldChange("title", e.target.value)}
                  className={`${inputClasses} ${errors.title ? "border-red-500" : ""}`}
                  placeholder="What needs to be done?"
                />
                {errors.title && <p className="text-red-500 text-xs mt-1">{errors.title}</p>}
              </div>

              <div>
                <label className="block text-sm font-medium text-zinc-700 dark:text-zinc-300 mb-1">
                  Description
                </label>
                <textarea
                  value={formData.description}
                  onChange={(e) => handleFieldChange("description", e.target.value)}
                  className={inputClasses}
                  placeholder="Additional details..."
                  rows={3}
                />
              </div>
            </div>
          </div>

          {/* Status & Priority Section */}
          <div className="border-t border-zinc-300 dark:border-zinc-700 pt-4">
            <h4 className="text-sm font-semibold text-zinc-900 dark:text-zinc-100 mb-4">
              Status & Priority
            </h4>
            <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
              <div>
                <label className="block text-sm font-medium text-zinc-700 dark:text-zinc-300 mb-1">
                  Status
                </label>
                <select
                  value={formData.current_status_id}
                  onChange={(e) => handleFieldChange("current_status_id", e.target.value)}
                  className={selectClasses}>
                  <option value="">Select status...</option>
                  {statuses.map((s) => (
                    <option key={s.id} value={s.id}>
                      {s.name}
                    </option>
                  ))}
                </select>
              </div>

              <div>
                <label className="block text-sm font-medium text-zinc-700 dark:text-zinc-300 mb-1">
                  Priority
                </label>
                <select
                  value={formData.priority_id}
                  onChange={(e) => handleFieldChange("priority_id", e.target.value)}
                  className={selectClasses}>
                  <option value="">Select priority...</option>
                  {priorities.map((p) => (
                    <option key={p.id} value={p.id}>
                      {p.name}
                    </option>
                  ))}
                </select>
              </div>

              <div>
                <label className="block text-sm font-medium text-zinc-700 dark:text-zinc-300 mb-1">
                  Complexity
                </label>
                <select
                  value={formData.complexity}
                  onChange={(e) => handleFieldChange("complexity", e.target.value)}
                  className={selectClasses}>
                  {COMPLEXITY_OPTIONS.map((c) => (
                    <option key={c.value} value={c.value}>
                      {c.label}
                    </option>
                  ))}
                </select>
              </div>
            </div>
          </div>

          {/* Dates Section */}
          <div className="border-t border-zinc-300 dark:border-zinc-700 pt-4">
            <h4 className="text-sm font-semibold text-zinc-900 dark:text-zinc-100 mb-4">Dates</h4>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div>
                <label className="block text-sm font-medium text-zinc-700 dark:text-zinc-300 mb-1">
                  Start Date
                </label>
                <input
                  type="date"
                  value={formData.start_date}
                  onChange={(e) => handleFieldChange("start_date", e.target.value)}
                  className={inputClasses}
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-zinc-700 dark:text-zinc-300 mb-1">
                  Due Date
                </label>
                <input
                  type="date"
                  value={formData.due_date}
                  onChange={(e) => handleFieldChange("due_date", e.target.value)}
                  className={inputClasses}
                />
              </div>
            </div>
          </div>

          {/* Hours Section */}
          <div className="border-t border-zinc-300 dark:border-zinc-700 pt-4">
            <h4 className="text-sm font-semibold text-zinc-900 dark:text-zinc-100 mb-4">Hours</h4>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div>
                <label className="block text-sm font-medium text-zinc-700 dark:text-zinc-300 mb-1">
                  Estimated Hours
                </label>
                <input
                  type="number"
                  value={formData.estimated_hours}
                  onChange={(e) => handleFieldChange("estimated_hours", e.target.value)}
                  className={inputClasses}
                  min="0"
                  step="0.5"
                  placeholder="0"
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-zinc-700 dark:text-zinc-300 mb-1">
                  Actual Hours
                </label>
                <input
                  type="number"
                  value={formData.actual_hours}
                  onChange={(e) => handleFieldChange("actual_hours", e.target.value)}
                  className={inputClasses}
                  min="0"
                  step="0.5"
                  placeholder="0"
                />
              </div>
            </div>
          </div>

          {/* Link to Entity Section */}
          <div className="border-t border-zinc-300 dark:border-zinc-700 pt-4">
            <h4 className="text-sm font-semibold text-zinc-900 dark:text-zinc-100 mb-4">
              Link to Entity
            </h4>

            {/* Display linked entity in edit mode */}
            {isEditMode && task?.entity?.type && task?.entity?.display_name && (
              <div className="mb-4 p-3 bg-zinc-50 dark:bg-zinc-800/50 rounded border border-zinc-200 dark:border-zinc-700">
                <p className="text-xs text-zinc-500 dark:text-zinc-400 mb-1">Linked To</p>
                <div className="flex items-center gap-2">
                  <Badge
                    size="sm"
                    variant="light"
                    color={
                      task.entity.type === "Account"
                        ? "blue"
                        : task.entity.type === "Lead"
                          ? "orange"
                          : "gray"
                    }>
                    {task.entity.type}
                  </Badge>
                  {task.entity.url ? (
                    <a
                      href={task.entity.url}
                      className="text-sm font-medium text-lime-600 dark:text-lime-400 hover:underline">
                      {task.entity.display_name}
                    </a>
                  ) : (
                    <span className="text-sm font-medium text-zinc-900 dark:text-zinc-100">
                      {task.entity.display_name}
                    </span>
                  )}
                </div>
              </div>
            )}

            {/* Entity linking controls (create mode only) */}
            {!isEditMode && (
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm font-medium text-zinc-700 dark:text-zinc-300 mb-1">
                    Entity Type
                  </label>
                  <select
                    value={entityType}
                    onChange={(e) => handleEntityTypeChange(e.target.value)}
                    className={selectClasses}>
                    {ENTITY_TYPES.map((t) => (
                      <option key={t.value} value={t.value}>
                        {t.label}
                      </option>
                    ))}
                  </select>
                </div>

                {entityType && (
                  <div className="relative">
                    <label className="block text-sm font-medium text-zinc-700 dark:text-zinc-300 mb-1">
                      Search {entityType}
                    </label>
                    <input
                      type="text"
                      value={entitySearch}
                      onChange={(e) => {
                        setEntitySearch(e.target.value);
                        if (!e.target.value) setEntityId("");
                      }}
                      className={inputClasses}
                      placeholder={`Search for ${entityType.toLowerCase()}...`}
                    />
                    {isSearching && <p className="text-xs text-zinc-500 mt-1">Searching...</p>}
                    {entityOptions.length > 0 && !entityId && (
                      <div className="absolute z-10 w-full mt-1 bg-white dark:bg-zinc-800 border border-zinc-300 dark:border-zinc-700 rounded shadow-lg max-h-48 overflow-auto">
                        {entityOptions.map((entity) => (
                          <button
                            key={`${entity.type}-${entity.id}`}
                            type="button"
                            onClick={() => handleEntitySelect(entity)}
                            className="w-full px-3 py-2 text-left hover:bg-zinc-100 dark:hover:bg-zinc-700 text-sm">
                            <span className="font-medium">{entity.name}</span>
                            {entity.type && (
                              <span className="ml-2 text-xs text-zinc-500">({entity.type})</span>
                            )}
                          </button>
                        ))}
                      </div>
                    )}
                    {entityId && (
                      <p className="text-xs text-lime-600 dark:text-lime-400 mt-1">
                        Linked to: {entitySearch}
                      </p>
                    )}
                  </div>
                )}
              </div>
            )}

            {/* Show "not linked" message in edit mode if no entity */}
            {isEditMode && !task?.entity?.type && (
              <p className="text-sm text-zinc-500 dark:text-zinc-400">
                This task is not linked to any entity.
              </p>
            )}
          </div>
        </div>

        <FormActions
          isEditMode={isEditMode}
          isSubmitting={isSubmitting}
          isDeleting={isDeleting}
          hasPendingChanges={hasPendingChanges}
          onClose={onClose}
          onDelete={onDelete ? handleDelete : undefined}
        />

        {isEditMode && task && (
          <Timestamps createdAt={task.created_at} updatedAt={task.updated_at} />
        )}
      </form>

      {errors._form && (
        <div className="mt-4 p-3 bg-red-100 dark:bg-red-900/30 text-red-600 dark:text-red-400 rounded">
          {errors._form}
        </div>
      )}

      {/* Comments Section - only in edit mode */}
      {isEditMode && task && (
        <div className="border-t border-zinc-200 dark:border-zinc-700 pt-6 mt-6">
          <CommentsSection entityType="Task" entityId={task.id} />
        </div>
      )}
    </div>
  );
};

export default TaskForm;
