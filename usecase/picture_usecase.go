package usecase

import (
	"Hai-Service/domain"
	"context"
	"errors"
)

type PictureUsecase struct {
	repo      domain.PictureRepository
	generator domain.ImageGeneratorClient
}

func NewPictureUsecase(repo domain.PictureRepository, generator domain.ImageGeneratorClient) *PictureUsecase {
	return &PictureUsecase{repo: repo, generator: generator}
}

type GeneratePictureInput struct {
	ImageBase64    string
	Prompt         string
	NegativePrompt string
	Size           string
	PromptExtend   bool
	Watermark      bool
	Model          string
	Seed           *int
}

func (u *PictureUsecase) GenerateAndSave(ctx context.Context, in GeneratePictureInput) (*domain.Picture, *domain.GenerateImageResult, error) {
	if in.ImageBase64 == "" {
		return nil, nil, errors.New("image base64 empty")
	}

	model := in.Model
	if model == "" {
		model = "wan2.5-i2i-preview"
	}
	size := in.Size
	if size == "" {
		size = "1280*1280"
	}

	res, err := u.generator.Generate(ctx, domain.GenerateImageRequest{
		ImageBase64:    in.ImageBase64,
		Model:          model,
		Prompt:         in.Prompt,
		NegativePrompt: in.NegativePrompt,
		Size:           size,
		PromptExtend:   in.PromptExtend,
		Watermark:      in.Watermark,
		Seed:           in.Seed,
	})
	if err != nil {
		return nil, nil, err
	}

	p := &domain.Picture{
		Prompt:   in.Prompt,
		ImageURL: res.ImageURL, // 固定保存白底图
	}
	if err := u.repo.Create(ctx, p); err != nil {
		return nil, nil, err
	}
	return p, res, nil
}

func (u *PictureUsecase) GetByID(ctx context.Context, id int64) (*domain.Picture, error) {
	return u.repo.GetByID(ctx, id)
}
