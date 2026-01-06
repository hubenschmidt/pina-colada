import { useState } from "react";
import { Select } from "@mantine/core";

export const StatusSelect = ({ value, options, colors = {}, onUpdate, field = "status" }) => {
  const [localValue, setLocalValue] = useState(value);
  const [loading, setLoading] = useState(false);

  const handleChange = async (newValue) => {
    if (!newValue || newValue === localValue || !onUpdate) return;

    setLoading(true);
    setLocalValue(newValue);

    try {
      await onUpdate(field, newValue);
    } catch {
      setLocalValue(value);
    } finally {
      setLoading(false);
    }
  };

  return (
    <Select
      size="xs"
      value={localValue}
      data={options}
      onChange={handleChange}
      disabled={loading || !onUpdate}
      allowDeselect={false}
      classNames={{
        input: `text-xs font-medium rounded border ${colors[localValue] || "bg-gray-100 text-gray-800"}`,
      }}
    />
  );
};
