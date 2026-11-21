import { afterEach, describe, expect, it, vi } from "vitest";

import { ApiError } from "@/lib/api/client";

import { findAlert, listAlerts } from "./repository";

const ALERT_ID =
  "sha256:335e1083e616ea97812124694f9fb9ec9c6c94bd4ce4c353e7e191b2a7c53d52";

const RAW_ALERT = {
  alert_id: ALERT_ID,
  claim_id: "CLM-1015",
  loss_type: "MOTOR",
  claimed_amount: 35000000,
  status: "OPEN",
  suppressed: false,
  created_at: "2026-11-20T12:39:34.259573Z",
  flags: [
    {
      rule_id: "B5",
      rule_version: 1,
      rule_name: "Coverage Limit Breach",
      severity: "High",
      trigger_value: "Rp 35.000.000 > sum insured Rp 20.000.000",
      evidence_refs: ["CLM-1015"],
      explanation: "Nominal klaim melebihi sum insured.",
    },
  ],
};

function jsonResponse(body: unknown, status = 200): Response {
  return new Response(JSON.stringify(body), {
    status,
    headers: { "Content-Type": "application/json" },
  });
}

afterEach(() => {
  vi.unstubAllGlobals();
});

describe("alert repository", () => {
  it("decodes the route param before calling the raw alert path", async () => {
    const fetchMock = vi.fn(async () => jsonResponse(RAW_ALERT));
    vi.stubGlobal("fetch", fetchMock);

    const alert = await findAlert("token", encodeURIComponent(ALERT_ID));

    expect(alert?.alertId).toBe(ALERT_ID);
    const [url] = fetchMock.mock.calls[0] as unknown as [string];
    expect(url).toBe(`http://localhost:8084/v1/alerts/${ALERT_ID}`);
  });

  it("rejects ids outside the allowed character set", async () => {
    const fetchMock = vi.fn();
    vi.stubGlobal("fetch", fetchMock);

    await expect(findAlert("token", "sha256:bad/id")).rejects.toMatchObject({
      status: 404,
    });
    expect(fetchMock).not.toHaveBeenCalled();
  });

  it("requests suppressed alerts in the queue listing", async () => {
    const fetchMock = vi.fn(async () =>
      jsonResponse({ alerts: [RAW_ALERT], count: 1 }),
    );
    vi.stubGlobal("fetch", fetchMock);

    const alerts = await listAlerts("token");

    expect(alerts).toHaveLength(1);
    const [url] = fetchMock.mock.calls[0] as unknown as [string];
    expect(url).toContain("include_suppressed=true");
  });

  it("maps an upstream 404 to an ApiError", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn(async () =>
        jsonResponse({ code: "NOT_FOUND", detail: "missing" }, 404),
      ),
    );

    await expect(findAlert("token", ALERT_ID)).rejects.toBeInstanceOf(
      ApiError,
    );
  });
});
