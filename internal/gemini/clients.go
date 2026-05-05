package gemini

import (
	"context"
	"encoding/json"
	"series-batch-go/internal/pkg/llm"

	"google.golang.org/genai"
)

type Client struct {
	client *genai.Client

	Model string
}

func NewClient(client *genai.Client, model string) *Client {
	return &Client{client: client, Model: model}
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
	} else {
		return &response, nil
	}
}

func (c *Client) RunSeriesNormalizeBatch(ctx context.Context, displayName string, req []*llm.SeriesNormalizeRequest) (string, error) {
	systemPrompt, err := ParseSeriesNormalizeSystemPrompt()
	if err != nil {
		return "", err
	}

	generateConfig := genai.GenerateContentConfig{
		SystemInstruction: genai.NewContentFromText(systemPrompt, genai.RoleUser),
		ResponseMIMEType:  "application/json",
	}

	var inlinedRequests []*genai.InlinedRequest
	for _, r := range req {
		if promptText, err := ConvertSeriesNormalizePrompt(r); err == nil {
			inlinedRequests = append(inlinedRequests, &genai.InlinedRequest{
				Contents: genai.Text(promptText),
				Config:   &generateConfig,
			})
		} else {
			return "", err
		}
	}

	jobResource := genai.BatchJobSource{InlinedRequests: inlinedRequests}
	jobConfig := genai.CreateBatchJobConfig{DisplayName: displayName}
	if job, err := c.client.Batches.Create(ctx, c.Model, &jobResource, &jobConfig); err == nil {
		return job.Name, nil
	} else {
		return "", err
	}
}

func (c *Client) GetSeriesNormalizeBatch(ctx context.Context, jobName string) (llm.JobStatus, []*llm.SeriesNormalizeResponse, error) {
	batch, err := c.client.Batches.Get(ctx, jobName, nil)
	if err != nil {
		return llm.JobStatusFailed, nil, err
	}

	var res []*llm.SeriesNormalizeResponse
	if batch.State == genai.JobStateSucceeded {
		for _, in := range batch.Dest.InlinedResponses {
			text := in.Response.Text()
			var r llm.SeriesNormalizeResponse
			if err := json.Unmarshal([]byte(text), &r); err != nil {
				return llm.JobStatusFailed, nil, err
			}
			res = append(res, &r)
		}
	}
	return ConvertJobStatus(batch.State), res, nil
}
