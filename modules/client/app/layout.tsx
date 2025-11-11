import type { Metadata } from "next";
import "@mantine/core/styles.css";
import "./globals.css";
import Header from "../components/Header";
import { NavProvider } from "../context/navContext";
import { MantineProvider } from "@mantine/core";

export const metadata: Metadata = {
  title: "PinaColada.co",
  description: "Software and AI Solutions",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en" className="scroll-smooth">
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
}
