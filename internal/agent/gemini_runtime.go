package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/levygit837-cyber/OrchestraOS/internal/apperrors"
	"github.com/levygit837-cyber/OrchestraOS/internal/domain"
	"google.golang.org/genai"
)

// GeminiRuntime implements the Runtime interface using Google AI Studio (Gemini API).
type GeminiRuntime struct {
	config    RuntimeConfig
	status    RuntimeStatus
	eventChan chan *domain.EventEnvelope
	stopChan  chan struct{}
	started   bool

	client *genai.Client
	model  string

	// Conversation history for multi-turn function calling.
	history []*genai.Content

	// Internal channel to receive tool execution results from SendEvent.
	toolResponseChan chan *genai.Content

	// Current step counter.
	currentStep int
}

// NewGeminiRuntime creates a new Gemini-backed agent runtime.
func NewGeminiRuntime() *GeminiRuntime {
	return &GeminiRuntime{
		eventChan:        make(chan *domain.EventEnvelope, 100),
		stopChan:         make(chan struct{}),
		toolResponseChan: make(chan *genai.Content, 1),
	}
}

// Start initializes the Gemini client and begins the inference loop.
func (g *GeminiRuntime) Start(ctx context.Context, config RuntimeConfig) error {
	if g.started {
		return fmt.Errorf("runtime already started")
	}

	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		apiKey = os.Getenv("GOOGLE_API_KEY")
	}
	if apiKey == "" {
		return apperrors.New(apperrors.CodeRuntime, "gemini_runtime.start", "GEMINI_API_KEY or GOOGLE_API_KEY environment variable is required")
	}

	g.model = os.Getenv("GEMINI_MODEL")
	if g.model == "" {
		g.model = "gemini-3-flash-preview"
	}

	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey: apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return apperrors.Wrap(apperrors.CodeRuntime, "gemini_runtime.new_client", err)
	}

	g.client = client
	g.config = config
	g.status = RuntimeStatus{
		State:       "starting",
		CurrentStep: 0,
	}
	g.started = true
	g.history = make([]*genai.Content, 0)

	g.emitEvent("agent.connected", "v1", map[string]interface{}{
		"agent_id": config.AgentID,
		"run_id":   config.RunID,
		"status":   "connected",
	})

	g.status.State = "running"
	g.emitEvent("agent.started", "v1", map[string]interface{}{
		"agent_id":            config.AgentID,
		"run_id":              config.RunID,
		"work_unit":           config.WorkUnitID,
		"prompt_hash":         config.PromptHash,
		"prompt_snapshot_id":  config.PromptSnapshotID,
		"toolset_snapshot_id": config.ToolsetSnapshotID,
		"toolset":             config.Toolset,
	})

	go g.heartbeatLoop()
	go g.inferenceLoop(ctx)

	return nil
}

// Stop halts the runtime.
func (g *GeminiRuntime) Stop(ctx context.Context) error {
	if !g.started {
		return nil
	}

	close(g.stopChan)
	g.status.State = "stopped"
	g.started = false

	g.emitEvent("agent.stopped", "v1", map[string]interface{}{
		"agent_id": g.config.AgentID,
		"run_id":   g.config.RunID,
		"reason":   "requested",
	})

	return nil
}

// SendEvent delivers external events (e.g., tool approvals) into the runtime.
func (g *GeminiRuntime) SendEvent(ctx context.Context, event *domain.EventEnvelope) error {
	switch event.Type {
	case "tool.approved":
		var payload struct {
			Tool   string                 `json:"tool"`
			Input  map[string]interface{} `json:"input"`
			Result map[string]interface{} `json:"result,omitempty"`
		}
		if err := json.Unmarshal(event.Payload, &payload); err != nil {
			payload.Tool = "unknown"
			payload.Input = map[string]interface{}{}
			payload.Result = map[string]interface{}{"status": "approved"}
		}
		if payload.Result == nil {
			payload.Result = map[string]interface{}{"status": "approved"}
		}

		funcResp := genai.NewContentFromFunctionResponse(payload.Tool, payload.Result, genai.RoleUser)
		select {
		case g.toolResponseChan <- funcResp:
		case <-ctx.Done():
			return ctx.Err()
		case <-g.stopChan:
			return fmt.Errorf("runtime stopped")
		}

	case "tool.denied":
		var payload struct {
			Tool   string `json:"tool"`
			Reason string `json:"reason"`
		}
		_ = json.Unmarshal(event.Payload, &payload)
		if payload.Tool == "" {
			payload.Tool = "unknown"
		}

		funcResp := genai.NewContentFromFunctionResponse(payload.Tool, map[string]interface{}{
			"status": "denied",
			"reason": payload.Reason,
		}, genai.RoleUser)
		select {
		case g.toolResponseChan <- funcResp:
		case <-ctx.Done():
			return ctx.Err()
		case <-g.stopChan:
			return fmt.Errorf("runtime stopped")
		}
	}
	return nil
}

