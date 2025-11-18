"use client";

import { useState } from "react";
import Link from "next/link";
import {
  ChevronRight,
  ChevronDown,
  PanelLeftClose,
  PanelLeft,
  Settings,
} from "lucide-react";
import { Stack, ScrollArea } from "@mantine/core";
import { useNavContext } from "../../context/navContext";
import { SET_SIDEBAR_COLLAPSED } from "../../reducers/navReducer";

export const Sidebar = () => {
  const [leadsExpanded, setLeadsExpanded] = useState(true);
  const { navState, dispatchNav } = useNavContext();
  const { sidebarCollapsed } = navState;

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
                type: SET_SIDEBAR_COLLAPSED,
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
                      className="block rounded px-3 py-2 text-sm text-zinc-600 dark:text-zinc-400 hover:bg-zinc-100 dark:hover:bg-zinc-800 hover:text-zinc-900 dark:hover:text-zinc-100"
                    >
                      Jobs
                    </Link>
                  </div>
                )}
              </div>
            </nav>
          )}
        </ScrollArea>

        {/* Settings Button at Bottom */}
        <div className="sticky bottom-0 bg-zinc-50 dark:bg-zinc-900 p-4 border-t border-zinc-200 dark:border-zinc-700">
          <Link
            href="/settings"
            className={`flex items-center rounded text-sm text-zinc-700 dark:text-zinc-300 hover:bg-zinc-100 dark:hover:bg-zinc-800 ${
              sidebarCollapsed
                ? "justify-center p-3"
                : "gap-3 px-3 py-2"
            }`}
            title="Settings"
          >
            <Settings className="h-5 w-5 flex-shrink-0" />
            {!sidebarCollapsed && <span>Settings</span>}
          </Link>
        </div>
      </Stack>
    </aside>
  );
};
