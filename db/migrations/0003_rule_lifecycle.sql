-- Rule lifecycle state. The pinned catalog stays immutable; lifecycle changes
-- pause or retire a rule by activating its kill switch, so the change takes
-- effect on the next evaluation without a deploy.
CREATE TABLE IF NOT EXISTS rules.lifecycle (
    tenant_id  uuid NOT NULL,
    rule_id    text NOT NULL,
    status     text NOT NULL CHECK (status IN ('ACTIVE', 'PAUSED', 'RETIRED')),
    reason     text NOT NULL DEFAULT '',
    changed_by uuid NOT NULL,
    changed_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (tenant_id, rule_id)
);

ALTER TABLE rules.lifecycle ENABLE ROW LEVEL SECURITY;
ALTER TABLE rules.lifecycle FORCE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS tenant_isolation ON rules.lifecycle;
CREATE POLICY tenant_isolation ON rules.lifecycle
    USING (tenant_id = nullif(current_setting('app.tenant_id', true), '')::uuid)
    WITH CHECK (tenant_id = nullif(current_setting('app.tenant_id', true), '')::uuid);
