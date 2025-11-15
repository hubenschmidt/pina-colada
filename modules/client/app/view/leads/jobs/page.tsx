"use client";

import { useRef } from "react";
import JobTracker, { JobTrackerRef } from "../../../../components/JobTracker/JobTracker";
import { Plus } from "lucide-react";

const JobsPage = () => {
  const jobTrackerRef = useRef<JobTrackerRef>(null);

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-2xl font-semibold text-zinc-900">Job Leads</h1>
        <button
          onClick={() => jobTrackerRef.current?.openJobForm()}
          className="flex items-center justify-center w-10 h-10 bg-black text-white rounded-lg hover:bg-zinc-800"
          aria-label="Add new job application"
        >
          <Plus size={20} />
        </button>
      </div>
      <JobTracker ref={jobTrackerRef} />
    </div>
  );
};

export default JobsPage;
