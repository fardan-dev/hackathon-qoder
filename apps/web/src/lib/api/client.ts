import { apiTimeoutMs } from "./config";

interface ProblemBody {
  code?: string;
  detail?: string;
  title?: string;
}

export class ApiError extends Error {
  readonly status: number;
  readonly code: string;

  constructor(status: number, code: string, message: string) {
    super(message);
    this.name = "ApiError";
    this.status = status;
    this.code = code;
  }
}

export function isUnauthorized(error: unknown): boolean {
  return error instanceof ApiError && error.status === 401;
}

export async function apiGet<T>(
  baseUrl: string,
  path: string,
  token: string,
): Promise<T> {
  let response: Response;

  try {
    response = await fetch(`${baseUrl}${path}`, {
      headers: {
        Authorization: `Bearer ${token}`,
        Accept: "application/json",
      },
      cache: "no-store",
      signal: AbortSignal.timeout(apiTimeoutMs()),
    });
  } catch {
    throw new ApiError(
      0,
      "UPSTREAM_UNAVAILABLE",
      "Layanan hulu tidak dapat dihubungi.",
    );
  }

  if (!response.ok) {
    let problem: ProblemBody = {};
    try {
      problem = (await response.json()) as ProblemBody;
    } catch {
      // Non-JSON error body; keep the status-based fallback.
    }

    throw new ApiError(
      response.status,
      problem.code ?? "UPSTREAM_ERROR",
      problem.detail ?? problem.title ?? `Permintaan gagal (${response.status}).`,
    );
  }

  return (await response.json()) as T;
}
