# Module: trigger

## Purpose

Manages configurable anomaly detection and threshold triggers for runs, agent sessions, and work units.

## File Map

| File | Responsibility |
|---|---|
| `doc.go` | Package godoc |
| `contract.go` | ModuleContract for LLM agents |
| `models.go` | Domain type aliases |
| `queries.go` | SQL constants |
| `repository.go` | Pure CRUD |
| `fetch.go` | `RequireByID` helper |
| `events.go` | Event type mapping |
| `service.go` | `TriggerService` with Create, EvaluateRun, EvaluateSession, EvaluateWorkUnit, Resolve, Dismiss, ListActive, ListByRun |
| `detectors.go` | Deterministic anomaly detectors |
| `thresholds.go` | ThresholdConfig defaults and validation |
| `validation.go` | Input validation |

## Dependencies

- `core/db`: transaction helpers
- `core/transition`: event emission
- `core/apperrors`: error handling
- `core/validation`: input validation
- `domain`: Trigger types

## Related Packages

- `run/`: triggers evaluate runs
- `agentsession/`: triggers evaluate sessions
- `workunit/`: triggers evaluate work units
