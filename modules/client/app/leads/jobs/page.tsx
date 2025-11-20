"use client";

import LeadTracker from "../../../components/LeadTracker/LeadTracker";
import { useLeadConfig } from "../../../components/config";

const JobsPage = () => {
  const jobConfig = useLeadConfig("job");

  return <LeadTracker config={jobConfig} />;
};

export default JobsPage;
