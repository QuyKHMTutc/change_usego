package ui

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"syscall"
	"unicode/utf8"
	"unsafe"

	"tools2/pkg/colors"
)

var (
	user32                      = syscall.NewLazyDLL("user32.dll")
	procGetConsoleWindow        = kernel32Proc.NewProc("GetConsoleWindow")
	procGetSystemMetrics        = user32.NewProc("GetSystemMetrics")
	procSetWindowPos            = user32.NewProc("SetWindowPos")
	procSetConsoleCP            = kernel32Proc.NewProc("SetConsoleCP")
	procSetConsoleOutputCP      = kernel32Proc.NewProc("SetConsoleOutputCP")
	procSetCurrentConsoleFontEx = kernel32Proc.NewProc("SetCurrentConsoleFontEx")
	ansiRegex                   = regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)
)

type coord struct {
	X int16
	Y int16
}

type consoleFontInfoEx struct {
	cbSize     uint32
	nFont      uint32
	dwFontSize coord
	fontFamily uint32
	fontWeight uint32
	faceName   [32]uint16
}

// SetupTerminalWindow tu dong phong to man hinh terminal (128 cot x 50 dong) va can giua man hinh
func SetupTerminalWindow() {
	// 1. Chuyen console codepage sang UTF-8 (65001)
	_, _, _ = procSetConsoleCP.Call(65001)
	_, _, _ = procSetConsoleOutputCP.Call(65001)

	// 2. Thiet lap font Consolas 17px ro net
	handle, err := syscall.GetStdHandle(syscall.STD_OUTPUT_HANDLE)
	if err == nil && handle != syscall.InvalidHandle {
		var font consoleFontInfoEx
		font.cbSize = uint32(unsafe.Sizeof(font))
		font.dwFontSize.Y = 16
		font.fontFamily = 54 // TMPF_VECTOR | FF_MODERN
		font.fontWeight = 400
		copy(font.faceName[:], syscall.StringToUTF16("Consolas"))
		_, _, _ = procSetCurrentConsoleFontEx.Call(uintptr(handle), 0, uintptr(unsafe.Pointer(&font)))
	}

	// 3. Goi mode con de co gian vua khit bang (124 cot x 48 dong)
	cmd := exec.Command("cmd", "/c", "mode con: cols=124 lines=48")
	_ = cmd.Run()

	// 4. Can giua va co gian kich thuoc cua so gon gang vua van voi bang (khong mo qua rong)
	hwnd, _, _ := procGetConsoleWindow.Call()
	if hwnd != 0 {
		sw, _, _ := procGetSystemMetrics.Call(0) // SM_CXSCREEN
		sh, _, _ := procGetSystemMetrics.Call(1) // SM_CYSCREEN

		winW := 1130
		if winW > int(sw)-30 {
			winW = int(sw) - 30
		}
		winH := 890
		if winH > int(sh)-50 {
			winH = int(sh) - 50
		}
		posX := (int(sw) - winW) / 2
		if posX < 0 {
			posX = 0
		}
		posY := (int(sh)-winH)/2 - 10
		if posY < 5 {
			posY = 5
		}

		// SWP_NOZORDER (0x0004) | SWP_SHOWWINDOW (0x0040)
		_, _, _ = procSetWindowPos.Call(hwnd, 0, uintptr(posX), uintptr(posY), uintptr(winW), uintptr(winH), 0x0044)
	}
}

// StripANSI loai bo tat ca ma mau ANSI de tinh toan chieu rong that
func StripANSI(s string) string {
	return ansiRegex.ReplaceAllString(s, "")
}

// VisibleWidth do so cot hien thi thuc te (tinh theo runes khong chua ANSI)
func VisibleWidth(s string) int {
	return utf8.RuneCountInString(StripANSI(s))
}

// PadRightVisible chen khoang trang ben phai de dat dung do rong cot
func PadRightVisible(s string, width int) string {
	w := VisibleWidth(s)
	if w >= width {
		return s
	}
	return s + strings.Repeat(" ", width-w)
}

