/**
 * Format phone number as user types.
 * Formats to +1-XXX-XXX-XXXX for US numbers.
 */
export const formatPhoneNumber = (value: string): string => {
  const digits = value.replace(/\D/g, "");

  // Remove leading 1 if present (we'll add +1 prefix)
  const nationalDigits = digits.startsWith("1") ? digits.slice(1) : digits;

  if (nationalDigits.length === 0) return "";
  if (nationalDigits.length <= 3) return `+1-${nationalDigits}`;
  if (nationalDigits.length <= 6) return `+1-${nationalDigits.slice(0, 3)}-${nationalDigits.slice(3)}`;

  return `+1-${nationalDigits.slice(0, 3)}-${nationalDigits.slice(3, 6)}-${nationalDigits.slice(6, 10)}`;
};
