package serialization

import (
	"encoding/json"

	"github.com/levygit837-cyber/OrchestraOS/internal/core/apperrors"
)

// MarshalPayload JSON-serializes a map. Returns `{}` for nil input.
func MarshalPayload(op string, payload map[string]interface{}) (json.RawMessage, error) {
	if payload == nil {
		payload = map[string]interface{}{}
	}
	bytes, err := json.Marshal(payload)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.CodeValidation, op, err)
	}
	return bytes, nil
}
