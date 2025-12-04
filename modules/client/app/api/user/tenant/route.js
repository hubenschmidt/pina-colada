function _optionalChain(ops) {
  let lastAccessLHS = undefined;
  let value = ops[0];
  let i = 1;
  while (i < ops.length) {
    const op = ops[i];
    const fn = ops[i + 1];
    i += 2;
    if ((op === "optionalAccess" || op === "optionalCall") && value == null) {
      return undefined;
    }
    if (op === "access" || op === "optionalAccess") {
      lastAccessLHS = value;
      value = fn(value);
    } else if (op === "call" || op === "optionalCall") {
      value = fn((...args) => value.call(lastAccessLHS, ...args));
      lastAccessLHS = undefined;
    }
  }
  return value;
}
import { NextResponse } from "next/server";
import { auth0 } from "../../../../lib/auth0";

export const GET = async () => {
  try {
    const session = await auth0.getSession();

    if (!_optionalChain([session, "optionalAccess", (_) => _.user])) {
      return NextResponse.json({ error: "Unauthorized" }, { status: 401 });
    }

    const userEmail = session.user.email;

    // Query backend to check if user has a tenant
    const backendUrl = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8000";
    const response = await fetch(`${backendUrl}/users/${encodeURIComponent(userEmail)}/tenant`, {
      headers: {
        "Content-Type": "application/json",
      },
    });

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
    return NextResponse.json({ error: "Failed to check tenant" }, { status: 500 });
  }
};
