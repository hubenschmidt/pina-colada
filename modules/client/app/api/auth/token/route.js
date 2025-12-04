import { auth0 } from "../../../../lib/auth0";
import { NextResponse } from "next/server";

export const GET = async () => {
  try {
    const { token } = await auth0.getAccessToken();

    return NextResponse.json({ accessToken: token });
  } catch (error) {
    console.error("Token error:", error);
    return NextResponse.json({ error: "Unable to get access token" }, { status: 401 });
  }
};
