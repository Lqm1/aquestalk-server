package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/Lqm1/aquestalk-server/pkg/aquestalk"
	"github.com/gin-gonic/gin"
)

// 許可されるvoiceのリスト
var allowedVoices = map[string]bool{
	"dvd":  true,
	"f1":   true,
	"f2":   true,
	"imd1": true,
	"jgr":  true,
	"m1":   true,
	"m2":   true,
	"r1":   true,
}

type SpeechRequest struct {
	Model          string  `json:"model" binding:"required"`
	Input          string  `json:"input" binding:"required"`
	Voice          string  `json:"voice" binding:"required"`
	ResponseFormat string  `json:"response_format,omitempty"`
	Speed          float64 `json:"speed,omitempty"`
}

func main() {
	r := gin.Default()

	r.POST("/v1/audio/speech", func(c *gin.Context) {
		var req SpeechRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// モデルのチェック
		if req.Model != "tts-1" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "only 'tts-1' model is supported",
			})
			return
		}

		// レスポンスフォーマットのチェック
		if req.ResponseFormat != "" && req.ResponseFormat != "wav" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "'response_format' must be 'wav'",
			})
			return
		}

		// Voiceのチェック
		if !allowedVoices[req.Voice] {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid voice specified",
			})
			return
		}

		// 入力テキストの長さチェック
		if len(req.Input) == 0 || len(req.Input) > 4096 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "input must be between 1 and 4096 characters",
			})
			return
		}

		// Speedのチェック
		if req.Speed != 0 && (req.Speed < 0.5 || req.Speed > 3.0) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "speed must be between 0.5 and 3.0",
			})
			return
		}

		// デフォルト速度設定
		speed := 1.0
		if req.Speed != 0 {
			speed = req.Speed
		}

		// AquesTalkの初期化
		aq, err := aquestalk.New(req.Voice)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": fmt.Sprintf("aquestalk init failed: %v", err),
			})
			return
		}
		defer aq.Close()

		// 速度を100倍して整数に変換（1.0 → 100, 2.0 → 200）
		speedParam := int(speed * 100)

		// 音声合成
		wav, err := aq.Synthe(req.Input, speedParam)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": fmt.Sprintf("synthesis failed: %v", err),
			})
			return
		}

		// 音声データを返却
		c.Data(http.StatusOK, "audio/wav", wav)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	r.Run(":" + port)
}
