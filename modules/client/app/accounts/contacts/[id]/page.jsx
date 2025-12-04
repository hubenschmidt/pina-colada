"use client";

import { useEffect, useState } from "react";
import { useParams, useRouter } from "next/navigation";
import { Center, Stack, Loader } from "@mantine/core";
import { usePageLoading } from "../../../../context/pageLoadingContext";
import ContactForm from "../../../../components/ContactForm/ContactForm";
import { getContact, updateContact, deleteContact } from "../../../../api";

const ContactDetailPage = () => {
  const params = useParams();
  const router = useRouter();
  const { dispatchPageLoading } = usePageLoading();
  const id = Number(params.id);

  const [contact, setContact] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  useEffect(() => {
    const fetchContact = async () => {
      try {
        const data = await getContact(id);
        setContact(data);
      } catch (err) {
        setError(err instanceof Error ? err.message : "Failed to load contact");
      } finally {
        setLoading(false);
        dispatchPageLoading({ type: "SET_PAGE_LOADING", payload: false });
      }
    };

    fetchContact();
  }, [id, dispatchPageLoading]);

  const handleClose = () => {
    router.push("/accounts/contacts");
  };

  const handleSave = async (data) => {
    await updateContact(id, data);
    router.push("/accounts/contacts");
  };

  const handleDelete = async () => {
    await deleteContact(id);
    router.push("/accounts/contacts");
  };

  if (loading) {
    return (
      <Center mih={400}>
        <Stack align="center" gap="md">
          <Loader size="xl" color="lime" />
        </Stack>
      </Center>
    );
  }

  if (error || !contact) {
    return (
      <div className="p-6">
        <p className="text-red-600 dark:text-red-400">{error || "Contact not found"}</p>
      </div>
    );
  }

  return (
    <ContactForm
      contact={contact}
      onSave={handleSave}
      onDelete={handleDelete}
      onClose={handleClose}
    />
  );
};

export default ContactDetailPage;
