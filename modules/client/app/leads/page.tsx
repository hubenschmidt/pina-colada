"use client";

import { useEffect, useState } from "react";
import { Tabs, Badge, Text, Group, Button, TextInput } from "@mantine/core";
import { Briefcase, Handshake, Target, DollarSign, Plus, Search } from "lucide-react";
import { useProjectContext } from "../../context/projectContext";
import { usePageLoading } from "../../context/pageLoadingContext";
import { getProjectLeads, getProjectDeals, ProjectLead, ProjectDeal } from "../../api";
import { DataTable, Column, PageData } from "../../components/DataTable/DataTable";
import Link from "next/link";

const EmptyCell = () => <span className="text-zinc-400">—</span>;

const renderDate = (value: string | null | undefined) =>
  value ? new Date(value).toLocaleDateString() : <EmptyCell />;

// Helper to wrap array data in PageData format for DataTable
const toPageData = <T,>(items: T[]): PageData<T> => ({
  items,
  currentPage: 1,
  totalPages: 1,
  total: items.length,
});

const leadColumns: Column<ProjectLead>[] = [
  {
    header: "Title",
    accessor: "title",
    sortable: true,
    sortKey: "title",
    width: "25%",
  },
  {
    header: "Account",
    accessor: "account_name",
    sortable: true,
    sortKey: "account_name",
    width: "20%",
    render: (lead) => lead.account_name || <EmptyCell />,
  },
  {
    header: "Status",
    accessor: "current_status",
    sortable: true,
    sortKey: "current_status",
    width: "15%",
    render: (lead) => lead.current_status || <EmptyCell />,
  },
  {
    header: "Source",
    accessor: "source",
    sortable: true,
    sortKey: "source",
    width: "15%",
    render: (lead) => lead.source || <EmptyCell />,
  },
  {
    header: "Created",
    accessor: "created_at",
    sortable: true,
    sortKey: "created_at",
    width: "15%",
    render: (lead) => renderDate(lead.created_at),
  },
];

const dealColumns: Column<ProjectDeal>[] = [
  {
    header: "Name",
    accessor: "name",
    sortable: true,
    sortKey: "name",
    width: "30%",
  },
  {
    header: "Status",
    accessor: "current_status",
    sortable: true,
    sortKey: "current_status",
    width: "20%",
    render: (deal) => deal.current_status || <EmptyCell />,
  },
  {
    header: "Value",
    accessor: "value_amount",
    sortable: true,
    sortKey: "value_amount",
    width: "20%",
    render: (deal) => deal.value_amount ? `$${deal.value_amount.toLocaleString()}` : <EmptyCell />,
  },
  {
    header: "Created",
    accessor: "created_at",
    sortable: true,
    sortKey: "created_at",
    width: "15%",
    render: (deal) => renderDate(deal.created_at),
  },
];

