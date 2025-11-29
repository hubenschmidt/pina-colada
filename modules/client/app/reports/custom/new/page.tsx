"use client";

import { useRouter } from "next/navigation";
import { Stack } from "@mantine/core";
import { ReportBuilder } from "../../../../components/Reports/ReportBuilder";
import { createSavedReport, ReportQueryRequest } from "../../../../api";

const NewReportPage = () => {
  const router = useRouter();

  const handleSave = async (name: string, description: string, query: ReportQueryRequest) => {
    await createSavedReport({
      name,
      description: description || undefined,
      query_definition: query,
    });
    router.push("/reports/custom");
  };

  return (
    <Stack gap="lg">
      <h1 className="text-xl font-bold text-zinc-900 dark:text-zinc-100">
        New Custom Report
      </h1>
      <ReportBuilder onSave={handleSave} />
    </Stack>
  );
};

export default NewReportPage;