// CenterVisible can giua chuoi trong khoang rong
func CenterVisible(s string, width int) string {
	w := VisibleWidth(s)
	if w >= width {
		return s
	}
	diff := width - w
	left := diff / 2
	right := diff - left
	return strings.Repeat(" ", left) + s + strings.Repeat(" ", right)
}

type MenuItem struct {
	ID    int
	Name  string
	Desc  string
	IsVIP bool
}

var MenuItems = []MenuItem{
	{1, "Làm sạch ký tự & Viết hoa chữ đầu", "Chỉ giữ chữ + số, viết hoa chữ cái đầu tiên", false},
	{2, "Chèn Username Email vào đầu Pass", "Xóa chữ trong pass, lấy Username dán lên đầu", false},
	{3, "Dồn số ra sau & nhân đôi 2 số đầu", "Làm sạch ký tự, hoa đầu, lấy 2 số đầu dán vào cuối", false},
	{4, "Đưa số lên đầu + Chữ viết hoa ra sau", "Tách số đưa lên trước, chữ viết hoa theo sau", false},
	{5, "Chữ Username + Số trong Mật khẩu", "Lấy chữ username ghép với toàn bộ số của pass", false},
	{6, "Dồn số ra sau + Chữ đầu Username vào đuôi", "Chuyển số ra sau, chữ đầu email in hoa vào đuôi", false},
	{7, "Đảo 2 chữ đầu Username + Giữ số Pass", "2 chữ đầu username đảo ngược in hoa + số pass cũ", false},
	{8, "Đảo ngược thứ tự các con số trong Pass", "Đảo ngược dãy số nhưng giữ nguyên vị trí chữ", false},
	{9, "Lặp / Nhân đôi số cuối mật khẩu", "Lặp 1-2 số cuối; nếu >2 số lặp 2 số đầu", false},
	{10, "2 chữ Username + Viết hoa toàn bộ", "2 chữ đầu email + pass hoa (>=4 số)", false},
	{11, "Lấy số Email thêm vào cuối Pass", "Lấy tất cả số & 2 số cuối email ghép vào đuôi", false},
	{12, "Mã hóa ký tự Pass (12 a, 12 i, 12 s)", "Gõ 12 a (a->@), 12 i (i->!), 12 s (s->$)", false},
	{13, "Viết hoa chữ cái sau Số cuối cùng", "Sau con số cuối cùng, phần chữ sẽ viết hoa", false},
	{14, "N Chữ đầu Email + Số (Gõ 14.<N> tùy ý)", "Lấy N chữ: 14.2 (2 chữ), 14.3, 14.5 (5 chữ)...", false},
	{15, "Ký tự đặc biệt TRƯỚC SỐ (15 <ký tự>)", "Ký tự tùy ý: 15 @, 15 $, 15 !, 15 %, 15 _...", false},
	{16, "Ký tự đặc biệt SAU SỐ (16 <ký tự>)", "Ký tự tùy ý: 16 @, 16 #, 16 !, 16 %, 16 _...", false},
	{17, "Tên Email viết hoa đầu + Số mật khẩu", "Lấy tên email ghép số (không có số chỉ lấy tên)", false},
	{18, "Số mật khẩu + Tên Email", "Số mật khẩu đứng trước, tên email đứng sau", false},
	{19, "Nhân đôi ký tự/số ở vị trí cuối cùng", "Nhân đôi ký tự cuối (123->1233, pass!->pass!!)", false},
	{20, "Chữ bất kỳ từ bàn phím + Số mật khẩu cũ", "Nhập chữ từ phím ghép số cũ (Ví dụ: 20 Aa | 20 Ab)", false},
	{21, "Viết Hoa / Thường chữ cái đầu Password", "Gõ 21 H (viết hoa chữ đầu) | 21 T (viết thường chữ đầu)", false},
	{22, "Lọc mật khẩu >= 6 ký tự", "Xóa dòng chỉ có mail, xóa Password < 6 ký tự", false},
	{23, "Xóa ký tự đặc biệt (Giữ chữ & số)", "Chỉ xóa ký tự đặc biệt, giữ nguyên chữ và số", false},
	{24, "Chỉ lấy chữ trong Pass (>= 6 chữ)", "Lấy chữ trong pass, xóa số & ký tự (giữ pass >= 6 chữ)", false},
	{25, "Tách file Domain & Phần còn lại", "Tách 2 file: 25_hotmail / 25_gmail VÀ 25_other.txt", true},
	{26, "Chỉ lấy số trong Pass (>= 6 số)", "Lấy số trong pass, xóa chữ & ký tự (giữ pass >= 6 số)", false},
	{27, "N Chữ Email + Số Pass (27 3D / 27 3C)", "Ghép ĐẦU (3D) / CUỐI (3C) + gộp 7 biến thể (!, @, $, *, #)", false},
	{28, "Tên Email làm Pass mới (6-15 chữ)", "Lấy chữ từ tên mail viết hoa đầu làm pass mới", false},
	{29, "Chèn ký tự đặc biệt tại vị trí bất kỳ", "Gõ 29 <vị_trí><ký_tự> (Ví dụ: 29 2@, 29 3$...)", false},
	{30, "Ghép 2 số cuối Pass thành Năm (19xx/20xx)", "50-99 ghép 1950-1999, 01-26 ghép 2001-2026 (loại còn lại)", false},
}

