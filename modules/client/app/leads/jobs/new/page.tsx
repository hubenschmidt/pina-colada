"use client";

import { useRouter } from "next/navigation";
import LeadForm from "../../../../components/LeadTracker/LeadForm";
import { useLeadFormConfig } from "../../../../components/LeadTracker/hooks/useLeadFormConfig";
import { createJob } from "../../../../api";

const NewJobPage = () => {
  const router = useRouter();
  const formConfig = useLeadFormConfig("job");

  const handleClose = () => {
    router.push("/leads/jobs");
  };

  const handleAdd = async (jobData: any) => {
    await createJob(jobData);
    router.push("/leads/jobs");
  };

  return (
    <LeadForm
      onClose={handleClose}
      onAdd={handleAdd}
      config={formConfig}
    />
  );
};

export default NewJobPage;
