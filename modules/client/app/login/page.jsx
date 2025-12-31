"use client";

import { useEffect } from "react";
import { useRouter } from "next/navigation";
import { ChevronRight, ChevronLeft, Loader2 } from "lucide-react";
import { useUserContext } from "../../context/userContext";
import Header from "../../components/Header/Header";

const LoginPage = () => {
  const { userState } = useUserContext();
  const { user, isLoading, isAuthed, tenantName } = userState;
  const router = useRouter();

  useEffect(() => {
    if (isLoading) return;
    if (!isAuthed) return;

    if (tenantName) {
      router.push("/chat");
      return;
    }
    router.push("/tenant/select");
  }, [isLoading, isAuthed, tenantName, router]);

  if (isLoading || isAuthed) {
    return (
      <div className="min-h-screen bg-zinc-800 flex items-center justify-center">
        <Loader2 className="h-8 w-8 animate-spin text-lime-500" />
      </div>
    );
  }

  return (
    <>
      <Header />
      <div className="min-h-screen bg-zinc-800 flex items-center justify-center p-4">
        <div className="flex flex-col items-center gap-6 max-w-md w-full">
          <div className="text-center mb-4">
            <h1 className="text-4xl font-bold text-transparent bg-clip-text bg-gradient-to-r from-lime-500 via-yellow-400 to-lime-500 mb-2">
              Demo AI-CRM
            </h1>
          </div>

          <div className="w-full bg-zinc-700 rounded-2xl shadow-xl p-8 border border-zinc-600">
            <div className="flex flex-col gap-4">
              <a
                href="/auth/login"
                className="inline-flex items-center justify-center gap-2 rounded-lg bg-gradient-to-r from-lime-500 to-yellow-400 px-6 py-3 text-sm font-semibold text-zinc-900 hover:brightness-95 transition-all">
                {/* <ChevronLeft size={18} /> */}
                Log in
              </a>

              {/* <a
                href="/auth/login?screen_hint=signup"
                className="inline-flex items-center justify-center gap-2 px-6 py-3 text-sm font-semibold text-zinc-700 bg-zinc-50 border-2 border-zinc-200 rounded-lg hover:border-zinc-300 hover:bg-zinc-100 transition-colors">
                Sign up
                <ChevronRight size={18} />
              </a> */}
            </div>

            <div className="mt-6 pt-6 border-t border-zinc-600">
              <p className="text-xs text-center text-zinc-400">
                You agree to our{" "}
                <a href="/terms" className="underline hover:text-zinc-200">
                  Terms of Service
                </a>{" "}
                and{" "}
                <a href="/privacy" className="underline hover:text-zinc-200">
                  Privacy Policy
                </a>
                .
              </p>
            </div>
          </div>

          <div className="mt-8 text-center">
            <p className="text-sm text-zinc-400">
              © {new Date().getFullYear()} PinaColada.co. All rights reserved.
            </p>
          </div>
        </div>
      </div>
    </>
  );
};

export default LoginPage;
