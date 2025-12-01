"use client";

import { useState, useEffect } from "react";
import { getEmployeeCountRanges, EmployeeCountRange } from "../../api";

interface EmployeeCountRangeSelectorProps {
  value: number | null;
  onChange: (value: number | null) => void;
}

const EmployeeCountRangeSelector = ({ value, onChange }: EmployeeCountRangeSelectorProps) => {
  const [ranges, setRanges] = useState<EmployeeCountRange[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const loadRanges = async () => {
      try {
        const data = await getEmployeeCountRanges();
        setRanges(data);
      } catch (error) {
        console.error("Failed to fetch employee count ranges:", error);
      } finally {
        setLoading(false);
      }
    };
    loadRanges();
  }, []);

  if (loading) {
    return (
      <div className="w-full px-3 py-2 border border-zinc-300 dark:border-zinc-700 rounded bg-zinc-100 dark:bg-zinc-800 text-zinc-500 dark:text-zinc-400">
        Loading...
      </div>
    );
  }

  return (
    <select
      value={value ?? ""}
      onChange={(e) => onChange(e.target.value ? Number(e.target.value) : null)}
      className={`w-full px-3 py-2 border border-zinc-300 dark:border-zinc-700 rounded focus:outline-none focus:ring-2 focus:ring-lime-500 bg-white dark:bg-zinc-800 ${value ? "text-zinc-900 dark:text-zinc-100" : "text-zinc-500 dark:text-zinc-400"}`}
    >
      <option value="">Select employee count...</option>
      {ranges.map((range) => (
        <option key={range.id} value={range.id} className="text-zinc-900 dark:text-zinc-100">
          {range.label}
        </option>
      ))}
    </select>
  );
};

export default EmployeeCountRangeSelector;
