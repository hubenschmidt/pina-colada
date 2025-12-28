"use client";

import { useEffect } from "react";
import { useRouter } from "next/navigation";
import LeadForm from "../../../../components/LeadTracker/LeadForm";
import { useLeadFormConfig } from "../../../../components/LeadTracker/hooks/useLeadFormConfig";
import { usePageLoading } from "../../../../context/pageLoadingContext";
import { createJob } from "../../../../api";

const NewJobPage = () => {
  const router = useRouter();
  const formConfig = useLeadFormConfig("job");
  const { dispatchPageLoading } = usePageLoading();

  useEffect(() => {
    dispatchPageLoading({ type: "SET_PAGE_LOADING", payload: false });
  }, [dispatchPageLoading]);

  const handleClose = () => {
    router.push("/leads/jobs");
  };

  const handleAdd = async (jobData) => {
    const job = await createJob(jobData);
    router.push("/leads/jobs");
    return job;
  };

  return <LeadForm onClose={handleClose} onAdd={handleAdd} config={formConfig} />;
};

export default NewJobPage;
