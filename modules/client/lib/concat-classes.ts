// Concat classes safely, filtering out falsy values
const cx = (...parts: Array<string | false | null | undefined>) =>
  parts.filter(Boolean).join(" ");

export default cx;
