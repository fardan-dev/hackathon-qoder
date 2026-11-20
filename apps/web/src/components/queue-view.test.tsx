import { render, screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import type { ReactNode } from "react";
import { describe, expect, it, vi } from "vitest";

import { AlertTable } from "./AlertTable";
import { FilterBar } from "./FilterBar";
import {
  buildQueueQuery,
  parseQueueFilters,
} from "@/lib/alerts/filters";
import { createMockAlerts } from "@/lib/alerts/mock-data";
import { buildQueue } from "@/lib/alerts/queue";

vi.mock("next/link", async () => {
  const { createElement } = await import("react");
  return {
    default: ({
      href,
      children,
      ...rest
    }: { href: string; children?: ReactNode } & Record<string, unknown>) =>
      createElement("a", { href, ...rest }, children),
  };
});

const NOW = new Date("2026-11-20T10:00:00+07:00");
const DEFAULT_FILTERS = parseQueueFilters({});
const ALERTS = createMockAlerts(NOW);

function renderQueue() {
  const { rows } = buildQueue(ALERTS, DEFAULT_FILTERS, NOW);
  const queueQuery = buildQueueQuery(DEFAULT_FILTERS);

  render(
    <>
      <FilterBar filters={DEFAULT_FILTERS} />
      <AlertTable rows={rows} queueQuery={queueQuery} />
    </>,
  );

  return rows;
}

describe("reviewer queue UI", () => {
  it("exposes labelled filter controls", () => {
    renderQueue();

    expect(screen.getByLabelText("Cari")).toBeInTheDocument();
    expect(screen.getByLabelText("Prioritas")).toHaveValue("default");
    expect(screen.getByLabelText("Status")).toHaveValue("default");
    expect(screen.getByLabelText("Line of business")).toHaveValue("all");
    expect(screen.getByLabelText("SLA")).toHaveValue("all");
    expect(
      screen.getByRole("button", { name: "Terapkan filter" }),
    ).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "Reset" })).toHaveAttribute(
      "href",
      "/queue",
    );
  });

  it("renders the default queue with priority sort and SLA indicators", () => {
    const rows = renderQueue();
    const table = screen.getByRole("table", { name: /Daftar alert/ });
    const bodyRows = within(table).getAllByRole("row").slice(1);

    expect(bodyRows).toHaveLength(rows.length);
    expect(bodyRows).toHaveLength(13);

    const firstRow = within(bodyRows[0] as HTMLElement);
    expect(
      firstRow.getByText("CLM-0003", { selector: ".claim-id" }),
    ).toBeInTheDocument();
    expect(
      firstRow.getByRole("link", { name: /Buka detail/ }),
    ).toHaveAttribute("href", "/alerts/ALT-1003");

    expect(screen.getByText("Lewat 6 jam")).toBeInTheDocument();
    expect(screen.getByText("Sisa 4 jam")).toBeInTheDocument();
    expect(screen.queryByText("CLM-0010")).not.toBeInTheDocument();
    expect(screen.queryByText("CLM-0011")).not.toBeInTheDocument();
  });

  it("keeps filter context in detail links", () => {
    const activeFilters = parseQueueFilters({
      priority: "HIGH",
      sla: "OVERDUE",
    });
    const { rows } = buildQueue(ALERTS, activeFilters, NOW);
    const queueQuery = buildQueueQuery(activeFilters);

    render(<AlertTable rows={rows} queueQuery={queueQuery} />);

    const links = screen.getAllByRole("link", { name: /Buka detail/ });
    expect(links).toHaveLength(3);
    expect(links[0]).toHaveAttribute(
      "href",
      `/alerts/ALT-1015?back=${encodeURIComponent(queueQuery)}`,
    );
  });

  it("is keyboard navigable from the first filter control", async () => {
    const user = userEvent.setup();
    renderQueue();

    await user.tab();
    expect(screen.getByLabelText("Cari")).toHaveFocus();

    await user.tab();
    expect(screen.getByLabelText("Prioritas")).toHaveFocus();
  });
});
