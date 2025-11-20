"use client";

import { useEffect } from "react";
import Image from "next/image";
import { useUser } from "@auth0/nextjs-auth0/client";
import { usePathname, useRouter } from "next/navigation";
import PageLoadingOverlay from "../components/PageLoadingOverlay";
import { useUserContext } from "../context/userContext";

export const RootLayoutClient = ({
  children,
}: {
  children: React.ReactNode;
}) => {
  const { user, isLoading } = useUser();
  const pathname = usePathname();
  const router = useRouter();
  const { userState } = useUserContext();

  // Centralized route protection and redirects
  useEffect(() => {
    if (isLoading) return;
    if (!userState.isAuthed) return;

    // Redirect authenticated users from home to appropriate page
    if (pathname === "/") {
      if (userState.tenantName) {
        router.push("/chat");
      } else {
        router.push("/tenant/select");
      }
      return;
    }

    // Redirect from tenant/select to chat if already has tenant
    if (pathname === "/tenant/select" && userState.tenantName) {
      router.push("/chat");
      return;
    }

    // Protect routes: redirect to /tenant/select if no tenantName
    const protectedRoutes = ["/chat", "/leads/jobs"];
    const isProtectedRoute = protectedRoutes.some((route) =>
      pathname.startsWith(route)
    );

    if (isProtectedRoute && !userState.tenantName) {
      router.push("/tenant/select");
    }
  }, [userState.isAuthed, userState.tenantName, pathname, router, isLoading]);

  // Show loading while auth is loading or authenticated user is on "/" (about to redirect)
  const isRedirecting = pathname === "/" && user;

  if (isLoading || isRedirecting) {
    return (
      <div className="flex min-h-screen items-center justify-center">
        <Image src="/icon.png" alt="Loading" width={200} height={200} />
      </div>
    );
  }

  return <PageLoadingOverlay>{children}</PageLoadingOverlay>;
};
