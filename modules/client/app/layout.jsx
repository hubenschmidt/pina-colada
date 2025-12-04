import "@mantine/core/styles.css";
import "./globals.css";
import { NavProvider } from "../context/navContext";
import { UserProvider } from "../context/userContext";
import { PageLoadingProvider } from "../context/pageLoadingContext";
import { ProjectProvider } from "../context/projectContext";
import { LookupsProvider } from "../context/lookupsContext";
import { PublicEnvScript } from "next-runtime-env";
import { Auth0Provider } from "@auth0/nextjs-auth0/client";
import { AuthStateManager } from "../components/AuthStateManager/AuthStateManager";
import { RootLayoutClient } from "./RootLayoutClient";
import { ThemeApplier } from "../components/ThemeApplier/ThemeApplier";
import { MantineThemeProvider } from "../components/MantineThemeProvider/MantineThemeProvider";

export const metadata = {
  title: "PinaColada.co",
  description: "Software and AI Solutions",
};

export default ({ children }) => {
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
                <ProjectProvider>
                  <LookupsProvider>
                    <PageLoadingProvider>
                      <RootLayoutClient>{children}</RootLayoutClient>
                    </PageLoadingProvider>
                  </LookupsProvider>
                </ProjectProvider>
              </NavProvider>
            </MantineThemeProvider>
          </UserProvider>
        </Auth0Provider>
      </body>
    </html>
  );
};
