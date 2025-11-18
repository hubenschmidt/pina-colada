"use client";

import { useUserContext } from "../../context/userContext";
import { useState, useEffect } from "react";
import { SET_THEME } from "../../reducers/userReducer";

export default function SettingsPage() {
  const { userState, dispatchUser } = useUserContext();
  const [userTheme, setUserTheme] = useState<string | null>(null);
  const [tenantTheme, setTenantTheme] = useState<string>("light");
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const fetchPrefs = async () => {
      if (!userState.isAuthed) return;

      try {
        const userPrefs = await fetch("/api/preferences/user");
        if (userPrefs.ok) {
          const data = await userPrefs.json();
          setUserTheme(data.theme);
        }

        if (userState.canEditTenantTheme) {
          const tenantPrefs = await fetch("/api/preferences/tenant");
          if (tenantPrefs.ok) {
            const data = await tenantPrefs.json();
            setTenantTheme(data.theme);
          }
        }
      } catch (error) {
        console.error("Failed to fetch preferences:", error);
      }

      setLoading(false);
    };

    fetchPrefs();
  }, [userState.isAuthed, userState.canEditTenantTheme]);

  const updateUserTheme = async (newTheme: "light" | "dark" | null) => {
    if (!userState.isAuthed) return;

    try {
      const res = await fetch("/api/preferences/user", {
        method: "PATCH",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ theme: newTheme }),
      });

      if (res.ok) {
        const data = await res.json();
        setUserTheme(newTheme);
        dispatchUser({
          type: SET_THEME,
          payload: {
            theme: data.effective_theme,
            canEditTenant: userState.canEditTenantTheme,
          },
        });
      }
    } catch (error) {
      console.error("Failed to update user theme:", error);
    }
  };

  const updateTenantTheme = async (newTheme: "light" | "dark") => {
    if (!userState.isAuthed || !userState.canEditTenantTheme) return;

    try {
      const res = await fetch("/api/preferences/tenant", {
        method: "PATCH",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ theme: newTheme }),
      });

      if (res.ok) {
        setTenantTheme(newTheme);

        if (userTheme === null) {
          dispatchUser({
            type: SET_THEME,
            payload: {
              theme: newTheme,
              canEditTenant: userState.canEditTenantTheme,
            },
          });
        }
      }
    } catch (error) {
      console.error("Failed to update tenant theme:", error);
    }
  };

  if (!userState.isAuthed) {
    return (
      <div className="p-8">
        <p>Please log in to access settings.</p>
      </div>
    );
  }

  if (loading) {
    return (
      <div className="p-8">
        <p>Loading preferences...</p>
      </div>
    );
  }

  return (
    <div className="p-8 max-w-2xl mx-auto">
      <h1 className="text-3xl font-bold mb-8">Settings</h1>

      <section className="mb-12">
        <h2 className="text-2xl font-semibold mb-4">Personal Theme</h2>
        <div className="space-y-3">
          <label className="flex items-center gap-3 cursor-pointer">
            <input
              type="radio"
              checked={userTheme === "light"}
              onChange={() => updateUserTheme("light")}
              className="w-4 h-4"
            />
            <span>Light</span>
          </label>
          <label className="flex items-center gap-3 cursor-pointer">
            <input
              type="radio"
              checked={userTheme === "dark"}
              onChange={() => updateUserTheme("dark")}
              className="w-4 h-4"
            />
            <span>Dark</span>
          </label>
          <label className="flex items-center gap-3 cursor-pointer">
            <input
              type="radio"
              checked={userTheme === null}
              onChange={() => updateUserTheme(null)}
              className="w-4 h-4"
            />
            <span>Inherit from organization</span>
          </label>
        </div>
      </section>

      {userState.canEditTenantTheme && (
        <section className="mb-12 p-6 border rounded-lg">
          <h2 className="text-2xl font-semibold mb-2">Organization Theme (Admin)</h2>
          <p className="text-sm mb-4 opacity-70">
            This affects all users who inherit the organization theme.
          </p>
          <div className="space-y-3">
            <label className="flex items-center gap-3 cursor-pointer">
              <input
                type="radio"
                checked={tenantTheme === "light"}
                onChange={() => updateTenantTheme("light")}
                className="w-4 h-4"
              />
              <span>Light</span>
            </label>
            <label className="flex items-center gap-3 cursor-pointer">
              <input
                type="radio"
                checked={tenantTheme === "dark"}
                onChange={() => updateTenantTheme("dark")}
                className="w-4 h-4"
              />
              <span>Dark</span>
            </label>
          </div>
        </section>
      )}
    </div>
  );
}
