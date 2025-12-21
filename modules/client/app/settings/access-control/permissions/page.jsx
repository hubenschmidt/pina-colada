"use client";

import { useEffect } from "react";
import { Tabs } from "@mantine/core";
import PermissionMatrix from "../../../../components/RbacAdmin/PermissionMatrix";
import { usePageLoading } from "context/pageLoadingContext";

const PermissionsPage = () => {
  const { dispatchPageLoading} = usePageLoading()

  useEffect(() => {
    dispatchPageLoading({ type: "SET_PAGE_LOADING", payload: false });
  }, [dispatchPageLoading]);

  return (
    <Tabs.Panel value="permissions" pt="md">
      <PermissionMatrix />
    </Tabs.Panel>
  );
};

export default PermissionsPage;
