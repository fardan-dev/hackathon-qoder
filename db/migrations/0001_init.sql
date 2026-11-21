-- ClaimShield thin-slice schema (Phase 0/1).
-- One PostgreSQL cluster, one schema per service. Services never read another
-- service's tables; cross-service data travels through APIs or Kafka events.

CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE SCHEMA IF NOT EXISTS identity;
CREATE SCHEMA IF NOT EXISTS ingestion;
CREATE SCHEMA IF NOT EXISTS rules;
CREATE SCHEMA IF NOT EXISTS alert;
CREATE SCHEMA IF NOT EXISTS cases;
CREATE SCHEMA IF NOT EXISTS audit;
CREATE SCHEMA IF NOT EXISTS metering;
CREATE SCHEMA IF NOT EXISTS reporting;

-- ---------------------------------------------------------------------------
-- identity
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS identity.users (
    id            uuid PRIMARY KEY,
    tenant_id     uuid NOT NULL,
    email         text NOT NULL,
    display_name  text NOT NULL,
    password_hash text NOT NULL,
    role          text NOT NULL CHECK (role IN ('admin', 'reviewer', 'supervisor')),
    created_at    timestamptz NOT NULL DEFAULT now(),
    UNIQUE (tenant_id, email)
);

CREATE TABLE IF NOT EXISTS identity.sessions (
    token_hash   text PRIMARY KEY,
    user_id      uuid NOT NULL REFERENCES identity.users (id) ON DELETE CASCADE,
    tenant_id    uuid NOT NULL,
    created_at   timestamptz NOT NULL DEFAULT now(),
    last_seen_at timestamptz NOT NULL DEFAULT now(),
    expires_at   timestamptz NOT NULL
);
CREATE INDEX IF NOT EXISTS sessions_user_idx ON identity.sessions (user_id);

-- ---------------------------------------------------------------------------
-- ingestion
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS ingestion.import_batch (
    id              uuid PRIMARY KEY,
    tenant_id       uuid NOT NULL,
    idempotency_key text NOT NULL,
    request_hash    text NOT NULL,
    file_name       text NOT NULL,
    file_sha256     text NOT NULL,
    status          text NOT NULL CHECK (status IN ('PROCESSING', 'PROCESSED', 'REJECTED', 'FAILED')),
    total_rows      integer NOT NULL DEFAULT 0,
    accepted_rows   integer NOT NULL DEFAULT 0,
    rejected_rows   integer NOT NULL DEFAULT 0,
    error_code      text,
    created_at      timestamptz NOT NULL DEFAULT now(),
    completed_at    timestamptz,
    UNIQUE (tenant_id, idempotency_key),
    UNIQUE (tenant_id, file_sha256)
);

CREATE TABLE IF NOT EXISTS ingestion.claim_raw (
    tenant_id      uuid NOT NULL,
    claim_id       text NOT NULL,
    batch_id       uuid NOT NULL,
    policy_id      text NOT NULL,
    claimant_id    text NOT NULL,
    claim_date     date NOT NULL,
    incident_date  date NOT NULL,
    loss_type      text NOT NULL CHECK (loss_type IN ('MOTOR', 'PROPERTY')),
    claimed_amount bigint NOT NULL CHECK (claimed_amount > 0),
    currency       text NOT NULL CHECK (currency = 'IDR'),
    provider_id    text NOT NULL,
    invoice_number text,
    status         text NOT NULL,
    rep_id         text,
    fingerprint    text NOT NULL,
    lineage        jsonb NOT NULL,
    created_at     timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (tenant_id, claim_id)
);
CREATE INDEX IF NOT EXISTS claim_batch_idx ON ingestion.claim_raw (tenant_id, batch_id);
CREATE INDEX IF NOT EXISTS claim_claimant_idx ON ingestion.claim_raw (tenant_id, claimant_id);
CREATE INDEX IF NOT EXISTS claim_policy_idx ON ingestion.claim_raw (tenant_id, policy_id);
CREATE INDEX IF NOT EXISTS claim_invoice_idx ON ingestion.claim_raw (tenant_id, invoice_number);