// ReceiveEvent reads events emitted by the runtime.
func (g *GeminiRuntime) ReceiveEvent(ctx context.Context) (*domain.EventEnvelope, error) {
	select {
	case event := <-g.eventChan:
		return event, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-g.stopChan:
		return nil, fmt.Errorf("runtime stopped")
	}
}

// Status returns the current runtime status.
func (g *GeminiRuntime) Status() RuntimeStatus {
	return g.status
}

func (g *GeminiRuntime) heartbeatLoop() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if !g.started {
				return
			}
			g.emitHeartbeat()
		case <-g.stopChan:
			return
		}
	}
}

func (g *GeminiRuntime) emitHeartbeat() {
	g.status.LastHeartbeat = time.Now().Unix()
	g.emitEvent("agent.heartbeat", "v1", map[string]interface{}{
		"agent_id": g.config.AgentID,
		"run_id":   g.config.RunID,
		"count":    g.status.LastHeartbeat,
	})
}

func (g *GeminiRuntime) inferenceLoop(ctx context.Context) {
	defer func() {
		if r := recover(); r != nil {
			g.emitEvent("agent.failed", "v1", map[string]interface{}{
				"agent_id": g.config.AgentID,
				"run_id":   g.config.RunID,
				"reason":   fmt.Sprintf("runtime panic: %v", r),
			})
			g.status.State = "failed"
		}
	}()

	maxSteps := g.config.MaxSteps
	if maxSteps <= 0 {
		maxSteps = 10
	}

	// Build initial user content from the composed prompt.
	userContent := genai.NewContentFromText(g.config.Prompt, genai.RoleUser)

	for g.currentStep < maxSteps && g.started {
		select {
		case <-g.stopChan:
			return
		default:
		}

		g.currentStep++
		g.status.CurrentStep = g.currentStep

		contents := make([]*genai.Content, 0, len(g.history)+1)
		contents = append(contents, g.history...)
		contents = append(contents, userContent)

		tools := g.buildTools()
		config := &genai.GenerateContentConfig{
			SystemInstruction: genai.NewContentFromText(g.config.SystemPrompt, genai.RoleUser),
			Temperature:       genai.Ptr(float32(0.2)),
		}
		if len(tools) > 0 {
			config.Tools = tools
		}

		resp, err := g.client.Models.GenerateContent(ctx, g.model, contents, config)
		if err != nil {
			g.emitEvent("agent.failed", "v1", map[string]interface{}{
				"agent_id": g.config.AgentID,
				"run_id":   g.config.RunID,
				"reason":   fmt.Sprintf("gemini api error: %v", err),
			})
			g.status.State = "failed"
			return
		}

		if resp == nil || len(resp.Candidates) == 0 {
			g.emitEvent("agent.failed", "v1", map[string]interface{}{
				"agent_id": g.config.AgentID,
				"run_id":   g.config.RunID,
				"reason":   "empty response from gemini api",
			})
			g.status.State = "failed"
			return
		}

		candidate := resp.Candidates[0]
		if candidate.Content == nil {
			g.emitEvent("agent.failed", "v1", map[string]interface{}{
				"agent_id": g.config.AgentID,
				"run_id":   g.config.RunID,
				"reason":   "candidate content is nil",
			})
			g.status.State = "failed"
			return
		}

		// Append model response to history.
		modelContent := &genai.Content{
			Role:  genai.RoleModel,
			Parts: candidate.Content.Parts,
		}
		g.history = append(g.history, userContent, modelContent)

		// Check for function calls.
		funcCalls := resp.FunctionCalls()
		if len(funcCalls) > 0 {
			for _, fc := range funcCalls {
				g.emitEvent("agent.tool_requested", "v1", map[string]interface{}{
					"agent_id": g.config.AgentID,
					"run_id":   g.config.RunID,
					"tool":     fc.Name,
					"input":    fc.Args,
					"reason":   fmt.Sprintf("Gemini model requested tool %s", fc.Name),
				})
			}

			// Wait for tool response(s) via SendEvent.
			var toolResponses []*genai.Content
			for range funcCalls {
				select {
				case toolResp := <-g.toolResponseChan:
					toolResponses = append(toolResponses, toolResp)
				case <-ctx.Done():
					g.emitEvent("agent.failed", "v1", map[string]interface{}{
						"agent_id": g.config.AgentID,
						"run_id":   g.config.RunID,
						"reason":   "timeout waiting for tool response",
					})
					g.status.State = "failed"
					return
				case <-g.stopChan:
					return
				}
			}

			// Next turn sends the function responses.
			userContent = genai.NewContentFromParts(toolResponses[0].Parts, genai.RoleUser)
			if len(toolResponses) > 1 {
				allParts := make([]*genai.Part, 0)
				for _, tr := range toolResponses {
					allParts = append(allParts, tr.Parts...)
				}
				userContent = genai.NewContentFromParts(allParts, genai.RoleUser)
			}
			continue
		}

		// Text response.
		text := resp.Text()
		if text != "" {
			g.emitEvent("agent.checkpoint_reached", "v1", map[string]interface{}{
				"checkpoint_id":   uuid.New().String(),
				"current_goal":    fmt.Sprintf("step %d/%d", g.currentStep, maxSteps),
				"minimal_summary": text,
				"ledger": map[string]interface{}{
					"current_goal":    fmt.Sprintf("step %d/%d", g.currentStep, maxSteps),
					"completed_goals": []string{},
					"pending_todos":   []string{},
					"blockers":        []string{},
					"risks":           []string{},
				},
				"files_read":     []string{},
				"files_modified": []string{},
				"evidence_refs":  []string{"gemini-runtime:response"},
			})
		}

		// Check if the model finished naturally.
		if candidate.FinishReason == genai.FinishReasonStop {
			g.emitEvent("agent.completed", "v1", map[string]interface{}{
				"status":    "completed",
				"summary":   text,
				"artifacts": []string{},
				"metrics": map[string]int{
					"steps_executed": g.currentStep,
				},
			})
			g.status.State = "completed"
			return
		}

		// For other finish reasons or if we want another turn, send an empty user prompt.
		userContent = genai.NewContentFromText("Continue.", genai.RoleUser)
	}

	// Max steps reached.
	g.emitEvent("agent.completed", "v1", map[string]interface{}{
		"status":    "completed",
		"summary":   "max steps reached",
		"artifacts": []string{},
		"metrics": map[string]int{
			"steps_executed": g.currentStep,
		},
	})
	g.status.State = "completed"
}

