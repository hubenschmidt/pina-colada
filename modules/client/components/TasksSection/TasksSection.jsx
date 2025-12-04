"use client";

import { useState, useEffect } from "react";
import Link from "next/link";
import { Plus, Trash2, Edit2, Check, X } from "lucide-react";
import {
  getTasksByEntity,
  getTaskStatuses,
  getTaskPriorities,
  createTask,
  updateTask,
  deleteTask,
} from "../../api";

const priorityColors = {
  Low: "bg-gray-100 text-gray-700 dark:bg-gray-800 dark:text-gray-300",
  Medium: "bg-blue-100 text-blue-700 dark:bg-blue-900 dark:text-blue-300",
  High: "bg-orange-100 text-orange-700 dark:bg-orange-900 dark:text-orange-300",
  Urgent: "bg-red-100 text-red-700 dark:bg-red-900 dark:text-red-300",
};

const COMPLEXITY_OPTIONS = [1, 2, 3, 5, 8, 13, 21];

const TasksSection = ({
  entityType,
  entityId,
  pendingTasks,
  onPendingTasksChange,
}) => {
  const isCreateMode = entityId === null;
  const [tasks, setTasks] = useState([]);
  const [statuses, setStatuses] = useState([]);
  const [priorities, setPriorities] = useState([]);
  const [isLoading, setIsLoading] = useState(false);
  const [showAddForm, setShowAddForm] = useState(false);
  const [error, setError] = useState(null);

  // New task form state
  const [newTaskTitle, setNewTaskTitle] = useState("");
  const [newTaskDescription, setNewTaskDescription] = useState("");
  const [newTaskStatusId, setNewTaskStatusId] = useState(null);
  const [newTaskPriorityId, setNewTaskPriorityId] = useState(null);
  const [newTaskDueDate, setNewTaskDueDate] = useState("");
  const [newTaskComplexity, setNewTaskComplexity] = useState(null);

  // Edit task form state
  const [editingTaskId, setEditingTaskId] = useState(null);
  const [editTitle, setEditTitle] = useState("");
  const [editDescription, setEditDescription] = useState("");
  const [editStatusId, setEditStatusId] = useState(null);
  const [editPriorityId, setEditPriorityId] = useState(null);
  const [editDueDate, setEditDueDate] = useState("");
  const [editComplexity, setEditComplexity] = useState(null);

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

  useEffect(() => {
    if (!isCreateMode && entityId) {
      fetchTasks();
    }
  }, [entityType, entityId, isCreateMode]);

  const fetchTasks = async () => {
    if (!entityId) return;
    setIsLoading(true);
    try {
      const data = await getTasksByEntity(entityType, entityId);
      setTasks(data.items);
    } catch (err) {
      console.error("Failed to fetch tasks:", err);
    } finally {
      setIsLoading(false);
    }
  };

  const resetNewTaskForm = () => {
    setNewTaskTitle("");
    setNewTaskDescription("");
    setNewTaskStatusId(null);
    setNewTaskPriorityId(null);
    setNewTaskDueDate("");
    setNewTaskComplexity(null);
    setShowAddForm(false);
  };

  const handleAddTask = async () => {
    if (!newTaskTitle.trim()) return;

    const taskData = {
      title: newTaskTitle.trim(),
      description: newTaskDescription.trim() || undefined,
      current_status_id: newTaskStatusId,
      priority_id: newTaskPriorityId,
      due_date: newTaskDueDate || undefined,
      complexity: newTaskComplexity,
    };

    if (isCreateMode && onPendingTasksChange) {
      onPendingTasksChange([...(pendingTasks || []), taskData]);
      resetNewTaskForm();
      return;
    }

    if (!entityId) return;

    setError(null);
    try {
      const task = await createTask({
        ...taskData,
        taskable_type: entityType,
        taskable_id: entityId,
      });
      setTasks([task, ...tasks]);
      resetNewTaskForm();
    } catch (err) {
      setError(err?.message || "Failed to add task");
    }
  };

  const handleUpdateTask = async (taskId) => {
    if (!editTitle.trim()) return;

    setError(null);
    try {
      const updated = await updateTask(taskId, {
        title: editTitle.trim(),
        description: editDescription.trim() || undefined,
        current_status_id: editStatusId,
        priority_id: editPriorityId,
        due_date: editDueDate || undefined,
        complexity: editComplexity,
      });
      setTasks(tasks.map((t) => (t.id === taskId ? updated : t)));
      cancelEditing();
    } catch (err) {
      setError(err?.message || "Failed to update task");
    }
  };

  const handleDeleteTask = async (taskId) => {
    setError(null);
    try {
      await deleteTask(taskId);
      setTasks(tasks.filter((t) => t.id !== taskId));
    } catch (err) {
      setError(err?.message || "Failed to delete task");
    }
  };

  const handleRemovePendingTask = (index) => {
    if (onPendingTasksChange && pendingTasks) {
      onPendingTasksChange(pendingTasks.filter((_, i) => i !== index));
    }
  };

  const startEditing = (task) => {
    setEditingTaskId(task.id);
    setEditTitle(task.title);
    setEditDescription(task.description || "");
    setEditStatusId(task.status?.id || null);
    setEditPriorityId(task.priority?.id || null);
    setEditDueDate(task.due_date || "");
    setEditComplexity(task.complexity);
  };

  const cancelEditing = () => {
    setEditingTaskId(null);
    setEditTitle("");
    setEditDescription("");
    setEditStatusId(null);
    setEditPriorityId(null);
    setEditDueDate("");
    setEditComplexity(null);
  };

  const inputClasses =
    "w-full px-3 py-2 border border-zinc-300 dark:border-zinc-700 rounded focus:outline-none focus:ring-2 focus:ring-lime-500 bg-white dark:bg-zinc-800 text-zinc-900 dark:text-zinc-100";

  const selectClasses =
    "px-3 py-2 border border-zinc-300 dark:border-zinc-700 rounded focus:outline-none focus:ring-2 focus:ring-lime-500 bg-white dark:bg-zinc-800 text-zinc-900 dark:text-zinc-100 text-sm";

  const renderTaskCard = (task) => {
    const isEditing = editingTaskId === task.id;
    const priorityClass = task.priority?.name
      ? priorityColors[task.priority.name] || priorityColors.Medium
      : "";

    return (
      <div
        key={task.id}
        className="p-3 border border-zinc-200 dark:border-zinc-700 rounded"
      >
        {isEditing ? (
          <div className="space-y-3">
            <input
              type="text"
              value={editTitle}
              onChange={(e) => setEditTitle(e.target.value)}
              className={inputClasses}
              placeholder="Task title..."
              autoFocus
            />

            <textarea
              value={editDescription}
              onChange={(e) => setEditDescription(e.target.value)}
              className={inputClasses}
              placeholder="Description (optional)..."
              rows={2}
            />

            <div className="grid grid-cols-2 md:grid-cols-4 gap-2">
              <select
                value={editStatusId || ""}
                onChange={(e) =>
                  setEditStatusId(
                    e.target.value ? Number(e.target.value) : null,
                  )
                }
                className={selectClasses}
              >
                <option value="">Status...</option>
                {statuses.map((s) => (
                  <option key={s.id} value={s.id}>
                    {s.name}
                  </option>
                ))}
              </select>
              <select
                value={editPriorityId || ""}
                onChange={(e) =>
                  setEditPriorityId(
                    e.target.value ? Number(e.target.value) : null,
                  )
                }
                className={selectClasses}
              >
                <option value="">Priority...</option>
                {priorities.map((p) => (
                  <option key={p.id} value={p.id}>
                    {p.name}
                  </option>
                ))}
              </select>
              <input
                type="date"
                value={editDueDate}
                onChange={(e) => setEditDueDate(e.target.value)}
                className={selectClasses}
              />

              <select
                value={editComplexity || ""}
                onChange={(e) =>
                  setEditComplexity(
                    e.target.value ? Number(e.target.value) : null,
                  )
                }
                className={selectClasses}
              >
                <option value="">Complexity...</option>
                {COMPLEXITY_OPTIONS.map((c) => (
                  <option key={c} value={c}>
                    {c}
                  </option>
                ))}
              </select>
            </div>
            <div className="flex gap-2">
              <button
                type="button"
                onClick={() => handleUpdateTask(task.id)}
                className="p-1 text-lime-600 hover:text-lime-700"
                title="Save"
              >
                <Check size={16} />
              </button>
              <button
                type="button"
                onClick={cancelEditing}
                className="p-1 text-zinc-400 hover:text-zinc-600"
                title="Cancel"
              >
                <X size={16} />
              </button>
            </div>
          </div>
        ) : (
          <>
            <div className="flex items-start justify-between gap-2">
              <div className="flex-1">
                <Link
                  href={`/tasks/${task.id}`}
                  className="text-sm font-medium text-zinc-900 dark:text-zinc-100 hover:text-zinc-500 dark:hover:text-zinc-400 hover:underline"
                >
                  {task.title}
                </Link>
                {task.description && (
                  <p className="text-sm text-zinc-600 dark:text-zinc-400 mt-1">
                    {task.description}
                  </p>
                )}
              </div>
              <div className="flex gap-1 shrink-0">
                <button
                  type="button"
                  onClick={() => startEditing(task)}
                  className="p-1 text-zinc-400 hover:text-zinc-600"
                  title="Edit"
                >
                  <Edit2 size={14} />
                </button>
                <button
                  type="button"
                  onClick={() => handleDeleteTask(task.id)}
                  className="p-1 text-zinc-400 hover:text-red-500"
                  title="Delete"
                >
                  <Trash2 size={14} />
                </button>
              </div>
            </div>
            <div className="flex flex-wrap items-center gap-2 mt-2">
              {task.status && (
                <span className="text-xs px-2 py-0.5 rounded bg-zinc-100 dark:bg-zinc-700 text-zinc-600 dark:text-zinc-300">
                  {task.status.name}
                </span>
              )}
              {task.priority && (
                <span
                  className={`text-xs px-2 py-0.5 rounded ${priorityClass}`}
                >
                  {task.priority.name}
                </span>
              )}
              {task.complexity && (
                <span className="text-xs px-2 py-0.5 rounded bg-purple-100 text-purple-700 dark:bg-purple-900 dark:text-purple-300">
                  {task.complexity} pts
                </span>
              )}
              {task.due_date && (
                <span className="text-xs text-zinc-500">
                  Due: {task.due_date}
                </span>
              )}
            </div>
          </>
        )}
      </div>
    );
  };

  const renderPendingTaskCard = (task, index) => {
    const priorityName = priorities.find(
      (p) => p.id === task.priority_id,
    )?.name;
    const statusName = statuses.find(
      (s) => s.id === task.current_status_id,
    )?.name;
    const priorityClass = priorityName
      ? priorityColors[priorityName] || ""
      : "";

    return (
      <div
        key={`pending-${index}`}
        className="p-3 border border-zinc-200 dark:border-zinc-700 rounded"
      >
        <div className="flex items-start justify-between gap-2">
          <div className="flex-1">
            <p className="text-sm font-medium text-zinc-900 dark:text-zinc-100">
              {task.title}
            </p>
            {task.description && (
              <p className="text-sm text-zinc-600 dark:text-zinc-400 mt-1">
                {task.description}
              </p>
            )}
          </div>
          <button
            type="button"
            onClick={() => handleRemovePendingTask(index)}
            className="p-1 text-zinc-400 hover:text-red-500"
            title="Remove"
          >
            <Trash2 size={14} />
          </button>
        </div>
        <div className="flex flex-wrap items-center gap-2 mt-2">
          <span className="text-xs text-zinc-500 italic">Pending</span>
          {statusName && (
            <span className="text-xs px-2 py-0.5 rounded bg-zinc-100 dark:bg-zinc-700 text-zinc-600 dark:text-zinc-300">
              {statusName}
            </span>
          )}
          {priorityName && (
            <span className={`text-xs px-2 py-0.5 rounded ${priorityClass}`}>
              {priorityName}
            </span>
          )}
          {task.complexity && (
            <span className="text-xs px-2 py-0.5 rounded bg-purple-100 text-purple-700 dark:bg-purple-900 dark:text-purple-300">
              {task.complexity} pts
            </span>
          )}
          {task.due_date && (
            <span className="text-xs text-zinc-500">Due: {task.due_date}</span>
          )}
        </div>
      </div>
    );
  };

  const renderAddForm = () => (
    <div className="mb-4 p-4 border border-zinc-200 dark:border-zinc-700 rounded-lg bg-zinc-50 dark:bg-zinc-800/50 space-y-3">
      <input
        type="text"
        value={newTaskTitle}
        onChange={(e) => setNewTaskTitle(e.target.value)}
        className={inputClasses}
        placeholder="Task title..."
        autoFocus
      />

      <textarea
        value={newTaskDescription}
        onChange={(e) => setNewTaskDescription(e.target.value)}
        className={inputClasses}
        placeholder="Description (optional)..."
        rows={2}
      />

      <div className="grid grid-cols-2 md:grid-cols-4 gap-2">
        <select
          value={newTaskStatusId || ""}
          onChange={(e) =>
            setNewTaskStatusId(e.target.value ? Number(e.target.value) : null)
          }
          className={selectClasses}
        >
          <option value="">Status...</option>
          {statuses.map((s) => (
            <option key={s.id} value={s.id}>
              {s.name}
            </option>
          ))}
        </select>
        <select
          value={newTaskPriorityId || ""}
          onChange={(e) =>
            setNewTaskPriorityId(e.target.value ? Number(e.target.value) : null)
          }
          className={selectClasses}
        >
          <option value="">Priority...</option>
          {priorities.map((p) => (
            <option key={p.id} value={p.id}>
              {p.name}
            </option>
          ))}
        </select>
        <input
          type="date"
          value={newTaskDueDate}
          onChange={(e) => setNewTaskDueDate(e.target.value)}
          className={selectClasses}
        />

        <select
          value={newTaskComplexity || ""}
          onChange={(e) =>
            setNewTaskComplexity(e.target.value ? Number(e.target.value) : null)
          }
          className={selectClasses}
        >
          <option value="">Complexity...</option>
          {COMPLEXITY_OPTIONS.map((c) => (
            <option key={c} value={c}>
              {c}
            </option>
          ))}
        </select>
      </div>
      <div className="flex gap-2">
        <button
          type="button"
          onClick={handleAddTask}
          disabled={!newTaskTitle.trim()}
          className="px-3 py-1 text-sm bg-lime-600 text-white rounded hover:bg-lime-700 disabled:opacity-50"
        >
          Add
        </button>
        <button
          type="button"
          onClick={resetNewTaskForm}
          className="px-3 py-1 text-sm bg-zinc-200 dark:bg-zinc-700 text-zinc-700 dark:text-zinc-300 rounded hover:bg-zinc-300 dark:hover:bg-zinc-600"
        >
          Cancel
        </button>
      </div>
    </div>
  );

  return (
    <div className="space-y-2">
      <div className="flex items-center justify-between mb-3">
        <span className="text-sm font-medium text-zinc-700 dark:text-zinc-300">
          Tasks
        </span>
        {!showAddForm && (
          <button
            type="button"
            onClick={() => setShowAddForm(true)}
            className="flex items-center gap-1 text-sm text-lime-600 dark:text-lime-400 hover:text-lime-700 dark:hover:text-lime-300"
          >
            <Plus size={16} />
            Add Task
          </button>
        )}
      </div>

      {error && (
        <div className="p-2 text-sm text-red-600 dark:text-red-400 bg-red-100 dark:bg-red-900/30 rounded">
          {error}
        </div>
      )}

      {showAddForm && renderAddForm()}

      {isLoading && <p className="text-sm text-zinc-500">Loading tasks...</p>}

      <div className="space-y-2">
        {tasks.map((task) => renderTaskCard(task))}
        {pendingTasks?.map((task, index) => renderPendingTaskCard(task, index))}
      </div>

      {!isLoading &&
        tasks.length === 0 &&
        (!pendingTasks || pendingTasks.length === 0) &&
        !showAddForm && (
          <p className="text-sm text-zinc-500 dark:text-zinc-400">
            No tasks yet.
          </p>
        )}
    </div>
  );
};

export default TasksSection;
