// Package validation provides input validation primitives.
//
// # Responsibility
// Validates primitive fields (UUIDs, text, lists, enums) at system boundaries.
// Every function returns an apperrors.Error with a descriptive message.
//
// # Key Types
//   - RequiredUUID, OptionalUUID: UUID validation
//   - RequiredText: non-empty trimmed string
//   - StringList: non-empty slice with no blank elements
//   - Priority: P0-P3 validation
//   - RiskLevel: low/medium/high/critical validation
//   - Runtime: codex_cli/fake/external/gemini validation
//
// # Dependencies
//   - core/apperrors: error typing
//
// # Related Packages
//   - All modules/: call these functions at the start of service methods
package validation
