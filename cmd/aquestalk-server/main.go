package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/Lqm1/aquestalk-server/pkg/aqkanji2koe"
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
	Input          string  `json:"input" binding:"required"`
	Model          string  `json:"model" binding:"required"`
	Voice          string  `json:"voice" binding:"required"`
	Instructions   string  `json:"instructions,omitempty"`
	ResponseFormat string  `json:"response_format,omitempty"`
	Speed          float64 `json:"speed,omitempty"`
	StreamFormat   string  `json:"stream_format,omitempty"`
}

func main() {
	// AqKanji2Koeの初期化（起動時に1回だけ）
	ak, err := aqkanji2koe.New()
	if err != nil {
		log.Fatalf("aqkanji2koe init failed: %v", err)
	}
	defer ak.Close()
	var akMu sync.Mutex

	// AquesTalkの全voice初期化（起動時に1回だけ）
	aqMap := make(map[string]*aquestalk.AquesTalk)
	aqMu := make(map[string]*sync.Mutex)
	for voice := range allowedVoices {
		aq, err := aquestalk.New(voice)
		if err != nil {
			log.Fatalf("aquestalk init failed for voice %s: %v", voice, err)
		}
		defer aq.Close()
		aqMap[voice] = aq
		aqMu[voice] = &sync.Mutex{}
	}

	log.Println("All engines initialized")

	r := gin.Default()
	r.TrustedPlatform = gin.PlatformCloudflare

	r.POST("/v1/audio/speech", func(c *gin.Context) {
		var req SpeechRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Inputのチェック
		if len(req.Input) == 0 || len(req.Input) > 4096 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "input must be between 1 and 4096 characters",
			})
			return
		}

		// Modelのチェック
		// if req.Model != "aquestalk" {
		// 	c.JSON(http.StatusBadRequest, gin.H{
		// 		"error": "only 'aquestalk' model is supported",
		// 	})
		// 	return
		// }

		// Voiceのチェック
		if !allowedVoices[req.Voice] {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid voice specified",
			})
			return
		}

		// ResponseFormatのチェック
		if req.ResponseFormat != "" && req.ResponseFormat != "wav" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "'response_format' must be 'wav'",
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

		// StreamFormatのチェック
		if req.StreamFormat != "" && req.StreamFormat != "audio" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "'stream_format' must be 'audio'",
			})
			return
		}

		// 入力テキストをかな音声記号列に変換
		akMu.Lock()
		koe, err := ak.Convert(req.Input)
		akMu.Unlock()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": fmt.Sprintf("convert failed: %v", err),
			})
			return
		}

		// 速度を100倍して整数に変換（1.0 → 100, 2.0 → 200）
		speed := 100
		if req.Speed != 0 {
			speed = int(req.Speed * 100)
		}

		// 音声合成
		aq := aqMap[req.Voice]
		mu := aqMu[req.Voice]
		mu.Lock()
		wav, err := aq.Synthe(koe, speed)
		mu.Unlock()
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
