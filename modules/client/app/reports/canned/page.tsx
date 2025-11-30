"use client";

import { useRouter } from "next/navigation";
import { Stack, Text, Badge, Group } from "@mantine/core";
import { BarChart3, Building2, Users, StickyNote, FolderKanban, LucideIcon } from "lucide-react";
import { useProjectContext } from "../../../context/projectContext";
import { DataTable, type PageData, type Column } from "../../../components/DataTable/DataTable";

interface CannedReport {
  id: string;
  title: string;
  description: string;
  href: string;
  icon: LucideIcon;
  projectFiltered: boolean;
}

const cannedReports: CannedReport[] = [
  {
    id: "lead-pipeline",
    title: "Lead Pipeline",
    description: "Lead counts by status, conversion rates, and source analysis",
    href: "/reports/canned/lead-pipeline",
    icon: BarChart3,
    projectFiltered: true,
  },
  {
    id: "account-overview",
    title: "Account Overview",
    description: "Organization and individual counts, industry and geographic distribution",
    href: "/reports/canned/account-overview",
    icon: Building2,
    projectFiltered: false,
  },
  {
    id: "contact-coverage",
    title: "Contact Coverage",
    description: "Contact density per organization, decision-maker ratios, coverage gaps",
    href: "/reports/canned/contact-coverage",
    icon: Users,
    projectFiltered: false,
  },
  {
    id: "notes-activity",
    title: "Notes Activity",
    description: "Notes distribution by entity type, recent activity, and coverage metrics",
    href: "/reports/canned/notes-activity",
    icon: StickyNote,
    projectFiltered: true,
  },
];

const CannedReportsPage = () => {
  const router = useRouter();
  const { projectState } = useProjectContext();
  const { selectedProject } = projectState;

  const data: PageData<CannedReport> = {
    items: cannedReports,
    currentPage: 1,
    totalPages: 1,
    total: cannedReports.length,
    pageSize: cannedReports.length,
  };

  const columns: Column<CannedReport>[] = [
    {
      header: "",
      width: 50,
      render: (report) => {
        const Icon = report.icon;
        return <Icon className="h-5 w-5 text-lime-600 dark:text-lime-400" />;
      },
    },
    {
      header: "Name",
      render: (report) => (
        <Text fw={500}>{report.title}</Text>
      ),
    },
    {
      header: "Description",
      render: (report) => (
        <Text c="dimmed" size="sm">
          {report.description}
        </Text>
      ),
    },
    {
      header: "Scope",
      width: 100,
      render: (report) =>
        report.projectFiltered ? (
          <Badge size="sm" variant="light" color={selectedProject ? "lime" : "gray"}>
            {selectedProject ? "Project" : "Global"}
          </Badge>
        ) : (
          <Badge size="sm" variant="light" color="gray">
            Global
          </Badge>
        ),
    },
  ];

  const handleRowClick = (report: CannedReport) => {
    router.push(report.href);
  };

  return (
    <Stack gap="lg">
      <Group justify="space-between">
        <h1 className="text-xl font-bold text-zinc-900 dark:text-zinc-100">
          Canned Reports
        </h1>
        {selectedProject ? (
          <Badge variant="light" color="lime" leftSection={<FolderKanban className="h-3 w-3" />}>
            {selectedProject.name}
          </Badge>
        ) : (
          <Badge variant="light" color="gray">
            Global
          </Badge>
        )}
      </Group>
      <Text c="dimmed" size="sm">
        Pre-built reports for quick insights into your data.
        {selectedProject
          ? " Reports marked with a project badge will be filtered to the selected project."
          : " Select a project to see project-specific data in supported reports."}
      </Text>

      <DataTable
        data={data}
        columns={columns}
        onRowClick={handleRowClick}
        rowKey={(report) => report.id}
        emptyText="No canned reports available"
      />
    </Stack>
  );
};

export default CannedReportsPage;
