import { NextResponse, NextRequest } from "next/server";
import { auth0 } from "../../../../lib/auth0";

const BACKEND_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8000";

export async function GET() {
  try {
    const { token } = await auth0.getAccessToken();

    const response = await fetch(`${BACKEND_URL}/preferences/tenant`, {
      headers: {
        Authorization: `Bearer ${token}`,
        "Content-Type": "application/json",
      },
    });

    if (!response.ok) {
      const error = await response.text();
      return NextResponse.json(
        { error: error || "Failed to fetch tenant preferences" },
        { status: response.status }
      );
    }

    const data = await response.json();
    return NextResponse.json(data);
  } catch (error) {
    console.error("Error fetching tenant preferences:", error);
    return NextResponse.json(
      { error: "Failed to fetch tenant preferences" },
      { status: 500 }
    );
  }
}

export async function PATCH(request: NextRequest) {
  try {
    const { token } = await auth0.getAccessToken();
    const body = await request.json();

    const response = await fetch(`${BACKEND_URL}/preferences/tenant`, {
      method: "PATCH",
      headers: {
        Authorization: `Bearer ${token}`,
        "Content-Type": "application/json",
      },
      body: JSON.stringify(body),
    });

    if (!response.ok) {
      const error = await response.text();
      return NextResponse.json(
        { error: error || "Failed to update tenant preferences" },
        { status: response.status }
      );
    }

    const data = await response.json();
    return NextResponse.json(data);
  } catch (error) {
    console.error("Error updating tenant preferences:", error);
    return NextResponse.json(
      { error: "Failed to update tenant preferences" },
      { status: 500 }
    );
  }
}
