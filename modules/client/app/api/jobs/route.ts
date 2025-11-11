import { NextRequest, NextResponse } from "next/server";

// In development, always use local Postgres
const USE_LOCAL_DB = process.env.NODE_ENV === "development";

async function getLocalPostgresClient() {
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

export async function GET(request: NextRequest) {
  const { searchParams } = new URL(request.url);
  const page = Math.max(parseInt(searchParams.get("page") || "1"), 1);
  const limit = Math.max(1, Math.min(parseInt(searchParams.get("limit") || "25"), 100));
  const orderBy = searchParams.get("orderBy") || "application_date";
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
      // Build WHERE clause for search
      let whereClause = "";
      let countParams: any[] = [];
      let queryParams: any[] = [];
      
      if (search) {
        whereClause = "WHERE company ILIKE $1 OR job_title ILIKE $1";
        const searchPattern = `%${search}%`;
        countParams = [searchPattern];
        queryParams = [searchPattern, limit, offset];
      } else {
        queryParams = [limit, offset];
      }

      // Get total count
      const countQuery = search
        ? `SELECT COUNT(*) FROM applied_jobs ${whereClause}`
        : "SELECT COUNT(*) FROM applied_jobs";
      const countResult = await client.query(countQuery, countParams);
      const total = parseInt(countResult.rows[0].count);

      // Build ORDER BY clause
      const orderByColumn = orderBy === "job_title" ? "job_title" :
                           orderBy === "company" ? "company" :
                           orderBy === "status" ? "status" :
                           "application_date";
      const orderClause = `ORDER BY ${orderByColumn} ${order}`;

      // Get paginated results
      const query = search
        ? `SELECT * FROM applied_jobs ${whereClause} ${orderClause} LIMIT $2 OFFSET $3`
        : `SELECT * FROM applied_jobs ${orderClause} LIMIT $1 OFFSET $2`;
      const result = await client.query(query, queryParams);
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
    
    // Apply search filter
    const ascending = order === "ASC";
    const orderColumn = orderBy === "job_title" ? "job_title" :
                       orderBy === "company" ? "company" :
                       orderBy === "status" ? "status" :
                       "application_date";
    
    let countQuery = supabase.from("applied_jobs").select("*", { count: "exact", head: true });
    let dataQuery = supabase.from("applied_jobs").select("*");
    
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

export async function POST(request: NextRequest) {
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
      const result = await client.query(
        `INSERT INTO applied_jobs (company, job_title, job_url, location, salary_range, notes, status, source)
         VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
         RETURNING *`,
        [
          body.company,
          body.job_title,
          body.job_url || null,
          body.location || null,
          body.salary_range || null,
          body.notes || null,
          body.status || "applied",
          body.source || "manual",
        ]
      );
      await client.end();
      return NextResponse.json(result.rows[0]);
    } catch (error) {
      await client.end();
      console.error("Error creating job in Postgres:", error);
      return NextResponse.json(
        { error: "Failed to create job" },
        { status: 500 }
      );
    }
  }

  // Production: Use Supabase
  try {
    const { supabase } = await import("../../../lib/supabase");
    const { data, error } = await supabase
      .from("applied_jobs")
      .insert(body)
      .select()
      .single();

    if (error) throw error;

    return NextResponse.json(data);
  } catch (error) {
    console.error("Error creating job:", error);
    return NextResponse.json(
      { error: "Failed to create job" },
      { status: 500 }
    );
  }
}

