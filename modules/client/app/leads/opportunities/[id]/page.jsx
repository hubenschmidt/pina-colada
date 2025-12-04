"use client";

import { useEffect, useState } from "react";
import { useParams, useRouter } from "next/navigation";
import { Center, Stack, Loader } from "@mantine/core";
import { usePageLoading } from "../../../../context/pageLoadingContext";
import LeadForm from "../../../../components/LeadTracker/LeadForm";
import { useLeadFormConfig } from "../../../../components/LeadTracker/hooks/useLeadFormConfig";
import {
  getOpportunity,
  updateOpportunity,
  deleteOpportunity,
} from "../../../../api";

const OpportunityDetailPage = () => {
  const params = useParams();
  const router = useRouter();
  const { dispatchPageLoading } = usePageLoading();
  const formConfig = useLeadFormConfig("opportunity");
  const id = params.id;

  const [opportunity, setOpportunity] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  useEffect(() => {
    const fetchOpportunity = async () => {
      try {
        const data = await getOpportunity(id);
        setOpportunity(data);
      } catch (err) {
        setError(
          err instanceof Error ? err.message : "Failed to load opportunity",
        );
      } finally {
        setLoading(false);
        dispatchPageLoading({ type: "SET_PAGE_LOADING", payload: false });
      }
    };

    fetchOpportunity();
  }, [id, dispatchPageLoading]);

  const handleClose = () => {
    router.push("/leads/opportunities");
  };

  const handleUpdate = async (opportunityId, updates) => {
    await updateOpportunity(opportunityId, updates);
    router.push("/leads/opportunities");
  };

  const handleDelete = async (opportunityId) => {
    await deleteOpportunity(opportunityId);
    router.push("/leads/opportunities");
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

  if (error || !opportunity) {
    return (
      <div className="p-6">
        <p className="text-red-600 dark:text-red-400">
          {error || "Opportunity not found"}
        </p>
      </div>
    );
  }

  return (
    <LeadForm
      onClose={handleClose}
      lead={opportunity}
      onUpdate={handleUpdate}
      onDelete={handleDelete}
      config={formConfig}
    />
  );
};

export default OpportunityDetailPage;