CREATE TABLE IF NOT EXISTS ingestion.invoice_raw (
    tenant_id      uuid NOT NULL,
    invoice_number text NOT NULL,
    provider_id    text NOT NULL,
    invoice_date   date NOT NULL,
    invoice_amount bigint NOT NULL CHECK (invoice_amount > 0),
    currency       text NOT NULL CHECK (currency = 'IDR'),
    claim_id       text NOT NULL,
    doc_sha256     text,
    batch_id       uuid NOT NULL,
    lineage        jsonb NOT NULL,
    created_at     timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (tenant_id, invoice_number, provider_id)
);
CREATE INDEX IF NOT EXISTS invoice_claim_idx ON ingestion.invoice_raw (tenant_id, claim_id);

CREATE TABLE IF NOT EXISTS ingestion.policy_raw (
    tenant_id            uuid NOT NULL,
    policy_id            text NOT NULL,
    product              text NOT NULL,
    loss_type_scope      text NOT NULL CHECK (loss_type_scope IN ('MOTOR', 'PROPERTY')),
    inception_date       date NOT NULL,
    coverage_added_date  date,
    expiry_date          date NOT NULL,
    status               text NOT NULL,
    sum_insured          bigint NOT NULL CHECK (sum_insured > 0),
    premium              bigint NOT NULL CHECK (premium > 0),
    reinstatement_date   date,
    asset_id             text,
    agent_id             text,
    batch_id             uuid NOT NULL,
    lineage              jsonb NOT NULL,
    created_at           timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (tenant_id, policy_id)
);

CREATE TABLE IF NOT EXISTS ingestion.provider_raw (
    tenant_id   uuid NOT NULL,
    provider_id text NOT NULL,
    name        text NOT NULL,
    type        text NOT NULL,
    region      text NOT NULL,
    batch_id    uuid NOT NULL,
    lineage     jsonb NOT NULL,
    created_at  timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (tenant_id, provider_id)
);

CREATE TABLE IF NOT EXISTS ingestion.identity_map (
    tenant_id      uuid NOT NULL,
    claimant_id    text NOT NULL,
    attribute_type text NOT NULL,
    attribute_hash text NOT NULL,
    batch_id       uuid NOT NULL,
    lineage        jsonb NOT NULL,
    created_at     timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (tenant_id, claimant_id, attribute_type)
);

