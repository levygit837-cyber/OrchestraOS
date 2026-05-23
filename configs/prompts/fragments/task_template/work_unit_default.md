# Task Prompt

Task: {{ .TaskTitle }}
Task ID: {{ .TaskID }}
Task Graph ID: {{ .TaskGraphID }}
Work Unit ID: {{ .WorkUnitID }}
Agent Profile: {{ .AgentProfile }}
System Persona: {{ .SystemPersona }}
Operating Mode: {{ .SystemOperatingMode }}
Technical Domain: {{ .SystemTechnicalDomain }}

## Objective
{{ .WorkUnitObjective }}

## System-Aligned Task Focus
{{ .TaskExecutionFocus }}

## Context
{{ .TaskDescription }}

## TaskPromptDecompose
This prompt is for exactly one WorkUnit from the TaskGraph.

- WorkUnit title: {{ .WorkUnitTitle }}
- WorkUnit ID: {{ .WorkUnitID }}
- Dependencies: {{ .DependsOnInline }}
- Do not take ownership of sibling WorkUnits.
- Do not edit paths outside Owned Paths.
- If another WorkUnit is required, emit a structured blocker instead of expanding scope.

## Owned Paths
{{ .OwnedPaths }}

## Read Paths
{{ .ReadPaths }}

## Acceptance Criteria
{{ .AcceptanceCriteria }}

## Validation Plan
{{ .ValidationPlan }}

## Toolset
{{ .Toolset }}

## System Constraints Applied To This Task
- Allowed operations: {{ .SystemAllowedOperations }}
- Denied operations: {{ .SystemDeniedOperations }}
- Approval required: {{ .SystemApprovalRequired }}
- Output contract: {{ .SystemOutputContract }}
- Category signature: {{ .SystemCategorySignature }}

## Initial Todo Ledger
- [ ] Confirm authorized scope and owned paths.
- [ ] Inspect the minimum required context.
- [ ] Implement or review the smallest sufficient change.
- [ ] Run the validation plan or record why it cannot run.
- [ ] Emit checkpoint evidence, remaining risks, and completion summary.
