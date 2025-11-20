import type { Metadata } from "next";
import "@mantine/core/styles.css";
import "./globals.css";
import { NavProvider } from "../context/navContext";
import { UserProvider } from "../context/userContext";
import { PageLoadingProvider } from "../context/pageLoadingContext";
import { PublicEnvScript } from "next-runtime-env";
import { Auth0Provider } from "@auth0/nextjs-auth0/client";
import { AuthStateManager } from "../components/AuthStateManager";
import { RootLayoutClient } from "./RootLayoutClient";
import { ThemeApplier } from "../components/ThemeApplier";
import { MantineThemeProvider } from "../components/MantineThemeProvider";

export const metadata: Metadata = {
  title: "PinaColada.co",
  description: "Software and AI Solutions",
};

export default ({ children }: Readonly<{ children: React.ReactNode }>) => {
  return (
    <html lang="en" className="scroll-smooth">
      <head>
        <PublicEnvScript />
      </head>
      <body>
        <Auth0Provider>
          <UserProvider>
            <AuthStateManager />
            <ThemeApplier />
            <MantineThemeProvider>
              <NavProvider>
                <PageLoadingProvider>
                  <RootLayoutClient>{children}</RootLayoutClient>
                </PageLoadingProvider>
              </NavProvider>
            </MantineThemeProvider>
          </UserProvider>
        </Auth0Provider>
      </body>
    </html>
  );
};
