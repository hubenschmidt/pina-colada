// types.ts
export type UserContextV1 = {
  schema: "user_context@v1";
  // Browser & OS
  ua: string;
  platform: string | undefined;
  vendor: string | undefined;
  languages: string[];
  tz: string;
  dnt: boolean | null;                // Do Not Track
  gpc: boolean | null;                // Global Privacy Control
  // Device capabilities
  deviceMemory?: number;              // GB (rounded)
  hardwareConcurrency?: number;       // CPU cores
  maxTouchPoints?: number;
  // Screen/viewport
  screen: { w: number; h: number; dpr: number; colorDepth: number };
  viewport: { w: number; h: number };
  prefersColorScheme?: "light" | "dark" | "no-preference";
  prefersReducedMotion?: boolean;
  // Network (no PII)
  connection?: { effectiveType?: string; downlink?: number; rtt?: number; saveData?: boolean };
  // Page context
  url: string;
  ref: string | null;
  utm?: Record<string, string>;
  // Storage support
  storage: { cookies: boolean; local: boolean; session: boolean; };
  // Optional – only with consent
  geo?: { lat: number; lon: number; accuracy?: number; source: "geolocation" | "ip" };
  battery?: { charging: boolean; level: number };  // 0..1
};
