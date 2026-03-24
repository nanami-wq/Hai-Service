package usecase

import (
	"Hai-Service/domain"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

type DashScopeImageClient struct {
	endpoint string
	apiKey   string
	httpCli  *http.Client
}

func NewDashScopeImageClient(endpoint, apiKey string) *DashScopeImageClient {
	cli := &http.Client{Timeout: 60 * time.Second}
	return &DashScopeImageClient{
		endpoint: endpoint,
		apiKey:   apiKey,
		httpCli:  cli,
	}
}

type dashReq struct {
	Model string `json:"model"`
	Input struct {
		Messages []struct {
			Role    string `json:"role"`
			Content []struct {
				Text string `json:"text"`
			} `json:"content"`
		} `json:"messages"`
	} `json:"input"`
	Parameters struct {
		NegativePrompt string `json:"negative_prompt,omitempty"`
		Size           string `json:"size,omitempty"`
		N              int    `json:"n,omitempty"`
		PromptExtend   bool   `json:"prompt_extend"`
		Watermark      bool   `json:"watermark"`
		Seed           *int   `json:"seed,omitempty"`
	} `json:"parameters,omitempty"`
}

type dashResp struct {
	Output struct {
		Choices []struct {
			FinishReason string `json:"finish_reason"`
			Message      struct {
				Role    string `json:"role"`
				Content []struct {
					Image string `json:"image"`
				} `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	} `json:"output"`
	Usage struct {
		Height     int `json:"height"`
		Width      int `json:"width"`
		ImageCount int `json:"image_count"`
	} `json:"usage"`
	RequestID string `json:"request_id"`

	Code    string `json:"code"`
	Message string `json:"message"`
}

func (c *DashScopeImageClient) Generate(ctx context.Context, req domain.GenerateImageRequest) (*domain.GenerateImageResult, error) {
	if c.endpoint == "" {
		return nil, errors.New("dashscope endpoint empty")
	}
	if c.apiKey == "" {
		return nil, errors.New("dashscope api key empty")
	}
	if req.Model == "" {
		return nil, errors.New("model empty")
	}
	if req.Prompt == "" {
		return nil, errors.New("prompt empty")
	}

	var dr dashReq
	dr.Model = req.Model
	dr.Input.Messages = make([]struct {
		Role    string `json:"role"`
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
	}, 1)
	dr.Input.Messages[0].Role = "user"
	dr.Input.Messages[0].Content = make([]struct {
		Text string `json:"text"`
	}, 1)
	dr.Input.Messages[0].Content[0].Text = req.Prompt

	dr.Parameters.NegativePrompt = req.NegativePrompt
	dr.Parameters.Size = req.Size
	dr.Parameters.N = 1
	dr.Parameters.PromptExtend = req.PromptExtend
	dr.Parameters.Watermark = req.Watermark
	dr.Parameters.Seed = req.Seed

	b, err := json.Marshal(&dr)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint, bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpCli.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var out dashResp
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		if out.Code != "" || out.Message != "" {
			return nil, fmt.Errorf("dashscope http %d: %s %s", resp.StatusCode, out.Code, out.Message)
		}
		return nil, fmt.Errorf("dashscope http %d: %s", resp.StatusCode, string(body))
	}
	if out.Code != "" {
		return nil, fmt.Errorf("dashscope error: %s %s", out.Code, out.Message)
	}
	if len(out.Output.Choices) == 0 || len(out.Output.Choices[0].Message.Content) == 0 || out.Output.Choices[0].Message.Content[0].Image == "" {
		return nil, errors.New("dashscope empty image url")
	}

	return &domain.GenerateImageResult{
		ImageURL:   out.Output.Choices[0].Message.Content[0].Image,
		RequestID:  out.RequestID,
		Width:      out.Usage.Width,
		Height:     out.Usage.Height,
		ImageCount: out.Usage.ImageCount,
	}, nil
}
