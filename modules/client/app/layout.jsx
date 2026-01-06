import "@mantine/core/styles.css";
import "./globals.css";
import { Space_Grotesk } from "next/font/google";
import { GoogleAnalytics } from "@next/third-parties/google"

const spaceGrotesk = Space_Grotesk({
  subsets: ["latin"],
  variable: "--font-space-grotesk",
});
import { NavProvider } from "../context/navContext";
import { UserProvider } from "../context/userContext";
import { PageLoadingProvider } from "../context/pageLoadingContext";
import { ProjectProvider } from "../context/projectContext";
import { LookupsProvider } from "../context/lookupsContext";
import { ConversationProvider } from "../context/conversationContext";
import { AgentConfigProvider } from "../context/agentConfigContext";
import { NotificationProvider } from "../context/notificationContext";
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
      <body className={spaceGrotesk.className}>
        <Auth0Provider>
          <UserProvider>
            <AuthStateManager />
            <ThemeApplier />
            <MantineThemeProvider>
              <NotificationProvider>
                <AgentConfigProvider>
                  <NavProvider>
                    <ProjectProvider>
                      <LookupsProvider>
                        <ConversationProvider>
                          <PageLoadingProvider>
                            <RootLayoutClient>{children}</RootLayoutClient>
                          </PageLoadingProvider>
                        </ConversationProvider>
                      </LookupsProvider>
                    </ProjectProvider>
                  </NavProvider>
                </AgentConfigProvider>
              </NotificationProvider>
            </MantineThemeProvider>
          </UserProvider>
        </Auth0Provider>
      </body>
      <GoogleAnalytics gaId="G-X4JY302VH2"/>
    </html>
  );
};
