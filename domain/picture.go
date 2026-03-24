package domain

import "context"

type Picture struct {
	ID       int64  `gorm:"primaryKey;autoIncrement"`
	Prompt   string `gorm:"type:text"`
	ImageURL string `gorm:"type:text"`
	UserID   int64  `gorm:"type:bigint unsigned;not null"`
}

type PictureRepository interface {
	Create(ctx context.Context, p *Picture) error
	GetByID(ctx context.Context, id int64) (*Picture, error)
}

type ImageGeneratorClient interface {
	Generate(ctx context.Context, req GenerateImageRequest) (*GenerateImageResult, error)
}

type GenerateImageRequest struct {
	ImageBase64    string
	Model          string
	Prompt         string
	NegativePrompt string
	Size           string
	PromptExtend   bool
	Watermark      bool
	Seed           *int
}

type FivePack struct {
	WhiteBG     string   `json:"white_bg"`
	Transparent string   `json:"transparent"`
	SceneImages []string `json:"scene_images"`
	EffectImage string   `json:"effect_image"`
}

type GenerateImageResult struct {
	// 为兼容原落库逻辑：默认放白底图
	ImageURL  string
	ImageURLs []string
	FivePack  *FivePack

	RequestID  string
	Width      int
	Height     int
	ImageCount int
}
