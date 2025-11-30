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

    if (userState.theme === "dark") {
      htmlElement.classList.add("dark");
    }

    if (userState.theme === "light") {
      htmlElement.classList.remove("dark");
    }
  }, [userState.theme]);

  return null;
};
