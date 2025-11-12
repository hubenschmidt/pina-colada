import { NextRequest, NextResponse } from "next/server";

// In development, always use local Postgres
const USE_LOCAL_DB = process.env.NODE_ENV === "development";

const getLocalPostgresClient = async () => {
  if (!USE_LOCAL_DB) return null;
  
  try {
    const { Client } = await import("pg");
    const client = new Client({
      host: process.env.POSTGRES_HOST || "localhost",
      port: parseInt(process.env.POSTGRES_PORT || "5432"),
      user: process.env.POSTGRES_USER || "postgres",
      password: process.env.POSTGRES_PASSWORD || "postgres",
      database: process.env.POSTGRES_DB || "pina_colada",
    });
    await client.connect();
    return client;
  } catch (error) {
    console.error("Failed to connect to local Postgres:", error);
    return null;
  }
}

export const GET = async (request: NextRequest) => {
  if (USE_LOCAL_DB) {
    const client = await getLocalPostgresClient();
    if (!client) {
      return NextResponse.json(
        { error: "Failed to connect to local Postgres" },
        { status: 500 }
      );
    }

    try {
      // Get the most recent job with a resume date, ordered by date DESC
      const result = await client.query(
        `SELECT resume FROM "Job" 
         WHERE resume IS NOT NULL 
         ORDER BY date DESC 
         LIMIT 1`
      );
      await client.end();

      if (result.rows.length === 0 || !result.rows[0].resume) {
        return NextResponse.json({ resume_date: null });
      }

      // Convert timestamp to date string (YYYY-MM-DD format)
      const date = new Date(result.rows[0].resume);
      const dateString = date.toISOString().split('T')[0];
      
      return NextResponse.json({ resume_date: dateString });
    } catch (error: any) {
      await client.end();
      console.error("Error fetching recent resume date from Postgres:", error);
      return NextResponse.json(
        { error: error.message || "Failed to fetch recent resume date" },
        { status: 500 }
      );
    }
  }

  // Production: Use Supabase
  try {
    const { supabase } = await import("../../../../lib/supabase");
    const { data, error } = await supabase
      .from("Job")
      .select("resume")
      .not("resume", "is", null)
      .order("date", { ascending: false })
      .limit(1)
      .maybeSingle();

    if (error || !data || !data.resume) {
      return NextResponse.json({ resume_date: null });
    }

    // Convert timestamp to date string (YYYY-MM-DD format)
    const date = new Date(data.resume);
    const dateString = date.toISOString().split('T')[0];
    
    return NextResponse.json({ resume_date: dateString });
  } catch (error: any) {
    console.error("Error fetching recent resume date:", error);
    return NextResponse.json(
      { error: error.message || "Failed to fetch recent resume date" },
      { status: 500 }
    );
  }
}

