package controller

import (
	"Hai-Service/config"
	"Hai-Service/core/libx"
	"Hai-Service/core/store/mysql"
	"Hai-Service/repository"
	"Hai-Service/usecase"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

type PictureController struct {
	uc *usecase.PictureUsecase
}

func NewPictureController() *PictureController {
	db, _ := mysql.InitMySQL()

	repo := repository.NewPictureRepo(db)
	generator := usecase.NewDashScopeImageClient(
		config.GetConfig().Picture.Endpoint,
		config.GetConfig().Picture.APIKey,
	)
	uc := usecase.NewPictureUsecase(repo, generator)
	return &PictureController{uc: uc}
}

type generatePictureReq struct {
	Prompt         string `json:"prompt" binding:"required"`
	NegativePrompt string `json:"negative_prompt"`
	Size           string `json:"size"`
	PromptExtend   *bool  `json:"prompt_extend"`
	Watermark      *bool  `json:"watermark"`
	Model          string `json:"model"`
	Seed           *int   `json:"seed"`
}

func (p *PictureController) Register(r *gin.RouterGroup) {
	r.POST("/pictures:generate", p.Generate)
	r.GET("/pictures/:id", p.GetByID)
}

func (p *PictureController) Generate(c *gin.Context) {
	var req generatePictureReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	promptExtend := true
	if req.PromptExtend != nil {
		promptExtend = *req.PromptExtend
	}
	watermark := false
	if req.Watermark != nil {
		watermark = *req.Watermark
	}

	pic, genRes, err := p.uc.GenerateAndSave(c.Request.Context(), usecase.GeneratePictureInput{
		Prompt:         req.Prompt,
		NegativePrompt: req.NegativePrompt,
		Size:           req.Size,
		PromptExtend:   promptExtend,
		Watermark:      watermark,
		Model:          req.Model,
		Seed:           req.Seed,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	libx.Ok(c, "创建成功", gin.H{
		"id":         pic.ID,
		"prompt":     pic.Prompt,
		"image_url":  pic.ImageURL,
		"request_id": genRes.RequestID,
		"usage": gin.H{
			"width":       genRes.Width,
			"height":      genRes.Height,
			"image_count": genRes.ImageCount,
		},
	})
}

func (p *PictureController) GetByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		libx.Err(c, http.StatusBadRequest, "invalid id", err)
		return
	}

	pic, err := p.uc.GetByID(c.Request.Context(), id)
	if err != nil {
		libx.Err(c, http.StatusInternalServerError, "failed to get picture", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":        pic.ID,
		"prompt":    pic.Prompt,
		"image_url": pic.ImageURL,
	})
}
