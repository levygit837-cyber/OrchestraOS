package agent

import (
	"context"
	"os"
	"testing"
	"time"
)

// TestGeminiRealInference faz uma chamada real à API Gemini para validar inferência.
// Requer GEMINI_API_KEY configurada. Usa gemini-2.5-flash como modelo.
func TestGeminiRealInference(t *testing.T) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		t.Skip("GEMINI_API_KEY não configurada — pulando teste de inferência real")
	}

	// Força o modelo padrão para o teste
	os.Setenv("GEMINI_MODEL", "gemini-3-flash-preview")

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	rt := NewGeminiRuntime()
	config := RuntimeConfig{
		RunID:        "run-test-001",
		WorkUnitID:   "wu-test-001",
		TaskID:       "task-test-001",
		AgentID:      "agent-test-001",
		Prompt:       "Respond with exactly the word 'INFERENCE_OK' and nothing else.",
		SystemPrompt: "You are a test agent. Follow instructions precisely.",
		TaskPrompt:   "Respond with exactly 'INFERENCE_OK'.",
		MaxSteps:     3,
		Timeout:      30,
	}

	t.Log("Iniciando GeminiRuntime...")
	if err := rt.Start(ctx, config); err != nil {
		t.Fatalf("Falha ao iniciar runtime: %v", err)
	}

	t.Log("Coletando eventos...")
	var gotCompleted bool
	for {
		event, err := rt.ReceiveEvent(ctx)
		if err != nil {
			t.Logf("ReceiveEvent retornou erro (esperado após completion): %v", err)
			break
		}

		t.Logf("Evento recebido: %s", event.Type)

		switch event.Type {
		case "agent.completed":
			gotCompleted = true
			t.Logf("Payload: %s", string(event.Payload))
			goto done
		case "agent.failed":
			t.Logf("Falha no runtime: %s", string(event.Payload))
			t.Fatalf("Runtime falhou durante inferência")
		}
	}
done:

	if !gotCompleted {
		t.Fatal("Não recebeu evento agent.completed")
	}

	status := rt.Status()
	if status.State != "completed" {
		t.Fatalf("Status final esperado 'completed', got '%s'", status.State)
	}

	t.Log("Inferência real concluída com sucesso!")
}
