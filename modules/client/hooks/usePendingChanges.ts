interface UsePendingChangesOptions {
  original: Record<string, unknown> | null;
  current: Record<string, unknown>;
  pendingDeletions?: unknown[];
  trackFields?: string[];
}

/**
 * Hook to detect if form has unsaved changes.
 * Compares current form data against original, and checks for pending deletions.
 */
export const usePendingChanges = ({
  original,
  current,
  pendingDeletions = [],
  trackFields,
}: UsePendingChangesOptions): boolean => {
  if (pendingDeletions.length > 0) return true;
  if (!original) return false;

  const fields = trackFields ?? Object.keys(current);
  return fields.some((key) => {
    const origVal = original[key] ?? "";
    const currVal = current[key] ?? "";
    return origVal !== currVal;
  });
};
