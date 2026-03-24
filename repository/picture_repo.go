package repository

import (
	"Hai-Service/domain"
	"context"
	"gorm.io/gorm"
)

type PictureRepo struct {
	db *gorm.DB
}

func NewPictureRepo(db *gorm.DB) domain.PictureRepository {
	return &PictureRepo{db: db}
}

func (r *PictureRepo) Create(ctx context.Context, p *domain.Picture) error {
	// 使用 WithContext 保持 context 传播
	return r.db.WithContext(ctx).Create(p).Error
}

func (r *PictureRepo) GetByID(ctx context.Context, id int64) (*domain.Picture, error) {
	var p domain.Picture
	if err := r.db.WithContext(ctx).First(&p, id).Error; err != nil {
		return nil, err
	}
	return &p, nil
}
