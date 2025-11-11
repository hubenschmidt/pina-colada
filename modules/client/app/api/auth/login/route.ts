import { NextRequest, NextResponse } from "next/server";

const AUTH_PASSWORD = process.env.JOBS_PASSWORD;

export async function POST(request: NextRequest) {
  try {
    const { password } = await request.json();

    if (!password) {
      return NextResponse.json(
        { error: "Password is required" },
        { status: 400 }
      );
    }

    if (password === AUTH_PASSWORD) {
      // Generate simple token
      const token = `auth_${Date.now()}_${Math.random()
        .toString(36)
        .substring(7)}`;

      return NextResponse.json({ success: true, token }, { status: 200 });
    }

    return NextResponse.json({ error: "Invalid password" }, { status: 401 });
  } catch (error) {
    console.error("Login error:", error);
    return NextResponse.json(
      { error: "Internal server error" },
      { status: 500 }
    );
  }
}
