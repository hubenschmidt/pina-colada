"use client";

import { useEffect, useState } from "react";
import { useParams, useRouter } from "next/navigation";
import { Center, Stack, Loader } from "@mantine/core";
import { usePageLoading } from "../../../../context/pageLoadingContext";
import LeadForm from "../../../../components/LeadTracker/LeadForm";
import { useLeadFormConfig } from "../../../../components/LeadTracker/hooks/useLeadFormConfig";
import { getPartnership, updatePartnership, deletePartnership } from "../../../../api";

const PartnershipDetailPage = () => {
  const params = useParams();
  const router = useRouter();
  const { dispatchPageLoading } = usePageLoading();
  const formConfig = useLeadFormConfig("partnership");
  const id = params.id;

  const [partnership, setPartnership] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  useEffect(() => {
    const fetchPartnership = async () => {
      try {
        const data = await getPartnership(id);
        setPartnership(data);
      } catch (err) {
        setError(err instanceof Error ? err.message : "Failed to load partnership");
      } finally {
        setLoading(false);
        dispatchPageLoading({ type: "SET_PAGE_LOADING", payload: false });
      }
    };

    fetchPartnership();
  }, [id, dispatchPageLoading]);

  const handleClose = () => {
    router.push("/leads/partnerships");
  };

  const handleUpdate = async (partnershipId, updates) => {
    await updatePartnership(partnershipId, updates);
    router.push("/leads/partnerships");
  };

  const handleDelete = async (partnershipId) => {
    await deletePartnership(partnershipId);
    router.push("/leads/partnerships");
  };

  if (loading) {
    return (
      <Center mih={400}>
        <Stack align="center" gap="md">
          <Loader size="xl" color="lime" />
        </Stack>
      </Center>);

  }

  if (error || !partnership) {
    return (
      <div className="p-6">
        <p className="text-red-600 dark:text-red-400">{error || "Partnership not found"}</p>
      </div>);

  }

  return (
    <LeadForm
      onClose={handleClose}
      lead={partnership}
      onUpdate={handleUpdate}
      onDelete={handleDelete}
      config={formConfig} />);


};

export default PartnershipDetailPage;