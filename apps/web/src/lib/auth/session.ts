import { cookies } from "next/headers";
import { redirect } from "next/navigation";
import { cache } from "react";

import { identityApiUrl } from "./config";
import { SESSION_COOKIE } from "./constants";

export type UserRole = "admin" | "reviewer" | "supervisor";

export interface SessionUser {
  userId: string;
  tenantId: string;
  email: string;
  displayName: string;
  role: UserRole;
}

interface PrincipalResponse {
  user_id: string;
  tenant_id: string;
  email: string;
  display_name: string;
  role: UserRole;
}

export async function getSessionToken(): Promise<string | undefined> {
  const store = await cookies();
  return store.get(SESSION_COOKIE)?.value;
}

export const getCurrentUser = cache(async (): Promise<SessionUser | null> => {
  const token = await getSessionToken();
  if (!token) return null;

  try {
    const response = await fetch(`${identityApiUrl()}/v1/me`, {
      headers: { Authorization: `Bearer ${token}` },
      cache: "no-store",
      signal: AbortSignal.timeout(5000),
    });
    if (!response.ok) return null;

    const principal = (await response.json()) as PrincipalResponse;
    return {
      userId: principal.user_id,
      tenantId: principal.tenant_id,
      email: principal.email,
      displayName: principal.display_name || principal.email,
      role: principal.role,
    };
  } catch {
    return null;
  }
});

export async function requireSession(
  nextPath?: string,
): Promise<{ token: string; user: SessionUser }> {
  const token = await getSessionToken();
  const user = token ? await getCurrentUser() : null;
  if (!token || !user) {
    redirect(
      nextPath
        ? `/login?next=${encodeURIComponent(nextPath)}`
        : "/login",
    );
  }
  return { token, user };
}
