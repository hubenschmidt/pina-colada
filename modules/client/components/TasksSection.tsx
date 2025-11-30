"use client";

import { useState, useEffect } from "react";
import { Plus, Trash2, Edit2, Check, X } from "lucide-react";
import {
  Task,
  TaskInput,
  getTasksByEntity,
  createTask,
  updateTask,
  deleteTask,
} from "../api";

interface TasksSectionProps {
  entityType: string;
  entityId: number | null;
  pendingTasks?: TaskInput[];
  onPendingTasksChange?: (tasks: TaskInput[]) => void;
}

const priorityColors: Record<string, string> = {
  Low: "bg-gray-100 text-gray-700 dark:bg-gray-800 dark:text-gray-300",
  Medium: "bg-blue-100 text-blue-700 dark:bg-blue-900 dark:text-blue-300",
  High: "bg-orange-100 text-orange-700 dark:bg-orange-900 dark:text-orange-300",
  Urgent: "bg-red-100 text-red-700 dark:bg-red-900 dark:text-red-300",
};

const TasksSection = ({
  entityType,
  entityId,
  pendingTasks,
  onPendingTasksChange,
}: TasksSectionProps) => {
  const isCreateMode = entityId === null;
  const [tasks, setTasks] = useState<Task[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [showAddForm, setShowAddForm] = useState(false);
  const [newTaskTitle, setNewTaskTitle] = useState("");
  const [newTaskDescription, setNewTaskDescription] = useState("");
  const [editingTaskId, setEditingTaskId] = useState<number | null>(null);
  const [editTitle, setEditTitle] = useState("");
  const [editDescription, setEditDescription] = useState("");
  const [error, setError] = useState<string | null>(null);

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

  const handleAddTask = async () => {
    if (!newTaskTitle.trim()) return;

    if (isCreateMode && onPendingTasksChange) {
      onPendingTasksChange([
        ...(pendingTasks || []),
        {
          title: newTaskTitle.trim(),
          description: newTaskDescription.trim() || undefined,
        },
      ]);
      setNewTaskTitle("");
      setNewTaskDescription("");
      setShowAddForm(false);
      return;
    }

    if (!entityId) return;

    setError(null);
    try {
      const task = await createTask({
        title: newTaskTitle.trim(),
        description: newTaskDescription.trim() || undefined,
        taskable_type: entityType,
        taskable_id: entityId,
      });
      setTasks([task, ...tasks]);
      setNewTaskTitle("");
      setNewTaskDescription("");
      setShowAddForm(false);
    } catch (err: any) {
      setError(err?.message || "Failed to add task");
    }
  };

  const handleUpdateTask = async (taskId: number) => {
    if (!editTitle.trim()) return;

    setError(null);
    try {
      const updated = await updateTask(taskId, {
        title: editTitle.trim(),
        description: editDescription.trim() || undefined,
      });
      setTasks(tasks.map((t) => (t.id === taskId ? updated : t)));
      setEditingTaskId(null);
      setEditTitle("");
      setEditDescription("");
    } catch (err: any) {
      setError(err?.message || "Failed to update task");
    }
  };

  const handleDeleteTask = async (taskId: number) => {
    setError(null);
    try {
      await deleteTask(taskId);
      setTasks(tasks.filter((t) => t.id !== taskId));
    } catch (err: any) {
      setError(err?.message || "Failed to delete task");
    }
  };

  const handleRemovePendingTask = (index: number) => {
    if (onPendingTasksChange && pendingTasks) {
      onPendingTasksChange(pendingTasks.filter((_, i) => i !== index));
    }
  };

  const startEditing = (task: Task) => {
    setEditingTaskId(task.id);
    setEditTitle(task.title);
    setEditDescription(task.description || "");
  };

  const cancelEditing = () => {
    setEditingTaskId(null);
    setEditTitle("");
    setEditDescription("");
  };

  const inputClasses =
    "w-full px-3 py-2 border border-zinc-300 dark:border-zinc-700 rounded focus:outline-none focus:ring-2 focus:ring-lime-500 bg-white dark:bg-zinc-800 text-zinc-900 dark:text-zinc-100";

  const renderTaskCard = (task: Task) => {
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
          <div className="space-y-2">
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
                <p className="text-sm font-medium text-zinc-900 dark:text-zinc-100">
                  {task.title}
                </p>
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
            <div className="flex items-center gap-2 mt-2">
              {task.status && (
                <span className="text-xs px-2 py-0.5 rounded bg-zinc-100 dark:bg-zinc-700 text-zinc-600 dark:text-zinc-300">
                  {task.status.name}
                </span>
              )}
              {task.priority && (
                <span className={`text-xs px-2 py-0.5 rounded ${priorityClass}`}>
                  {task.priority.name}
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

  const renderPendingTaskCard = (task: TaskInput, index: number) => (
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
      <span className="text-xs text-zinc-500 italic mt-1 block">Pending</span>
    </div>
  );

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
          onClick={() => {
            setShowAddForm(false);
            setNewTaskTitle("");
            setNewTaskDescription("");
          }}
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

      {isLoading && (
        <p className="text-sm text-zinc-500">Loading tasks...</p>
      )}

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
