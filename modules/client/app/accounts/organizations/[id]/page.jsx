"use client";

import { useEffect, useState } from "react";
import { useParams, useRouter } from "next/navigation";
import { Center, Stack, Loader } from "@mantine/core";
import { usePageLoading } from "../../../../context/pageLoadingContext";
import AccountForm from "../../../../components/AccountForm/AccountForm";
import { getOrganization, updateOrganization, deleteOrganization } from "../../../../api";

const OrganizationDetailPage = () => {
  const params = useParams();
  const router = useRouter();
  const { dispatchPageLoading } = usePageLoading();
  const id = Number(params.id);

  const [organization, setOrganization] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  useEffect(() => {
    const fetchOrganization = async () => {
      try {
        const data = await getOrganization(id);
        setOrganization(data);
      } catch (err) {
        setError(err instanceof Error ? err.message : "Failed to load organization");
      } finally {
        setLoading(false);
        dispatchPageLoading({ type: "SET_PAGE_LOADING", payload: false });
      }
    };

    fetchOrganization();
  }, [id, dispatchPageLoading]);

  const handleClose = () => {
    router.push("/accounts/organizations");
  };

  const handleUpdate = async (orgId, updates) => {
    await updateOrganization(orgId, updates);
    router.push("/accounts/organizations");
  };

  const handleDelete = async (orgId) => {
    await deleteOrganization(orgId);
    router.push("/accounts/organizations");
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

  if (error || !organization) {
    return (
      <div className="p-6">
        <p className="text-red-600 dark:text-red-400">{error || "Organization not found"}</p>
      </div>
    );
  }

  return (
    <AccountForm
      type="organization"
      onClose={handleClose}
      account={organization}
      onUpdate={handleUpdate}
      onDelete={handleDelete}
    />
  );
};

export default OrganizationDetailPage;
