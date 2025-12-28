"use client";

import { useEffect } from "react";
import { useRouter } from "next/navigation";
import { usePageLoading } from "../../../context/pageLoadingContext";

const AccessControlPage = () => {
  const router = useRouter();
  const { dispatchPageLoading } = usePageLoading();
  
  useEffect(() => {
    router.replace("/settings/access-control/roles");
  }, [router]);

  useEffect(() => {
    dispatchPageLoading({ type: "SET_PAGE_LOADING", payload: false });
  }, [dispatchPageLoading]);

  return null;
};

export default AccessControlPage;
