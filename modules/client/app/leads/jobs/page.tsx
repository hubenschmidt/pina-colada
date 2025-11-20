"use client";

import { useEffect } from "react";
import LeadTracker from "../../../components/LeadTracker/LeadTracker";
import { useLeadConfig } from "../../../components/config";
import { usePageLoading } from "../../../context/pageLoadingContext";

const JobsPage = () => {
  const jobConfig = useLeadConfig("job");
  const { dispatchPageLoading } = usePageLoading();

  useEffect(() => {
    dispatchPageLoading({ type: "SET_PAGE_LOADING", payload: false });
  }, [dispatchPageLoading]);

  return <LeadTracker config={jobConfig} />;
};

export default JobsPage;
