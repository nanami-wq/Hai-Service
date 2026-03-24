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
	"strings"
	"time"
)

type DashScopeI2IClient struct {
	endpoint string
	apiKey   string
	httpCli  *http.Client
}

func NewDashScopeI2IClient(endpoint, apiKey string) *DashScopeI2IClient {
	return &DashScopeI2IClient{
		endpoint: endpoint,
		apiKey:   apiKey,
		httpCli:  &http.Client{Timeout: 120 * time.Second},
	}
}

func (c *DashScopeI2IClient) Generate(ctx context.Context, req domain.GenerateImageRequest) (*domain.GenerateImageResult, error) {
	if c == nil {
		return nil, errors.New("nil dashscope client")
	}
	if req.ImageBase64 == "" {
		return nil, errors.New("image base64 empty")
	}
	if !strings.HasPrefix(req.ImageBase64, "data:") || !strings.Contains(req.ImageBase64, ";base64,") {
		return nil, errors.New("invalid image_base64 format, expected data:{mime};base64,{data}")
	}

	fp, err := GenerateFivePack(ctx, c, FivePackInput{
		ImageBase64:    req.ImageBase64,
		Model:          req.Model,
		Prompt:         req.Prompt,
		Size:           req.Size,
		Seed:           req.Seed,
		NegativePrompt: req.NegativePrompt,
		PromptExtend:   req.PromptExtend,
		Watermark:      req.Watermark,
	})
	if err != nil {
		return nil, err
	}

	return &domain.GenerateImageResult{
		ImageURL:   fp.WhiteBG,
		ImageURLs:  fp.AllImageURLs,
		FivePack:   &domain.FivePack{WhiteBG: fp.WhiteBG, Transparent: fp.Transparent, SceneImages: fp.SceneImages, EffectImage: fp.EffectImage},
		RequestID:  fp.RequestID,
		ImageCount: 5,
	}, nil
}

type I2ICreateTaskRequest struct {
	Model  string `json:"model"`
	Input  I2IIn  `json:"input"`
	Params I2IPar `json:"parameters,omitempty"`
}

type I2IIn struct {
	Prompt string   `json:"prompt,omitempty"`
	Images []string `json:"images"`
}

type I2IPar struct {
	NegativePrompt string `json:"negative_prompt,omitempty"`
	Size           string `json:"size,omitempty"`
	N              int    `json:"n,omitempty"`
	PromptExtend   bool   `json:"prompt_extend"`
	Watermark      bool   `json:"watermark"`
	Seed           *int   `json:"seed,omitempty"`
}

type i2iCreateResp struct {
	Output struct {
		TaskStatus string `json:"task_status"`
		TaskID     string `json:"task_id"`
	} `json:"output"`
	RequestID string `json:"request_id"`
	Code      string `json:"code"`
	Message   string `json:"message"`
}

type i2iGetResp struct {
	Output struct {
		TaskStatus string `json:"task_status"`
		TaskID     string `json:"task_id"`
		Results    []struct {
			URL string `json:"url"`
		} `json:"results"`
	} `json:"output"`
	RequestID string `json:"request_id"`
	Code      string `json:"code"`
	Message   string `json:"message"`
}

