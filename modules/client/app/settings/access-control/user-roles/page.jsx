"use client";

import { useEffect } from "react";
import { Tabs } from "@mantine/core";
import UserRoleAssignment from "../../../../components/RbacAdmin/UserRoleAssignment";
import { usePageLoading } from "../../../../context/pageLoadingContext";

const UserRolesPage = () => {
  const { dispatchPageLoading } = usePageLoading();

  useEffect(() => {
    dispatchPageLoading({ type: "SET_PAGE_LOADING", payload: false });
  }, [dispatchPageLoading]);

  return (
    <Tabs.Panel value="user-roles" pt="md">
      <UserRoleAssignment />
    </Tabs.Panel>
  );
};

export default UserRolesPage;
