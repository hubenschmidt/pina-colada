"use client";

import { useEffect } from "react";
import LeadTracker from "../../../components/LeadTracker/LeadTracker";
import { useLeadTrackerConfig } from "../../../components/LeadTracker/hooks/useLeadTrackerConfig";
import { usePageLoading } from "../../../context/pageLoadingContext";

const PartnershipsPage = () => {
  const partnershipConfig = useLeadTrackerConfig("partnership");
  const { dispatchPageLoading } = usePageLoading();

  useEffect(() => {
    dispatchPageLoading({ type: "SET_PAGE_LOADING", payload: false });
  }, [dispatchPageLoading]);

  return <LeadTracker config={partnershipConfig} />;
};

export default PartnershipsPage;