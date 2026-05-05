package gemini

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"series-batch-go/internal/pkg/llm"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"google.golang.org/genai"
)

func GenFileName() string {
	uid, _ := uuid.NewRandom()
	uidStr := uid.String()

	return strings.ReplaceAll(uidStr, "-", "")
}

type Client struct {
	key    string
	client *genai.Client

	Model string

	genFileName func() string
}

func NewClient(key string, client *genai.Client, model string) *Client {
	return &Client{
		key:         key,
		client:      client,
		Model:       model,
		genFileName: GenFileName,
	}
}

func (c *Client) RunSeriesNormalize(ctx context.Context, req *llm.SeriesNormalizeRequest) (*llm.SeriesNormalizeResponse, error) {
	systemPrompt, err := ParseSeriesNormalizeSystemPrompt()
	if err != nil {
		return nil, err
	}

	generateConfig := genai.GenerateContentConfig{
		SystemInstruction: genai.NewContentFromText(systemPrompt, genai.RoleUser),
		ResponseMIMEType:  "application/json",
	}
	promptText, err := ConvertSeriesNormalizePrompt(req)
	if err != nil {
		return nil, err
	}
	prompt := genai.Text(promptText)
	content, err := c.client.Models.GenerateContent(ctx, c.Model, prompt, &generateConfig)
	if err != nil {
		return nil, err
	}

	var response llm.SeriesNormalizeResponse
	if err = json.Unmarshal([]byte(content.Text()), &response); err != nil {
		return nil, err
	}
	return &response, nil
}

func (c *Client) RunSeriesNormalizeBatch(ctx context.Context, displayName string, requests []*llm.SeriesNormalizeRequest) (string, error) {
	systemPrompt, err := ParseSeriesNormalizeSystemPrompt()
	if err != nil {
		return "", err
	}

	generationConfig := &genai.GenerationConfig{
		ResponseMIMEType: "application/json",
	}

	contentBuilder := strings.Builder{}
	for i, request := range requests {
		promptText, err := ConvertSeriesNormalizePrompt(request)
		if err != nil {
			return "", err
		}

		batchRequest := &BatchRequest{
			Key: strconv.Itoa(i),
			Request: &GenerateContentRequest{
				Contents:          genai.Text(promptText),
				SystemInstruction: genai.NewContentFromText(systemPrompt, genai.RoleUser),
				GenerationConfig:  generationConfig,
			},
		}

		batchRequestStr, _ := json.Marshal(batchRequest)
		contentBuilder.Write(batchRequestStr)
		contentBuilder.WriteString("\n")
	}

	contentText := contentBuilder.String()
	filePath, err := createRequestJsonl(c.genFileName(), contentText)
	if err != nil {
		return "", err
	}
	defer func() {
		_ = os.Remove(filePath)
	}()

	cloudJsonl, err := c.client.Files.UploadFromPath(ctx, filePath, &genai.UploadFileConfig{
		MIMEType: "jsonl",
		Name:     c.genFileName(),
	})
	if err != nil {
		return "", err
	}

	jobResource := genai.BatchJobSource{
		FileName: cloudJsonl.Name,
	}
	jobConfig := genai.CreateBatchJobConfig{
		DisplayName: displayName,
	}

	if job, err := c.client.Batches.Create(ctx, c.Model, &jobResource, &jobConfig); err == nil {
		return job.Name, nil
	} else {
		return "", err
	}
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

func (c *Client) GetSeriesNormalizeBatch(ctx context.Context, jobName string) (llm.JobStatus, []*llm.SeriesNormalizeBatchResult, error) {
	batch, err := c.client.Batches.Get(ctx, jobName, nil)
	if err != nil {
		return llm.JobStatusFailed, nil, err
	}

	var responses []*llm.SeriesNormalizeBatchResult
	if batch.State == genai.JobStateSucceeded {
		outputFile, err := c.client.Files.Get(ctx, batch.Dest.FileName, nil)
		if err != nil {
			return llm.JobStatusFailed, nil, err
		}

		var batchResps []*BatchResponse
		consumer := func(line []byte) error {
			var resp BatchResponse
			if err = json.Unmarshal(line, &resp); err != nil {
				return err
			}
			batchResps = append(batchResps, &resp)
			return nil
		}

		if err = downloadOutputFile(outputFile.DownloadURI, c.key, consumer); err != nil {
			return llm.JobStatusFailed, nil, err
		}

		for _, batchResp := range batchResps {
			var resp llm.SeriesNormalizeResponse
			if batchResp.Response == nil {
				return llm.JobStatusFailed, nil, fmt.Errorf("batch name \"%s\", key \"%s\" response is nil", batch.Name, batchResp.Key)
			}
			if err = json.Unmarshal([]byte(batchResp.Response.Text()), &resp); err != nil {
				return llm.JobStatusFailed, nil, err
			}

			responses = append(responses, &llm.SeriesNormalizeBatchResult{
				Key:      batchResp.Key,
				Response: resp,
			})
		}
	}
	return ConvertJobStatus(batch.State), responses, nil
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
