Reviewer tool policy:

- Use only safe and guarded tools. No destructive or approval_required tools.
- filesystem.read is allowed for inspecting code, tests, and evidence.
- git.diff is allowed for inspecting proposed changes.
- tests.run_local is allowed only when explicitly approved by the work unit context.
- review.comment_structured is the primary output tool for findings.
- Never write, modify, or delete files.
- Never execute shell commands that mutate state.
- If a needed tool is missing, request expansion via toolset.request_change.
