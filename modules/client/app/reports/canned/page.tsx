"use client";

import Link from "next/link";
import { Stack, Card, Text, SimpleGrid } from "@mantine/core";
import { BarChart3, Building2, Users } from "lucide-react";

const cannedReports = [
  {
    title: "Lead Pipeline",
    description: "Lead counts by status, conversion rates, and source analysis",
    href: "/reports/canned/lead-pipeline",
    icon: BarChart3,
  },
  {
    title: "Account Overview",
    description: "Organization and individual counts, industry and geographic distribution",
    href: "/reports/canned/account-overview",
    icon: Building2,
  },
  {
    title: "Contact Coverage",
    description: "Contact density per organization, decision-maker ratios, coverage gaps",
    href: "/reports/canned/contact-coverage",
    icon: Users,
  },
];

const CannedReportsPage = () => {
  return (
    <Stack gap="lg">
      <h1 className="text-xl font-bold text-zinc-900 dark:text-zinc-100">
        Canned Reports
      </h1>
      <Text c="dimmed" size="sm">
        Pre-built reports for quick insights into your data
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
                <report.icon className="h-8 w-8 text-lime-600 dark:text-lime-400" />
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
