package domain

import "context"

type Picture struct {
	ID       int64  `gorm:"primaryKey;autoIncrement"`
	Prompt   string `gorm:"type:text"`
	ImageURL string `gorm:"type:text"`
}

// PictureRepository 仓库接口由 domain 层定义，便于向上层注入与替换实现。
type PictureRepository interface {
	Create(ctx context.Context, p *Picture) error
	GetByID(ctx context.Context, id int64) (*Picture, error)
}

// ImageGeneratorClient 由 domain 层定义，usecase 通过接口调用外部生成服务。
type ImageGeneratorClient interface {
	Generate(ctx context.Context, req GenerateImageRequest) (*GenerateImageResult, error)
}

// GenerateImageRequest 对齐 DashScope 入参（简化为常用字段）。
type GenerateImageRequest struct {
	Model          string
	Prompt         string
	NegativePrompt string
	Size           string
	PromptExtend   bool
	Watermark      bool
	Seed           *int
}

// GenerateImageResult 对齐 DashScope 出参（取 image url + request id）。
type GenerateImageResult struct {
	ImageURL   string
	RequestID  string
	Width      int
	Height     int
	ImageCount int
}
