-- Hardening and read-model additions for the thin slice.

-- Invoice facts needed to reproduce rules A4/A5/A6 from history.
ALTER TABLE rules.claim_history ADD COLUMN IF NOT EXISTS invoice_date date;
ALTER TABLE rules.claim_history ADD COLUMN IF NOT EXISTS invoice_amount bigint;

-- Metering publishes events through an outbox like every other writer.
CREATE TABLE IF NOT EXISTS metering.outbox_event (
    id              uuid PRIMARY KEY,
    tenant_id       uuid NOT NULL,
    event_type      text NOT NULL,
    aggregate_key   text NOT NULL,
    payload         jsonb NOT NULL,
    status          text NOT NULL DEFAULT 'PENDING' CHECK (status IN ('PENDING', 'PUBLISHED', 'FAILED')),
    attempts        integer NOT NULL DEFAULT 0,
    next_attempt_at timestamptz NOT NULL DEFAULT now(),
    created_at      timestamptz NOT NULL DEFAULT now(),
    published_at    timestamptz
);
CREATE INDEX IF NOT EXISTS metering_outbox_pending_idx
    ON metering.outbox_event (status, next_attempt_at) WHERE status = 'PENDING';

-- Reporting read models, populated from Kafka events.
CREATE TABLE IF NOT EXISTS reporting.alert_snapshot (
    tenant_id      uuid NOT NULL,
    alert_id       text NOT NULL,
    claim_id       text NOT NULL,
    loss_type      text NOT NULL,
    severity       text NOT NULL,
    severity_rank  smallint NOT NULL,
    flag_count     integer NOT NULL,
    claimed_amount bigint NOT NULL,
    flags          jsonb NOT NULL,
    suppressed     boolean NOT NULL DEFAULT false,
    status         text NOT NULL DEFAULT 'OPEN',
    sla_due_at     timestamptz NOT NULL,
    created_at     timestamptz NOT NULL,
    PRIMARY KEY (tenant_id, alert_id)
);
CREATE INDEX IF NOT EXISTS reporting_alert_created_idx
    ON reporting.alert_snapshot (tenant_id, created_at);
CREATE INDEX IF NOT EXISTS reporting_alert_lob_idx
    ON reporting.alert_snapshot (tenant_id, loss_type, created_at);

CREATE TABLE IF NOT EXISTS reporting.decision_snapshot (
    tenant_id       uuid NOT NULL,
    case_id         text NOT NULL,
    outcome         text NOT NULL,
    reason_code     text NOT NULL,
    value_protected bigint NOT NULL DEFAULT 0,
    value_recovered bigint NOT NULL DEFAULT 0,
    claim_ids       jsonb NOT NULL,
    decided_at      timestamptz NOT NULL,
    PRIMARY KEY (tenant_id, case_id)
);
CREATE INDEX IF NOT EXISTS reporting_decision_decided_idx
    ON reporting.decision_snapshot (tenant_id, decided_at);

-- Evidence is immutable; nothing can be attached after the package is frozen.
CREATE OR REPLACE FUNCTION cases.deny_evidence_mutation() RETURNS trigger
LANGUAGE plpgsql AS $$
BEGIN
    RAISE EXCEPTION 'evidence is append-only';
END;
$$;

CREATE OR REPLACE FUNCTION cases.guard_frozen_evidence() RETURNS trigger
LANGUAGE plpgsql AS $$
BEGIN
    IF EXISTS (SELECT 1 FROM cases.evidence_package WHERE case_id = NEW.case_id) THEN
        RAISE EXCEPTION 'evidence package is frozen';
    END IF;
    RETURN NEW;
END;
$$;

DROP TRIGGER IF EXISTS evidence_immutable ON cases.evidence;
CREATE TRIGGER evidence_immutable
    BEFORE UPDATE OR DELETE ON cases.evidence
    FOR EACH ROW EXECUTE FUNCTION cases.deny_evidence_mutation();

DROP TRIGGER IF EXISTS evidence_frozen_guard ON cases.evidence;
CREATE TRIGGER evidence_frozen_guard
    BEFORE INSERT ON cases.evidence
    FOR EACH ROW EXECUTE FUNCTION cases.guard_frozen_evidence();

CREATE OR REPLACE FUNCTION cases.deny_package_mutation() RETURNS trigger
LANGUAGE plpgsql AS $$
BEGIN
    RAISE EXCEPTION 'evidence package is immutable';
END;
$$;

DROP TRIGGER IF EXISTS package_immutable ON cases.evidence_package;
CREATE TRIGGER package_immutable
    BEFORE UPDATE OR DELETE ON cases.evidence_package
    FOR EACH ROW EXECUTE FUNCTION cases.deny_package_mutation();

-- RLS for the new reporting read models.
DO $$
DECLARE
    tbl text;
BEGIN
    FOREACH tbl IN ARRAY ARRAY['reporting.alert_snapshot', 'reporting.decision_snapshot'] LOOP
        EXECUTE format('ALTER TABLE %s ENABLE ROW LEVEL SECURITY', tbl);
        EXECUTE format('ALTER TABLE %s FORCE ROW LEVEL SECURITY', tbl);
        EXECUTE format('DROP POLICY IF EXISTS tenant_isolation ON %s', tbl);
        EXECUTE format(
            'CREATE POLICY tenant_isolation ON %s USING (tenant_id = nullif(current_setting(''app.tenant_id'', true), '''')::uuid) WITH CHECK (tenant_id = nullif(current_setting(''app.tenant_id'', true), '''')::uuid)',
            tbl
        );
    END LOOP;
END;
$$;
