package contracts_test

import (
	"bytes"
	"encoding/json"
	"io/fs"
	"strings"
	"testing"

	"github.com/levygit837-cyber/OrchestraOS/contracts"
	"github.com/santhosh-tekuri/jsonschema/v6"
)

type schemaCase struct {
	name           string
	path           string
	valid          string
	requiredField  string
	enumField      string
	invalidEnumVal string
}

func TestSchemasCompile(t *testing.T) {
	err := fs.WalkDir(contracts.Schemas, contracts.SchemaRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".schema.json") {
			return nil
		}

		t.Run(path, func(t *testing.T) {
			compileSchema(t, path)
		})

		return nil
	})
	if err != nil {
		t.Fatalf("walk schemas: %v", err)
	}
}

func TestSchemasAcceptMinimalValidInstances(t *testing.T) {
	for _, tc := range schemaCases() {
		t.Run(tc.name, func(t *testing.T) {
			schema := compileSchema(t, tc.path)
			if err := schema.Validate(decodeJSON(t, tc.valid)); err != nil {
				t.Fatalf("expected valid instance: %v", err)
			}
		})
	}
}

func TestSchemasRejectMissingRequiredFields(t *testing.T) {
	for _, tc := range schemaCases() {
		t.Run(tc.name, func(t *testing.T) {
			schema := compileSchema(t, tc.path)
			instance := decodeObject(t, tc.valid)
			delete(instance, tc.requiredField)

			if err := schema.Validate(instance); err == nil {
				t.Fatalf("expected missing required field %q to be rejected", tc.requiredField)
			}
		})
	}
}

func TestSchemasRejectInvalidEnums(t *testing.T) {
	for _, tc := range schemaCases() {
		if tc.enumField == "" {
			continue
		}
		t.Run(tc.name, func(t *testing.T) {
			schema := compileSchema(t, tc.path)
			instance := decodeObject(t, tc.valid)
			instance[tc.enumField] = tc.invalidEnumVal

			if err := schema.Validate(instance); err == nil {
				t.Fatalf("expected invalid enum value for %q to be rejected", tc.enumField)
			}
		})
	}
}

func TestSchemasRejectUnknownProperties(t *testing.T) {
	for _, tc := range schemaCases() {
		t.Run(tc.name, func(t *testing.T) {
			schema := compileSchema(t, tc.path)
			instance := decodeObject(t, tc.valid)
			instance["unexpected_field"] = "must be rejected"

			if err := schema.Validate(instance); err == nil {
				t.Fatalf("expected unknown property to be rejected")
			}
		})
	}
}

func TestEventEnvelopeAllowsTaskEventsWithoutRun(t *testing.T) {
	schema := compileSchema(t, "schemas/protocol/event-envelope.schema.json")

	instance := decodeObject(t, `{
		"id": "evt_001",
		"type": "task.created",
		"version": "v1",
		"task_id": "task_001",
		"sequence": 1,
		"priority": "notification",
		"requires_ack": false,
		"created_at": "2026-05-03T12:00:00Z",
		"payload": {}
	}`)

	if err := schema.Validate(instance); err != nil {
		t.Fatalf("expected task event without run_id to be valid: %v", err)
	}
}

func TestEventEnvelopeRejectsEmptyOptionalRun(t *testing.T) {
	schema := compileSchema(t, "schemas/protocol/event-envelope.schema.json")

	instance := decodeObject(t, `{
		"id": "evt_001",
		"type": "task.created",
		"version": "v1",
		"task_id": "task_001",
		"run_id": "",
		"sequence": 1,
		"priority": "notification",
		"requires_ack": false,
		"created_at": "2026-05-03T12:00:00Z",
		"payload": {}
	}`)

	if err := schema.Validate(instance); err == nil {
		t.Fatalf("expected empty run_id to be rejected when present")
	}
}

func TestEventEnvelopeRequiresRunForRuntimeEvents(t *testing.T) {
	schema := compileSchema(t, "schemas/protocol/event-envelope.schema.json")

	instance := decodeObject(t, `{
		"id": "evt_001",
		"type": "agent.started",
		"version": "v1",
		"task_id": "task_001",
		"sequence": 1,
		"priority": "notification",
		"requires_ack": false,
		"created_at": "2026-05-03T12:00:00Z",
		"payload": {}
	}`)

	if err := schema.Validate(instance); err == nil {
		t.Fatalf("expected agent event without run_id to be rejected")
	}
}

