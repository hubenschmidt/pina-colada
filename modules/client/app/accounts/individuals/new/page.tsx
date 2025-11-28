"use client";

import { useRouter } from "next/navigation";
import AccountForm from "../../../../components/AccountForm/AccountForm";
import { createIndividual } from "../../../../api";

const NewIndividualPage = () => {
  const router = useRouter();

  const handleClose = () => {
    router.push("/accounts/individuals");
  };

  const handleAdd = async (data: any) => {
    const created = await createIndividual(data);
    router.push("/accounts/individuals");
    return created;
  };

  return (
    <AccountForm
      type="individual"
      onClose={handleClose}
      onAdd={handleAdd}
    />
  );
};

export default NewIndividualPage;
