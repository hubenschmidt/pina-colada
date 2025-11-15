"use client";

import { useState } from "react";
import Link from "next/link";
import Image from "next/image";
import { ChevronRight, ChevronDown } from "lucide-react";
import { useUser } from "@auth0/nextjs-auth0/client";
import { useUserContext } from "../../context/userContext";

export default function ViewLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  const [leadsExpanded, setLeadsExpanded] = useState(true);
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
      <aside className="w-64 border-r border-zinc-200 bg-zinc-50 p-4">
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
                  href="/view/leads/jobs"
                  className="block rounded px-3 py-2 text-sm text-zinc-600 hover:bg-zinc-100 hover:text-zinc-900"
                >
                  Jobs
                </Link>
              </div>
            )}
          </div>
        </nav>
      </aside>

      {/* Main Content */}
      <main className="flex-1 p-8">{children}</main>
    </div>
  );
}