CREATE TABLE IF NOT EXISTS ingestion.claim_lineage (
    lineage_id  uuid PRIMARY KEY,
    tenant_id   uuid NOT NULL,
    claim_id    text NOT NULL,
    field_name  text NOT NULL,
    source      text NOT NULL,
    source_row  integer NOT NULL,
    transform   text NOT NULL,
    recorded_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS claim_lineage_claim_idx ON ingestion.claim_lineage (tenant_id, claim_id, field_name);

CREATE TABLE IF NOT EXISTS ingestion.import_error (
    error_id   uuid PRIMARY KEY,
    tenant_id  uuid NOT NULL,
    batch_id   uuid NOT NULL,
    dataset    text NOT NULL,
    row_number integer NOT NULL,
    field_name text,
    code       text NOT NULL,
    message    text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS import_error_batch_idx ON ingestion.import_error (tenant_id, batch_id);

CREATE TABLE IF NOT EXISTS ingestion.outbox_event (
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
CREATE INDEX IF NOT EXISTS ingestion_outbox_pending_idx
    ON ingestion.outbox_event (status, next_attempt_at) WHERE status = 'PENDING';

CREATE TABLE IF NOT EXISTS ingestion.processed_event (
    consumer     text NOT NULL,
    event_id     text NOT NULL,
    processed_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (consumer, event_id)
);

-- ---------------------------------------------------------------------------
-- rules
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS rules.rule_pack (
    version      text PRIMARY KEY,
    engine_version text NOT NULL,
    status       text NOT NULL CHECK (status IN ('ACTIVE', 'RETIRED')),
    created_at   timestamptz NOT NULL DEFAULT now(),
    retired_at   timestamptz
);

CREATE TABLE IF NOT EXISTS rules.catalog (
    rule_id          text NOT NULL,
    version          integer NOT NULL,
    family           text NOT NULL,
    name             text NOT NULL,
    severity         text NOT NULL CHECK (severity IN ('High', 'Medium', 'Low')),
    params           jsonb NOT NULL,
    params_version   integer NOT NULL,
    status           text NOT NULL CHECK (status IN ('ACTIVE', 'GATED_OFF')),
    owner            text NOT NULL,
    mapping_typology text[] NOT NULL,
    activated_at     timestamptz,
    PRIMARY KEY (rule_id, version)
);

CREATE TABLE IF NOT EXISTS rules.kill_switch (
    switch_id  uuid PRIMARY KEY,
    tenant_id  uuid,
    scope      text NOT NULL CHECK (scope IN ('rule', 'family', 'global')),
    target     text NOT NULL,
    enabled    boolean NOT NULL,
    updated_by uuid NOT NULL,
    updated_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (tenant_id, scope, target)
);

CREATE TABLE IF NOT EXISTS rules.evaluation (
    evaluation_id    uuid PRIMARY KEY,
    tenant_id        uuid NOT NULL,
    batch_id         uuid NOT NULL,
    claim_id         text NOT NULL,
    rule_pack_version text NOT NULL,
    engine_version   text NOT NULL,
    config_version   text NOT NULL,
    as_of            timestamptz NOT NULL,
    flags            jsonb NOT NULL,
    not_evaluated    jsonb NOT NULL,
    output_hash      text NOT NULL,
    evaluated_at     timestamptz NOT NULL DEFAULT now(),
    UNIQUE (tenant_id, claim_id, rule_pack_version, engine_version, config_version)
);
CREATE INDEX IF NOT EXISTS evaluation_batch_idx ON rules.evaluation (tenant_id, batch_id);

CREATE TABLE IF NOT EXISTS rules.claim_history (
    tenant_id      uuid NOT NULL,
    claim_id       text NOT NULL,
    batch_id       uuid NOT NULL,
    claimant_id    text NOT NULL,
    policy_id      text NOT NULL,
    provider_id    text NOT NULL,
    invoice_number text,
    loss_type      text NOT NULL,
    claim_date     date NOT NULL,
    incident_date  date NOT NULL,
    claimed_amount bigint NOT NULL,
    status         text NOT NULL,
    fingerprint    text NOT NULL,
    doc_sha256     text,
    asset_id       text,
    sum_insured    bigint NOT NULL,
    inception_date date NOT NULL,
    coverage_added_date date,
    expiry_date    date NOT NULL,
    reinstatement_date date,
    policy_status  text NOT NULL,
    identity_attrs jsonb NOT NULL,
    needs_info_count integer NOT NULL DEFAULT 0,
    prior_cleared_reason text,
    prior_confirmed boolean NOT NULL DEFAULT false,
    recorded_at    timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (tenant_id, claim_id)
);
CREATE INDEX IF NOT EXISTS claim_history_claimant_idx ON rules.claim_history (tenant_id, claimant_id);
CREATE INDEX IF NOT EXISTS claim_history_invoice_idx ON rules.claim_history (tenant_id, invoice_number);

CREATE TABLE IF NOT EXISTS rules.outbox_event (
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
CREATE INDEX IF NOT EXISTS rules_outbox_pending_idx
    ON rules.outbox_event (status, next_attempt_at) WHERE status = 'PENDING';

CREATE TABLE IF NOT EXISTS rules.processed_event (
    consumer     text NOT NULL,
    event_id     text NOT NULL,
    processed_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (consumer, event_id)
);

-- ---------------------------------------------------------------------------
-- alert
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS alert.bundle (
    alert_id        text PRIMARY KEY,
    tenant_id       uuid NOT NULL,
    claim_id        text NOT NULL,
    batch_id        uuid NOT NULL,
    loss_type       text NOT NULL,
    severity        text NOT NULL CHECK (severity IN ('High', 'Medium', 'Low')),
    severity_rank   smallint NOT NULL,
    flag_count      integer NOT NULL,
    claimed_amount  bigint NOT NULL,
    fingerprint     text NOT NULL,
    flags           jsonb NOT NULL,
    status          text NOT NULL DEFAULT 'OPEN',
    suppressed      boolean NOT NULL DEFAULT false,
    suppression_id  uuid,
    created_at      timestamptz NOT NULL DEFAULT now(),
    sla_due_at      timestamptz NOT NULL,
    UNIQUE (tenant_id, claim_id)
);
CREATE INDEX IF NOT EXISTS alert_queue_idx
    ON alert.bundle (tenant_id, severity_rank, flag_count DESC, claimed_amount DESC, alert_id);

CREATE TABLE IF NOT EXISTS alert.suppression_projection (
    tenant_id     uuid NOT NULL,
    fingerprint   text NOT NULL,
    suppression_id uuid NOT NULL,
    expires_at    timestamptz NOT NULL,
    revoked_at    timestamptz,
    PRIMARY KEY (tenant_id, fingerprint)
);

CREATE TABLE IF NOT EXISTS alert.outbox_event (
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
CREATE INDEX IF NOT EXISTS alert_outbox_pending_idx
    ON alert.outbox_event (status, next_attempt_at) WHERE status = 'PENDING';

CREATE TABLE IF NOT EXISTS alert.processed_event (
    consumer     text NOT NULL,
    event_id     text NOT NULL,
    processed_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (consumer, event_id)
);

-- ---------------------------------------------------------------------------
-- cases
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS cases.case_group (
    case_id    uuid PRIMARY KEY,
    tenant_id  uuid NOT NULL,
    title      text NOT NULL,
    summary    text NOT NULL DEFAULT '',
    status     text NOT NULL DEFAULT 'OPEN' CHECK (status IN ('OPEN', 'NEEDS_INFO', 'CLOSED')),
    priority   smallint NOT NULL DEFAULT 2,
    owner_id   uuid,
    version    integer NOT NULL DEFAULT 1,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    closed_at  timestamptz
);
CREATE INDEX IF NOT EXISTS case_queue_idx ON cases.case_group (tenant_id, status, priority, created_at);

CREATE TABLE IF NOT EXISTS cases.case_claim (
    tenant_id uuid NOT NULL,
    case_id   uuid NOT NULL REFERENCES cases.case_group (case_id),
    claim_id  text NOT NULL,
    PRIMARY KEY (case_id, claim_id)
);

CREATE TABLE IF NOT EXISTS cases.evidence (
    evidence_id      uuid PRIMARY KEY,
    tenant_id        uuid NOT NULL,
    case_id          uuid NOT NULL REFERENCES cases.case_group (case_id),
    artifact_type    text NOT NULL,
    source           text NOT NULL,
    content_sha256   text NOT NULL,
    chain_of_custody jsonb NOT NULL,
    notes            text NOT NULL DEFAULT '',
    created_at       timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS evidence_case_idx ON cases.evidence (tenant_id, case_id);

CREATE TABLE IF NOT EXISTS cases.evidence_package (
    package_id     uuid PRIMARY KEY,
    tenant_id      uuid NOT NULL,
    case_id        uuid NOT NULL REFERENCES cases.case_group (case_id),
    package_sha256 text NOT NULL,
    artifact_count integer NOT NULL,
    frozen_at      timestamptz NOT NULL DEFAULT now(),
    UNIQUE (case_id)
);

CREATE TABLE IF NOT EXISTS cases.decision (
    decision_id     uuid PRIMARY KEY,
    tenant_id       uuid NOT NULL,
    case_id         uuid NOT NULL REFERENCES cases.case_group (case_id),
    outcome         text NOT NULL CHECK (outcome IN ('CONFIRMED', 'CLEARED', 'NEEDS_INFO')),
    reason_code     text NOT NULL,
    reviewer_id     uuid NOT NULL,
    notes           text NOT NULL DEFAULT '',
    value_protected bigint,
    value_recovered bigint,
    decided_at      timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS decision_case_idx ON cases.decision (tenant_id, case_id, decided_at);

CREATE TABLE IF NOT EXISTS cases.suppression (
    suppression_id uuid PRIMARY KEY,
    tenant_id      uuid NOT NULL,
    claim_id       text NOT NULL,
    fingerprint    text NOT NULL,
    reason         text NOT NULL,
    approved_by    uuid NOT NULL,
    expires_at     timestamptz NOT NULL,
    created_at     timestamptz NOT NULL DEFAULT now(),
    revoked_at     timestamptz,
    revoked_by     uuid
);
CREATE INDEX IF NOT EXISTS suppression_fingerprint_idx ON cases.suppression (tenant_id, fingerprint);

CREATE TABLE IF NOT EXISTS cases.alert_projection (
    tenant_id      uuid NOT NULL,
    alert_id       text NOT NULL,
    claim_id       text NOT NULL,
    fingerprint    text NOT NULL,
    loss_type      text NOT NULL,
    severity       text NOT NULL,
    claimed_amount bigint NOT NULL,
    priority_rank  smallint NOT NULL,
    flags          jsonb NOT NULL,
    created_at     timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (tenant_id, claim_id)
);

CREATE TABLE IF NOT EXISTS cases.outbox_event (
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
CREATE INDEX IF NOT EXISTS cases_outbox_pending_idx
    ON cases.outbox_event (status, next_attempt_at) WHERE status = 'PENDING';

CREATE TABLE IF NOT EXISTS cases.processed_event (
    consumer     text NOT NULL,
    event_id     text NOT NULL,
    processed_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (consumer, event_id)
);

-- ---------------------------------------------------------------------------
-- audit
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS audit.audit_event (
    seq          bigserial PRIMARY KEY,
    event_id     uuid NOT NULL UNIQUE,
    tenant_id    uuid NOT NULL,
    actor        text NOT NULL,
    action       text NOT NULL,
    object_type  text NOT NULL,
    object_id    text NOT NULL,
    occurred_at  timestamptz NOT NULL,
    payload      jsonb NOT NULL,
    prev_hash    text NOT NULL,
    hash         text NOT NULL
);
CREATE INDEX IF NOT EXISTS audit_object_idx ON audit.audit_event (tenant_id, object_type, object_id, seq);
CREATE INDEX IF NOT EXISTS audit_tenant_idx ON audit.audit_event (tenant_id, seq);

CREATE TABLE IF NOT EXISTS audit.processed_event (
    consumer     text NOT NULL,
    event_id     text NOT NULL,
    processed_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (consumer, event_id)
);

-- ---------------------------------------------------------------------------
-- metering
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS metering.usage_event (
    usage_event_id  uuid PRIMARY KEY,
    tenant_id       uuid NOT NULL,
    source_event_id text NOT NULL,
    metric          text NOT NULL,
    quantity        integer NOT NULL CHECK (quantity >= 0),
    period          text NOT NULL,
    occurred_at     timestamptz NOT NULL,
    payload         jsonb NOT NULL,
    UNIQUE (tenant_id, source_event_id, metric)
);
CREATE INDEX IF NOT EXISTS usage_event_period_idx ON metering.usage_event (tenant_id, period);

CREATE TABLE IF NOT EXISTS metering.usage_record (
    tenant_id       uuid NOT NULL,
    period          text NOT NULL,
    formula_version text NOT NULL,
    metrics         jsonb NOT NULL,
    checksum        text NOT NULL,
    closed_by       uuid,
    closed_at       timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (tenant_id, period)
);

CREATE TABLE IF NOT EXISTS metering.value_report (
    report_id       uuid PRIMARY KEY,
    tenant_id       uuid NOT NULL,
    period          text NOT NULL,
    formula_version text NOT NULL,
    payload         jsonb NOT NULL,
    checksum        text NOT NULL,
    generated_at    timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS metering.billing_export (
    export_id  uuid PRIMARY KEY,
    tenant_id  uuid NOT NULL,
    period     text NOT NULL,
    version    text NOT NULL,
    payload    jsonb NOT NULL,
    checksum   text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (tenant_id, period, version)
);

CREATE TABLE IF NOT EXISTS metering.processed_event (
    consumer     text NOT NULL,
    event_id     text NOT NULL,
    processed_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (consumer, event_id)
);

-- ---------------------------------------------------------------------------
-- reporting (read model; derived, not source of truth)
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS reporting.coverage_typology (
    typology_key text PRIMARY KEY,
    label        text NOT NULL,
    in_scope     boolean NOT NULL,
    phase01_covered boolean NOT NULL,
    rule_ids     text[] NOT NULL,
    residual_risk text,
    sort_order   integer NOT NULL
);

-- ---------------------------------------------------------------------------
-- Immutability guards
-- ---------------------------------------------------------------------------
CREATE OR REPLACE FUNCTION audit.deny_mutation() RETURNS trigger
LANGUAGE plpgsql AS $$
BEGIN
    RAISE EXCEPTION 'append-only table: % is not mutable', TG_TABLE_NAME;
END;
$$;

DROP TRIGGER IF EXISTS audit_event_immutable ON audit.audit_event;
CREATE TRIGGER audit_event_immutable
    BEFORE UPDATE OR DELETE ON audit.audit_event
    FOR EACH ROW EXECUTE FUNCTION audit.deny_mutation();

DROP TRIGGER IF EXISTS claim_lineage_immutable ON ingestion.claim_lineage;
CREATE TRIGGER claim_lineage_immutable
    BEFORE UPDATE OR DELETE ON ingestion.claim_lineage
    FOR EACH ROW EXECUTE FUNCTION audit.deny_mutation();

DROP TRIGGER IF EXISTS decision_immutable ON cases.decision;
CREATE TRIGGER decision_immutable
    BEFORE UPDATE OR DELETE ON cases.decision
    FOR EACH ROW EXECUTE FUNCTION audit.deny_mutation();

CREATE OR REPLACE FUNCTION metering.deny_closed_record_mutation() RETURNS trigger
LANGUAGE plpgsql AS $$
BEGIN
    IF OLD.closed_at IS NOT NULL THEN
        RAISE EXCEPTION 'usage record % is closed and immutable', OLD.period;
    END IF;
    RETURN NEW;
END;
$$;

DROP TRIGGER IF EXISTS usage_record_immutable ON metering.usage_record;
CREATE TRIGGER usage_record_immutable
    BEFORE UPDATE ON metering.usage_record
    FOR EACH ROW EXECUTE FUNCTION metering.deny_closed_record_mutation();

-- ---------------------------------------------------------------------------
-- Row level security (defence in depth; services also filter by tenant_id)
-- ---------------------------------------------------------------------------
DO $$
DECLARE
    t record;
    tables text[] := ARRAY[
        'ingestion.import_batch', 'ingestion.claim_raw', 'ingestion.invoice_raw',
        'ingestion.policy_raw', 'ingestion.provider_raw', 'ingestion.identity_map',
        'ingestion.claim_lineage', 'ingestion.import_error',
        'rules.evaluation', 'rules.claim_history',
        'alert.bundle', 'alert.suppression_projection',
        'cases.case_group', 'cases.case_claim', 'cases.evidence',
        'cases.evidence_package', 'cases.decision', 'cases.suppression',
        'cases.alert_projection',
        'audit.audit_event',
        'metering.usage_event', 'metering.usage_record',
        'metering.value_report', 'metering.billing_export'
    ];
    tbl text;
BEGIN
    FOREACH tbl IN ARRAY tables LOOP
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