// getFeatureThemeColor phoi mau hien dai theo tung cum chuc nang (Modern Functional Palette)
func getFeatureThemeColor(fid int) string {
	switch {
	case fid >= 1 && fid <= 5:
		return colors.MintGreen // Nhom 1: Lam sach & Chuan hoa
	case fid >= 6 && fid <= 10:
		return colors.MangoGold // Nhom 2: Dao nguoc & Nhan doi so
	case fid >= 11 && fid <= 15:
		return colors.IceBlue // Nhom 3: Ma hoa & Ghep chu so
	case fid >= 16 && fid <= 20:
		return colors.Lavender // Nhom 4: Ky tu dac biet & To hop phim
	case fid >= 21 && fid <= 25:
		return colors.CyanNeon // Nhom 5: Bo loc chat luong & Tach Domain VIP
	default:
		return colors.RoseNeon // Nhom 6: AI / Tai cau truc am tiet nang cao
	}
}

// DisplayTableMenu in bang menu 3 cot voi ty le va khung vien chuan xac 100%
func DisplayTableMenu() {
	wSTT := 6
	wName := 43
	wDesc := 60

	bColor := colors.BorderCyan
	vBar := bColor + "│" + colors.Reset

	// Cong 2 khoang trang dem (1 trai, 1 phai) cho moi cot de vien va noi dung khop tuyet doi
	bTop := fmt.Sprintf("  %s┌%s┬%s┬%s┐%s", bColor, strings.Repeat("─", wSTT+2), strings.Repeat("─", wName+2), strings.Repeat("─", wDesc+2), colors.Reset)
	bMid := fmt.Sprintf("  %s├%s┼%s┼%s┤%s", bColor, strings.Repeat("─", wSTT+2), strings.Repeat("─", wName+2), strings.Repeat("─", wDesc+2), colors.Reset)
	bBot := fmt.Sprintf("  %s└%s┴%s┴%s┘%s", bColor, strings.Repeat("─", wSTT+2), strings.Repeat("─", wName+2), strings.Repeat("─", wDesc+2), colors.Reset)
	bSep := fmt.Sprintf("  %s├%s┼%s┼%s┤%s", colors.BorderDim, strings.Repeat("─", wSTT+2), strings.Repeat("─", wName+2), strings.Repeat("─", wDesc+2), colors.Reset)
	footerLine := fmt.Sprintf("  %s%s%s", bColor, strings.Repeat("─", (wSTT+2)+(wName+2)+(wDesc+2)+4), colors.Reset)

	hSTT := fmt.Sprintf("%s%sSTT%s", colors.Bold, colors.CyanNeon, colors.Reset)
	hName := fmt.Sprintf("%s%sCHỨC NĂNG%s", colors.Bold, colors.MangoGold, colors.Reset)
	hDesc := fmt.Sprintf("%s%sMÔ TẢ%s", colors.Bold, colors.MintGreen, colors.Reset)

	headerLine := fmt.Sprintf("  %s %s %s %s %s %s %s",
		vBar, CenterVisible(hSTT, wSTT),
		vBar, CenterVisible(hName, wName),
		vBar, CenterVisible(hDesc, wDesc),
		vBar,
	)

	fmt.Printf("  %s%s[Bước 2] DANH MỤC 30 CHỨC NĂNG XỬ LÝ COMBO%s\n", colors.Bold, colors.CyanNeon, colors.Reset)
	fmt.Println(bTop)
	fmt.Println(headerLine)
	fmt.Println(bMid)

	for _, item := range MenuItems {
		fid := item.ID
		sttStr := fmt.Sprintf("%s[%s%s%2d%s%s]%s", colors.BorderDim, colors.Reset, colors.CyanNeon, fid, colors.Reset, colors.BorderDim, colors.Reset)

		themeColor := getFeatureThemeColor(fid)
		nameStr := themeColor + item.Name + colors.Reset

		descStr := colors.TextMuted + item.Desc + colors.Reset

		row := fmt.Sprintf("  %s %s %s %s %s %s %s",
			vBar, CenterVisible(sttStr, wSTT),
			vBar, PadRightVisible(nameStr, wName),
			vBar, PadRightVisible(descStr, wDesc),
			vBar,
		)
		fmt.Println(row)

		if fid == 5 || fid == 10 || fid == 15 || fid == 20 || fid == 25 {
			fmt.Println(bSep)
		}
	}

	fmt.Println(bBot)
	fmt.Printf("  %s[+] Chạy nhiều chức năng: Ghép bằng dấu '|' (Ví dụ: %s12 a | 14.2 | 21 H | 25.1 | 27 3D%s)%s\n",
		colors.CyanNeon, colors.MangoAmber, colors.CyanNeon, colors.Reset)
	fmt.Printf("  %s[C]%s %sĐổi sang mở file khác         %s[E]%s %sThoát chương trình       %s(Giữ Ctrl + lăn chuột để chỉnh cỡ chữ)%s\n",
		colors.CyanNeon, colors.Reset, colors.TextRegular,
		colors.MangoRed, colors.Reset, colors.TextRegular,
		colors.TextMuted, colors.Reset,
	)
	fmt.Println(footerLine)
}

