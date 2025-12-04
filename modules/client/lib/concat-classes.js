// Concat classes safely, filtering out falsy values
const cx = (...parts) =>
  parts.filter(Boolean).join(" ");

export default cx;
