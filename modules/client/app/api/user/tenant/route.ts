import { NextResponse } from "next/server";
import { auth0 } from "../../../../lib/auth0";

export const GET = async () => {
  try {
    const session = await auth0.getSession();

    if (!session?.user) {
      return NextResponse.json({ error: "Unauthorized" }, { status: 401 });
    }

    const userEmail = session.user.email;

    // Query backend to check if user has a tenant
    const backendUrl =
      process.env.NEXT_PUBLIC_API_URL || "http://localhost:8000";
    const response = await fetch(
      `${backendUrl}/users/${encodeURIComponent(userEmail)}/tenant`,
      {
        headers: {
          "Content-Type": "application/json",
        },
      }
    );

    if (!response.ok) {
      if (response.status === 404) {
        return NextResponse.json({ hasTenant: false, tenant: null });
      }
      throw new Error(`Backend returned ${response.status}`);
    }

    const data = await response.json();
    return NextResponse.json({ hasTenant: true, tenant: data });
  } catch (error) {
    console.error("Error checking user tenant:", error);
    return NextResponse.json(
      { error: "Failed to check tenant" },
      { status: 500 }
    );
  }
};
