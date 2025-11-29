"use client";

import Link from "next/link";
import { Stack, Card, Text, SimpleGrid, Badge, Group } from "@mantine/core";
import { BarChart3, Building2, Users, StickyNote, FolderKanban } from "lucide-react";
import { useProjectContext } from "../../../context/projectContext";

const cannedReports = [
  {
    title: "Lead Pipeline",
    description: "Lead counts by status, conversion rates, and source analysis",
    href: "/reports/canned/lead-pipeline",
    icon: BarChart3,
    projectFiltered: true,
  },
  {
    title: "Account Overview",
    description: "Organization and individual counts, industry and geographic distribution",
    href: "/reports/canned/account-overview",
    icon: Building2,
    projectFiltered: false,
  },
  {
    title: "Contact Coverage",
    description: "Contact density per organization, decision-maker ratios, coverage gaps",
    href: "/reports/canned/contact-coverage",
    icon: Users,
    projectFiltered: false,
  },
  {
    title: "Notes Activity",
    description: "Notes distribution by entity type, recent activity, and coverage metrics",
    href: "/reports/canned/notes-activity",
    icon: StickyNote,
    projectFiltered: true,
  },
];

const CannedReportsPage = () => {
  const { projectState } = useProjectContext();
  const { selectedProject } = projectState;

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

      <SimpleGrid cols={{ base: 1, sm: 2, lg: 3 }} spacing="lg">
        {cannedReports.map((report) => (
          <Link key={report.href} href={report.href}>
            <Card
              shadow="sm"
              padding="lg"
              radius="md"
              withBorder
              className="hover:border-lime-500 dark:hover:border-lime-400 transition-colors cursor-pointer h-full"
            >
              <Stack gap="sm">
                <Group justify="space-between">
                  <report.icon className="h-8 w-8 text-lime-600 dark:text-lime-400" />
                  {report.projectFiltered && (
                    <Badge size="xs" variant="light" color={selectedProject ? "lime" : "gray"}>
                      {selectedProject ? "Project" : "Global"}
                    </Badge>
                  )}
                </Group>
                <Text fw={600} size="lg">
                  {report.title}
                </Text>
                <Text c="dimmed" size="sm">
                  {report.description}
                </Text>
              </Stack>
            </Card>
          </Link>
        ))}
      </SimpleGrid>
    </Stack>
  );
};

export default CannedReportsPage;
