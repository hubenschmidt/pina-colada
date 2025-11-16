"use client";

import LeadTracker from "../../../components/LeadTracker/LeadTracker";
import { useLeadConfig } from "../../../components/config";

const JobsPage = () => {
  const jobConfig = useLeadConfig("job");

  return (
    <div>
      <LeadTracker config={jobConfig} />
    </div>
  );
};

export default JobsPage;
