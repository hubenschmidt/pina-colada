"use client";

import { useState } from "react";
import Link from "next/link";
import {
  ChevronRight,
  ChevronDown,
  PanelLeftClose,
  PanelLeft,
  Building2,
  Users,
  Contact,
  Briefcase,
  BarChart3,
  Wrench,
  FolderKanban,
  Check,
} from "lucide-react";
import { Stack, ScrollArea, Select } from "@mantine/core";
import { useNavContext } from "../../context/navContext";
import { useProjectContext } from "../../context/projectContext";

export const Sidebar = () => {
  const [leadsExpanded, setLeadsExpanded] = useState(true);
  const [accountsExpanded, setAccountsExpanded] = useState(true);
  const [reportsExpanded, setReportsExpanded] = useState(true);
  const { navState, dispatchNav } = useNavContext();
  const { sidebarCollapsed } = navState;
  const { projectState, selectProject } = useProjectContext();
  const { projects, selectedProject } = projectState;

  return (
    <aside
      className={`h-screen border-r border-zinc-200 dark:border-zinc-700 bg-zinc-50 dark:bg-zinc-900 transition-all duration-300 ${
        sidebarCollapsed ? "w-14" : "w-64"
      }`}
    >
      <Stack h="100%" gap={0} justify="space-between">
        {/* Toggle Button */}
        <div className="p-4 pb-0 flex justify-end">
          <button
            onClick={() =>
              dispatchNav({
                type: "SET_SIDEBAR_COLLAPSED",
                payload: !sidebarCollapsed,
              })
            }
            className="flex items-center justify-center rounded p-2 text-zinc-700 dark:text-zinc-300 hover:bg-zinc-100 dark:hover:bg-zinc-800"
            title={sidebarCollapsed ? "Expand sidebar" : "Collapse sidebar"}
          >
            {sidebarCollapsed ? (
              <PanelLeft className="h-5 w-5" />
            ) : (
              <PanelLeftClose className="h-5 w-5" />
            )}
          </button>
        </div>

        {/* Navigation Links */}
        <ScrollArea flex={1} type="auto">
          {!sidebarCollapsed && (
            <nav className="p-4 space-y-2">
              {/* Projects Section */}
              <div className="mb-4">
                <Link
                  href="/projects"
                  className="flex items-center gap-2 rounded px-3 py-2 text-sm font-medium text-zinc-700 dark:text-zinc-300 hover:bg-zinc-100 dark:hover:bg-zinc-800"
                >
                  <FolderKanban className="h-4 w-4" />
                  Projects
                </Link>
                {projects.filter((p) => p.status !== "Inactive").length > 0 && (
                  <div className="mt-2 px-3">
                    <Select
                      size="xs"
                      placeholder="Select project..."
                      value={selectedProject?.id?.toString() || null}
                      onChange={(value) => {
                        const project = projects.find((p) => p.id.toString() === value);
                        selectProject(project || null);
                      }}
                      data={projects
                        .filter((p) => p.status !== "Inactive")
                        .map((p) => ({
                          value: p.id.toString(),
                          label: p.name,
                        }))}
                      clearable
                      comboboxProps={{ withinPortal: false }}
                    />
                  </div>
                )}
              </div>

              {/* Accounts Section */}
              <div>
                <button
                  onClick={() => setAccountsExpanded(!accountsExpanded)}
                  className="flex w-full items-center gap-2 rounded px-3 py-2 text-sm font-medium text-zinc-700 dark:text-zinc-300 hover:bg-zinc-100 dark:hover:bg-zinc-800"
                >
                  {accountsExpanded ? (
                    <ChevronDown className="h-4 w-4" />
                  ) : (
                    <ChevronRight className="h-4 w-4" />
                  )}
                  Accounts
                </button>
                {accountsExpanded && (
                  <div className="ml-6 mt-1 space-y-1">
                    <Link
                      href="/accounts/organizations"
                      className="flex items-center gap-2 rounded px-3 py-2 text-sm text-zinc-600 dark:text-zinc-400 hover:bg-zinc-100 dark:hover:bg-zinc-800 hover:text-zinc-900 dark:hover:text-zinc-100"
                    >
                      <Building2 className="h-4 w-4" />
                      Organizations
                    </Link>
                    <Link
                      href="/accounts/individuals"
                      className="flex items-center gap-2 rounded px-3 py-2 text-sm text-zinc-600 dark:text-zinc-400 hover:bg-zinc-100 dark:hover:bg-zinc-800 hover:text-zinc-900 dark:hover:text-zinc-100"
                    >
                      <Users className="h-4 w-4" />
                      Individuals
                    </Link>
                    <Link
                      href="/accounts/contacts"
                      className="flex items-center gap-2 rounded px-3 py-2 text-sm text-zinc-600 dark:text-zinc-400 hover:bg-zinc-100 dark:hover:bg-zinc-800 hover:text-zinc-900 dark:hover:text-zinc-100"
                    >
                      <Contact className="h-4 w-4" />
                      Contacts
                    </Link>
                  </div>
                )}
              </div>

              {/* Leads Section */}
              <div>
                <button
                  onClick={() => setLeadsExpanded(!leadsExpanded)}
                  className="flex w-full items-center gap-2 rounded px-3 py-2 text-sm font-medium text-zinc-700 dark:text-zinc-300 hover:bg-zinc-100 dark:hover:bg-zinc-800"
                >
                  {leadsExpanded ? (
                    <ChevronDown className="h-4 w-4" />
                  ) : (
                    <ChevronRight className="h-4 w-4" />
                  )}
                  Leads
                </button>
                {leadsExpanded && (
                  <div className="ml-6 mt-1 space-y-1">
                    <Link
                      href="/leads/jobs"
                      className="flex items-center gap-2 rounded px-3 py-2 text-sm text-zinc-600 dark:text-zinc-400 hover:bg-zinc-100 dark:hover:bg-zinc-800 hover:text-zinc-900 dark:hover:text-zinc-100"
                    >
                      <Briefcase className="h-4 w-4" />
                      Jobs
                    </Link>
                  </div>
                )}
              </div>

              {/* Reports Section */}
              <div>
                <button
                  onClick={() => setReportsExpanded(!reportsExpanded)}
                  className="flex w-full items-center gap-2 rounded px-3 py-2 text-sm font-medium text-zinc-700 dark:text-zinc-300 hover:bg-zinc-100 dark:hover:bg-zinc-800"
                >
                  {reportsExpanded ? (
                    <ChevronDown className="h-4 w-4" />
                  ) : (
                    <ChevronRight className="h-4 w-4" />
                  )}
                  Reports
                </button>
                {reportsExpanded && (
                  <div className="ml-6 mt-1 space-y-1">
                    <Link
                      href="/reports/canned"
                      className="flex items-center gap-2 rounded px-3 py-2 text-sm text-zinc-600 dark:text-zinc-400 hover:bg-zinc-100 dark:hover:bg-zinc-800 hover:text-zinc-900 dark:hover:text-zinc-100"
                    >
                      <BarChart3 className="h-4 w-4" />
                      Canned
                    </Link>
                    <Link
                      href="/reports/custom"
                      className="flex items-center gap-2 rounded px-3 py-2 text-sm text-zinc-600 dark:text-zinc-400 hover:bg-zinc-100 dark:hover:bg-zinc-800 hover:text-zinc-900 dark:hover:text-zinc-100"
                    >
                      <Wrench className="h-4 w-4" />
                      Custom
                    </Link>
                  </div>
                )}
              </div>
            </nav>
          )}
        </ScrollArea>
      </Stack>
    </aside>
  );
};
