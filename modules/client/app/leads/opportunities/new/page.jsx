"use client";

import { useRouter } from "next/navigation";
import LeadForm from "../../../../components/LeadTracker/LeadForm";
import { useLeadFormConfig } from "../../../../components/LeadTracker/hooks/useLeadFormConfig";
import { createOpportunity } from "../../../../api";

const NewOpportunityPage = () => {
  const router = useRouter();
  const formConfig = useLeadFormConfig("opportunity");

  const handleClose = () => {
    router.push("/leads");
  };

  const handleAdd = async (data) => {
    const opportunity = await createOpportunity(data);
    router.push("/leads");
    return opportunity;
  };

  return (
    <LeadForm
      onClose={handleClose}
      onAdd={handleAdd}
      config={formConfig} />);


};

export default NewOpportunityPage;