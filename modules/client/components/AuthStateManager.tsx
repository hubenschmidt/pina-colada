"use client";

import { useEffect } from "react";
import { useUser } from "@auth0/nextjs-auth0/client";
import { useUserContext } from "../context/userContext";
import { SET_USER, SET_TENANT_NAME, SET_AUTHED, SET_THEME } from "../reducers/userReducer";
import { getMe } from "../api";

/**
 * Manages authentication state by syncing Auth0 session with backend user data
 * Runs on app load to restore auth state after page refresh
 */
export const AuthStateManager = () => {
  const { user: auth0User, isLoading } = useUser();
  const { userState, dispatchUser } = useUserContext();

  useEffect(() => {
    const restoreAuthState = async () => {
      // Don't do anything while Auth0 is loading
      if (isLoading) return;

      // User is logged in with Auth0
      if (auth0User) {
        // Only fetch if we don't already have user data
        if (!userState.isAuthed) {
          try {
            const response = await getMe();

            dispatchUser({
              type: SET_USER,
              payload: response.user,
            });

            // Find tenant name from user's tenant_id or current_tenant_id
            const tenantId = response.current_tenant_id || response.user.tenant_id;
            if (tenantId && response.tenants.length > 0) {
              const tenant = response.tenants.find(t => t.id === tenantId);
              if (tenant) {
                dispatchUser({
                  type: SET_TENANT_NAME,
                  payload: tenant.name,
                });
              }
            }

            dispatchUser({
              type: SET_AUTHED,
              payload: true,
            });

            // Fetch user preferences
            try {
              const prefsResponse = await fetch("/api/preferences/user");
              if (prefsResponse.ok) {
                const prefs = await prefsResponse.json();
                dispatchUser({
                  type: SET_THEME,
                  payload: {
                    theme: prefs.effective_theme,
                    canEditTenant: prefs.can_edit_tenant,
                  },
                });
              }
            } catch (error) {
              console.error("Failed to fetch user preferences:", error);
            }
          } catch (error) {
            console.error("Failed to restore auth state:", error);
          }
        }
      }
    };

    restoreAuthState();
  }, [auth0User, isLoading, userState.isAuthed, dispatchUser]);

  return null; // This component doesn't render anything
};
