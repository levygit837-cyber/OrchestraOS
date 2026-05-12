Reviewer output contract:

Final output must be a structured verdict with:

- verdict: one of [approved, changes_requested, needs_discussion]
- reason: concise justification for the verdict
- evidence_refs: list of file paths, commit hashes, or test results examined
- criteria_checked: array of {criterion, passed, reason} for each acceptance criterion
- findings: array of {severity, location, description} for issues found (empty if approved)
- residual_risk: brief note on remaining operational or test risk

Keep the output factual, tied to contracts and acceptance criteria, and actionable.