func (g *GeminiRuntime) buildTools() []*genai.Tool {
	if len(g.config.Toolset) == 0 {
		return nil
	}

	decls := make([]*genai.FunctionDeclaration, 0, len(g.config.Toolset))
	for _, toolName := range g.config.Toolset {
		decls = append(decls, &genai.FunctionDeclaration{
			Name:        sanitizeFunctionName(toolName),
			Description: fmt.Sprintf("OrchestraOS tool: %s", toolName),
			Parameters: &genai.Schema{
				Type: genai.TypeObject,
				Properties: map[string]*genai.Schema{
					"reason": {
						Type:        genai.TypeString,
						Description: "Reason for calling this tool.",
					},
				},
			},
		})
	}

	if len(decls) == 0 {
		return nil
	}

	return []*genai.Tool{{
		FunctionDeclarations: decls,
	}}
}

func sanitizeFunctionName(name string) string {
	// Gemini function names must match ^[a-zA-Z_][a-zA-Z0-9_\-\.]{0,64}$
	out := make([]byte, 0, len(name))
	for i, c := range name {
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || c == '_' {
			out = append(out, byte(c))
		} else if (c >= '0' && c <= '9') || c == '-' || c == '.' {
			if i == 0 {
				out = append(out, '_')
			}
			out = append(out, byte(c))
		} else {
			out = append(out, '_')
		}
	}
	if len(out) > 64 {
		out = out[:64]
	}
	if len(out) == 0 {
		return "tool"
	}
	return string(out)
}

func (g *GeminiRuntime) emitEvent(eventType, version string, payload interface{}) {
	payloadBytes, _ := json.Marshal(payload)

	event := &domain.EventEnvelope{
		ID:          uuid.New().String(),
		Type:        eventType,
		Version:     version,
		TaskID:      g.config.TaskID,
		RunID:       g.config.RunID,
		WorkUnitID:  g.config.WorkUnitID,
		AgentID:     g.config.AgentID,
		Sequence:    0,
		Priority:    domain.EventPriorityNotification,
		RequiresAck: false,
		CreatedAt:   time.Now(),
		Payload:     payloadBytes,
	}

	select {
	case g.eventChan <- event:
	default:
	}
}
