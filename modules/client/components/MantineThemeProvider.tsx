"use client";

import { MantineProvider, createTheme, MantineColorScheme } from "@mantine/core";
import { useEffect, useState } from "react";
import { useUserContext } from "../context/userContext";

const theme = createTheme({
  primaryColor: "gray",
});

export const MantineThemeProvider = ({ children }: { children: React.ReactNode }) => {
  const { userState } = useUserContext();
  const [colorScheme, setColorScheme] = useState<MantineColorScheme>("light");

  useEffect(() => {
    setColorScheme(userState.theme === "dark" ? "dark" : "light");
  }, [userState.theme]);

  return (
    <MantineProvider theme={theme} forceColorScheme={colorScheme}>
      {children}
    </MantineProvider>
  );
};
