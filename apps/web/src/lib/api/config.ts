const DEFAULT_TIMEOUT_MS = 5000;

export function alertApiUrl(): string {
  return process.env.ALERT_API_URL ?? "http://localhost:8084";
}

export function caseApiUrl(): string {
  return process.env.CASE_API_URL ?? "http://localhost:8085";
}

export function reportingApiUrl(): string {
  return process.env.REPORTING_API_URL ?? "http://localhost:8088";
}

export function apiTimeoutMs(): number {
  const configured = Number(process.env.WEB_API_TIMEOUT_MS);
  return Number.isFinite(configured) && configured > 0
    ? configured
    : DEFAULT_TIMEOUT_MS;
}
