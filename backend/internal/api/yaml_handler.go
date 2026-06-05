package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v3"

	"github.com/2754302155/novel-to-screenplay-ai-toolkit/backend/internal/domain"
	"github.com/2754302155/novel-to-screenplay-ai-toolkit/backend/internal/schema"
)

type ValidateYAMLRequest struct {
	YAML string `json:"yaml"`
}

func registerYAMLRoutes(router *gin.RouterGroup) {
	router.POST("/yaml/validate", func(ctx *gin.Context) {
		var request ValidateYAMLRequest
		if err := ctx.ShouldBindJSON(&request); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"code":    "INVALID_REQUEST",
				"message": "请求格式不正确，请提交 YAML 文本。",
			})
			return
		}

		var draft domain.ScreenplayDraft
		if err := yaml.Unmarshal([]byte(request.YAML), &draft); err != nil {
			ctx.JSON(http.StatusOK, gin.H{
				"valid": false,
				"issues": []schema.ValidationIssue{
					{
						Path:    "yaml",
						Message: "YAML 解析失败：" + err.Error(),
					},
				},
				"quality_report": domain.QualityReport{
					Warnings:            []string{"YAML 无法解析。"},
					HumanReviewRequired: []string{"修复 YAML 语法后重新校验。"},
				},
			})
			return
		}

		result := schema.ValidateDraft(draft)
		ctx.JSON(http.StatusOK, gin.H{
			"valid":          result.Valid,
			"issues":         result.Issues,
			"quality_report": schema.GenerateQualityReport(draft, draft.Source.Chapters, result),
		})
	})
}
