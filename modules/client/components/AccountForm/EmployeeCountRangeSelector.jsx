"use client";

import { useEffect } from "react";
import { useLookupsContext } from "../../context/lookupsContext";
import { getEmployeeCountRanges } from "../../api";
import { fetchOnce } from "../../lib/lookup-cache";

const EmployeeCountRangeSelector = ({ value, onChange }) => {
  const { lookupsState, dispatchLookups } = useLookupsContext();
  const { employeeCountRanges, loaded } = lookupsState;

  useEffect(() => {
    if (loaded.employeeCountRanges) return;
    fetchOnce("employeeCountRanges", getEmployeeCountRanges).then((data) => {
      dispatchLookups({ type: "SET_EMPLOYEE_COUNT_RANGES", payload: data });
    });
  }, [loaded.employeeCountRanges, dispatchLookups]);

  if (!loaded.employeeCountRanges) {
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
      className={`w-full px-3 py-2 border border-zinc-300 dark:border-zinc-700 rounded focus:outline-none focus:ring-2 focus:ring-lime-500 bg-white dark:bg-zinc-800 ${value ? "text-zinc-900 dark:text-zinc-100" : "text-zinc-500 dark:text-zinc-400"}`}>
      <option value="">Select employee count...</option>
      {employeeCountRanges.map((range) => (
        <option key={range.id} value={range.id} className="text-zinc-900 dark:text-zinc-100">
          {range.label}
        </option>
      ))}
    </select>
  );
};

export default EmployeeCountRangeSelector;
