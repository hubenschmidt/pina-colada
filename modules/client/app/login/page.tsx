"use client";

import { useUser } from "@auth0/nextjs-auth0/client";
import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { ChevronRight, ChevronLeft } from "lucide-react";
import { checkUserTenant } from "../../api";
import Header from "../../components/Header/Header";

const LoginPage = () => {
  const { user, isLoading } = useUser();
  const router = useRouter();
  const [checkingTenant, setCheckingTenant] = useState(false);

  // If already authenticated, check for tenant and redirect
  useEffect(() => {
    if (!isLoading && user) {
      setCheckingTenant(true);

      checkUserTenant(user)
        .then((tenant) => {
          if (tenant && tenant.id) {
            router.push("/chat");
            return;
          }
          router.push("/tenant/select");
        })
        .catch((error) => {
          // On error (404 or no tenant), default to tenant selection
          router.push("/tenant/select");
        })
        .finally(() => {
          setCheckingTenant(false);
        });
    }
  }, [user, isLoading, router]);

  // Show loading state while checking auth or tenant
  if (isLoading || checkingTenant) {
    return null;
  }

  // Already authenticated (should be redirecting)
  if (user) {
    return null;
  }

  // Unauthenticated: show login UI
  return (
    <>
      <Header />
      <div className="min-h-screen bg-gradient-to-br from-blue-50 to-blue-100 flex items-center justify-center p-4">
      <div className="flex flex-col items-center gap-6 max-w-md w-full">
        <div className="text-center mb-4">
          <h1 className="text-4xl font-bold text-transparent bg-clip-text bg-gradient-to-r from-lime-500 via-yellow-400 to-lime-500 mb-2">
            PinaColada
          </h1>
        </div>

        <div className="w-full bg-white rounded-2xl shadow-xl p-8 border border-blue-100">
          <div className="flex flex-col gap-4">
            {/* Login button */}
            <a
              href="/auth/login"
              className="inline-flex items-center justify-center gap-2 rounded-lg bg-gradient-to-r from-lime-500 to-yellow-400 px-6 py-3 text-sm font-semibold text-blue-900 hover:brightness-95 transition-all"
            >
              <ChevronLeft size={18} />
              Log in
            </a>

            {/* Sign up button */}
            <a
              href="/auth/login?screen_hint=signup"
              className="inline-flex items-center justify-center gap-2 px-6 py-3 text-sm font-semibold text-blue-700 bg-blue-50 border-2 border-blue-200 rounded-lg hover:border-blue-300 hover:bg-blue-100 transition-colors"
            >
              Sign up
              <ChevronRight size={18} />
            </a>
          </div>

          <div className="mt-6 pt-6 border-t border-blue-100">
            <p className="text-xs text-center text-blue-600">
              By continuing, you agree to our{" "}
              <a href="/terms" className="underline hover:text-blue-800">
                Terms of Service
              </a>{" "}
              and{" "}
              <a href="/privacy" className="underline hover:text-blue-800">
                Privacy Policy
              </a>
              .
            </p>
          </div>
        </div>

        {/* Footer */}
        <div className="mt-8 text-center">
          <p className="text-sm text-blue-600">
            © {new Date().getFullYear()} PinaColada.co. All rights reserved.
          </p>
        </div>
        </div>
      </div>
    </>
  );
};

export default LoginPage;
