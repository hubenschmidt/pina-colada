"use client";

import { useRouter } from "next/navigation";
import { usePageLoading } from "../../../../context/pageLoadingContext";
import ContactForm from "../../../../components/ContactForm";
import { createContact, Contact } from "../../../../api";
import { useEffect } from "react";

const NewContactPage = () => {
  const router = useRouter();
  const { dispatchPageLoading } = usePageLoading();

  useEffect(() => {
    dispatchPageLoading({ type: "SET_PAGE_LOADING", payload: false });
  }, [dispatchPageLoading]);

  const handleClose = () => {
    router.push("/accounts/contacts");
  };

  const handleSave = async (data: Partial<Contact>) => {
    await createContact(data);
    router.push("/accounts/contacts");
  };

  return <ContactForm onSave={handleSave} onClose={handleClose} />;
};

export default NewContactPage;
