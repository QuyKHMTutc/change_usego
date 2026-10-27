package ui

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"tools2/pkg/colors"
	"tools2/pkg/features"
)

var (
	kernel32Proc = syscall.NewLazyDLL("kernel32.dll")
	procBeep     = kernel32Proc.NewProc("Beep")
)

// PlayCompletionSound phát âm thanh thông báo nhẹ nhàng khi xong
func PlayCompletionSound() {
	_, _, _ = procBeep.Call(587, 100)
	_, _, _ = procBeep.Call(880, 150)
}

// ClearScreen xoa sach toan bo man hinh console va scrollback buffer
func ClearScreen() {
	cmd := exec.Command("cmd", "/c", "cls")
	cmd.Stdout = os.Stdout
	_ = cmd.Run()
	fmt.Print("\033[H\033[2J\033[3J")
}

func PrintBanner() {
	ClearScreen()
	b1 := colors.MangoRed + `   ▄▄▄▄███▄▄▄▄      ▄████████ ███▄▄▄▄      ▄██████▄   ▄██████▄  ` + colors.Reset
	b2 := colors.MangoOrange + ` ▄██▀▀█████▀▀██▄   ███    ███ ███▀▀▀██▄   ███    ███ ███    ███ ` + colors.Reset
	b3 := colors.MangoAmber + ` ███   ███   ███   ███    ███ ███   ███   ███    █▀  ███    ███ ` + colors.Reset
	b4 := colors.MangoGold + ` ███   ███   ███   ███    ███ ███   ███  ▄███        ███    ███ ` + colors.Reset
	b5 := colors.MangoYellow + ` ███   ███   ███ ▀███████████ ███   ███ ▀▀███ ████▄  ███    ███ ` + colors.Reset
	b6 := colors.Yellow + ` ███   ███   ███   ███    ███ ███   ███   ███    ███ ███    ███ ` + colors.Reset
	b7 := colors.White + `  ▀█   ███   █▀    ███    ███  ▀█   █▀    ███    ███  ▀██████▀  ` + colors.Reset
	b8 := colors.White + `                   ███    ███                ▀██████▀           ` + colors.Reset

	fmt.Println(b1)
	fmt.Println(b2)
	fmt.Println(b3)
	fmt.Println(b4)
	fmt.Println(b5)
	fmt.Println(b6)
	fmt.Println(b7)
	fmt.Println(b8)
	fmt.Printf("%s        [+] NATIVE GO ULTRA ENGINE V2 - 200M+ CPM  |  Goroutines Zero-IPC  |  Tự dọn rác .tmp%s\n", colors.CyanNeon, colors.Reset)
}

func PrintMenu() {
	fmt.Printf("%s  DANH SÁCH 30 CHỨC NĂNG TỐI ƯU SIÊU TỐC:%s\n", colors.MangoGold, colors.Reset)
	fmt.Printf("%s  ─────────────────────────────────────────────────────────────────────────────%s\n", colors.TextMuted, colors.Reset)

	col1 := features.Features[:15]
	col2 := features.Features[15:]

	for i := 0; i < len(col1); i++ {
		f1 := col1[i]
		line1 := fmt.Sprintf(" %s[%2d]%s %-32s", colors.CyanNeon, f1.ID, colors.Reset, f1.Name)
		if len(f1.Name) > 32 {
			line1 = fmt.Sprintf(" %s[%2d]%s %s", colors.CyanNeon, f1.ID, colors.Reset, f1.Name[:32])
		}

		if i < len(col2) {
			f2 := col2[i]
			line2 := fmt.Sprintf(" %s[%2d]%s %s", colors.CyanNeon, f2.ID, colors.Reset, f2.Name)
			fmt.Printf(" %-45s │ %s\n", line1, line2)
		} else {
			fmt.Printf(" %s\n", line1)
		}
	}
	fmt.Printf("%s  ─────────────────────────────────────────────────────────────────────────────%s\n", colors.TextMuted, colors.Reset)
	fmt.Printf("  %s[+] Chạy nhiều mục dùng dấu '|' (Ví dụ: 12 a | 14.2 | 21 H | 25.1 | 27 3D)%s\n", colors.MintGreen, colors.Reset)
	fmt.Printf("  %s[C] Đổi sang mở file khác             [E] Thoát chương trình%s\n\n", colors.MangoAmber, colors.Reset)
}
