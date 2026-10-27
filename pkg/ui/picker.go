package ui

import (
	"syscall"
	"unsafe"
)

var (
	comdlg32            = syscall.NewLazyDLL("comdlg32.dll")
	procGetOpenFileName = comdlg32.NewProc("GetOpenFileNameW")
)

type tagOFNW struct {
	lStructSize       uint32
	hwndOwner         uintptr
	hInstance         uintptr
	lpstrFilter       *uint16
	lpstrCustomFilter *uint16
	nMaxCustFilter    uint32
	nFilterIndex      uint32
	lpstrFile         *uint16
	nMaxFile          uint32
	lpstrFileTitle    *uint16
	nMaxFileTitle     uint32
	lpstrInitialDir   *uint16
	lpstrTitle        *uint16
	Flags             uint32
	nFileOffset       uint16
	nFileExtension    uint16
	lpstrDefExt       *uint16
	lCustData         uintptr
	lpfnHook          uintptr
	lpTemplateName    *uint16
	pvReserved        uintptr
	dwReserved        uint32
	flagsEx           uint32
}

const (
	OFN_FILEMUSTEXIST   = 0x00001000
	OFN_PATHMUSTEXIST   = 0x00000800
	OFN_HIDEREADONLY    = 0x00000004
	OFN_EXPLORER        = 0x00080000
)

// OpenFileDialog mở cửa sổ chọn file chuẩn của Windows Explorer trong 0.001 giây
func OpenFileDialog(title string) (string, error) {
	fileBuf := make([]uint16, 32768)
	filter, _ := syscall.UTF16PtrFromString("Text Files (*.txt)\x00*.txt\x00All Files (*.*)\x00*.*\x00\x00")
	titlePtr, _ := syscall.UTF16PtrFromString(title)

	var ofn tagOFNW
	ofn.lStructSize = uint32(unsafe.Sizeof(ofn))
	ofn.lpstrFilter = filter
	ofn.lpstrFile = &fileBuf[0]
	ofn.nMaxFile = uint32(len(fileBuf))
	ofn.lpstrTitle = titlePtr
	ofn.Flags = OFN_FILEMUSTEXIST | OFN_PATHMUSTEXIST | OFN_HIDEREADONLY | OFN_EXPLORER

	r1, _, _ := procGetOpenFileName.Call(uintptr(unsafe.Pointer(&ofn)))
	if r1 == 0 {
		return "", nil // Người dùng hủy chọn
	}

	return syscall.UTF16ToString(fileBuf), nil
}

