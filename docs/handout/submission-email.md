# Submission email draft

**To:** Fian (a.putrifortunella@alibaba-inc.com), Muchlis (muchlis.insani@alibaba-inc.com)
**Subject:** FSI Mini-Hackathon submission — Sentinel: insurance claim fraud triage
**Attachment:** `docs/handout/Sentinel-Insurance-Fraud-Hackathon.pptx` (14 slides, English)

---

Dear Fian and Muchlis,

Thank you for the workshop and for the take-home challenge. Here is the submission from Team7 FSI.

**What it is.** Sentinel is a human-in-the-loop triage layer for P&C insurance claims. It imports
claims, invoices, policies and providers, evaluates 33 deterministic rules, and routes flagged
claims to a reviewer with the rule, the trigger value, the evidence references and a plain-language
explanation attached to every flag. The system never rejects a claim, never changes claim status
and never emits a risk score — a reviewer always decides. All data is synthetic.

**Deliverables.**
- Deck (attached): 14 slides, business-first — problem and market, positioning, pricing and
  packaging, unit economics, go-to-market — with one technical page on the deterministic
  pipeline and controls, plus the demo, coverage evidence and reproduction steps.
- Source code: https://github.com/fardan-dev/hackathon-qoder — `README.md` has the quick start and
  the evidence index; `docs/QA-Acceptance.md` maps every claim to an acceptance criterion or a test.
- Cloud: the TRD reference design targets Alibaba Cloud Jakarta (`ap-southeast-5`, 3 AZ). The Alibaba
  account was disabled during the hackathon, so the Phase 1 landing zone was implemented on Google
  Cloud Jakarta (`asia-southeast2`, 3 zones) with the same failure-domain pattern; the Alibaba design
  is kept as the TRD reference.

**Verified in this tree.**
- 33/33 planted positives flagged (100% recall, rule families A–G).
- 0 Must-rule false positives across 14 clean controls; `make fp-gate` regenerates the report.
- 131 audit events with a valid hash chain; decisions and suppressions are audited.
- Determinism diff 0: same dataset + rule pack = same output hash.
- 52/52 web tests, lint, typecheck and production build pass.
- Clean-environment acceptance: `docker compose down -v`, then `make up && make seed && make demo`.

**How to reproduce.**
```
cp .env.example .env        # set demo secrets
make up && make seed        # stack, users, coverage matrix
make datasetgen             # deterministic CSVs
make demo                   # five-minute acceptance path
```

**Business case.** Pricing is per claim scored (platform subscription by volume band plus per-claim),
never per alert, because charging for alerts would reward noise. The numbers in the deck
(IDR 315m illustrative ARR, 70% gross margin, 13.6-month CAC payback) are formula tests only; the BRD
marks every input as TBD until two paid pilots and 10–15 willingness-to-pay interviews.

**Next step.** A pilot with a named SIU sponsor, an agreed baseline and success criteria measured
before the start: flag precision, review time, value protected and audit completeness.

Happy to walk through the demo or the evidence in more detail.

Best regards,

Fardan Wardhana
Team7 FSI
fardan.wardhana@gmail.com
https://github.com/fardan-dev/hackathon-qoder
