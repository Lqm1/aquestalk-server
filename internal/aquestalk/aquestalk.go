package aquestalk

import (
	"fmt"
	"path/filepath"
	"syscall"
	"unsafe"

	"golang.org/x/text/encoding/japanese"
)

// AquesTalk DLLハンドルと関数プロシージャを保持する構造体
type AquesTalk struct {
	dll          *syscall.DLL
	syntheProc   *syscall.Proc
	freeWaveProc *syscall.Proc
}

// 新しいAquesTalkインスタンスを作成（指定された話者用のDLLをロード）
func New(voice, baseDir string) (*AquesTalk, error) {
	dllPath := filepath.Join(baseDir, "dll", "aquestalk", voice, "AquesTalk.dll")
	dll, err := syscall.LoadDLL(dllPath)
	if err != nil {
		return nil, fmt.Errorf("DLL load error: %w", err)
	}

	syntheProc, err := dll.FindProc("AquesTalk_Synthe")
	if err != nil {
		dll.Release()
		return nil, fmt.Errorf("AquesTalk_Synthe not found: %w", err)
	}

	freeWaveProc, err := dll.FindProc("AquesTalk_FreeWave")
	if err != nil {
		dll.Release()
		return nil, fmt.Errorf("AquesTalk_FreeWave not found: %w", err)
	}

	return &AquesTalk{
		dll:          dll,
		syntheProc:   syntheProc,
		freeWaveProc: freeWaveProc,
	}, nil
}

// リソースの解放
func (a *AquesTalk) Close() error {
	if a.dll != nil {
		a.dll.Release()
		a.dll = nil
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
