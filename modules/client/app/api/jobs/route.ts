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
  const { searchParams } = new URL(request.url);
  const page = Math.max(parseInt(searchParams.get("page") || "1"), 1);
  const limit = Math.max(1, Math.min(parseInt(searchParams.get("limit") || "25"), 100));
  const orderBy = searchParams.get("orderBy") || "date";
  const order = searchParams.get("order")?.toUpperCase() === "ASC" ? "ASC" : "DESC";
  const search = searchParams.get("search")?.trim() || "";
  const offset = (page - 1) * limit;

  if (USE_LOCAL_DB) {
    const client = await getLocalPostgresClient();
    if (!client) {
      return NextResponse.json(
        { error: "Failed to connect to local Postgres" },
        { status: 500 }
      );
    }

    try {
      const getOrderColumn = (by: string): string => {
        if (by === "job_title") return "job_title"
        if (by === "company") return "company"
        if (by === "status") return "status"
        if (by === "date") return "date"
        if (by === "resume") return "resume"
        return "date"
      }
      const orderByColumn = getOrderColumn(orderBy)
      const orderClause = `ORDER BY ${orderByColumn} ${order}`;

      if (!search) {
        const countResult = await client.query('SELECT COUNT(*) FROM "Job"');
        const total = parseInt(countResult.rows[0].count);
        const result = await client.query(
          `SELECT * FROM "Job" ${orderClause} LIMIT $1 OFFSET $2`,
          [limit, offset]
        );
        await client.end();

        const totalPages = Math.max(1, Math.ceil(total / limit));
        return NextResponse.json({
          items: result.rows,
          currentPage: page,
          totalPages,
          total,
          pageSize: limit,
        });
      }

      const whereClause = "WHERE company ILIKE $1 OR job_title ILIKE $1";
      const searchPattern = `%${search}%`;
      
      const countResult = await client.query(
        `SELECT COUNT(*) FROM "Job" ${whereClause}`,
        [searchPattern]
      );
      const total = parseInt(countResult.rows[0].count);

      const result = await client.query(
        `SELECT * FROM "Job" ${whereClause} ${orderClause} LIMIT $2 OFFSET $3`,
        [searchPattern, limit, offset]
      );
      await client.end();

      const totalPages = Math.max(1, Math.ceil(total / limit));
      return NextResponse.json({
        items: result.rows,
        currentPage: page,
        totalPages,
        total,
        pageSize: limit,
      });
    } catch (error) {
      await client.end();
      console.error("Error fetching jobs from Postgres:", error);
      return NextResponse.json(
        { error: "Failed to fetch jobs" },
        { status: 500 }
      );
    }
  }

  // Production: Use Supabase
  try {
    const { supabase } = await import("../../../lib/supabase");
    
    const getOrderColumn = (by: string): string => {
      if (by === "job_title") return "job_title"
      if (by === "company") return "company"
      if (by === "status") return "status"
      if (by === "date") return "date"
      if (by === "resume") return "resume"
      return "date"
    }
    const ascending = order === "ASC";
    const orderColumn = getOrderColumn(orderBy)
    
    let countQuery = supabase.from("Job").select("*", { count: "exact", head: true });
    let dataQuery = supabase.from("Job").select("*");
    
    if (search) {
      const searchPattern = `%${search}%`;
      countQuery = countQuery.or(`company.ilike."${searchPattern}",job_title.ilike."${searchPattern}"`);
      dataQuery = dataQuery.or(`company.ilike."${searchPattern}",job_title.ilike."${searchPattern}"`);
    }
    
    // Get total count
    const { count } = await countQuery;
    
    const { data, error } = await dataQuery
      .order(orderColumn, { ascending })
      .range(offset, offset + limit - 1);

    if (error) throw error;

    const total = count || 0;
    const totalPages = Math.max(1, Math.ceil(total / limit));
    
    return NextResponse.json({
      items: data || [],
      currentPage: page,
      totalPages,
      total,
      pageSize: limit,
    });
  } catch (error) {
    console.error("Error fetching jobs:", error);
    return NextResponse.json(
      { error: "Failed to fetch jobs" },
      { status: 500 }
    );
  }
}

export const POST = async (request: NextRequest) => {
  const body = await request.json();

  // Sanitize date fields: convert empty strings to null
  const sanitizedBody = { ...body };
  if (sanitizedBody.date === '') sanitizedBody.date = null;
  if (sanitizedBody.resume === '') sanitizedBody.resume = null;

  if (USE_LOCAL_DB) {
    const client = await getLocalPostgresClient();
    if (!client) {
      return NextResponse.json(
        { error: "Failed to connect to local Postgres" },
        { status: 500 }
      );
    }

    try {
      const result = await client.query(
        `INSERT INTO "Job" (company, job_title, date, job_url, salary_range, notes, resume, status, source)
         VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
         RETURNING *`,
        [
          sanitizedBody.company,
          sanitizedBody.job_title,
          sanitizedBody.date || new Date().toISOString(),
          sanitizedBody.job_url || null,
          sanitizedBody.salary_range || null,
          sanitizedBody.notes || null,
          sanitizedBody.resume || null,
          sanitizedBody.status || "applied",
          sanitizedBody.source || "manual",
        ]
      );
      await client.end();
      return NextResponse.json(result.rows[0]);
    } catch (error: any) {
      await client.end();
      console.error("Error creating job in Postgres:", error);
      return NextResponse.json(
        { error: error.message || "Failed to create job" },
        { status: 500 }
      );
    }
  }

  // Production: Use Supabase
  try {
    const { supabase } = await import("../../../lib/supabase");
    const { data, error } = await supabase
      .from("Job")
      .insert(sanitizedBody)
      .select()
      .single();

    if (error) throw error;

    return NextResponse.json(data);
  } catch (error: any) {
    console.error("Error creating job:", error);
    return NextResponse.json(
      { error: error.message || "Failed to create job" },
      { status: 500 }
    );
  }
}

