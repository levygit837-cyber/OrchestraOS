// Package apperrors provides typed error codes for the entire system.
//
// # Responsibility
// Defines canonical error codes (InvalidInput, NotFound, Conflict, Persistence,
// InvalidTransition, Runtime, External, Internal, Validation) and a structured
// Error type that carries code, operation, and wrapped cause.
//
// # Key Types
//   - Error: structured error with Code, Op, Msg, and Cause
//   - Code*: constant error codes
//   - New, Wrap: constructors
//
// # Dependencies
//   - None (this is a leaf package)
//
// # Related Packages
//   - All other packages: use apperrors for consistent error typing
package apperrors
