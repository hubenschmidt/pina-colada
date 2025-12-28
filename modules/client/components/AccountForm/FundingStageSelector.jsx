"use client";

import { useEffect } from "react";
import { useLookupsContext } from "../../context/lookupsContext";
import { getFundingStages } from "../../api";
import { fetchOnce } from "../../lib/lookup-cache";

const FundingStageSelector = ({ value, onChange }) => {
  const { lookupsState, dispatchLookups } = useLookupsContext();
  const { fundingStages, loaded } = lookupsState;

  useEffect(() => {
    if (loaded.fundingStages) return;
    fetchOnce("fundingStages", getFundingStages).then((data) => {
      dispatchLookups({ type: "SET_FUNDING_STAGES", payload: data });
    });
  }, [loaded.fundingStages, dispatchLookups]);

  if (!loaded.fundingStages) {
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
      <option value="">Select funding stage...</option>
      {fundingStages.map((stage) => (
        <option key={stage.id} value={stage.id} className="text-zinc-900 dark:text-zinc-100">
          {stage.label}
        </option>
      ))}
    </select>
  );
};

export default FundingStageSelector;
