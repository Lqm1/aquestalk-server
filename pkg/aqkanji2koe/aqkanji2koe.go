package aqkanji2koe

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"unsafe"
)

// DLLを埋め込むための設定
//
//go:embed bin/*
var dllFS embed.FS

// エラーコード
const (
	ErrNone           = 0 // 成功
	ErrArgument       = 1 // 引数エラー（NULLポインタ等）
	ErrNotInitialized = 2 // 未初期化（aqk2k_create を呼んでいない）
	ErrBufferTooSmall = 3 // バッファ不足
	ErrProcessing     = 4 // 処理エラー
)

// 最小バッファサイズ
const minBufSize = 256

type AqKanji2Koe struct {
	dll              *syscall.DLL
	createProc       *syscall.Proc
	releaseProc      *syscall.Proc
	convertProc      *syscall.Proc
	convertRomanProc *syscall.Proc
	tempDir          string
}

func New() (*AqKanji2Koe, error) {
	dllData, err := dllFS.ReadFile("bin/AqKanji2Koe.dll")
	if err != nil {
		return nil, fmt.Errorf("DLL not found: %w", err)
	}

	// 一時ディレクトリの作成
	tempDir, err := os.MkdirTemp("", "aqkanji2koe-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp dir: %w", err)
	}

	// 一時ファイルにDLLを書き出し
	tempDLLPath := filepath.Join(tempDir, "AqKanji2Koe.dll")
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
	createProc, err := dll.FindProc("aqk2k_create")
	if err != nil {
		dll.Release()
		os.RemoveAll(tempDir)
		return nil, fmt.Errorf("aqk2k_create not found: %w", err)
	}

	releaseProc, err := dll.FindProc("aqk2k_release")
	if err != nil {
		dll.Release()
		os.RemoveAll(tempDir)
		return nil, fmt.Errorf("aqk2k_release not found: %w", err)
	}

	convertProc, err := dll.FindProc("aqk2k_convert")
	if err != nil {
		dll.Release()
		os.RemoveAll(tempDir)
		return nil, fmt.Errorf("aqk2k_convert not found: %w", err)
	}

	convertRomanProc, err := dll.FindProc("aqk2k_convert_roman")
	if err != nil {
		dll.Release()
		os.RemoveAll(tempDir)
		return nil, fmt.Errorf("aqk2k_convert_roman not found: %w", err)
	}

	aq := &AqKanji2Koe{
		dll:              dll,
		createProc:       createProc,
		releaseProc:      releaseProc,
		convertProc:      convertProc,
		convertRomanProc: convertRomanProc,
		tempDir:          tempDir,
	}

	// 初期化
	ret, _, _ := aq.createProc.Call()
	if ret != ErrNone {
		dll.Release()
		os.RemoveAll(tempDir)
		return nil, fmt.Errorf("aqk2k_create failed (code: %d)", ret)
	}

	return aq, nil
}

func (a *AqKanji2Koe) Close() error {
	if a.dll != nil {
		a.releaseProc.Call()
		a.dll.Release()
		a.dll = nil
	}
	if a.tempDir != "" {
		os.RemoveAll(a.tempDir)
		a.tempDir = ""
	}
	return nil
}

// Convert は漢字かな混じりテキストをかな音声記号列に変換する（UTF-8）
func (a *AqKanji2Koe) Convert(text string) (string, error) {
	return a.callConvert(a.convertProc, text)
}

// ConvertRoman は漢字かな混じりテキストをローマ字音声記号列に変換する（UTF-8）
func (a *AqKanji2Koe) ConvertRoman(text string) (string, error) {
	return a.callConvert(a.convertRomanProc, text)
}

// callConvert は convert / convert_roman の共通処理
func (a *AqKanji2Koe) callConvert(proc *syscall.Proc, text string) (string, error) {
	cInput, err := syscall.BytePtrFromString(text)
	if err != nil {
		return "", fmt.Errorf("invalid input: %w", err)
	}

	// 入力テキストの2倍以上、最低256バイト
	bufSize := max(len(text)*2, minBufSize)

	buf := make([]byte, bufSize)

	ret, _, _ := proc.Call(
		uintptr(unsafe.Pointer(cInput)),
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(len(buf)),
	)

	if ret != ErrNone {
		return "", fmt.Errorf("convert failed (code: %d)", ret)
	}

	// NULL終端までの文字列を取得
	n := 0
	for n < len(buf) && buf[n] != 0 {
		n++
	}
	return string(buf[:n]), nil
}
