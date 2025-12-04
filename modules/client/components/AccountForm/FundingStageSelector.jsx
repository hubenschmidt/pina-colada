"use client";

import { useState, useEffect } from "react";
import { getFundingStages } from "../../api";






const FundingStageSelector = ({ value, onChange }) => {
  const [stages, setStages] = useState([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const loadStages = async () => {
      try {
        const data = await getFundingStages();
        setStages(data);
      } catch (error) {
        console.error("Failed to fetch funding stages:", error);
      } finally {
        setLoading(false);
      }
    };
    loadStages();
  }, []);

  if (loading) {
    return (
      <div className="w-full px-3 py-2 border border-zinc-300 dark:border-zinc-700 rounded bg-zinc-100 dark:bg-zinc-800 text-zinc-500 dark:text-zinc-400">
        Loading...
      </div>);

  }

  return (
    <select
      value={value ?? ""}
      onChange={(e) => onChange(e.target.value ? Number(e.target.value) : null)}
      className={`w-full px-3 py-2 border border-zinc-300 dark:border-zinc-700 rounded focus:outline-none focus:ring-2 focus:ring-lime-500 bg-white dark:bg-zinc-800 ${value ? "text-zinc-900 dark:text-zinc-100" : "text-zinc-500 dark:text-zinc-400"}`}>

      <option value="">Select funding stage...</option>
      {stages.map((stage) =>
      <option key={stage.id} value={stage.id} className="text-zinc-900 dark:text-zinc-100">
          {stage.label}
        </option>
      )}
    </select>);

};

export default FundingStageSelector;