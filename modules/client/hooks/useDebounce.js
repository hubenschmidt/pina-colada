import { useRef, useCallback, useEffect } from "react";

export const DEBOUNCE_MS = {
  SEARCH: 300,
  PREVIEW: 500,
};

export const useDebounce = (fn, delay = DEBOUNCE_MS.SEARCH) => {
  const timeoutRef = useRef(null);

  const cleanup = useCallback(() => {
    if (timeoutRef.current) {
      clearTimeout(timeoutRef.current);
      timeoutRef.current = null;
    }
  }, []);

  useEffect(() => cleanup, [cleanup]);

  const debounced = useCallback(
    (...args) => {
      cleanup();
      timeoutRef.current = setTimeout(() => fn(...args), delay);
    },
    [fn, delay, cleanup]
  );

  debounced.cancel = cleanup;

  return debounced;
};
