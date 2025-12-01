"use client";

import { useEffect } from "react";
import { useUser } from "@auth0/nextjs-auth0/client";
import { useUserContext } from "../../context/userContext";
import { getMe, getUserPreferences } from "../../api";
import { SET_THEME, SET_LOADING } from "../../reducers/userReducer";

/**
 * Manages authentication state by syncing Auth0 session with backend user data
 * Runs on app load to restore auth state after page refresh
 */
export const AuthStateManager = () => {
  const { user: auth0User, isLoading } = useUser();
  const { userState, dispatchUser } = useUserContext();

  // Sync Auth0 loading state to context
  useEffect(() => {
    dispatchUser({ type: SET_LOADING, payload: isLoading });
  }, [isLoading, dispatchUser]);

  useEffect(() => {
    const restoreAuthState = async () => {
      if (isLoading) return;
      if (!auth0User) return;
      if (userState.isAuthed) return;

      try {
        const response = await getMe();

        dispatchUser({
          type: "SET_USER",
          payload: response.user,
        });

        const tenantId = response.current_tenant_id || response.user.tenant_id;
        const tenant = response.tenants.find((t) => t.id === tenantId);

        if (tenant) {
          dispatchUser({
            type: "SET_TENANT_NAME",
            payload: tenant.name,
          });
        }

        dispatchUser({
          type: "SET_AUTHED",
          payload: true,
        });

        // Load user preferences (theme)
        try {
          const prefs = await getUserPreferences();
          dispatchUser({
            type: SET_THEME,
            payload: {
              theme: prefs.effective_theme,
              canEditTenant: prefs.can_edit_tenant,
            },
          });
        } catch (prefError) {
          console.error("Failed to load preferences:", prefError);
        }
      } catch (error) {
        console.error("Failed to restore auth state:", error);
        // Still mark as authed if Auth0 session exists - let routes handle errors
        dispatchUser({
          type: "SET_AUTHED",
          payload: true,
        });
      }
    };

    restoreAuthState();
  }, [auth0User, isLoading, userState.isAuthed, dispatchUser]);

  return null; // This component doesn't render anything
};
