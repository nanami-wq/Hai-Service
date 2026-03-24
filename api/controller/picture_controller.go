package controller

import (
	"Hai-Service/config"
	"Hai-Service/core/libx"
	"Hai-Service/core/store/mysql"
	"Hai-Service/repository"
	"Hai-Service/usecase"
	"encoding/base64"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"mime/multipart"
	"net/http"
	"strconv"
)

type PictureController struct {
	uc *usecase.PictureUsecase
}

func NewPictureController() *PictureController {
	db, _ := mysql.InitMySQL()

	repo := repository.NewPictureRepo(db)
	generator := usecase.NewDashScopeI2IClient(
		config.GetConfig().Picture.Endpoint,
		config.GetConfig().Picture.APIKey,
	)
	uc := usecase.NewPictureUsecase(repo, generator)
	return &PictureController{uc: uc}
}

func (p *PictureController) Register(r *gin.RouterGroup) {
	r.POST("/pictures/generate", p.Generate)
	r.GET("/pictures/:id", p.GetByID)
}

type generatePictureForm struct {
	Prompt         string `form:"prompt"`
	NegativePrompt string `form:"negative_prompt"`
	Size           string `form:"size"`
	PromptExtend   *bool  `form:"prompt_extend"`
	Model          string `form:"model"`
	Seed           *int   `form:"seed"`
}

func fileToDataBase64(fh *multipart.FileHeader) (string, error) {
	f, err := fh.Open()
	if err != nil {
		return "", err
	}
	defer f.Close()

	b, err := io.ReadAll(f)
	if err != nil {
		return "", err
	}

	ct := fh.Header.Get("Content-Type")
	if ct == "" {
		ct = "application/octet-stream"
	}
	enc := base64.StdEncoding.EncodeToString(b)
	return fmt.Sprintf("data:%s;base64,%s", ct, enc), nil
}

func (p *PictureController) Generate(c *gin.Context) {
	var form generatePictureForm
	if err := c.ShouldBind(&form); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	fh, err := c.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "image required"})
		return
	}

	imgBase64, err := fileToDataBase64(fh)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	promptExtend := true
	if form.PromptExtend != nil {
		promptExtend = *form.PromptExtend
	}

	_, genRes, err := p.uc.GenerateAndSave(c.Request.Context(), usecase.GeneratePictureInput{
		ImageBase64:    imgBase64,
		Prompt:         form.Prompt,
		NegativePrompt: form.NegativePrompt,
		Size:           form.Size,
		PromptExtend:   promptExtend,
		Watermark:      false,
		Model:          form.Model,
		Seed:           form.Seed,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	libx.Ok(c, "创建成功", gin.H{
		"request_id": genRes.RequestID,
		"images":     genRes.ImageURLs,
		"five_pack":  genRes.FivePack,
		"usage": gin.H{
			"width":       genRes.Width,
			"height":      genRes.Height,
			"image_count": genRes.ImageCount, // 固定为 5
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

	libx.Ok(c, "查询成功", gin.H{
		"id":        pic.ID,
		"prompt":    pic.Prompt,
		"image_url": pic.ImageURL,
	})
}
