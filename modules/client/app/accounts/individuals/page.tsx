"use client";

import { useEffect } from "react";
import { usePageLoading } from "../../../context/pageLoadingContext";
import { AccountList } from "../../../components/AccountList";
import { useAccountListConfig } from "../../../components/AccountList/hooks/useAccountListConfig";

const IndividualsPage = () => {
  const config = useAccountListConfig("individual");
  const { dispatchPageLoading } = usePageLoading();

  useEffect(() => {
    dispatchPageLoading({ type: "SET_PAGE_LOADING", payload: false });
  }, [dispatchPageLoading]);

  return <AccountList config={config} />;
};

export default IndividualsPage;