func compileSchema(t *testing.T, path string) *jsonschema.Schema {
	t.Helper()

	raw, err := contracts.Schemas.ReadFile(path)
	if err != nil {
		t.Fatalf("read schema %s: %v", path, err)
	}

	doc, err := jsonschema.UnmarshalJSON(bytes.NewReader(raw))
	if err != nil {
		t.Fatalf("unmarshal schema %s: %v", path, err)
	}

	compiler := jsonschema.NewCompiler()
	compiler.DefaultDraft(jsonschema.Draft2020)
	compiler.AssertFormat()
	if err := compiler.AddResource(path, doc); err != nil {
		t.Fatalf("add schema resource %s: %v", path, err)
	}

	schema, err := compiler.Compile(path)
	if err != nil {
		t.Fatalf("compile schema %s: %v", path, err)
	}

	return schema
}

func decodeJSON(t *testing.T, raw string) any {
	t.Helper()

	decoder := json.NewDecoder(strings.NewReader(raw))
	decoder.UseNumber()

	var value any
	if err := decoder.Decode(&value); err != nil {
		t.Fatalf("decode json: %v", err)
	}
	return value
}

func decodeObject(t *testing.T, raw string) map[string]any {
	t.Helper()

	value, ok := decodeJSON(t, raw).(map[string]any)
	if !ok {
		t.Fatalf("expected json object")
	}
	return value
}

