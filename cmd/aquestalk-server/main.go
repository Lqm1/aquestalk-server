package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/c7e715d1b04b17683718fb1e8944cc28/aquestalk-server/internal/aquestalk"
)

func main() {
	// 話者とベースディレクトリの指定
	voice := "f1"                   // 切り替え可能: m1, f2, etc.
	baseDir, _ := filepath.Abs(".") // プロジェクトルートを想定

	aq, err := aquestalk.New(voice, baseDir)
	if err != nil {
		fmt.Printf("Initialization failed: %v\n", err)
		os.Exit(1)
	}
	defer aq.Close()

	// 音声合成の実行
	wav, err := aq.Synthe("こんにちわ。", 100)
	if err != nil {
		fmt.Printf("Synthesis error: %v\n", err)
		os.Exit(1)
	}

	// ファイル保存
	if err := os.WriteFile("output.wav", wav, 0644); err != nil {
		fmt.Printf("File write error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Successfully generated output.wav")
}
