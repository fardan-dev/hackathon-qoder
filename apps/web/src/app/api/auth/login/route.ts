import { NextResponse, type NextRequest } from "next/server";

import { identityApiUrl, sessionCookieOptions } from "@/lib/auth/config";
import { SESSION_COOKIE } from "@/lib/auth/constants";

const DEFAULT_NEXT = "/queue";

function safeNext(value: FormDataEntryValue | null): string {
  if (typeof value !== "string") return DEFAULT_NEXT;
  if (!value.startsWith("/") || value.startsWith("//")) return DEFAULT_NEXT;
  return value;
}

function errorRedirect(request: NextRequest, code: string, next: string) {
  const url = new URL("/login", request.url);
  url.searchParams.set("error", code);
  if (next !== DEFAULT_NEXT) url.searchParams.set("next", next);
  return NextResponse.redirect(url, 303);
}

export async function POST(request: NextRequest) {
  const form = await request.formData();
  const email = String(form.get("email") ?? "").trim();
  const password = String(form.get("password") ?? "");
  const next = safeNext(form.get("next"));

  if (!email || !password) {
    return errorRedirect(request, "required", next);
  }

  let response: Response;
  try {
    response = await fetch(`${identityApiUrl()}/v1/auth/login`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ email, password }),
      cache: "no-store",
      signal: AbortSignal.timeout(8000),
    });
  } catch {
    return errorRedirect(request, "unavailable", next);
  }

  if (!response.ok) {
    const code =
      response.status === 429
        ? "rate"
        : response.status === 401 || response.status === 400
          ? "invalid"
          : "unavailable";
    return errorRedirect(request, code, next);
  }

  const payload = (await response.json()) as {
    token?: string;
    expires_at?: string;
  };
  if (!payload.token) {
    return errorRedirect(request, "unavailable", next);
  }

  const expiresAt = payload.expires_at
    ? new Date(payload.expires_at).getTime()
    : Number.NaN;
  const maxAge = Number.isFinite(expiresAt)
    ? Math.max(60, Math.floor((expiresAt - Date.now()) / 1000))
    : 60 * 60 * 8;

  const redirectResponse = NextResponse.redirect(
    new URL(next, request.url),
    303,
  );
  redirectResponse.cookies.set(
    SESSION_COOKIE,
    payload.token,
    sessionCookieOptions(maxAge),
  );
  return redirectResponse;
}
