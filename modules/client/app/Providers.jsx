"use client";

import { Auth0Provider } from "@auth0/nextjs-auth0/client";
import { UserProvider } from "../context/userContext";
import { NavProvider } from "../context/navContext";
import { PageLoadingProvider } from "../context/pageLoadingContext";
import { ProjectProvider } from "../context/projectContext";
import { LookupsProvider } from "../context/lookupsContext";
import { ConversationProvider } from "../context/conversationContext";
import { AgentConfigProvider } from "../context/agentConfigContext";
import { NotificationProvider } from "../context/notificationContext";
import { AuthStateManager } from "../components/AuthStateManager/AuthStateManager";
import { ThemeApplier } from "../components/ThemeApplier/ThemeApplier";
import { MantineThemeProvider } from "../components/MantineThemeProvider/MantineThemeProvider";

const compose = (...providers) => providers.reduce(
  (Acc, [Provider, props]) => ({ children }) => (
    <Acc>
      <Provider {...props}>{children}</Provider>
    </Acc>
  ),
  ({ children }) => <>{children}</>
);

// Auth layer: Auth0 + user state + side-effect components
const AuthProviders = compose(
  [Auth0Provider, {}],
  [UserProvider, {}],
);

// UI layer: theme + notifications + Mantine
const UIProviders = compose(
  [MantineThemeProvider, {}],
  [NotificationProvider, {}],
);

// App state: nav, project, lookups, conversations, agent config, page loading
const AppStateProviders = compose(
  [AgentConfigProvider, {}],
  [NavProvider, {}],
  [ProjectProvider, {}],
  [LookupsProvider, {}],
  [ConversationProvider, {}],
  [PageLoadingProvider, {}],
);

const Providers = ({ children }) => (
  <AuthProviders>
    <AuthStateManager />
    <ThemeApplier />
    <UIProviders>
      <AppStateProviders>
        {children}
      </AppStateProviders>
    </UIProviders>
  </AuthProviders>
);

export default Providers;
