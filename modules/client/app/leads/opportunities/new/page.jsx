"use client";

import { useEffect } from "react";
import { useRouter } from "next/navigation";
import LeadForm from "../../../../components/LeadTracker/LeadForm";
import { useLeadFormConfig } from "../../../../components/LeadTracker/hooks/useLeadFormConfig";
import { usePageLoading } from "../../../../context/pageLoadingContext";
import { createOpportunity } from "../../../../api";

const NewOpportunityPage = () => {
  const router = useRouter();
  const formConfig = useLeadFormConfig("opportunity");
  const { dispatchPageLoading } = usePageLoading();

  useEffect(() => {
    dispatchPageLoading({ type: "SET_PAGE_LOADING", payload: false });
  }, [dispatchPageLoading]);

  const handleClose = () => {
    router.push("/leads/opportunities");
  };

  const handleAdd = async (data) => {
    const opportunity = await createOpportunity(data);
    router.push("/leads/opportunities");
    return opportunity;
  };

  return <LeadForm onClose={handleClose} onAdd={handleAdd} config={formConfig} />;
};

export default NewOpportunityPage;
