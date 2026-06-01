package batch

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"google.golang.org/genai"
)

const (
	ModelGemini3_5Flash     = "gemini-3.5-flash"
	ModelGemini3_1FlashLite = "gemini-3.1-flash-lite"
)

type Gemini struct {
	key    string
	client *genai.Client

	Model string

	generateFileName func() string
}

func NewGemini(key string, client *genai.Client, model string) *Gemini {
	return &Gemini{
		key:              key,
		client:           client,
		Model:            model,
		generateFileName: defaultGenerateFileName,
	}
}

func (ai *Gemini) RequestSeriesNormalizeBatch(ctx context.Context, displayName string, requests []SeriesNormalizeRequest) (string, error) {
	systemPrompt := parseSeriesNormalizeSystemPrompt()
	generationConfig := &genai.GenerationConfig{
		ResponseMIMEType: "application/json",
	}

	contentBuilder := strings.Builder{}
	for i, request := range requests {
		promptText := convertSeriesNormalizePrompt(request)

		batchRequest := jobRequest{
			Key: strconv.Itoa(i),
			Request: contentRequest{
				Contents:          genai.Text(promptText),
				SystemInstruction: genai.NewContentFromText(systemPrompt, genai.RoleUser),
				GenerationConfig:  generationConfig,
			},
		}

		batchRequestJson, _ := json.Marshal(batchRequest)
		contentBuilder.Write(batchRequestJson)
		contentBuilder.WriteString("\n")
	}

	path, err := createRequestJsonl(ai.generateFileName(), contentBuilder.String())
	if err != nil {
		return "", err
	}
	defer func() {
		_ = os.Remove(path)
	}()

	cloudJsonl, err := ai.client.Files.UploadFromPath(ctx, path, &genai.UploadFileConfig{
		MIMEType: "jsonl",
		Name:     ai.generateFileName(),
	})
	if err != nil {
		return "", err
	}

	jobSource := genai.BatchJobSource{
		FileName: cloudJsonl.Name,
	}
	jobConfig := genai.CreateBatchJobConfig{
		DisplayName: displayName,
	}

	if job, err := ai.client.Batches.Create(ctx, ai.Model, &jobSource, &jobConfig); err == nil {
		return job.Name, nil
	} else {
		return "", err
	}
}

func (ai *Gemini) GetSeriesNormalizeBatch(ctx context.Context, jobName string) (Status, []SeriesNormalizeBatchResult, error) {
	batch, err := ai.client.Batches.Get(ctx, jobName, nil)
	if err != nil {
		return StatusFailed, nil, err
	}

	var responses []SeriesNormalizeBatchResult
	if batch.State == genai.JobStateSucceeded {
		output, err := ai.client.Files.Get(ctx, batch.Dest.FileName, nil)
		if err != nil {
			return StatusFailed, nil, err
		}

		var batchResponses []jobResponse
		consumer := func(line []byte) error {
			var resp jobResponse
			if err = json.Unmarshal(line, &resp); err != nil {
				return err
			}
			batchResponses = append(batchResponses, resp)
			return nil
		}

		if err = downloadOutputFile(output.DownloadURI, ai.key, consumer); err != nil {
			return StatusFailed, nil, err
		}

		for _, batchResponse := range batchResponses {
			var resp SeriesNormalizeResponse
			if batchResponse.Response == nil {
				return StatusFailed, nil, fmt.Errorf("batch name \"%s\", key \"%s\" response is nil", batch.Name, batchResponse.Key)
			}
			if err = json.Unmarshal([]byte(batchResponse.Response.Text()), &resp); err != nil {
				return StatusFailed, nil, err
			}
			responses = append(responses, SeriesNormalizeBatchResult{
				Key:      batchResponse.Key,
				Response: resp,
			})
		}
	}

	return convertGeminiJobStatus(batch.State), responses, nil
}

func createRequestJsonl(filename, content string) (string, error) {
	filePath := filepath.Join(os.TempDir(), filename+".jsonl")
	localJsonl, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer func() {
		_ = localJsonl.Close()
	}()

	bfw := bufio.NewWriter(localJsonl)
	if _, err = bfw.WriteString(content); err != nil {
		return "", err
	}
	if err = bfw.Flush(); err != nil {
		return "", err
	}

	return filePath, nil
}

func downloadOutputFile(path, key string, consumer func([]byte) error) error {
	req, err := http.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return err
	}
	req.Header.Add("x-goog-api-key", key)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Bytes()
		if err = consumer(line); err != nil {
			return err
		}
	}
	return nil
}

func defaultGenerateFileName() string {
	uid, _ := uuid.NewRandom()
	return strings.ReplaceAll(uid.String(), "-", "")
}

func convertSeriesNormalizePrompt(request SeriesNormalizeRequest) string {
	path := filepath.Join("internal/batch/prompt/normalize_prompt_tmpl.md")
	md, err := os.ReadFile(path)
	if err != nil {
		return ""
	}

	tmpl, err := template.New("prompt").Parse(string(md))
	if err != nil {
		return ""
	}

	var buf bytes.Buffer
	if err = tmpl.Execute(&buf, request); err != nil {
		return ""
	}
	return buf.String()
}

func parseSeriesNormalizeSystemPrompt() string {
	path := filepath.Join("internal/batch/prompt/normalize_system_instruction.md")
	md, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return string(md)
}

func convertGeminiJobStatus(state genai.JobState) Status {
	switch state {
	case genai.JobStatePending, genai.JobStateQueued:
		return StatusPending
	case genai.JobStateRunning:
		return StatusRunning
	case genai.JobStateCancelling, genai.JobStateCancelled, genai.JobStateExpired, genai.JobStatePaused:
		return StatusCancelled
	case genai.JobStateFailed:
		return StatusFailed
	case genai.JobStateSucceeded:
		return StatusDone
	default:
		return StatusUndefined
	}
}

type jobRequest struct {
	Key     string         `json:"key"`
	Request contentRequest `json:"request"`
}

type jobResponse struct {
	Key      string                         `json:"key"`
	Response *genai.GenerateContentResponse `json:"response"`
}

type contentRequest struct {
	Contents          []*genai.Content `json:"contents"`
	SystemInstruction *genai.Content   `json:"systemInstruction"`

	GenerationConfig *genai.GenerationConfig `json:"generationConfig"`
}
