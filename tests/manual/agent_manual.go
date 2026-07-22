package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/levygit837-cyber/OrchestraOS/internal/domain"
	"github.com/levygit837-cyber/OrchestraOS/internal/provider/deepseek"
	"github.com/levygit837-cyber/OrchestraOS/internal/provider/gemini"
	"github.com/levygit837-cyber/OrchestraOS/internal/runtime"
)

func main() {
	ctx := context.Background()

	// WorkUnit and Task for testing
	task := &domain.Task{
		ID:          "test-task-001",
		Title:       "Testar agente LLM",
		Description: "Verificar se o agente consegue responder a uma pergunta simples de forma direta e correta.",
		AcceptanceCriteria: []string{
			"Responder com uma saudação",
			"Mencionar o nome do modelo",
			"Resposta deve ter menos de 200 palavras",
		},
	}

	wu := &domain.WorkUnit{
		ID:        "test-wu-001",
		TaskID:    task.ID,
		Title:     "Saudação e identificação",
		Objective: "Responda com uma saudação amigável e diga qual modelo você é. Seja breve.",
		Status:    domain.WorkUnitStatusCreated,
	}

	// Test DeepSeek
	deepseekKey := os.Getenv("DEEPSEEK_API_KEY")
	if deepseekKey == "" {
		log.Println("[SKIP] DEEPSEEK_API_KEY não está definida")
	} else {
		fmt.Println("\n========== DEEPSEEK ==========")
		testAgent(ctx, "deepseek", "deepseek-v4-flash", deepseekKey, wu, task)
	}

	// Test Gemini models
	geminiKey := os.Getenv("GEMINI_API_KEY")
	if geminiKey == "" {
		log.Println("[SKIP] GEMINI_API_KEY não está definida")
		return
	}

	geminiModels := []string{
		"gemini-2.5-pro",
		"gemini-2.5-flash",
		"gemini-3.5-flash",
		"gemini-3-pro-preview",
	}

	for _, model := range geminiModels {
		fmt.Printf("\n========== GEMINI (%s) ==========\n", model)
		testAgent(ctx, "gemini", model, geminiKey, wu, task)
		// Small delay between requests to avoid rate limiting
		time.Sleep(2 * time.Second)
	}
}

func testAgent(ctx context.Context, provider, model, apiKey string, wu *domain.WorkUnit, task *domain.Task) {
	var rt domain.Runtime

	switch provider {
	case "deepseek":
		rt = deepseek.New(runtime.Config{
			APIKey: apiKey,
			Model:  model,
		})
	case "gemini":
		rt = gemini.New(runtime.Config{
			APIKey: apiKey,
			Model:  model,
		})
	default:
		log.Fatalf("Provedor desconhecido: %s", provider)
	}

	start := time.Now()
	result, err := rt.Execute(ctx, wu, task)
	elapsed := time.Since(start)

	if err != nil {
		fmt.Printf("❌ ERRO: %v\n", err)
		return
	}

	fmt.Printf("✅ SUCESSO\n")
	fmt.Printf("   Provedor:   %s\n", result.Provider)
	fmt.Printf("   Modelo:     %s\n", result.Model)
	fmt.Printf("   Tokens:     %d\n", result.TokensUsed)
	fmt.Printf("   Latência:   %v\n", elapsed)
	fmt.Printf("   Resposta:   %s\n", truncate(result.Output, 300))
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}
