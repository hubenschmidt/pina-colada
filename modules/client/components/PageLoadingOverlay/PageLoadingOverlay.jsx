"use client";

import Image from "next/image";
import { usePathname } from "next/navigation";
import { usePageLoading } from "../../context/pageLoadingContext";

// Routes that should NOT show the loading overlay
const EXCLUDED_ROUTES = ["/", "/privacy", "/terms", "/login"];

const PageLoadingOverlay = ({ children }) => {
  const pathname = usePathname();
  const { pageLoadingState } = usePageLoading();

  const isExcluded = EXCLUDED_ROUTES.some(
    (route) => pathname === route || pathname.startsWith(`${route}/`)
  );

  const showOverlay = !isExcluded && pageLoadingState.isPageLoading;

  return (
    <>
      {showOverlay && (
        <div className="flex min-h-screen items-center justify-center">
          <Image src="/loading-icon.png" alt="Loading" width={200} height={200} />
        </div>
      )}
      <div style={{ display: showOverlay ? "none" : "block" }}>{children}</div>
    </>
  );
};

export default PageLoadingOverlay;
