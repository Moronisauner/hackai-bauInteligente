// Package assistant implementa o assistente conversacional de planejamento:
// conduz uma conversa com o cliente para coletar objetivo (meta, prazo) e, ao
// final, propõe um conjunto de alocações por conta-fonte (PRD RF-03/RF-04).
//
// O provedor de LLM é encapsulado aqui — nada da API pública (handlers, DTOs)
// depende de um provedor específico. A implementação fala o protocolo de chat
// compatível com OpenAI (POST {base}/chat/completions, Bearer opcional), que
// atende tanto provedores em nuvem quanto servidores locais com endpoint /v1.
package assistant

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Role identifica o autor de uma mensagem do chat.
type Role string

const (
	RoleSystem    Role = "system"
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
)

// Message é uma mensagem trocada na conversa.
type Message struct {
	Role    Role   `json:"role"`
	Content string `json:"content"`
}

// Client conversa com o provedor de LLM. Stateless: cada chamada carrega o
// histórico completo.
type Client struct {
	BaseURL string // ex: https://api.groq.com/openai/v1
	Model   string
	APIKey  string // opcional; quando vazio, não envia Authorization
	HTTP    *http.Client
}

// NewClient cria um Client. O timeout cobre confortavelmente provedores em
// nuvem; servidores locais lentos podem exigir mais.
func NewClient(baseURL, model, apiKey string) *Client {
	return &Client{
		BaseURL: baseURL,
		Model:   model,
		APIKey:  apiKey,
		HTTP:    &http.Client{Timeout: 60 * time.Second},
	}
}

// chatRequest/chatResponse espelham o protocolo de chat compatível com OpenAI.
type chatRequest struct {
	Model          string         `json:"model"`
	Messages       []Message      `json:"messages"`
	Temperature    float64        `json:"temperature"`
	MaxTokens      int            `json:"max_tokens,omitempty"`
	Stream         bool           `json:"stream"`
	ResponseFormat *responseFormat `json:"response_format,omitempty"`
}

type responseFormat struct {
	Type string `json:"type"` // "json_object" para modo JSON
}

type chatResponse struct {
	Choices []struct {
		Message Message `json:"message"`
	} `json:"choices"`
}

// complete envia o histórico e devolve o conteúdo bruto da resposta do modelo.
// jsonMode liga a decodificação estruturada (saída garantidamente em JSON).
// maxTokens limita a geração (0 = sem limite explícito).
func (c *Client) complete(ctx context.Context, messages []Message, jsonMode bool, maxTokens int) (string, error) {
	reqBody := chatRequest{
		Model:       c.Model,
		Messages:    messages,
		Temperature: 0, // determinístico
		MaxTokens:   maxTokens,
		Stream:      false,
	}
	if jsonMode {
		reqBody.ResponseFormat = &responseFormat{Type: "json_object"}
	}

	buf, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("assistant: marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+"/chat/completions", bytes.NewReader(buf))
	if err != nil {
		return "", fmt.Errorf("assistant: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if c.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.APIKey)
	}

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return "", fmt.Errorf("assistant: call llm: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("assistant: llm returned status %d", resp.StatusCode)
	}

	var out chatResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", fmt.Errorf("assistant: decode response: %w", err)
	}
	if len(out.Choices) == 0 {
		return "", fmt.Errorf("assistant: llm returned no choices")
	}
	return out.Choices[0].Message.Content, nil
}
