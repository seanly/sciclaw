package main

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/sipeed/picoclaw/pkg/providers"
)

type scriptedPhiEvalProvider struct {
	responses []*providers.LLMResponse
	errs      []error
	callIndex int
}

func (p *scriptedPhiEvalProvider) Chat(_ context.Context, _ []providers.Message, _ []providers.ToolDefinition, _ string, _ map[string]interface{}) (*providers.LLMResponse, error) {
	idx := p.callIndex
	p.callIndex++
	if idx < len(p.errs) && p.errs[idx] != nil {
		return nil, p.errs[idx]
	}
	if idx >= len(p.responses) {
		return &providers.LLMResponse{}, nil
	}
	return p.responses[idx], nil
}

func TestRunPhiTextEval_PassesOnExpectedToken(t *testing.T) {
	provider := &scriptedPhiEvalProvider{
		responses: []*providers.LLMResponse{{
			Content:      "READY_LOCAL_TEST",
			FinishReason: "stop",
			Diagnostics: &providers.ResponseDiagnostics{
				ContentSource: "thinking",
			},
		}},
	}

	result := runPhiTextEval(provider, "qwen3.5:9b", 2*time.Second)
	if !result.Passed {
		t.Fatalf("expected pass, got %+v", result)
	}
	if result.Diagnostics == nil || result.Diagnostics.ContentSource != "thinking" {
		t.Fatalf("diagnostics=%+v", result.Diagnostics)
	}
}

func TestRunPhiJSONEval_FailsOnInvalidJSON(t *testing.T) {
	provider := &scriptedPhiEvalProvider{
		responses: []*providers.LLMResponse{{
			Content:      "```json\n{\"status\":\"ok\"}\n```",
			FinishReason: "stop",
		}},
	}

	result := runPhiJSONEval(provider, "qwen3.5:9b", 2*time.Second)
	if result.Passed {
		t.Fatalf("expected failure, got %+v", result)
	}
	if result.Note == "" {
		t.Fatalf("expected parse note, got %+v", result)
	}
}

func TestRunPhiToolEval_PassesWithRecoveredToolCall(t *testing.T) {
	provider := &scriptedPhiEvalProvider{
		responses: []*providers.LLMResponse{
			{
				Content:      "",
				FinishReason: "tool_calls",
				ToolCalls: []providers.ToolCall{{
					ID:   "call_1",
					Name: "word_count",
					Arguments: map[string]interface{}{
						"text": "alpha beta gamma delta",
					},
				}},
				Diagnostics: &providers.ResponseDiagnostics{
					ToolCallSource: "thinking",
				},
			},
			{
				Content:      "4",
				FinishReason: "stop",
			},
		},
	}

	result := runPhiToolEval(provider, "qwen3.5:9b", 2*time.Second)
	if !result.Passed {
		t.Fatalf("expected pass, got %+v", result)
	}
	if result.ToolCalls != 1 {
		t.Fatalf("toolCalls=%d want 1", result.ToolCalls)
	}
	if result.Diagnostics == nil || result.Diagnostics.ToolCallSource != "thinking" {
		t.Fatalf("diagnostics=%+v", result.Diagnostics)
	}
}

func TestRunPhiToolEval_ReportsProviderError(t *testing.T) {
	provider := &scriptedPhiEvalProvider{
		errs: []error{errors.New("timeout")},
	}

	result := runPhiToolEval(provider, "qwen3.5:9b", 2*time.Second)
	if result.Passed {
		t.Fatalf("expected failure, got %+v", result)
	}
	if result.Note != "timeout" {
		t.Fatalf("note=%q want timeout", result.Note)
	}
}
