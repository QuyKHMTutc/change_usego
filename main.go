package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"tools2/pkg/colors"
	"tools2/pkg/features"
	"tools2/pkg/processor"
	"tools2/pkg/ui"
)

var cleanNameRegex = regexp.MustCompile(`[\\/*?:"<>|]`)

func getResultsDir(inputPath string) string {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}
	baseResults := filepath.Join(home, "Desktop", "results")
	if inputPath != "" {
		baseName := strings.TrimSuffix(filepath.Base(inputPath), filepath.Ext(inputPath))
		safeSub := cleanNameRegex.ReplaceAllString(baseName, "_")
		safeSub = strings.TrimSpace(safeSub)
		if safeSub == "" {
			safeSub = "output"
		}
		// Cấu trúc: <Tên_File>_<Ngày>_<Giờ> (Ví dụ: hotmail 1_27-10-2026_11h45m12)
		timestamp := time.Now().Format("02-01-2006_15h04m05")
		folderName := fmt.Sprintf("%s_%s", safeSub, timestamp)
		outDir := filepath.Join(baseResults, folderName)
		_ = os.MkdirAll(outDir, 0755)
		return outDir
	}
	_ = os.MkdirAll(baseResults, 0755)
	return baseResults
}

func getSafeOutputFilename(outDir string, feat features.Feature, arg string) string {
	fid := feat.ID
	cleanArg := strings.TrimSpace(arg)

	var filename string
	switch fid {
	case 11:
		val := strings.ToLower(cleanArg)
		if val == "all" || val == "tatca" {
			filename = "11.all.txt"
		} else if val == "2" || val == "last2" {
			filename = "11.2.txt"
		} else if val != "" {
			safe := cleanNameRegex.ReplaceAllString(val, "_")
			filename = fmt.Sprintf("11_%s.txt", safe)
		} else {
			filename = "11.txt"
		}
	case 12:
		val := cleanArg
		if val == "" {
			val = "a"
		}
		safe := cleanNameRegex.ReplaceAllString(val, "_")
		filename = fmt.Sprintf("12_%s.txt", safe)
	case 14:
		val := cleanArg
		if val == "" {
			val = "2"
		}
		safe := cleanNameRegex.ReplaceAllString(val, "_")
		filename = fmt.Sprintf("14.%s.txt", safe)
	case 21:
		val := strings.ToUpper(cleanArg)
		if val != "T" {
			val = "H"
		}
		filename = fmt.Sprintf("21_%s.txt", val)
	case 25:
		val := strings.ToLower(cleanArg)
		val = strings.Trim(val, ".@")
		if val == "2" || strings.Contains(val, "gmail") {
			filename = "25_gmail.txt"
		} else if val == "1" || val == "hotmail" || val == "outlook" || val == "" {
			filename = "25_hotmail.txt"
		} else {
			safe := cleanNameRegex.ReplaceAllString(val, "_")
			filename = fmt.Sprintf("25_%s.txt", safe)
		}
	case 27:
		val := strings.ToUpper(cleanArg)
		pos := "D"
		if strings.Contains(val, "C") {
			pos = "C"
		}
		reNum := regexp.MustCompile(`\d+`)
		m := reNum.FindString(val)
		n := "3"
		if m != "" {
			n = m
		}
		filename = fmt.Sprintf("27_%s%s.txt", n, pos)
	case 29:
		val := cleanArg
		if val == "" {
			val = "2@"
		}
		safe := cleanNameRegex.ReplaceAllString(val, "_")
		filename = fmt.Sprintf("29_%s.txt", safe)
	case 30:
		filename = "30_year.txt"
	default:
		if cleanArg != "" {
			safe := cleanNameRegex.ReplaceAllString(cleanArg, "_")
			filename = fmt.Sprintf("%d_%s.txt", fid, safe)
		} else {
			filename = fmt.Sprintf("%d.txt", fid)
		}
	}

	return filepath.Join(outDir, filename)
}