const LeadsPage = () => {
  const { projectState } = useProjectContext();
  const { selectedProject } = projectState;
  const { dispatchPageLoading } = usePageLoading();

  const [activeTab, setActiveTab] = useState<string | null>("jobs");
  const [leads, setLeads] = useState<ProjectLead[]>([]);
  const [deals, setDeals] = useState<ProjectDeal[]>([]);
  const [loading, setLoading] = useState(false);
  const [searchQuery, setSearchQuery] = useState("");

  useEffect(() => {
    dispatchPageLoading({ type: "SET_PAGE_LOADING", payload: false });
  }, [dispatchPageLoading]);

  useEffect(() => {
    if (!selectedProject) {
      setLeads([]);
      setDeals([]);
      return;
    }

    const fetchData = async () => {
      setLoading(true);
      try {
        const [leadsData, dealsData] = await Promise.all([
          getProjectLeads(selectedProject.id),
          getProjectDeals(selectedProject.id),
        ]);
        setLeads(leadsData);
        setDeals(dealsData);
      } catch (error) {
        console.error("Failed to fetch leads/deals:", error);
      } finally {
        setLoading(false);
      }
    };

    fetchData();
  }, [selectedProject]);

  // Filter leads by search query
  const filterLeads = (items: ProjectLead[]) => {
    if (!searchQuery.trim()) return items;
    const query = searchQuery.toLowerCase();
    return items.filter(
      (item) =>
        item.title?.toLowerCase().includes(query) ||
        item.account_name?.toLowerCase().includes(query) ||
        item.description?.toLowerCase().includes(query)
    );
  };

  const filterDeals = (items: ProjectDeal[]) => {
    if (!searchQuery.trim()) return items;
    const query = searchQuery.toLowerCase();
    return items.filter(
      (item) =>
        item.name?.toLowerCase().includes(query) ||
        item.description?.toLowerCase().includes(query)
    );
  };

  const jobs = filterLeads(leads.filter((l) => l.type === "Job"));
  const opportunities = filterLeads(leads.filter((l) => l.type === "Opportunity"));
  const partnerships = filterLeads(leads.filter((l) => l.type === "Partnership"));
  const filteredDeals = filterDeals(deals);

  // Clear search when switching tabs
  const handleTabChange = (value: string | null) => {
    setActiveTab(value);
    setSearchQuery("");
  };

  // Get placeholder based on active tab
  const getSearchPlaceholder = () => {
    switch (activeTab) {
      case "jobs": return "Search jobs...";
      case "opportunities": return "Search opportunities...";
      case "partnerships": return "Search partnerships...";
      case "deals": return "Search deals...";
      default: return "Search...";
    }
  };

  if (!selectedProject) {
    return (
      <div className="p-6">
        <Text c="dimmed" ta="center" py="xl">
          Select a project to view leads and deals
        </Text>
      </div>
    );
  }

  return (
    <div className="p-6">
      <Group justify="space-between" mb="md">
        <Text size="xl" fw={600}>
          {selectedProject.name} - Leads & Deals
        </Text>
        <TextInput
          placeholder={getSearchPlaceholder()}
          leftSection={<Search size={16} />}
          value={searchQuery}
          onChange={(e) => setSearchQuery(e.currentTarget.value)}
          w={300}
        />
      </Group>

      <Tabs value={activeTab} onChange={handleTabChange}>
        <Tabs.List>
          <Tabs.Tab
            value="jobs"
            leftSection={<Briefcase size={16} />}
            rightSection={<Badge size="sm" variant="light" color="blue">{jobs.length}</Badge>}
          >
            Jobs
          </Tabs.Tab>
          <Tabs.Tab
            value="opportunities"
            leftSection={<Target size={16} />}
            rightSection={<Badge size="sm" variant="light" color="green">{opportunities.length}</Badge>}
          >
            Opportunities
          </Tabs.Tab>
          <Tabs.Tab
            value="partnerships"
            leftSection={<Handshake size={16} />}
            rightSection={<Badge size="sm" variant="light" color="purple">{partnerships.length}</Badge>}
          >
            Partnerships
          </Tabs.Tab>
          <Tabs.Tab
            value="deals"
            leftSection={<DollarSign size={16} />}
            rightSection={<Badge size="sm" variant="light" color="orange">{deals.length}</Badge>}
          >
            Deals
          </Tabs.Tab>
        </Tabs.List>

        <Tabs.Panel value="jobs" pt="md">
          <Group justify="flex-end" mb="sm">
            <Button
              component={Link}
              href="/leads/jobs/new"
              leftSection={<Plus size={16} />}
              size="sm"
            >
              Add Job
            </Button>
          </Group>
          {loading ? (
            <Text c="dimmed">Loading...</Text>
          ) : jobs.length === 0 ? (
            <Text c="dimmed" ta="center" py="xl">
              {searchQuery ? "No jobs match your search" : "No jobs in this project yet"}
            </Text>
          ) : (
            <DataTable
              data={toPageData(jobs)}
              columns={leadColumns}
              onRowClick={(job) => window.location.href = `/leads/jobs/${job.id}`}
            />
          )}
        </Tabs.Panel>

        <Tabs.Panel value="opportunities" pt="md">
          {loading ? (
            <Text c="dimmed">Loading...</Text>
          ) : opportunities.length === 0 ? (
            <Text c="dimmed" ta="center" py="xl">
              {searchQuery ? "No opportunities match your search" : "No opportunities in this project yet"}
            </Text>
          ) : (
            <DataTable
              data={toPageData(opportunities)}
              columns={leadColumns}
              onRowClick={(opp) => window.location.href = `/leads/opportunities/${opp.id}`}
            />
          )}
        </Tabs.Panel>

        <Tabs.Panel value="partnerships" pt="md">
          {loading ? (
            <Text c="dimmed">Loading...</Text>
          ) : partnerships.length === 0 ? (
            <Text c="dimmed" ta="center" py="xl">
              {searchQuery ? "No partnerships match your search" : "No partnerships in this project yet"}
            </Text>
          ) : (
            <DataTable
              data={toPageData(partnerships)}
              columns={leadColumns}
              onRowClick={(p) => window.location.href = `/leads/partnerships/${p.id}`}
            />
          )}
        </Tabs.Panel>

        <Tabs.Panel value="deals" pt="md">
          {loading ? (
            <Text c="dimmed">Loading...</Text>
          ) : filteredDeals.length === 0 ? (
            <Text c="dimmed" ta="center" py="xl">
              {searchQuery ? "No deals match your search" : "No deals in this project yet"}
            </Text>
          ) : (
            <DataTable
              data={toPageData(filteredDeals)}
              columns={dealColumns}
              onRowClick={(deal) => window.location.href = `/deals/${deal.id}`}
            />
          )}
        </Tabs.Panel>
      </Tabs>
    </div>
  );
};

export default LeadsPage;
