import { NextResponse, type NextRequest } from "next/server";

import { requestOrigin } from "@/lib/api/request-url";
import { identityApiUrl, sessionCookieOptions } from "@/lib/auth/config";
import { SESSION_COOKIE } from "@/lib/auth/constants";

export async function POST(request: NextRequest) {
  const token = request.cookies.get(SESSION_COOKIE)?.value;

  if (token) {
    try {
      await fetch(`${identityApiUrl()}/v1/auth/logout`, {
        method: "POST",
        headers: { Authorization: `Bearer ${token}` },
        cache: "no-store",
        signal: AbortSignal.timeout(5000),
      });
    } catch {
      // The local session is cleared even if identity is unreachable.
    }
  }

  const response = NextResponse.redirect(
    new URL("/login", requestOrigin(request)),
    303,
  );
  response.cookies.set(SESSION_COOKIE, "", sessionCookieOptions(0));
  return response;
}
