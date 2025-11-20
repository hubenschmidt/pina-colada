"use client";

import { useUser } from "@auth0/nextjs-auth0/client";
// import LeadTracker from '../../components/LeadTracker/LeadTracker' # disable for now but keep this here
import { JobTracker } from "../../components/JobTracker/JobTracker";
import { useLeadConfig } from "../../components/config";
import { LogOut, Loader } from "lucide-react";
import Header from "../../components/Header";

const JobsPage = () => {
  const { user, isLoading } = useUser();
  const jobConfig = useLeadConfig("job");

  if (isLoading) {
    return (
      <div className="min-h-screen bg-blue-50 flex items-center justify-center">
        <Loader size={48} className="text-blue-600 animate-spin" />
      </div>
    );
  }

  // Middleware should redirect to /login if not authenticated
  // But show a fallback just in case
  if (!user) {
    return (
      <div className="min-h-screen bg-blue-50 flex items-center justify-center">
        <p className="text-zinc-600">Redirecting to login...</p>
      </div>
    );
  }

  // disable LeadTracker but do not delete it
  return (
    <>
      <Header />
      <main className="min-h-screen bg-blue-50 py-8 px-4">
      <div className="max-w-6xl mx-auto">
        <div className="flex justify-end mb-4">
          <a
            href="/auth/logout"
            className="flex items-center gap-2 px-4 py-2 bg-white border border-zinc-300 rounded-lg text-zinc-700 hover:bg-zinc-50 transition-colors"
          >
            <LogOut size={18} />
            Logout
          </a>
        </div>
        <JobTracker />
        {/* <LeadTracker config={jobConfig} /> */}
        </div>
      </main>
    </>
  );
};

export default JobsPage;
