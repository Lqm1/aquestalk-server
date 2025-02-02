package aquestalk

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"unsafe"

	"golang.org/x/text/encoding/japanese"
)

// DLLを埋め込むための設定（プロジェクトのディレクトリ構造に合わせて調整）
//
//go:embed bin/*/AquesTalk.dll
var dllFS embed.FS

type AquesTalk struct {
	dll          *syscall.DLL
	syntheProc   *syscall.Proc
	freeWaveProc *syscall.Proc
	tempDir      string // 一時ディレクトリの保持用
}

func New(voice string) (*AquesTalk, error) {
	// 埋め込みDLLのパスを構築
	dllPathInEmbed := fmt.Sprintf("bin/%s/AquesTalk.dll", voice)
	dllData, err := dllFS.ReadFile(dllPathInEmbed)
	if err != nil {
		return nil, fmt.Errorf("DLL not found for voice %s: %w", voice, err)
	}

	// 一時ディレクトリの作成
	tempDir, err := os.MkdirTemp("", "aquestalk-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp dir: %w", err)
	}

	// 一時ファイルにDLLを書き出し
	tempDLLPath := filepath.Join(tempDir, "AquesTalk.dll")
	if err := os.WriteFile(tempDLLPath, dllData, 0644); err != nil {
		os.RemoveAll(tempDir)
		return nil, fmt.Errorf("failed to write DLL: %w", err)
	}

	// DLLの読み込み
	dll, err := syscall.LoadDLL(tempDLLPath)
	if err != nil {
		os.RemoveAll(tempDir)
		return nil, fmt.Errorf("DLL load error: %w", err)
	}

	// プロシージャの取得
	syntheProc, err := dll.FindProc("AquesTalk_Synthe")
	if err != nil {
		dll.Release()
		os.RemoveAll(tempDir)
		return nil, fmt.Errorf("AquesTalk_Synthe not found: %w", err)
	}

	freeWaveProc, err := dll.FindProc("AquesTalk_FreeWave")
	if err != nil {
		dll.Release()
		os.RemoveAll(tempDir)
		return nil, fmt.Errorf("AquesTalk_FreeWave not found: %w", err)
	}

	return &AquesTalk{
		dll:          dll,
		syntheProc:   syntheProc,
		freeWaveProc: freeWaveProc,
		tempDir:      tempDir,
	}, nil
}

func (a *AquesTalk) Close() error {
	if a.dll != nil {
		a.dll.Release()
		a.dll = nil
	}
	if a.tempDir != "" {
		os.RemoveAll(a.tempDir) // 一時ディレクトリを削除
		a.tempDir = ""
	}
	return nil
}

// 音声合成を実行
func (a *AquesTalk) Synthe(koe string, speed int) ([]byte, error) {
	enc := japanese.ShiftJIS.NewEncoder()
	koe, err := enc.String(koe)
	if err != nil {
		return nil, fmt.Errorf("failed to convert koe to sjis: %w", err)
	}

	ckoe, err := syscall.BytePtrFromString(koe)
	if err != nil {
		return nil, fmt.Errorf("invalid parameter: %w", err)
	}

	var size int
	ret, _, _ := a.syntheProc.Call(
		uintptr(unsafe.Pointer(ckoe)),
		uintptr(speed),
		uintptr(unsafe.Pointer(&size)),
	)

	if ret == 0 {
		return nil, fmt.Errorf("synthesis failed (code: %d)", size)
	}

	// WAVデータのコピーと解放
	wavData := unsafe.Slice((*byte)(unsafe.Pointer(ret)), size)
	data := make([]byte, len(wavData))
	copy(data, wavData)
	a.freeWaveProc.Call(ret)

	return data, nil
}
