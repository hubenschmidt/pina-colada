"use client";

import { useEffect, useState } from "react";
import { useRouter, useParams } from "next/navigation";
import { Stack, Center, Loader, Text } from "@mantine/core";
import { ReportBuilder } from "../../../../components/Reports/ReportBuilder";
import {
  getSavedReport,
  updateSavedReport,
  SavedReport,
  ReportQueryRequest,
} from "../../../../api";

const EditReportPage = () => {
  const router = useRouter();
  const params = useParams();
  const reportId = Number(params.id);

  const [report, setReport] = useState<SavedReport | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchReport = async () => {
      try {
        const data = await getSavedReport(reportId);
        setReport(data);
      } catch (err) {
        setError(err instanceof Error ? err.message : "Failed to load report");
      } finally {
        setLoading(false);
      }
    };
    fetchReport();
  }, [reportId]);

  const handleSave = async (name: string, description: string, query: ReportQueryRequest) => {
    await updateSavedReport(reportId, {
      name,
      description: description || undefined,
      query_definition: query,
    });
    router.push("/reports/custom");
  };

  if (loading) {
    return (
      <Center mih={400}>
        <Stack align="center" gap="md">
          <Loader size="xl" color="lime" />
          <Text c="dimmed">Loading report...</Text>
        </Stack>
      </Center>
    );
  }

  if (error || !report) {
    return (
      <Stack gap="lg">
        <h1 className="text-xl font-bold text-zinc-900 dark:text-zinc-100">
          Edit Report
        </h1>
        <Text c="red">{error || "Report not found"}</Text>
      </Stack>
    );
  }

  return (
    <Stack gap="lg">
      <h1 className="text-xl font-bold text-zinc-900 dark:text-zinc-100">
        Edit Report: {report.name}
      </h1>
      <ReportBuilder
        initialQuery={report.query_definition}
        reportName={report.name}
        reportDescription={report.description || ""}
        onSave={handleSave}
      />
    </Stack>
  );
};

export default EditReportPage;
