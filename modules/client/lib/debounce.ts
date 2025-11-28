export const debounce = <T extends unknown[]>(
  fn: (...args: T) => void,
  wait: number
): ((...args: T) => void) => {
  let timeoutId: ReturnType<typeof setTimeout> | null = null;

  return (...args: T) => {
    if (timeoutId) clearTimeout(timeoutId);
    timeoutId = setTimeout(() => fn(...args), wait);
  };
};
