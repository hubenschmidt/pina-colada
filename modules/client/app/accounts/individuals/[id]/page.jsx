"use client";

import { useEffect, useState } from "react";
import { useParams, useRouter } from "next/navigation";
import { Center, Stack, Loader } from "@mantine/core";
import { usePageLoading } from "../../../../context/pageLoadingContext";
import AccountForm from "../../../../components/AccountForm/AccountForm";
import { getIndividual, updateIndividual, deleteIndividual } from "../../../../api";

const IndividualDetailPage = () => {
  const params = useParams();
  const router = useRouter();
  const { dispatchPageLoading } = usePageLoading();
  const id = Number(params.id);

  const [individual, setIndividual] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  useEffect(() => {
    const fetchIndividual = async () => {
      try {
        const data = await getIndividual(id);
        setIndividual(data);
      } catch (err) {
        setError(err instanceof Error ? err.message : "Failed to load individual");
      } finally {
        setLoading(false);
        dispatchPageLoading({ type: "SET_PAGE_LOADING", payload: false });
      }
    };

    fetchIndividual();
  }, [id, dispatchPageLoading]);

  const handleClose = () => {
    router.push("/accounts/individuals");
  };

  const handleUpdate = async (indId, updates) => {
    await updateIndividual(indId, updates);
    router.push("/accounts/individuals");
  };

  const handleDelete = async (indId) => {
    await deleteIndividual(indId);
    router.push("/accounts/individuals");
  };

  if (loading) {
    return (
      <Center mih={400}>
        <Stack align="center" gap="md">
          <Loader size="xl" color="lime" />
        </Stack>
      </Center>);

  }

  if (error || !individual) {
    return (
      <div className="p-6">
        <p className="text-red-600 dark:text-red-400">{error || "Individual not found"}</p>
      </div>);

  }

  return (
    <AccountForm
      type="individual"
      onClose={handleClose}
      account={individual}
      onUpdate={handleUpdate}
      onDelete={handleDelete} />);


};

export default IndividualDetailPage;