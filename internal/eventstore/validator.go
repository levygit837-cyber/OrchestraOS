package eventstore

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/levygit837-cyber/OrchestraOS/contracts"
	jsonschema "github.com/santhosh-tekuri/jsonschema/v6"
)

// Validator handles JSON Schema validation for events
type Validator struct {
	schema *jsonschema.Schema
}

// NewValidator creates a new event envelope validator
func NewValidator() (*Validator, error) {
	schemaBytes, err := contracts.ReadSchema("protocol/event-envelope.schema.json")
	if err != nil {
		return nil, fmt.Errorf("failed to read event envelope schema: %w", err)
	}

	compiler := jsonschema.NewCompiler()
	compiler.DefaultDraft(jsonschema.Draft2020)

	schemaDoc, err := jsonschema.UnmarshalJSON(bytes.NewReader(schemaBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal event envelope schema: %w", err)
	}

	if err := compiler.AddResource("event-envelope.schema.json", schemaDoc); err != nil {
		return nil, fmt.Errorf("failed to add schema to compiler: %w", err)
	}

	schema, err := compiler.Compile("event-envelope.schema.json")
	if err != nil {
		return nil, fmt.Errorf("failed to compile schema: %w", err)
	}

	return &Validator{schema: schema}, nil
}

// Validate validates an event envelope against the JSON schema
func (v *Validator) Validate(envelope interface{}) error {
	var data interface{}

	switch e := envelope.(type) {
	case []byte:
		if err := json.Unmarshal(e, &data); err != nil {
			return fmt.Errorf("failed to unmarshal envelope: %w", err)
		}
	case string:
		if err := json.Unmarshal([]byte(e), &data); err != nil {
			return fmt.Errorf("failed to unmarshal envelope: %w", err)
		}
	case map[string]interface{}:
		data = e
	default:
		// Try to marshal and unmarshal to get map representation
		bytes, err := json.Marshal(envelope)
		if err != nil {
			return fmt.Errorf("failed to marshal envelope: %w", err)
		}
		if err := json.Unmarshal(bytes, &data); err != nil {
			return fmt.Errorf("failed to unmarshal envelope: %w", err)
		}
	}

	if err := v.schema.Validate(data); err != nil {
		return fmt.Errorf("schema validation failed: %w", err)
	}

	return nil
}
