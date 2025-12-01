"use client";

import { useRouter } from "next/navigation";
import LeadForm from "../../../../components/LeadTracker/LeadForm";
import { useLeadFormConfig } from "../../../../components/LeadTracker/hooks/useLeadFormConfig";
import { createPartnership } from "../../../../api";

const NewPartnershipPage = () => {
  const router = useRouter();
  const formConfig = useLeadFormConfig("partnership");

  const handleClose = () => {
    router.push("/leads");
  };

  const handleAdd = async (data: any) => {
    const partnership = await createPartnership(data);
    router.push("/leads");
    return partnership;
  };

  return (
    <LeadForm
      onClose={handleClose}
      onAdd={handleAdd}
      config={formConfig}
    />
  );
};

export default NewPartnershipPage;