func promptCustomArg(featID int, reader *bufio.Reader) string {
	switch featID {
	case 11:
		fmt.Printf("\n  %s> Chọn chế độ lấy số email %s(all: Tất cả số | 2: 2 số cuối | Enter: Lấy cả hai):%s ",
			colors.CyanNeon, colors.TextMuted, colors.Reset)
	case 12:
		fmt.Printf("\n  %s> Nhập ký tự muốn mã hóa %s(Chọn: a [a->@], i [i->!], s [s->$]):%s ",
			colors.CyanNeon, colors.TextMuted, colors.Reset)
	case 14:
		fmt.Printf("\n  %s> Nhập số lượng chữ cái đầu email muốn lấy %s(Ví dụ: 2, 3, 4, 5... bất kỳ):%s ",
			colors.CyanNeon, colors.TextMuted, colors.Reset)
	case 15:
		fmt.Printf("\n  %s> Nhập ký tự đặc biệt muốn chèn TRƯỚC số %s(Ví dụ: @, $, #, !, %%, _... bất kỳ):%s ",
			colors.CyanNeon, colors.TextMuted, colors.Reset)
	case 16:
		fmt.Printf("\n  %s> Nhập ký tự đặc biệt muốn chèn SAU số %s(Ví dụ: @, #, $, !, %%, _... bất kỳ):%s ",
			colors.CyanNeon, colors.TextMuted, colors.Reset)
	case 20:
		fmt.Printf("\n  %s> Nhập chuỗi ký tự muốn ghép cho chức năng [20] %s(Ví dụ: Aa):%s ",
			colors.CyanNeon, colors.TextMuted, colors.Reset)
	case 21:
		fmt.Printf("\n  %s> Chọn chế độ đổi chữ cái đầu %s(H: Viết HOA chữ đầu | T: Viết THƯỜNG chữ đầu):%s ",
			colors.CyanNeon, colors.TextMuted, colors.Reset)
	case 25:
		fmt.Printf("\n  %s> Nhập domain muốn tách %s(1: Hotmail/Outlook | 2: Gmail | hoặc domain bất kỳ ví dụ: yahoo.com):%s ",
			colors.CyanNeon, colors.TextMuted, colors.Reset)
	case 27:
		fmt.Printf("\n  %s> Nhập số ký tự và vị trí %s(Ví dụ: 3D cho vào đầu, 3C cho vào cuối):%s ",
			colors.CyanNeon, colors.TextMuted, colors.Reset)
	case 29:
		fmt.Printf("\n  %s> Nhập vị trí và ký tự muốn chèn %s(Ví dụ: 2@ chèn @ sau chữ 1, 3$ chèn $ sau chữ 2):%s ",
			colors.CyanNeon, colors.TextMuted, colors.Reset)
	default:
		fmt.Printf("\n  %s> Nhập chuỗi ký tự muốn ghép cho chức năng [%d] %s(Ví dụ: Aa):%s ",
			colors.CyanNeon, featID, colors.TextMuted, colors.Reset)
	}

	line, err := reader.ReadString('\n')
	if err != nil {
		return ""
	}
	val := strings.TrimSpace(line)

	if val == "" {
		switch featID {
		case 12:
			return "a"
		case 14:
			return "2"
		case 15:
			return "@"
		case 16:
			return "@"
		case 20:
			return "Aa"
		case 21:
			return "H"
		case 25:
			return "1"
		case 27:
			return "3D"
		case 29:
			return "2@"
		}
	}

	return val
}

