import { redirect } from "next/navigation";

import { isUnauthorized } from "./client";

export async function redirectOnUnauthorized<T>(
  promise: Promise<T>,
  nextPath: string,
): Promise<T> {
  try {
    return await promise;
  } catch (error) {
    if (isUnauthorized(error)) {
      redirect(`/login?next=${encodeURIComponent(nextPath)}`);
    }
    throw error;
  }
}
