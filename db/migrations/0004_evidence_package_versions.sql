-- AC-10: an evidence package records the engine and rule pack that produced
-- the case so a frozen package remains reproducible.
ALTER TABLE cases.evidence_package
    ADD COLUMN IF NOT EXISTS engine_version text NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS rule_pack_version text NOT NULL DEFAULT '';
