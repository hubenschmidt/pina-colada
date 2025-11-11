import { NextRequest, NextResponse } from "next/server";

// In development, always use local Postgres
const USE_LOCAL_DB = process.env.NODE_ENV === "development";

const getLocalPostgresClient = async () => {
  if (!USE_LOCAL_DB) return null;
  
  try {
    const { Client } = await import("pg");
    const client = new Client({
      host: process.env.POSTGRES_HOST || "postgres",
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

export const PUT = async (
  request: NextRequest,
  { params }: { params: Promise<{ id: string }> }
) => {
  const { id } = await params;
  const body = await request.json();

  if (USE_LOCAL_DB) {
    const client = await getLocalPostgresClient();
    if (!client) {
      return NextResponse.json(
        { error: "Failed to connect to local Postgres" },
        { status: 500 }
      );
    }

    try {
      const updates: string[] = [];
      const values: any[] = [];
      let paramIndex = 1;

      if (body.company !== undefined) {
        updates.push(`company = $${paramIndex++}`);
        values.push(body.company);
      }
      if (body.date !== undefined) {
        updates.push(`date = $${paramIndex++}`);
        values.push(body.date);
      }
      if (body.resume !== undefined) {
        updates.push(`resume = $${paramIndex++}`);
        values.push(body.resume);
      }
      if (body.job_title !== undefined) {
        updates.push(`job_title = $${paramIndex++}`);
        values.push(body.job_title);
      }
      if (body.status !== undefined) {
        updates.push(`status = $${paramIndex++}`);
        values.push(body.status);
      }
      if (body.job_url !== undefined) {
        updates.push(`job_url = $${paramIndex++}`);
        values.push(body.job_url);
      }
      if (body.salary_range !== undefined) {
        updates.push(`salary_range = $${paramIndex++}`);
        values.push(body.salary_range);
      }
      if (body.notes !== undefined) {
        updates.push(`notes = $${paramIndex++}`);
        values.push(body.notes);
      }

      updates.push(`updated_at = NOW()`);
      values.push(id);

      const result = await client.query(
        `UPDATE "Job" SET ${updates.join(", ")} WHERE id = $${paramIndex} RETURNING *`,
        values
      );
      await client.end();

      if (result.rows.length === 0) {
        return NextResponse.json({ error: "Job not found" }, { status: 404 });
      }

      return NextResponse.json(result.rows[0]);
    } catch (error) {
      await client.end();
      console.error("Error updating job in Postgres:", error);
      return NextResponse.json(
        { error: "Failed to update job" },
        { status: 500 }
      );
    }
  }

  // Production: Use Supabase
  try {
    const { supabase } = await import("../../../../lib/supabase");
    const { data, error } = await supabase
      .from("Job")
      .update(body)
      .eq("id", id)
      .select()
      .single();

    if (error) throw error;

    return NextResponse.json(data);
  } catch (error) {
    console.error("Error updating job:", error);
    return NextResponse.json(
      { error: "Failed to update job" },
      { status: 500 }
    );
  }
}

export const DELETE = async (
  request: NextRequest,
  { params }: { params: Promise<{ id: string }> }
) => {
  const { id } = await params;
  
  if (USE_LOCAL_DB) {
    const client = await getLocalPostgresClient();
    if (!client) {
      return NextResponse.json(
        { error: "Failed to connect to local Postgres" },
        { status: 500 }
      );
    }

    try {
      await client.query('DELETE FROM "Job" WHERE id = $1', [id]);
      await client.end();
      return NextResponse.json({ success: true });
    } catch (error) {
      await client.end();
      console.error("Error deleting job from Postgres:", error);
      return NextResponse.json(
        { error: "Failed to delete job" },
        { status: 500 }
      );
    }
  }

  // Production: Use Supabase
  try {
    const { supabase } = await import("../../../../lib/supabase");
    const { error } = await supabase
      .from("Job")
      .delete()
      .eq("id", id);

    if (error) throw error;

    return NextResponse.json({ success: true });
  } catch (error) {
    console.error("Error deleting job:", error);
    return NextResponse.json(
      { error: "Failed to delete job" },
      { status: 500 }
    );
  }
}

