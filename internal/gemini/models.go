package gemini

import "google.golang.org/genai"

type BatchRequest struct {
	Key     string                  `json:"key"`
	Request *GenerateContentRequest `json:"request"`
}

type BatchResponse struct {
	Key      string                         `json:"key"`
	Response *genai.GenerateContentResponse `json:"response"`
}

type GenerateContentRequest struct {
	Contents          []*genai.Content `json:"contents"`
	SystemInstruction *genai.Content   `json:"systemInstruction"`

	GenerationConfig *genai.GenerationConfig `json:"generationConfig"`
}
