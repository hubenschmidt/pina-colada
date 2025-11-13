import type { Metadata } from "next";
import "@mantine/core/styles.css";
import "./globals.css";
import Header from "../components/Header";
import { NavProvider } from "../context/navContext";
import { MantineProvider } from "@mantine/core";
import { PublicEnvScript } from "next-runtime-env";
import { Auth0Provider } from "@auth0/nextjs-auth0/client";

export const metadata: Metadata = {
  title: "PinaColada.co",
  description: "Software and AI Solutions",
};

const RootLayout = ({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) => {
  return (
    <html lang="en" className="scroll-smooth">
      <head>
        <PublicEnvScript />
      </head>
      <body>
        <Auth0Provider>
          <MantineProvider>
            <NavProvider>
              <Header />
              {children}
            </NavProvider>
          </MantineProvider>
        </Auth0Provider>
      </body>
    </html>
  );
};

export default RootLayout
