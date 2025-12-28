"use client";

import { useEffect } from "react";
import { Tabs } from "@mantine/core";
import RoleList from "../../../../components/RbacAdmin/RoleList";
import { usePageLoading } from "context/pageLoadingContext";

const RolesPage = () => {
  const { dispatchPageLoading} = usePageLoading()

  useEffect(() => {
    dispatchPageLoading({ type: "SET_PAGE_LOADING", payload: false });
  }, [dispatchPageLoading]);

  return (
    <Tabs.Panel value="roles" pt="md">
      <RoleList />
    </Tabs.Panel>
  );
};

export default RolesPage;