// PrintCelebrationBox in khung chuc mung hoan tat va liet ke cac file da tao
func PrintCelebrationBox(createdFiles []string, outDir string) {
	border := colors.MintGreen
	totalTasks := len(createdFiles)

	fmt.Printf("\n  %s╔════════════════════════════════════════════════════════════════╗%s\n", border, colors.Reset)
	fmt.Printf("  %s║%s%s       ĐÃ HOÀN TẤT TẤT CẢ %2d TÁC VỤ ĐƯỢC CHỌN !              %s%s║%s\n",
		border, colors.Bold, colors.White, totalTasks, colors.Reset, border, colors.Reset)
	fmt.Printf("  %s╠════════════════════════════════════════════════════════════════╣%s\n", border, colors.Reset)
	fmt.Printf("  %s║%s  %sLưu tại thư mục riêng:%s %s%s%s%s\n",
		border, colors.Reset, colors.TextMuted, colors.Reset, colors.Bold, colors.MangoGold, outDir, colors.Reset)

	for _, cf := range createdFiles {
		fmt.Printf("  %s║%s    %s[+]%s %s%-26s%s\n",
			border, colors.Reset, colors.MintGreen, colors.Reset, colors.MangoAmber, cf, colors.Reset)
	}
	fmt.Printf("  %s╚════════════════════════════════════════════════════════════════╝%s\n\n", border, colors.Reset)
}
