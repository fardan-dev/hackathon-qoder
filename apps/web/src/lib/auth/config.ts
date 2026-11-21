export function identityApiUrl(): string {
  return process.env.IDENTITY_API_URL ?? "http://localhost:8081";
}

export function sessionCookieOptions(maxAgeSeconds: number) {
  return {
    httpOnly: true,
    sameSite: "lax" as const,
    secure: process.env.SESSION_COOKIE_SECURE === "true",
    path: "/",
    maxAge: maxAgeSeconds,
  };
}
