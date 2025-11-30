"use client";

import { useState } from "react";
import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";
import {
  ChevronRight,
  ChevronDown,
  PanelLeftClose,
  PanelLeft,
  Building2,
  Briefcase,
  BarChart3,
  Wrench,
  FolderKanban,
  Lightbulb,
  Handshake,
  DollarSign,
  PenLine,
  User,
  FileText,
  Library,
  Clipboard,
  CheckSquare,
  Files,
} from "lucide-react";
import { Stack, ScrollArea, Select } from "@mantine/core";
import { useNavContext } from "../../context/navContext";
import { useProjectContext } from "../../context/projectContext";

export const Sidebar = () => {
  const [leadsExpanded, setLeadsExpanded] = useState(false);
  const [accountsExpanded, setAccountsExpanded] = useState(false);
  const [reportsExpanded, setReportsExpanded] = useState(false);
  const pathname = usePathname();
  const router = useRouter();
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
              <div>
                {projects.filter((p) => p.status !== "Inactive").length > 0 && (
                  <div className="mb-4 px-3">
                    <Select
                      size="xs"
                      placeholder="Select project..."
                      value={selectedProject?.id?.toString() || null}
                      onChange={(value) => {
                        const project = projects.find(
                          (p) => p.id.toString() === value
                        );
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
                <button
                  onClick={() => router.push("/projects")}
                  className={`flex w-full items-center gap-2 rounded px-3 py-2 text-sm text-zinc-600 dark:text-zinc-400 hover:bg-zinc-100 dark:hover:bg-zinc-800 ${
                    pathname === "/projects" ? "font-bold" : "font-normal"
                  }`}
                >
                  <FolderKanban className="h-4 w-4 text-lime-600 dark:text-lime-400" />
                  Projects
                </button>
              </div>

              {/* Accounts Section */}
              <div>
                <button
                  onClick={() => setAccountsExpanded(!accountsExpanded)}
                  className={`flex w-full items-center gap-2 rounded px-3 py-2 text-sm text-zinc-600 dark:text-zinc-400 hover:bg-zinc-100 dark:hover:bg-zinc-800 ${
                    pathname?.startsWith("/accounts")
                      ? "font-bold"
                      : "font-normal"
                  }`}
                >
                  <Library className="h-4 w-4 text-lime-600 dark:text-lime-400" />
                  Accounts
                  <span className="ml-auto">
                    {accountsExpanded ? (
                      <ChevronDown className="h-4 w-4" />
                    ) : (
                      <ChevronRight className="h-4 w-4" />
                    )}
                  </span>
                </button>
                {accountsExpanded && (
                  <div className="ml-6 mt-1 space-y-1">
                    <Link
                      href="/accounts/organizations"
                      className={`flex items-center gap-2 rounded px-3 py-2 text-sm text-zinc-600 dark:text-zinc-400 hover:bg-zinc-100 dark:hover:bg-zinc-800 hover:text-zinc-900 dark:hover:text-zinc-100 ${
                        pathname === "/accounts/organizations"
                          ? "font-bold"
                          : "font-normal"
                      }`}
                    >
                      <Building2 className="h-4 w-4 text-lime-600 dark:text-lime-400" />
                      Organizations
                    </Link>
                    <Link
                      href="/accounts/individuals"
                      className={`flex items-center gap-2 rounded px-3 py-2 text-sm text-zinc-600 dark:text-zinc-400 hover:bg-zinc-100 dark:hover:bg-zinc-800 hover:text-zinc-900 dark:hover:text-zinc-100 ${
                        pathname === "/accounts/individuals"
                          ? "font-bold"
                          : "font-normal"
                      }`}
                    >
                      <User className="h-4 w-4 text-lime-600 dark:text-lime-400" />
                      Individuals
                    </Link>
                    <Link
                      href="/accounts/contacts"
                      className={`flex items-center gap-2 rounded px-3 py-2 text-sm text-zinc-600 dark:text-zinc-400 hover:bg-zinc-100 dark:hover:bg-zinc-800 hover:text-zinc-900 dark:hover:text-zinc-100 ${
                        pathname === "/accounts/contacts"
                          ? "font-bold"
                          : "font-normal"
                      }`}
                    >
                      <PenLine className="h-4 w-4 text-lime-600 dark:text-lime-400" />
                      Contacts
                    </Link>
                  </div>
                )}
              </div>

              {/* Leads Section */}
              <div>
                <button
                  onClick={() => setLeadsExpanded(!leadsExpanded)}
                  className={`flex w-full items-center gap-2 rounded px-3 py-2 text-sm text-zinc-600 dark:text-zinc-400 hover:bg-zinc-100 dark:hover:bg-zinc-800 ${
                    pathname?.startsWith("/leads") ? "font-bold" : "font-normal"
                  }`}
                >
                  <Briefcase className="h-4 w-4 text-lime-600 dark:text-lime-400" />
                  Leads
                  <span className="ml-auto">
                    {leadsExpanded ? (
                      <ChevronDown className="h-4 w-4" />
                    ) : (
                      <ChevronRight className="h-4 w-4" />
                    )}
                  </span>
                </button>
                {leadsExpanded && (
                  <div className="ml-6 mt-1 space-y-1">
                    <Link
                      href="/leads/jobs"
                      className={`flex items-center gap-2 rounded px-3 py-2 text-sm text-zinc-600 dark:text-zinc-400 hover:bg-zinc-100 dark:hover:bg-zinc-800 hover:text-zinc-900 dark:hover:text-zinc-100 ${
                        pathname?.startsWith("/leads/jobs")
                          ? "font-bold"
                          : "font-normal"
                      }`}
                    >
                      <Clipboard className="h-4 w-4 text-lime-600 dark:text-lime-400" />
                      Jobs
                    </Link>
                    <Link
                      href="/leads/opportunities"
                      className={`flex items-center gap-2 rounded px-3 py-2 text-sm text-zinc-600 dark:text-zinc-400 hover:bg-zinc-100 dark:hover:bg-zinc-800 hover:text-zinc-900 dark:hover:text-zinc-100 ${
                        pathname?.startsWith("/leads/opportunities")
                          ? "font-bold"
                          : "font-normal"
                      }`}
                    >
                      <Lightbulb className="h-4 w-4 text-lime-600 dark:text-lime-400" />
                      Opportunities
                    </Link>
                    <Link
                      href="/leads/partnerships"
                      className={`flex items-center gap-2 rounded px-3 py-2 text-sm text-zinc-600 dark:text-zinc-400 hover:bg-zinc-100 dark:hover:bg-zinc-800 hover:text-zinc-900 dark:hover:text-zinc-100 ${
                        pathname?.startsWith("/leads/partnerships")
                          ? "font-bold"
                          : "font-normal"
                      }`}
                    >
                      <Handshake className="h-4 w-4 text-lime-600 dark:text-lime-400" />
                      Partnerships
                    </Link>
                  </div>
                )}
              </div>

              {/* Tasks */}
              <button
                onClick={() => router.push("/tasks")}
                className={`flex w-full items-center gap-2 rounded px-3 py-2 text-sm text-zinc-600 dark:text-zinc-400 hover:bg-zinc-100 dark:hover:bg-zinc-800 ${
                  pathname === "/tasks" ? "font-bold" : "font-normal"
                }`}
              >
                <CheckSquare className="h-4 w-4 text-lime-600 dark:text-lime-400" />
                Tasks
              </button>

              {/* Documents */}
              <button
                onClick={() => router.push("/assets/documents")}
                className={`flex w-full items-center gap-2 rounded px-3 py-2 text-sm text-zinc-600 dark:text-zinc-400 hover:bg-zinc-100 dark:hover:bg-zinc-800 ${
                  pathname === "/assets/documents" ? "font-bold" : "font-normal"
                }`}
              >
                <Files className="h-4 w-4 text-lime-600 dark:text-lime-400" />
                Documents
              </button>

              {/* Deals - disabled for now, will implement later */}
              {/* {selectedProject && (
                <button
                  onClick={() => router.push("/deals")}
                  className={`flex w-full items-center gap-2 rounded px-3 py-2 text-sm text-zinc-600 dark:text-zinc-400 hover:bg-zinc-100 dark:hover:bg-zinc-800 ${
                    pathname?.startsWith("/deals") ? "font-bold" : "font-normal"
                  }`}
                >
                  <DollarSign className="h-4 w-4 text-lime-600 dark:text-lime-400" />
                  Deals
                </button>
              )} */}

              {/* Reports Section */}
              <div>
                <button
                  onClick={() => setReportsExpanded(!reportsExpanded)}
                  className={`flex w-full items-center gap-2 rounded px-3 py-2 text-sm text-zinc-600 dark:text-zinc-400 hover:bg-zinc-100 dark:hover:bg-zinc-800 ${
                    pathname?.startsWith("/reports")
                      ? "font-bold"
                      : "font-normal"
                  }`}
                >
                  <BarChart3 className="h-4 w-4 text-lime-600 dark:text-lime-400" />
                  Reports
                  <span className="ml-auto">
                    {reportsExpanded ? (
                      <ChevronDown className="h-4 w-4" />
                    ) : (
                      <ChevronRight className="h-4 w-4" />
                    )}
                  </span>
                </button>
                {reportsExpanded && (
                  <div className="ml-6 mt-1 space-y-1">
                    <Link
                      href="/reports/canned"
                      className={`flex items-center gap-2 rounded px-3 py-2 text-sm text-zinc-600 dark:text-zinc-400 hover:bg-zinc-100 dark:hover:bg-zinc-800 hover:text-zinc-900 dark:hover:text-zinc-100 ${
                        pathname === "/reports/canned"
                          ? "font-bold"
                          : "font-normal"
                      }`}
                    >
                      <FileText className="h-4 w-4 text-lime-600 dark:text-lime-400" />
                      Canned
                    </Link>
                    <Link
                      href="/reports/custom"
                      className={`flex items-center gap-2 rounded px-3 py-2 text-sm text-zinc-600 dark:text-zinc-400 hover:bg-zinc-100 dark:hover:bg-zinc-800 hover:text-zinc-900 dark:hover:text-zinc-100 ${
                        pathname === "/reports/custom"
                          ? "font-bold"
                          : "font-normal"
                      }`}
                    >
                      <Wrench className="h-4 w-4 text-lime-600 dark:text-lime-400" />
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
