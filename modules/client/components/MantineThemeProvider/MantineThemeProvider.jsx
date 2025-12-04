"use client";

import { MantineProvider, createTheme } from "@mantine/core";
import { useEffect, useState } from "react";
import { useUserContext } from "../../context/userContext";

const theme = createTheme({
  primaryColor: "gray"
});

export const MantineThemeProvider = ({ children }) => {
  const { userState } = useUserContext();
  const [colorScheme, setColorScheme] = useState("light");

  useEffect(() => {
    setColorScheme(userState.theme === "light" ? "light" : "dark");
  }, [userState.theme]);

  return (
    <MantineProvider theme={theme} forceColorScheme={colorScheme}>
      {children}
    </MantineProvider>);

};