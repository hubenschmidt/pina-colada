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
  const { user: auth0User, isLoading: auth0Loading } = useUser();
  const { userState, dispatchUser } = useUserContext();

  useEffect(() => {
    const restoreAuthState = async () => {
      // Still waiting for Auth0 to check session
      if (auth0Loading) return;

      // No Auth0 user - not authenticated, done loading
      if (!auth0User) {
        dispatchUser({ type: SET_LOADING, payload: false });
        return;
      }

      // Already restored auth state
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
      } finally {
        // Done loading - either authed or errored
        dispatchUser({ type: SET_LOADING, payload: false });
      }
    };

    restoreAuthState();
  }, [auth0User, auth0Loading, userState.isAuthed, dispatchUser]);

  return null;
};
