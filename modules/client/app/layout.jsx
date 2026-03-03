import "@mantine/core/styles.css";
import "./globals.css";
import { Space_Grotesk } from "next/font/google";
import { GoogleAnalytics } from "@next/third-parties/google"
import { PublicEnvScript } from "next-runtime-env";
import { RootLayoutClient } from "./RootLayoutClient";
import Providers from "./Providers";

const spaceGrotesk = Space_Grotesk({
  subsets: ["latin"],
  variable: "--font-space-grotesk",
});

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
        <Providers>
          <RootLayoutClient>{children}</RootLayoutClient>
        </Providers>
      </body>
      <GoogleAnalytics gaId="G-E8GJ07YCY3"/>
    </html>
  );
};
