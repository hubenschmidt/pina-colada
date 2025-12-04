"use client";

import { useEffect, useState } from "react";
import { useParams, useRouter } from "next/navigation";
import { Center, Stack, Loader } from "@mantine/core";
import { usePageLoading } from "../../../../context/pageLoadingContext";
import LeadForm from "../../../../components/LeadTracker/LeadForm";
import { useLeadFormConfig } from "../../../../components/LeadTracker/hooks/useLeadFormConfig";
import { getJob, updateJob, deleteJob } from "../../../../api";

const JobDetailPage = () => {
  const params = useParams();
  const router = useRouter();
  const { dispatchPageLoading } = usePageLoading();
  const formConfig = useLeadFormConfig("job");
  const id = params.id;

  const [job, setJob] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  useEffect(() => {
    const fetchJob = async () => {
      try {
        const data = await getJob(id);
        setJob(data);
      } catch (err) {
        setError(err instanceof Error ? err.message : "Failed to load job");
      } finally {
        setLoading(false);
        dispatchPageLoading({ type: "SET_PAGE_LOADING", payload: false });
      }
    };

    fetchJob();
  }, [id, dispatchPageLoading]);

  const handleClose = () => {
    router.push("/leads/jobs");
  };

  const handleUpdate = async (jobId, updates) => {
    await updateJob(jobId, updates);
    router.push("/leads/jobs");
  };

  const handleDelete = async (jobId) => {
    await deleteJob(jobId);
    router.push("/leads/jobs");
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

  if (error || !job) {
    return (
      <div className="p-6">
        <p className="text-red-600 dark:text-red-400">
          {error || "Job not found"}
        </p>
      </div>
    );
  }

  return (
    <LeadForm
      onClose={handleClose}
      lead={job}
      onUpdate={handleUpdate}
      onDelete={handleDelete}
      config={formConfig}
    />
  );
};

export default JobDetailPage;
