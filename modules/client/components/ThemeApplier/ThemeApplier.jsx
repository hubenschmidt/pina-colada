"use client";

import { useEffect } from "react";
import { useUserContext } from "../../context/userContext";

/**
 * Applies theme class to <html> element based on user preferences
 */
export const ThemeApplier = () => {
  const { userState } = useUserContext();

  useEffect(() => {
    const htmlElement = document.documentElement;
    htmlElement.classList.remove("dark");

    if (userState.theme === "dark") {
      htmlElement.classList.add("dark");
    }
  }, [userState.theme]);

  return null;
};
