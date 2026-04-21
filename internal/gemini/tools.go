package gemini

import (
	"bytes"
	"fmt"
	"google.golang.org/genai"
	"html/template"
	"os"
	"path/filepath"
	"series-batch-go/internal/pkg/llm"
)

const (
	ModelGemini3FlashPreview      = "gemini-3-flash-preview"
	ModelGemini31ProPreview       = "gemini-3.1-pro-preview"
	ModelGemini31FlashLitePreview = "gemini-3.1-flash-lite-preview"
)

func ConvertSeriesNormalizePrompt(req *llm.SeriesNormalizeRequest) (string, error) {
	tmplPath := filepath.Join("internal/gemini/markdown/normalize_prompt_tmpl.md")
	tmplMD, err := os.ReadFile(tmplPath)
	if err != nil {
		return "", fmt.Errorf("failed to read prompt template: %w", err)
	}
	tmpl, err := template.New("prompt").Parse(string(tmplMD))
	if err != nil {
		return "", fmt.Errorf("failed to parse prompt template: %w", err)
	}

	var buf bytes.Buffer
	if err = tmpl.Execute(&buf, req); err != nil {
		return "", fmt.Errorf("failed to execute prompt template: %w", err)
	}
	return string(buf.Bytes()), nil
}

func ParseSeriesNormalizeSystemPrompt() (string, error) {
	systemPromptPath := filepath.Join("internal/gemini/markdown/normalize_system_instruction.md")
	systemPromptMD, err := os.ReadFile(systemPromptPath)
	if err != nil {
		return "", fmt.Errorf("failed to read system prompt: %w", err)
	}
	return string(systemPromptMD), nil
}

func ConvertJobStatus(state genai.JobState) llm.JobStatus {
	switch state {
	case genai.JobStatePending, genai.JobStateQueued:
		return llm.JobStatusPending
	case genai.JobStateRunning:
		return llm.JobStatusRunning
	case genai.JobStateCancelling, genai.JobStateCancelled, genai.JobStateExpired, genai.JobStatePaused:
		return llm.JobStatusCanceled
	case genai.JobStateFailed:
		return llm.JobStatusFailed
	case genai.JobStateSucceeded:
		return llm.JobStatusDone
	default:
		return llm.JobStatusUndefined
	}
}