func schemaCases() []schemaCase {
	return []schemaCase{
		{
			name:           "Task",
			path:           "schemas/domain/task.schema.json",
			requiredField:  "title",
			enumField:      "status",
			invalidEnumVal: "done",
			valid: `{
				"id": "task_001",
				"title": "Criar contratos M0",
				"description": "Primeiro corte executavel de contratos.",
				"status": "created",
				"priority": "P1",
				"risk_level": "medium",
				"created_from_message_id": "msg_001",
				"acceptance_criteria": ["Schemas executaveis existem"],
				"created_at": "2026-05-03T12:00:00Z",
				"updated_at": "2026-05-03T12:00:00Z"
			}`,
		},
		{
			name:           "TaskGraph",
			path:           "schemas/domain/task-graph.schema.json",
			requiredField:  "planner_strategy",
			enumField:      "status",
			invalidEnumVal: "draft",
			valid: `{
				"id": "graph_001",
				"task_id": "task_001",
				"version": 1,
				"status": "active",
				"planner_strategy": "local_heuristic_v1",
				"rationale": "Local decomposition from acceptance criteria.",
				"created_by": "test",
				"node_count": 2,
				"edge_count": 1,
				"created_at": "2026-05-03T12:00:00Z",
				"updated_at": "2026-05-03T12:00:00Z"
			}`,
		},
		{
			name:           "Run",
			path:           "schemas/domain/run.schema.json",
			requiredField:  "attempt",
			enumField:      "status",
			invalidEnumVal: "started",
			valid: `{
				"id": "run_001",
				"task_id": "task_001",
				"work_unit_id": "wu_001",
				"status": "running",
				"attempt": 1,
				"started_at": "2026-05-03T12:05:00Z",
				"finished_at": null,
				"result": null,
				"failure_reason": null
			}`,
		},
		{
			name:           "WorkUnit",
			path:           "schemas/domain/work-unit.schema.json",
			requiredField:  "objective",
			enumField:      "status",
			invalidEnumVal: "waiting",
			valid: `{
				"id": "wu_001",
				"task_id": "task_001",
				"task_graph_id": "graph_001",
				"title": "Criar schemas de dominio",
				"objective": "Definir contratos executaveis do M0.",
				"assigned_agent_profile": "codex",
				"status": "planned",
				"owned_paths": ["contracts/schemas/"],
				"read_paths": ["docs/contracts/json-schemas.md"],
				"acceptance_criteria": ["Schemas compilam"],
				"validation_plan": ["go test ./..."],
				"depends_on": []
			}`,
		},
		{
			name:           "Agent",
			path:           "schemas/domain/agent.schema.json",
			requiredField:  "runtime_type",
			enumField:      "runtime_type",
			invalidEnumVal: "codex_web",
			valid: `{
				"id": "agent_001",
				"name": "Codex Worker",
				"profile": "general_engineering",
				"capabilities": ["code_edit", "test_run"],
				"allowed_tools": ["shell.read", "shell.test"],
				"default_prompt_fragments": ["project_instructions"],
				"runtime_type": "codex_cli"
			}`,
		},
		{
			name:           "AgentSession",
			path:           "schemas/domain/agent-session.schema.json",
			requiredField:  "connection_id",
			enumField:      "status",
			invalidEnumVal: "connected",
			valid: `{
				"id": "as_001",
				"agent_id": "agent_001",
				"run_id": "run_001",
				"sandbox_id": "sandbox_001",
				"connection_id": "conn_001",
				"status": "running",
				"last_heartbeat_at": "2026-05-03T12:06:00Z",
				"last_checkpoint_at": null
			}`,
		},
		{
			name:           "PromptFragment",
			path:           "schemas/domain/prompt-fragment.schema.json",
			requiredField:  "body_hash",
			enumField:      "kind",
			invalidEnumVal: "freeform",
			valid: `{
				"id": "fragment.policy.global",
				"version": "1.0.0",
				"category": "policy.global",
				"kind": "global_policy",
				"title": "Global Prompt Policy",
				"priority": 100,
				"exclusive_group": "global_policy",
				"body_hash": "sha256:68b9a174cc95bea839d40eb07ddbd7c5a17feb7f900d61ca1d613205f9da1384",
				"metadata_hash": "sha256:ef73569da6f8fcf17d738ddea2361a2b0b64a51071184667990a98d0be293c88",
				"body": "Repository is source of truth.",
				"applies_when": {},
				"requires": [],
				"conflicts_with": [],
				"allows": ["read_authorized_context"],
				"denies": ["raise_autonomy_level"],
				"approval_required": ["network_access"],
				"autonomy_level": 0,
				"created_at": "2026-05-03T12:00:00Z",
				"updated_at": "2026-05-03T12:00:00Z"
			}`,
		},
		{
			name:          "PromptSnapshot",
			path:          "schemas/domain/prompt-snapshot.schema.json",
			requiredField: "combined_prompt_hash",
			valid: `{
				"id": "ps_001",
				"run_id": "run_001",
				"work_unit_id": "wu_001",
				"agent_session_id": "as_001",
				"system_prompt": "System prompt",
				"task_prompt": "Task prompt",
				"combined_prompt": "System prompt\n\nTask prompt",
				"system_prompt_hash": "sha256:68b9a174cc95bea839d40eb07ddbd7c5a17feb7f900d61ca1d613205f9da1384",
				"task_prompt_hash": "sha256:ef73569da6f8fcf17d738ddea2361a2b0b64a51071184667990a98d0be293c88",
				"combined_prompt_hash": "sha256:1ac4d66aa4e5adc769f549f82588b64265113d38d6be2a17cff8cd328d716124",
				"composition_hash": "sha256:8880a90cbad0dd05c597004a491c3cb2a3b7bad661ceabf839ffdfb154a80706",
				"category_signature": "sha256:abd5318eefb44a71edf41079221c27a1822a3d7997058934082bd7a277e93165",
				"fragment_refs": [{
					"id": "fragment.policy.global",
					"version": "1.0.0",
					"category": "policy.global",
					"kind": "global_policy",
					"order": 1,
					"body_hash": "sha256:68b9a174cc95bea839d40eb07ddbd7c5a17feb7f900d61ca1d613205f9da1384",
					"metadata_hash": "sha256:ef73569da6f8fcf17d738ddea2361a2b0b64a51071184667990a98d0be293c88",
					"title": "Global Prompt Policy"
				}],
				"assembly_order": ["fragment.policy.global@1.0.0"],
				"variables_applied": {"TaskID": "task_001"},
				"count_used": 1,
				"first_used_at": "2026-05-03T12:00:00Z",
				"last_used_at": "2026-05-03T12:00:00Z",
				"created_at": "2026-05-03T12:00:00Z"
			}`,
		},
		{
			name:           "ToolsetSnapshot",
			path:           "schemas/domain/toolset-snapshot.schema.json",
			requiredField:  "tools",
			enumField:      "tools",
			invalidEnumVal: "not-used",
			valid: `{
				"id": "ts_001",
				"run_id": "run_001",
				"agent_session_id": "as_001",
				"tools": [{
					"name": "filesystem.read",
					"scope": "approved_read_paths",
					"risk": "safe",
					"reason": "Read authorized context."
				}],
				"created_reason": "minimum toolset for code_worker profile",
				"created_at": "2026-05-03T12:00:00Z"
			}`,
		},
		{
			name:           "EventEnvelope",
			path:           "schemas/protocol/event-envelope.schema.json",
			requiredField:  "sequence",
			enumField:      "priority",
			invalidEnumVal: "urgent",
			valid: `{
				"id": "evt_001",
				"type": "task.created",
				"version": "v1",
				"task_id": "task_001",
				"run_id": "run_001",
				"work_unit_id": "wu_001",
				"agent_id": "agent_001",
				"trace_id": "trace_001",
				"span_id": "span_001",
				"sequence": 0,
				"priority": "notification",
				"requires_ack": false,
				"created_at": "2026-05-03T12:00:00Z",
				"payload": {}
			}`,
		},
	}
}
