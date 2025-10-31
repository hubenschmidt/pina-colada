import type { Metadata } from "next";
import "./globals.css";
import Header from "../components/Header";
import { NavProvider } from "../context/navContext";

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
        <NavProvider>
          <Header />
          {children}
        </NavProvider>
      </body>
    </html>
  );
}
