// Tiny local helper to concat classes safely
const cx = (...parts: Array<string | false | null | undefined>) => {
  return parts.filter(Boolean).join(" ");
};

export default cx
