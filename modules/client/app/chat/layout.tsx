import { AuthenticatedLayout } from "../../components/AuthenticatedLayout/AuthenticatedLayout";

<<<<<<< HEAD
export default ({ children }: { children: React.ReactNode }) => {
  return <AuthenticatedLayout>{children}</AuthenticatedLayout>;
};
=======
import { useState } from "react";
import Link from "next/link";
import Image from "next/image";
import {
  ChevronRight,
  ChevronDown,
  PanelLeftClose,
  PanelLeft,
} from "lucide-react";
import { useUser } from "@auth0/nextjs-auth0/client";
import { useUserContext } from "../../context/userContext";

const ChatLayout = ({
  children,
}: {
  children: React.ReactNode;
}) => {
  const [leadsExpanded, setLeadsExpanded] = useState(true);
  const [sidebarCollapsed, setSidebarCollapsed] = useState(true);
  const { isLoading } = useUser();
  const { userState } = useUserContext();

  // Show loading state while Auth0 or user data is loading
  if (isLoading || !userState.isAuthed) {
    return (
      <div className="flex min-h-screen items-center justify-center">
        <Image src="/icon.png" alt="Loading" width={200} height={200} />
      </div>
    );
  }

  return (
    <div className="flex min-h-screen">
      {/* Sidebar */}
      <aside
        className={`border-r border-zinc-200 bg-zinc-50 transition-all duration-300 ${
          sidebarCollapsed ? "w-14" : "w-64"
        }`}
      >
        <div className="flex h-full flex-col p-4">
          {/* Toggle Button */}
          <div className="mb-4 flex justify-end">
            <button
              onClick={() => setSidebarCollapsed(!sidebarCollapsed)}
              className="flex items-center justify-center rounded p-2 text-zinc-700 hover:bg-zinc-100"
              title={sidebarCollapsed ? "Expand sidebar" : "Collapse sidebar"}
            >
              {sidebarCollapsed ? (
                <PanelLeft className="h-5 w-5" />
              ) : (
                <PanelLeftClose className="h-5 w-5" />
              )}
            </button>
          </div>

          {!sidebarCollapsed && (
            <nav className="space-y-2">
              {/* Leads Section */}
              <div>
                <button
                  onClick={() => setLeadsExpanded(!leadsExpanded)}
                  className="flex w-full items-center gap-2 rounded px-3 py-2 text-sm font-medium text-zinc-700 hover:bg-zinc-100"
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
                      className="block rounded px-3 py-2 text-sm text-zinc-600 hover:bg-zinc-100 hover:text-zinc-900"
                    >
                      Jobs
                    </Link>
                  </div>
                )}
              </div>
            </nav>
          )}
        </div>
      </aside>

      {/* Main Content */}
      <main className="flex-1">{children}</main>
    </div>
  );
};

export default ChatLayout;
>>>>>>> 6ad842a46d25e1578737259ee70a54dccfbf1861
