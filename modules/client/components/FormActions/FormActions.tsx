"use client";

import { Plus, Save, Trash2 } from "lucide-react";

interface FormActionsProps {
  isEditMode: boolean;
  isSubmitting: boolean;
  isDeleting?: boolean;
  hasPendingChanges?: boolean;
  onClose: () => void;
  onDelete?: () => void;
  submitButtonText?: string;
  cancelButtonText?: string;
  variant?: "default" | "compact";
}

const FormActions = ({
  isEditMode,
  isSubmitting,
  isDeleting = false,
  hasPendingChanges = false,
  onClose,
  onDelete,
  submitButtonText,
  cancelButtonText,
  variant = "default",
}: FormActionsProps) => {
  const paddingY = variant === "compact" ? "py-2" : "py-3";
  const rounded = variant === "compact" ? "rounded" : "rounded-lg";
  const marginTop = variant === "compact" ? "pt-4" : "mt-6";

  const getSubmitText = () => {
    if (isSubmitting) {
      return isEditMode ? "Saving..." : "Creating...";
    }
    if (submitButtonText) return submitButtonText;
    return isEditMode ? "Save" : "Create";
  };

  return (
    <div className={`flex gap-3 ${marginTop}`}>
      <button
        type="submit"
        disabled={isSubmitting}
        className={`flex items-center gap-2 px-6 ${paddingY} ${rounded} font-semibold disabled:opacity-50 disabled:cursor-not-allowed ${
          hasPendingChanges
            ? "bg-lime-600 text-white border border-lime-600 hover:bg-lime-700"
            : "bg-white dark:bg-zinc-800 text-zinc-900 dark:text-zinc-100 border border-zinc-300 dark:border-zinc-700 hover:bg-zinc-50 dark:hover:bg-zinc-700"
        }`}
      >
        {isEditMode ? <Save size={18} /> : <Plus size={18} />}
        {getSubmitText()}
      </button>

      <button
        type="button"
        onClick={onClose}
        className={`px-6 ${paddingY} bg-zinc-200 dark:bg-zinc-800 text-zinc-700 dark:text-zinc-300 ${rounded} hover:bg-zinc-300 dark:hover:bg-zinc-700 font-semibold`}
      >
        {cancelButtonText || "Cancel"}
      </button>

      {isEditMode && onDelete && (
        <button
          type="button"
          onClick={onDelete}
          className={`flex items-center gap-2 px-6 ${paddingY} ${rounded} font-semibold ml-auto ${
            isDeleting
              ? "bg-red-600 text-white hover:bg-red-700"
              : "bg-red-100 dark:bg-red-900/30 text-red-600 dark:text-red-400 hover:bg-red-200 dark:hover:bg-red-900/50"
          }`}
        >
          <Trash2 size={18} />
          {isDeleting ? "Confirm Delete" : "Delete"}
        </button>
      )}
    </div>
  );
};

export default FormActions;

