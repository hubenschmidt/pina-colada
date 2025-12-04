"use client";

import { useEffect } from "react";
import Image from "next/image";
import { usePathname, useRouter } from "next/navigation";
import PageLoadingOverlay from "../components/PageLoadingOverlay/PageLoadingOverlay";
import { useUserContext } from "../context/userContext";

export const RootLayoutClient = ({ children }) => {
  const pathname = usePathname();
  const router = useRouter();
  const { userState } = useUserContext();
  const { user, isLoading, isAuthed, tenantName } = userState;

  // Centralized route protection and redirects
  useEffect(() => {
    if (isLoading) return;
    if (!isAuthed) return;

    // Redirect authenticated users from home to appropriate page
    if (pathname === "/") {
      if (tenantName) {
        router.push("/chat");
        return;
      }
      router.push("/tenant/select");
      return;
    }

    // Redirect from tenant/select to chat if already has tenant
    if (pathname === "/tenant/select" && tenantName) {
      router.push("/chat");
      return;
    }

    // Protect routes: redirect to /tenant/select if no tenantName
    const protectedRoutes = ["/chat", "/leads/jobs"];
    const isProtectedRoute = protectedRoutes.some((route) =>
      pathname.startsWith(route),
    );

    if (isProtectedRoute && !tenantName) {
      router.push("/tenant/select");
    }
  }, [isAuthed, tenantName, pathname, router, isLoading]);

  // Show loading while auth is loading or authenticated user is on "/" (about to redirect)
  const isRedirecting = pathname === "/" && user;

  if (isLoading || isRedirecting) {
    return (
      <div className="flex min-h-screen items-center justify-center">
        <Image
          src="/loading-icon.png"
          alt="Loading"
          width={200}
          height={200}
          priority
          unoptimized
        />
      </div>
    );
  }

  return <PageLoadingOverlay>{children}</PageLoadingOverlay>;
};