func main() {
	ui.SetupTerminalWindow()

	home, _ := os.UserHomeDir()
	baseResults := filepath.Join(home, "Desktop", "results")
	if count, freed := processor.CleanupAllStaleTempFiles(baseResults); count > 0 {
		fmt.Printf("  %s[*] Đã tự động dọn dẹp %d file tạm cũ, giải phóng %.2f MB dung lượng.%s\n",
			colors.MintGreen, count, float64(freed)/(1024*1024), colors.Reset)
	}

	var currentFile string
	if len(os.Args) > 1 && os.Args[1] != "" {
		if _, err := os.Stat(os.Args[1]); err == nil {
			currentFile = os.Args[1]
		}
	}

	reader := bufio.NewReader(os.Stdin)

	for currentFile == "" {
		ui.ClearScreen()
		ui.PrintBanner()
		fmt.Printf("  %s[>] Đang mở cửa sổ chọn file dữ liệu đầu vào...%s\n", colors.CyanNeon, colors.Reset)
		chosen, err := ui.OpenFileDialog("Chọn file dữ liệu đầu vào (.txt)")
		if err == nil && chosen != "" {
			currentFile = chosen
			ui.SetupTerminalWindow()
			break
		}

		fmt.Printf("  %s[!] Chưa chọn file qua cửa sổ. Nhập đường dẫn thủ công (hoặc gõ E để thoát):%s ", colors.Yellow, colors.Reset)
		line, err := reader.ReadString('\n')
		if err != nil {
			return
		}
		line = strings.TrimSpace(line)
		line = strings.Trim(line, "\"'")
		if strings.EqualFold(line, "E") || strings.EqualFold(line, "exit") {
			return
		}
		if _, err := os.Stat(line); err == nil {
			currentFile = line
			break
		}
		fmt.Printf("  %s[!] Đường dẫn file không tồn tại, vui lòng thử lại!%s\n", colors.MangoRed, colors.Reset)
	}

	for {
		ui.ClearScreen()
		ui.PrintBanner()

		fi, err := os.Stat(currentFile)
		if err != nil {
			fmt.Printf("  %s[!] Lỗi đọc file: %v%s\n", colors.MangoRed, err, colors.Reset)
			currentFile = ""
			for currentFile == "" {
				chosen, _ := ui.OpenFileDialog("Chọn file dữ liệu đầu vào (.txt)")
				if chosen != "" {
					currentFile = chosen
				}
			}
			continue
		}

		fileMB := float64(fi.Size()) / (1024 * 1024)
		fmt.Printf("  %s[+] File hiện tại:%s %s%s%s (%s%.2f MB%s)\n",
			colors.MangoAmber, colors.Reset,
			colors.Bold, filepath.Base(currentFile), colors.Reset,
			colors.CyanNeon, fileMB, colors.Reset,
		)

		ui.DisplayTableMenu()

		fmt.Printf("  %s> Chọn chức năng (chạy nhiều mục dùng dấu |):%s ", colors.MangoGold, colors.Reset)
		choiceStr, err := reader.ReadString('\n')
		if err != nil {
			return
		}
		choiceStr = strings.TrimSpace(choiceStr)

		if choiceStr == "" {
			continue
		}

		if strings.EqualFold(choiceStr, "E") || strings.EqualFold(choiceStr, "exit") {
			fmt.Printf("\n  %sCảm ơn anh đã sử dụng TOOLS2! Hẹn gặp lại anh.%s\n\n", colors.MintGreen, colors.Reset)
			return
		}

		if strings.EqualFold(choiceStr, "C") {
			chosen, _ := ui.OpenFileDialog("Chọn file dữ liệu mới (.txt)")
			if chosen != "" {
				currentFile = chosen
			}
			continue
		}

		tasks, errMsg := features.ParseChoiceString(choiceStr)
		if errMsg != "" {
			fmt.Printf("\n  %s[!] %s%s\n", colors.MangoRed, errMsg, colors.Reset)
			fmt.Println("  Bấm phím Enter để tiếp tục...")
			_, _ = reader.ReadString('\n')
			continue
		}

		// Xử lý các chức năng cần tham số bổ sung nếu người dùng chưa nhập
		for i := range tasks {
			if tasks[i].Feat.HasArg && strings.TrimSpace(tasks[i].Arg) == "" {
				tasks[i].Arg = promptCustomArg(tasks[i].Feat.ID, reader)
			}
		}

		outDir := getResultsDir(currentFile)
		fmt.Printf("\n  %s===> BẮT ĐẦU XỬ LÝ %d TÁC VỤ...%s\n\n", colors.CyanNeon, len(tasks), colors.Reset)

		var createdFiles []string

		for idx, t := range tasks {
			fmt.Printf("  %s[%d/%d] Đang xử lý [%d]%s %s\n",
				colors.CyanNeon, idx+1, len(tasks), t.Feat.ID, colors.Reset, t.Feat.Name)

			if t.Feat.ID == 25 {
				outMatch := getSafeOutputFilename(outDir, t.Feat, t.Arg)
				outOtherName := "25_other.txt"
				if strings.Contains(strings.ToLower(filepath.Base(outMatch)), "gmail") {
					outOtherName = "25_other_gmail.txt"
				}
				outOther := filepath.Join(outDir, outOtherName)

				fmt.Printf("  Output ghi ra: %s%s%s\n", colors.CyanNeon, filepath.Base(outMatch), colors.Reset)
				fmt.Printf("  Output còn lại: %s%s%s\n", colors.MangoGold, filepath.Base(outOther), colors.Reset)

				if err := processor.ProcessDomainSplit(currentFile, outMatch, outOther, t.Arg); err != nil {
					fmt.Printf("  %s[!] Lỗi xử lý [%d]: %v%s\n", colors.MangoRed, t.Feat.ID, err, colors.Reset)
				} else {
					createdFiles = append(createdFiles, filepath.Base(outMatch), filepath.Base(outOther))
				}
			} else {
				outPath := getSafeOutputFilename(outDir, t.Feat, t.Arg)
				fmt.Printf("  Output ghi ra: %s%s%s\n", colors.CyanNeon, filepath.Base(outPath), colors.Reset)
				if err := processor.ProcessGeneric(currentFile, outPath, t.Feat, t.Arg); err != nil {
					fmt.Printf("  %s[!] Lỗi xử lý [%d]: %v%s\n", colors.MangoRed, t.Feat.ID, err, colors.Reset)
				} else {
					createdFiles = append(createdFiles, filepath.Base(outPath))
				}
			}
		}

		if len(createdFiles) > 0 {
			ui.PlayCompletionSound()
			ui.PrintCelebrationBox(createdFiles, outDir)
		} else {
			fmt.Printf("\n  %s[!] Không có file kết quả nào được tạo (dữ liệu rỗng hoặc đã bị lọc hết).%s\n\n", colors.Yellow, colors.Reset)
		}

		fmt.Printf("  %s[O]%s %sMở thư mục kết quả%s   %s│%s   %s[C]%s %sĐổi file combo%s   %s│%s   %s[Enter]%s %sđể tiếp tục:%s ",
			colors.MangoGold, colors.Reset, colors.White, colors.Reset,
			colors.BorderCyan, colors.Reset,
			colors.CyanNeon, colors.Reset, colors.White, colors.Reset,
			colors.BorderCyan, colors.Reset,
			colors.MintGreen, colors.Reset, colors.TextRegular, colors.Reset)

		postChoice, err := reader.ReadString('\n')
		if err != nil {
			return
		}
		postChoice = strings.TrimSpace(strings.ToUpper(postChoice))

		if postChoice == "O" {
			_ = exec.Command("explorer.exe", outDir).Start()
		} else if postChoice == "C" {
			chosen, _ := ui.OpenFileDialog("Chọn file dữ liệu mới (.txt)")
			if chosen != "" {
				currentFile = chosen
				ui.SetupTerminalWindow()
			}
		}
	}
}
