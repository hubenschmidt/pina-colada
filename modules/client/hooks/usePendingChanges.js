function _nullishCoalesce(lhs, rhsFn) {
  if (lhs != null) {
    return lhs;
  } else {
    return rhsFn();
  }
}

/**
 * Hook to detect if form has unsaved changes.
 * Compares current form data against original, and checks for pending deletions.
 */
export const usePendingChanges = ({ original, current, pendingDeletions = [], trackFields }) => {
  if (pendingDeletions.length > 0) return true;
  if (!original) return false;

  const fields = _nullishCoalesce(trackFields, () => Object.keys(current));
  return fields.some((key) => {
    const origVal = _nullishCoalesce(original[key], () => "");
    const currVal = _nullishCoalesce(current[key], () => "");
    return origVal !== currVal;
  });
};
