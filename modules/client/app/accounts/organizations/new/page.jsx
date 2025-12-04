"use client";

import { useRouter } from "next/navigation";
import AccountForm from "../../../../components/AccountForm/AccountForm";
import { createOrganization } from "../../../../api";

const NewOrganizationPage = () => {
  const router = useRouter();

  const handleClose = () => {
    router.push("/accounts/organizations");
  };

  const handleAdd = async (data) => {
    const created = await createOrganization(data);
    router.push("/accounts/organizations");
    return created;
  };

  return (
    <AccountForm
      type="organization"
      onClose={handleClose}
      onAdd={handleAdd} />);


};

export default NewOrganizationPage;