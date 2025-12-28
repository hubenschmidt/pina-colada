"use client";

import { useEffect } from "react";
import { usePageLoading } from "../../../context/pageLoadingContext";
import AccountList from "../../../components/AccountList/AccountList";
import { useAccountListConfig } from "../../../components/AccountList/hooks/useAccountListConfig";

const ContactsPage = () => {
  const config = useAccountListConfig("contact");
  const { dispatchPageLoading } = usePageLoading();

  useEffect(() => {
    dispatchPageLoading({ type: "SET_PAGE_LOADING", payload: false });
  }, [dispatchPageLoading]);

  return <AccountList config={config} />;
};

export default ContactsPage;