func (c *DashScopeI2IClient) CreateTask(ctx context.Context, req I2ICreateTaskRequest) (string, error) {
	if c.endpoint == "" {
		return "", errors.New("dashscope endpoint empty")
	}
	if c.apiKey == "" {
		return "", errors.New("dashscope api key empty")
	}
	if req.Model == "" {
		return "", errors.New("model empty")
	}
	if len(req.Input.Images) == 0 || req.Input.Images[0] == "" {
		return "", errors.New("input image empty")
	}
	if req.Params.N == 0 {
		req.Params.N = 1
	}

	b, err := json.Marshal(&req)
	if err != nil {
		return "", err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint, bytes.NewReader(b))
	if err != nil {
		return "", err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
	httpReq.Header.Set("X-DashScope-Async", "enable")

	resp, err := c.httpCli.Do(httpReq)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var out i2iCreateResp
	if err := json.Unmarshal(body, &out); err != nil {
		return "", err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 || out.Code != "" {
		if out.Code != "" || out.Message != "" {
			return "", fmt.Errorf("dashscope create http %d: %s %s", resp.StatusCode, out.Code, out.Message)
		}
		return "", fmt.Errorf("dashscope create http %d: %s", resp.StatusCode, string(body))
	}
	if out.Output.TaskID == "" {
		return "", errors.New("dashscope create: empty task_id")
	}
	return out.Output.TaskID, nil
}

func (c *DashScopeI2IClient) GetTask(ctx context.Context, taskID string) (string, []string, string, error) {
	if taskID == "" {
		return "", nil, "", errors.New("task id empty")
	}

	taskURL := "https://dashscope.aliyuncs.com/api/v1/tasks/" + taskID

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, taskURL, nil)
	if err != nil {
		return "", nil, "", err
	}
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpCli.Do(httpReq)
	if err != nil {
		return "", nil, "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var out i2iGetResp
	if err := json.Unmarshal(body, &out); err != nil {
		return "", nil, "", err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 || out.Code != "" {
		if out.Code != "" || out.Message != "" {
			return "", nil, out.RequestID, fmt.Errorf("dashscope get http %d: %s %s", resp.StatusCode, out.Code, out.Message)
		}
		return "", nil, out.RequestID, fmt.Errorf("dashscope get http %d: %s", resp.StatusCode, string(body))
	}

	var urls []string
	for _, r := range out.Output.Results {
		if r.URL != "" {
			urls = append(urls, r.URL)
		}
	}
	return out.Output.TaskStatus, urls, out.RequestID, nil
}

func (c *DashScopeI2IClient) PollUntilDone(ctx context.Context, taskID string, interval, timeout time.Duration) ([]string, string, error) {
	if interval <= 0 {
		interval = 2 * time.Second
	}
	if timeout <= 0 {
		timeout = 2 * time.Minute
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	t := time.NewTicker(interval)
	defer t.Stop()

	for {
		status, urls, requestID, err := c.GetTask(ctx, taskID)
		if err != nil {
			return nil, requestID, err
		}

		switch status {
		case "SUCCEEDED":
			if len(urls) == 0 {
				return nil, requestID, errors.New("task succeeded but empty results")
			}
			return urls, requestID, nil
		case "FAILED", "CANCELED", "UNKNOWN":
			return nil, requestID, fmt.Errorf("task %s status %s", taskID, status)
		}

		select {
		case <-ctx.Done():
			return nil, requestID, fmt.Errorf("poll timeout: %w", ctx.Err())
		case <-t.C:
		}
	}
}

type FivePackResult struct {
	WhiteBG      string   `json:"white_bg"`
	Transparent  string   `json:"transparent"`
	SceneImages  []string `json:"scene_images"`
	EffectImage  string   `json:"effect_image"`
	AllTaskIDs   []string `json:"task_ids"`
	AllImageURLs []string `json:"all_image_urls"`
	RequestID    string   `json:"request_id"`
}

type FivePackInput struct {
	ImageBase64    string
	Model          string
	Prompt         string
	Size           string
	Seed           *int
	NegativePrompt string
	PromptExtend   bool
	Watermark      bool
}

func BuildFivePrompts(userPrompt string) (white, transparent, scene1, scene2, effect string) {
	base := userPrompt
	if base == "" {
		base = "保持主体清晰，细节真实，风格一致"
	}
	white = base + "，商品白底图，纯白背景，正面居中，高质感棚拍光，无多余道具，无文字无水印"
	transparent = base + "，主体抠图效果，透明背景，边缘干净，无阴影底板，无背景元素"
	scene1 = base + "，真实生活场景展示，光线自然，背景与主体协调，不遮挡主体，无文字无水印"
	scene2 = base + "，电商风场景展示，氛围感布光，突出主体卖点，背景简洁不抢主体，无文字无水印"
	effect = base + "，功效展示图，突出使用效果前后对比或功效表达，画面简洁，信息图风格但不要文字，无水印"
	return
}

// GenerateFivePack 固定生成 5 张：白底、透明、场景1、场景2、功效。
// 实现策略：先创建 5 个任务，再依次轮询取回结果，稳定且易控。
func GenerateFivePack(ctx context.Context, cli *DashScopeI2IClient, in FivePackInput) (*FivePackResult, error) {
	if cli == nil {
		return nil, errors.New("nil client")
	}
	if in.ImageBase64 == "" {
		return nil, errors.New("image base64 empty")
	}
	if in.Size == "" {
		in.Size = "1280*1280"
	}

	model := in.Model
	if model == "" {
		model = "wan2.5-i2i-preview"
	}

	whiteP, transP, scene1P, scene2P, effP := BuildFivePrompts(in.Prompt)

	type job struct {
		name   string
		prompt string
	}
	jobs := []job{
		{name: "white", prompt: whiteP},
		{name: "transparent", prompt: transP},
		{name: "scene1", prompt: scene1P},
		{name: "scene2", prompt: scene2P},
		{name: "effect", prompt: effP},
	}

	taskIDByName := make(map[string]string, len(jobs))
	allTaskIDs := make([]string, 0, len(jobs))

	for _, j := range jobs {
		taskID, err := cli.CreateTask(ctx, I2ICreateTaskRequest{
			Model: model,
			Input: I2IIn{
				Prompt: j.prompt,
				Images: []string{in.ImageBase64},
			},
			Params: I2IPar{
				NegativePrompt: in.NegativePrompt,
				Size:           in.Size,
				N:              1,
				PromptExtend:   in.PromptExtend,
				Watermark:      in.Watermark,
				Seed:           in.Seed,
			},
		})
		if err != nil {
			return nil, fmt.Errorf("create task %s: %w", j.name, err)
		}
		taskIDByName[j.name] = taskID
		allTaskIDs = append(allTaskIDs, taskID)
	}

	var requestID string
	getOne := func(name string) (string, error) {
		tid := taskIDByName[name]
		urls, rid, err := cli.PollUntilDone(ctx, tid, 2*time.Second, 3*time.Minute)
		if rid != "" && requestID == "" {
			requestID = rid
		}
		if err != nil {
			return "", fmt.Errorf("poll %s: %w", name, err)
		}
		if len(urls) == 0 || urls[0] == "" {
			return "", fmt.Errorf("poll %s: empty result url", name)
		}
		return urls[0], nil
	}

	whiteURL, err := getOne("white")
	if err != nil {
		return nil, err
	}
	transURL, err := getOne("transparent")
	if err != nil {
		return nil, err
	}
	scene1URL, err := getOne("scene1")
	if err != nil {
		return nil, err
	}
	scene2URL, err := getOne("scene2")
	if err != nil {
		return nil, err
	}
	effectURL, err := getOne("effect")
	if err != nil {
		return nil, err
	}

	allURLs := []string{whiteURL, transURL, scene1URL, scene2URL, effectURL}

	return &FivePackResult{
		WhiteBG:      whiteURL,
		Transparent:  transURL,
		SceneImages:  []string{scene1URL, scene2URL},
		EffectImage:  effectURL,
		AllTaskIDs:   allTaskIDs,
		AllImageURLs: allURLs,
		RequestID:    requestID,
	}, nil
}
