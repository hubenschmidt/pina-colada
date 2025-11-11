import type { Metadata } from "next";
import "@mantine/core/styles.css";
import "./globals.css";
import Header from "../components/Header";
import { NavProvider } from "../context/navContext";
import { MantineProvider } from "@mantine/core";
import { PublicEnvScript } from "next-runtime-env";

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
        <MantineProvider>
          <NavProvider>
            <Header />
            {children}
          </NavProvider>
        </MantineProvider>
      </body>
    </html>
  );
};

export default RootLayout
