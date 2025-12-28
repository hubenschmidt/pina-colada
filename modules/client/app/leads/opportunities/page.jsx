"use client";

import { useEffect } from "react";
import LeadTracker from "../../../components/LeadTracker/LeadTracker";
import { useLeadTrackerConfig } from "../../../components/LeadTracker/hooks/useLeadTrackerConfig";
import { usePageLoading } from "../../../context/pageLoadingContext";

const OpportunitiesPage = () => {
  const opportunityConfig = useLeadTrackerConfig("opportunity");
  const { dispatchPageLoading } = usePageLoading();

  useEffect(() => {
    dispatchPageLoading({ type: "SET_PAGE_LOADING", payload: false });
  }, [dispatchPageLoading]);

  return <LeadTracker config={opportunityConfig} />;
};

export default OpportunitiesPage;
