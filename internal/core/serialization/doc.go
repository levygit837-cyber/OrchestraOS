// Package serialization provides event-payload marshalling helpers.
//
// # Responsibility
// Converts Go values into JSON payloads suitable for event envelopes.
// Thin wrapper around encoding/json with error typing.
//
// # Key Types
//   - MarshalPayload: serialises any value to JSON bytes
//
// # Dependencies
//   - core/apperrors: error typing
//
// # Related Packages
//   - core/orchestration/: uses MarshalPayload when building transition events
//   - modules/*: service methods use it to persist structured data
package serialization
