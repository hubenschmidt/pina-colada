"use client";

import { useEffect } from "react";
import { useRouter } from "next/navigation";
import LeadForm from "../../../../components/LeadTracker/LeadForm";
import { useLeadFormConfig } from "../../../../components/LeadTracker/hooks/useLeadFormConfig";
import { usePageLoading } from "../../../../context/pageLoadingContext";
import { createPartnership } from "../../../../api";

const NewPartnershipPage = () => {
  const router = useRouter();
  const formConfig = useLeadFormConfig("partnership");
  const { dispatchPageLoading } = usePageLoading();

  useEffect(() => {
    dispatchPageLoading({ type: "SET_PAGE_LOADING", payload: false });
  }, [dispatchPageLoading]);

  const handleClose = () => {
    router.push("/leads/partnerships");
  };

  const handleAdd = async (data) => {
    const partnership = await createPartnership(data);
    router.push("/leads/partnerships");
    return partnership;
  };

  return <LeadForm onClose={handleClose} onAdd={handleAdd} config={formConfig} />;
};

export default NewPartnershipPage;
