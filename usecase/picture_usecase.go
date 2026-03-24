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
	Prompt         string
	NegativePrompt string
	Size           string
	PromptExtend   bool
	Watermark      bool
	Model          string
	Seed           *int
}

func (u *PictureUsecase) GenerateAndSave(ctx context.Context, in GeneratePictureInput) (*domain.Picture, *domain.GenerateImageResult, error) {
	if in.Prompt == "" {
		return nil, nil, errors.New("prompt empty")
	}
	model := in.Model
	if model == "" {
		model = "qwen-image-2.0-pro"
	}
	size := in.Size
	if size == "" {
		size = "2048*2048"
	}

	res, err := u.generator.Generate(ctx, domain.GenerateImageRequest{
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
		ImageURL: res.ImageURL,
	}
	if err := u.repo.Create(ctx, p); err != nil {
		return nil, nil, err
	}
	return p, res, nil
}

func (u *PictureUsecase) GetByID(ctx context.Context, id int64) (*domain.Picture, error) {
	return u.repo.GetByID(ctx, id)
}
