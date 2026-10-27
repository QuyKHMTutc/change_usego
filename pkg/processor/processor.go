package processor

import (
	"strconv"
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
	"unsafe"

	"tools2/pkg/colors"
	"tools2/pkg/features"
	"tools2/pkg/transforms"
)

var (
	kernel32                       = syscall.NewLazyDLL("kernel32.dll")
	procGetConsoleScreenBufferInfo = kernel32.NewProc("GetConsoleScreenBufferInfo")
)

type coord struct {
	X int16
	Y int16
}

type smallRect struct {
	Left   int16
	Top    int16
	Right  int16
	Bottom int16
}

type consoleScreenBufferInfo struct {
	Size              coord
	CursorPosition    coord
	Attributes        uint16
	Window            smallRect
	MaximumWindowSize coord
}

func getTerminalWidth() int {
	handle, err := syscall.GetStdHandle(syscall.STD_OUTPUT_HANDLE)
	if err != nil {
		return 80
	}
	var csbi consoleScreenBufferInfo
	r1, _, _ := procGetConsoleScreenBufferInfo.Call(uintptr(handle), uintptr(unsafe.Pointer(&csbi)))
	if r1 == 0 {
		return 80
	}
	w := int(csbi.Window.Right - csbi.Window.Left + 1)
	if w <= 0 {
		return 80
	}
	return w
}

type ByteRange struct {
	Index int
	Start int64
	End   int64
}

// SplitByteRanges chia file thành numChunks dải byte độc lập căn đúng ranh giới ký tự xuống dòng (\n)
func SplitByteRanges(filePath string, numChunks int) ([]ByteRange, error) {
	fi, err := os.Stat(filePath)
	if err != nil {
		return nil, err
	}
	fileSize := fi.Size()
	if fileSize == 0 || numChunks <= 1 {
		return []ByteRange{{Index: 0, Start: 0, End: fileSize}}, nil
	}

	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	step := fileSize / int64(numChunks)
	startsMap := make(map[int64]bool)
	startsMap[0] = true

	buf := make([]byte, 1024)
	for i := 1; i < numChunks; i++ {
		pos := int64(i) * step
		if _, err := f.Seek(pos, io.SeekStart); err != nil {
			continue
		}
		r := bufio.NewReader(f)
		_, err := r.ReadBytes('\n')
		if err != nil {
			continue
		}
		curPos, _ := f.Seek(0, io.SeekCurrent)
		// bù lại phần buffered chưa đọc
		actualStart := curPos - int64(r.Buffered())
		if actualStart < fileSize {
			startsMap[actualStart] = true
		}
	}
	_ = buf

	var starts []int64
	for s := range startsMap {
		starts = append(starts, s)
	}
	sort.Slice(starts, func(i, j int) bool { return starts[i] < starts[j] })

	var ranges []ByteRange
	for i := 0; i < len(starts); i++ {
		st := starts[i]
		var en int64
		if i+1 < len(starts) {
			en = starts[i+1]
		} else {
			en = fileSize
		}
		if st < en {
			ranges = append(ranges, ByteRange{
				Index: len(ranges),
				Start: st,
				End:   en,
			})
		}
	}
	return ranges, nil
}

// CountExactLines đếm chính xác số dòng bằng Multi-goroutine siêu tốc
func CountExactLines(filePath string) int64 {
	fi, err := os.Stat(filePath)
	if err != nil || fi.Size() == 0 {
		return 0
	}
	fileSize := fi.Size()

	threads := runtime.NumCPU()
	if threads < 4 {
		threads = 4
	}
	if threads > 32 {
		threads = 32
	}

	step := fileSize / int64(threads)
	var wg sync.WaitGroup
	var totalLines int64

	for i := 0; i < threads; i++ {
		st := int64(i) * step
		en := int64(i+1) * step
		if i == threads-1 {
			en = fileSize
		}

		wg.Add(1)
		go func(start, end int64) {
			defer wg.Done()
			f, err := os.Open(filePath)
			if err != nil {
				return
			}
			defer f.Close()

			if _, err := f.Seek(start, io.SeekStart); err != nil {
				return
			}

			buf := make([]byte, 4*1024*1024)
			rem := end - start
			var count int64

			for rem > 0 {
				toRead := int64(len(buf))
				if rem < toRead {
					toRead = rem
				}
				n, err := f.Read(buf[:toRead])
				if n > 0 {
					for j := 0; j < n; j++ {
						if buf[j] == '\n' {
							count++
						}
					}
					rem -= int64(n)
				}
				if err != nil {
					break
				}
			}
			atomic.AddInt64(&totalLines, count)
		}(st, en)
	}

	wg.Wait()

	// Kiểm tra ký tự cuối cùng
	f, err := os.Open(filePath)
	if err == nil {
		defer f.Close()
		lastByte := make([]byte, 1)
		if _, err := f.Seek(fileSize-1, io.SeekStart); err == nil {
			if _, err := f.Read(lastByte); err == nil {
				if lastByte[0] != '\n' && lastByte[0] != '\r' {
					totalLines++
				}
			}
		}
	}

	return totalLines
}


func formatCommas(n int64) string {
	in := strconv.FormatInt(n, 10)
	var out []byte
	l := len(in)
	for i, c := range in {
		if i > 0 && (l-i)%3 == 0 {
			out = append(out, ',')
		}
		out = append(out, byte(c))
	}
	return string(out)
}

func formatLinesCount(cnt int64, isMillion bool) string {
	if cnt >= 1_000_000 || isMillion {
		val := float64(cnt) / 1_000_000.0
		return fmt.Sprintf("%.2fM", val)
	} else if cnt >= 1_000 {
		val := float64(cnt) / 1_000.0
		return fmt.Sprintf("%.1fK", val)
	}
	return fmt.Sprintf("%d", cnt)
}

func renderProgressBar(
	doneBytes, totalBytes int64,
	doneTasks, totalTasks int,
	elapsed float64,
	doneLines, totalLines int64,
	label string,
) {
	var pct int64 = 100
	if totalBytes > 0 {
		pct = doneBytes * 100 / totalBytes
		if pct > 100 {
			pct = 100
		}
	}

	doneMB := float64(doneBytes) / (1024 * 1024)
	totalMB := float64(totalBytes) / (1024 * 1024)

	speed := 0.0
	if elapsed > 0 {
		speed = doneMB / elapsed
	}

	remBytes := totalBytes - doneBytes
	if remBytes < 0 {
		remBytes = 0
	}
	remTime := 0.0
	if speed > 0 {
		remTime = float64(remBytes) / (speed * 1024 * 1024)
	}

	var etaStr string
	if remTime > 60 {
		etaStr = fmt.Sprintf("%dm%02ds", int(remTime)/60, int(remTime)%60)
	} else {
		etaStr = fmt.Sprintf("%ds", int(remTime))
	}

	mbDisp := fmt.Sprintf("(%.1f / %.1f MB)", doneMB, totalMB)

	var cpm int64
	if elapsed > 0 && doneLines > 0 {
		cpm = int64(float64(doneLines*60) / elapsed)
	}

	cleanLbl := label
	if strings.Contains(strings.ToLower(label), "tách") {
		cleanLbl = "Tách domain"
	} else if cleanLbl == "" {
		cleanLbl = "Đang xử lý"
	}

	barLen := 20
	filled := int(pct * int64(barLen) / 100)
	if filled > barLen {
		filled = barLen
	}
	bar := colors.MangoOrange + strings.Repeat("█", filled) + colors.Reset + colors.White + strings.Repeat("░", barLen-filled) + colors.Reset

	fmt.Printf(
		"\r  %s[>] %s:%s [%s] %s%s%3d%%%s %s | %s%s%s cpm%s | %sETA: %-4s%s\033[K",
		colors.CyanNeon, cleanLbl, colors.Reset,
		bar,
		colors.Bold, colors.MangoYellow, pct, colors.Reset,
		mbDisp,
		colors.Bold, colors.MangoGold, formatCommas(cpm), colors.Reset,
		colors.TextMuted, etaStr, colors.Reset,
	)
}

// MergeTempFiles gộp các file part tạm thành file kết quả cuối cùng với tốc độ SSD tối đa và thanh tiến trình trực quan
func MergeTempFiles(tempPaths []string, finalPath string, label string) error {
	var totalMergeBytes int64
	for _, p := range tempPaths {
		if fi, err := os.Stat(p); err == nil {
			totalMergeBytes += fi.Size()
		}
	}

	fout, err := os.Create(finalPath)
	if err != nil {
		return err
	}
	defer fout.Close()

	numParts := len(tempPaths)
	if label == "" {
		fmt.Printf("  %s> Đang gộp %d phân đoạn thành file kết quả (SSD Block Merge)...%s\n", colors.CyanNeon, numParts, colors.Reset)
	} else {
		fmt.Printf("  %s> Đang gộp %d phân đoạn cho [%s] (SSD Block Merge)...%s\n", colors.CyanNeon, numParts, label, colors.Reset)
	}

	showProgress := totalMergeBytes > 5*1024*1024
	var mergedBytes int64
	t0 := time.Now()
	buf := make([]byte, 32*1024*1024) // 32MB buffer

	for _, p := range tempPaths {
		fin, err := os.Open(p)
		if err != nil {
			continue
		}
		for {
			n, rErr := fin.Read(buf)
			if n > 0 {
				_, _ = fout.Write(buf[:n])
				mergedBytes += int64(n)

				if showProgress {
					el := time.Since(t0).Seconds()
					if el < 0.001 {
						el = 0.001
					}
					speedMB := (float64(mergedBytes) / (1024 * 1024)) / el
					pct := int64(100)
					if totalMergeBytes > 0 {
						pct = mergedBytes * 100 / totalMergeBytes
						if pct > 100 {
							pct = 100
						}
					}
					mbDone := float64(mergedBytes) / (1024 * 1024)
					mbTotal := float64(totalMergeBytes) / (1024 * 1024)
					remMB := mbTotal - mbDone
					if remMB < 0 {
						remMB = 0
					}
					remSec := 0
					if speedMB > 0 {
						remSec = int(remMB / speedMB)
					}

					barLen := 20
					filled := int(pct * int64(barLen) / 100)
					if filled > barLen {
						filled = barLen
					}
					bar := colors.CyanNeon + strings.Repeat("█", filled) + colors.Reset + colors.White + strings.Repeat("░", barLen-filled) + colors.Reset

					fmt.Printf(
						"\r  %s[*] Đang gộp file:%s [%s] %s%s%3d%%%s (%.1f/%.1f MB) | %s%s%.1f MB/s%s | %sETA: %ds%s\033[K",
						colors.CyanNeon, colors.Reset,
						bar,
						colors.Bold, colors.CyanNeon, pct, colors.Reset,
						mbDone, mbTotal,
						colors.Bold, colors.MintGreen, speedMB, colors.Reset,
						colors.TextMuted, remSec, colors.Reset,
					)
				}
			}
			if rErr != nil {
				break
			}
		}
		fin.Close()
		_ = os.Remove(p)
	}

	if showProgress {
		el := time.Since(t0).Seconds()
		if el < 0.001 {
			el = 0.001
		}
		speedMB := (float64(mergedBytes) / (1024 * 1024)) / el
		mbTotal := float64(totalMergeBytes) / (1024 * 1024)
		bar := colors.CyanNeon + strings.Repeat("█", 20) + colors.Reset
		fmt.Printf(
			"\r  %s[*] Đang gộp file:%s [%s] %s%s100%%%s (%.1f/%.1f MB) | %s%s%.1f MB/s%s | %sETA: 0s%s\033[K\n",
			colors.CyanNeon, colors.Reset,
			bar,
			colors.Bold, colors.CyanNeon, colors.Reset,
			mbTotal, mbTotal,
			colors.Bold, colors.MintGreen, speedMB, colors.Reset,
			colors.TextMuted, colors.Reset,
		)
	}

	return nil
}

// CleanupStaleTempFiles dọn dẹp các file tạm .tmp cũ từ lần chạy trước
func CleanupStaleTempFiles(dir string) (int, int64) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return 0, 0
	}
	cnt := 0
	var bytesFreed int64
	for _, e := range entries {
		if !e.IsDir() {
			name := e.Name()
			if strings.HasPrefix(name, ".tmp_") || strings.HasSuffix(name, ".tmp") {
				info, err := e.Info()
				if err == nil {
					bytesFreed += info.Size()
				}
				_ = os.Remove(filepath.Join(dir, name))
				cnt++
			}
		}
	}
	return cnt, bytesFreed
}

// MergeTempFiles gộp các file part tạm thành file kết quả cuối cùng với tốc độ SSD tối đa (32MB block copy)
func ProcessGeneric(inputPath, outputPath string, feat features.Feature, arg string) error {
	startTime := time.Now()
	fi, err := os.Stat(inputPath)
	if err != nil {
		return err
	}
	fileSize := fi.Size()
	if fileSize == 0 {
		return fmt.Errorf("file rỗng (0 bytes)")
	}

	numWorkers := runtime.NumCPU()
	if numWorkers > 4 {
		numWorkers = numWorkers - 1
	}

	targetChunkSize := int64(48 * 1024 * 1024)
	numTasks := int(fileSize / targetChunkSize)
	if numTasks < numWorkers {
		numTasks = numWorkers
	}
	if numTasks > numWorkers*16 {
		numTasks = numWorkers * 16
	}

	ranges, err := SplitByteRanges(inputPath, numTasks)
	if err != nil {
		return err
	}
	actualTasks := len(ranges)

	tempDir := filepath.Dir(outputPath)
	baseOut := filepath.Base(outputPath)
	CleanupStaleTempFiles(tempDir)

	inMB := float64(fileSize) / (1024 * 1024)
	fmt.Printf("  %s> Kích thước file: %s%s%.2f MB%s %s| Chia thành %d dải byte độc lập%s\n",
		colors.MangoAmber, colors.Bold, colors.TextWhite, inMB, colors.Reset, colors.TextMuted, actualTasks, colors.Reset)
	fmt.Printf("  %s> Khởi chạy %d Goroutines CPU song song (Native Zero-Allocation Engine)...%s\n",
		colors.MangoGold, numWorkers, colors.Reset)

	exactLines := CountExactLines(inputPath)

	var doneBytes int64
	var doneTasks int64
	var totalValid int64
	var totalProcessed int64

	tempPaths := make([]string, actualTasks)
	for i := range ranges {
		tempPaths[i] = filepath.Join(tempDir, fmt.Sprintf(".tmp_%s_part_%d.tmp", baseOut, i))
	}

	// Kênh điều phối công việc
	taskChan := make(chan ByteRange, actualTasks)
	for _, r := range ranges {
		taskChan <- r
	}
	close(taskChan)

	var wg sync.WaitGroup
	renderProgressBar(0, fileSize, 0, actualTasks, 0.001, 0, exactLines, "Đang xử lý")

	// Ticker cập nhật giao diện 80ms
	ticker := time.NewTicker(80 * time.Millisecond)
	stopProgress := make(chan struct{})

	go func() {
		for {
			select {
			case <-ticker.C:
				el := time.Since(startTime).Seconds()
				db := atomic.LoadInt64(&doneBytes)
				dt := int(atomic.LoadInt64(&doneTasks))
				dl := atomic.LoadInt64(&totalProcessed)
				renderProgressBar(db, fileSize, dt, actualTasks, el, dl, exactLines, "Đang xử lý")
			case <-stopProgress:
				return
			}
		}
	}()

	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			readBuf := make([]byte, 16*1024*1024)
			writeBuf := make([]byte, 4*1024*1024)

			for r := range taskChan {
				fin, err := os.Open(inputPath)
				if err != nil {
					continue
				}

				fout, err := os.Create(tempPaths[r.Index])
				if err != nil {
					fin.Close()
					continue
				}

				_, _ = fin.Seek(r.Start, io.SeekStart)
				bytesToRead := r.End - r.Start
				var bytesRead int64
				var leftover []byte

				bw := bufio.NewWriterSize(fout, len(writeBuf))
				var taskValid int64
				var taskLines int64

				for bytesRead < bytesToRead {
					chunkLen := int64(len(readBuf))
					if bytesToRead-bytesRead < chunkLen {
						chunkLen = bytesToRead - bytesRead
					}
					n, err := fin.Read(readBuf[:chunkLen])
					if n > 0 {
						bytesRead += int64(n)
						data := readBuf[:n]
						if len(leftover) > 0 {
							data = append(leftover, data...)
							leftover = nil
						}

						lines := strings.Split(string(data), "\n")
						if bytesRead < bytesToRead {
							leftover = []byte(lines[len(lines)-1])
							lines = lines[:len(lines)-1]
						}

						for _, raw := range lines {
							line := strings.TrimSpace(raw)
							if len(line) == 0 || len(line) > 1000 {
								continue
							}
							taskLines++
							u, p, ok := transforms.ExtractCleanCredentials(line)
							if !ok {
								continue
							}
							cleanLine := u + ":" + p
							results := feat.Handler(cleanLine, arg)
							for _, res := range results {
								bw.WriteString(res + "\n")
								taskValid++
							}
						}
					}
					if err != nil {
						break
					}
				}

				if len(leftover) > 0 && len(leftover) <= 1000 {
					line := strings.TrimSpace(string(leftover))
					if len(line) > 0 {
						taskLines++
						if u, p, ok := transforms.ExtractCleanCredentials(line); ok {
							cleanLine := u + ":" + p
							results := feat.Handler(cleanLine, arg)
							for _, res := range results {
								bw.WriteString(res + "\n")
								taskValid++
							}
						}
					}
				}

				_ = bw.Flush()
				fout.Close()
				fin.Close()

				atomic.AddInt64(&doneBytes, bytesToRead)
				atomic.AddInt64(&doneTasks, 1)
				atomic.AddInt64(&totalValid, taskValid)
				atomic.AddInt64(&totalProcessed, taskLines)
			}
		}()
	}

	wg.Wait()
	ticker.Stop()
	close(stopProgress)

	el := time.Since(startTime).Seconds()
	renderProgressBar(fileSize, fileSize, actualTasks, actualTasks, el, exactLines, exactLines, "Đang xử lý")
	fmt.Println()

	_ = MergeTempFiles(tempPaths, outputPath, "")

	outMB := 0.0
	if fi, err := os.Stat(outputPath); err == nil {
		outMB = float64(fi.Size()) / (1024 * 1024)
	}

	border := colors.MangoOrange
	avgCPM := int64(0)
	if el > 0 {
		avgCPM = int64(float64(exactLines*60) / el)
	}
	speedMB := inMB / el

	fmt.Printf("\n  %s╔════════════════════════════════════════════════════════════════╗%s\n", border, colors.Reset)
	fmt.Printf("  %s║%s%s               XỬ LÝ DỮ LIỆU HOÀN TẤT THÀNH CÔNG !              %s%s║%s\n", border, colors.Bold, colors.White, colors.Reset, border, colors.Reset)
	fmt.Printf("  %s╠════════════════════════════════════════════════════════════════╣%s\n", border, colors.Reset)
	fmt.Printf("  %s║%s  %sFile đầu vào:%s     %.2f MB (%s dòng)\n", border, colors.Reset, colors.White, colors.Reset, inMB, formatLinesCount(exactLines, true))
	fmt.Printf("  %s║%s  %sFile kết quả:%s     %s%s%s%s (%.2f MB | %s dòng)\n", border, colors.Reset, colors.White, colors.Reset, colors.Bold, colors.CyanNeon, baseOut, colors.Reset, outMB, formatLinesCount(totalValid, true))
	fmt.Printf("  %s║%s  %sHuy động phần cứng:%s %s%s%d luồng CPU%s song song (Native Go Engine)\n", border, colors.Reset, colors.White, colors.Reset, colors.Bold, colors.MintGreen, numWorkers, colors.Reset)
	fmt.Printf("  %s║%s  %sTốc độ trung bình:%s %s%s%s CPM%s (%.2f MB/s)\n", border, colors.Reset, colors.White, colors.Reset, colors.Bold, colors.MangoGold, formatLinesCount(avgCPM, true), colors.Reset, speedMB)
	fmt.Printf("  %s║%s  %sThời gian chạy:%s   %s%.2f giây%s\n", border, colors.Reset, colors.White, colors.Reset, colors.MangoAmber, el, colors.Reset)
	fmt.Printf("  %s╚════════════════════════════════════════════════════════════════╝%s\n\n", border, colors.Reset)

	return nil
}

// ProcessDomainSplit xử lý chức năng 25: Tách đôi file theo Domain thành 2 file song song (Matched và Other)
func ProcessDomainSplit(inputPath, outputMatched, outputOther string, domainArg string) error {
	startTime := time.Now()
	fi, err := os.Stat(inputPath)
	if err != nil {
		return err
	}
	fileSize := fi.Size()
	if fileSize == 0 {
		return fmt.Errorf("file rỗng (0 bytes)")
	}

	numWorkers := runtime.NumCPU()
	if numWorkers > 4 {
		numWorkers = numWorkers - 1
	}

	targetChunkSize := int64(48 * 1024 * 1024)
	numTasks := int(fileSize / targetChunkSize)
	if numTasks < numWorkers {
		numTasks = numWorkers
	}
	if numTasks > numWorkers*16 {
		numTasks = numWorkers * 16
	}

	ranges, err := SplitByteRanges(inputPath, numTasks)
	if err != nil {
		return err
	}
	actualTasks := len(ranges)

	tempDir := filepath.Dir(outputMatched)
	baseMatch := filepath.Base(outputMatched)
	baseOther := filepath.Base(outputOther)
	CleanupStaleTempFiles(tempDir)

	inMB := float64(fileSize) / (1024 * 1024)
	fmt.Printf("  %s> Kích thước file: %s%s%.2f MB%s %s| Chia thành %d dải byte độc lập%s\n",
		colors.MangoAmber, colors.Bold, colors.TextWhite, inMB, colors.Reset, colors.TextMuted, actualTasks, colors.Reset)
	fmt.Printf("  %s> Khởi chạy %d Goroutines CPU song song (Tách kép: %s & %s)...%s\n",
		colors.MangoGold, numWorkers, baseMatch, baseOther, colors.Reset)

	exactLines := CountExactLines(inputPath)

	var doneBytes int64
	var doneTasks int64
	var totalValidMatch int64
	var totalValidOther int64
	var totalProcessed int64

	tempPathsMatch := make([]string, actualTasks)
	tempPathsOther := make([]string, actualTasks)
	for i := range ranges {
		tempPathsMatch[i] = filepath.Join(tempDir, fmt.Sprintf(".tmp_%s_part_%d.tmp", baseMatch, i))
		tempPathsOther[i] = filepath.Join(tempDir, fmt.Sprintf(".tmp_%s_part_%d.tmp", baseOther, i))
	}

	taskChan := make(chan ByteRange, actualTasks)
	for _, r := range ranges {
		taskChan <- r
	}
	close(taskChan)

	var wg sync.WaitGroup
	renderProgressBar(0, fileSize, 0, actualTasks, 0.001, 0, exactLines, "Tách domain")

	ticker := time.NewTicker(80 * time.Millisecond)
	stopProgress := make(chan struct{})

	go func() {
		for {
			select {
			case <-ticker.C:
				el := time.Since(startTime).Seconds()
				db := atomic.LoadInt64(&doneBytes)
				dt := int(atomic.LoadInt64(&doneTasks))
				dl := atomic.LoadInt64(&totalProcessed)
				renderProgressBar(db, fileSize, dt, actualTasks, el, dl, exactLines, "Tách domain")
			case <-stopProgress:
				return
			}
		}
	}()

	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			readBuf := make([]byte, 16*1024*1024)
			writeBuf := make([]byte, 4*1024*1024)

			for r := range taskChan {
				fin, err := os.Open(inputPath)
				if err != nil {
					continue
				}

				foutMatch, err := os.Create(tempPathsMatch[r.Index])
				if err != nil {
					fin.Close()
					continue
				}
				foutOther, err := os.Create(tempPathsOther[r.Index])
				if err != nil {
					foutMatch.Close()
					fin.Close()
					continue
				}

				_, _ = fin.Seek(r.Start, io.SeekStart)
				bytesToRead := r.End - r.Start
				var bytesRead int64
				var leftover []byte

				bwMatch := bufio.NewWriterSize(foutMatch, len(writeBuf))
				bwOther := bufio.NewWriterSize(foutOther, len(writeBuf))
				var taskMatch int64
				var taskOther int64
				var taskLines int64

				for bytesRead < bytesToRead {
					chunkLen := int64(len(readBuf))
					if bytesToRead-bytesRead < chunkLen {
						chunkLen = bytesToRead - bytesRead
					}
					n, err := fin.Read(readBuf[:chunkLen])
					if n > 0 {
						bytesRead += int64(n)
						data := readBuf[:n]
						if len(leftover) > 0 {
							data = append(leftover, data...)
							leftover = nil
						}

						lines := strings.Split(string(data), "\n")
						if bytesRead < bytesToRead {
							leftover = []byte(lines[len(lines)-1])
							lines = lines[:len(lines)-1]
						}

						for _, raw := range lines {
							line := strings.TrimSpace(raw)
							if len(line) == 0 || len(line) > 1000 {
								continue
							}
							taskLines++
							u, p, ok := transforms.ExtractCleanCredentials(line)
							if !ok || !transforms.IsValidEmail(u) || !transforms.IsCleanPassword(p) {
								continue
							}

							cleanLine := u + ":" + p
							if transforms.CheckDomainMatch(u, domainArg) {
								bwMatch.WriteString(cleanLine + "\n")
								taskMatch++
							} else {
								bwOther.WriteString(cleanLine + "\n")
								taskOther++
							}
						}
					}
					if err != nil {
						break
					}
				}

				if len(leftover) > 0 && len(leftover) <= 1000 {
					line := strings.TrimSpace(string(leftover))
					if len(line) > 0 {
						taskLines++
						if u, p, ok := transforms.ExtractCleanCredentials(line); ok {
							if transforms.IsValidEmail(u) && transforms.IsCleanPassword(p) {
								cleanLine := u + ":" + p
								if transforms.CheckDomainMatch(u, domainArg) {
									bwMatch.WriteString(cleanLine + "\n")
									taskMatch++
								} else {
									bwOther.WriteString(cleanLine + "\n")
									taskOther++
								}
							}
						}
					}
				}

				_ = bwMatch.Flush()
				_ = bwOther.Flush()
				foutMatch.Close()
				foutOther.Close()
				fin.Close()

				atomic.AddInt64(&doneBytes, bytesToRead)
				atomic.AddInt64(&doneTasks, 1)
				atomic.AddInt64(&totalValidMatch, taskMatch)
				atomic.AddInt64(&totalValidOther, taskOther)
				atomic.AddInt64(&totalProcessed, taskLines)
			}
		}()
	}

	wg.Wait()
	ticker.Stop()
	close(stopProgress)

	el := time.Since(startTime).Seconds()
	renderProgressBar(fileSize, fileSize, actualTasks, actualTasks, el, exactLines, exactLines, "Tách domain")
	fmt.Println()

	_ = MergeTempFiles(tempPathsMatch, outputMatched, filepath.Base(outputMatched))

	_ = MergeTempFiles(tempPathsOther, outputOther, filepath.Base(outputOther))

	outMBMatch := 0.0
	if fi, err := os.Stat(outputMatched); err == nil {
		outMBMatch = float64(fi.Size()) / (1024 * 1024)
	}
	outMBOther := 0.0
	if fi, err := os.Stat(outputOther); err == nil {
		outMBOther = float64(fi.Size()) / (1024 * 1024)
	}

	border := colors.MangoOrange
	avgCPM := int64(0)
	if el > 0 {
		avgCPM = int64(float64(exactLines*60) / el)
	}
	speedMB := inMB / el

	fmt.Printf("\n  %s╔════════════════════════════════════════════════════════════════╗%s\n", border, colors.Reset)
	fmt.Printf("  %s║%s%s         TÁCH DOMAIN & PHẦN CÒN LẠI HOÀN TẤT THÀNH CÔNG !       %s%s║%s\n", border, colors.Bold, colors.White, colors.Reset, border, colors.Reset)
	fmt.Printf("  %s╠════════════════════════════════════════════════════════════════╣%s\n", border, colors.Reset)
	fmt.Printf("  %s║%s  %sFile đầu vào:%s     %.2f MB (%s dòng)\n", border, colors.Reset, colors.White, colors.Reset, inMB, formatLinesCount(exactLines, true))
	fmt.Printf("  %s║%s  %sFile Domain:%s      %s%s%s%s (%.2f MB | %s dòng)\n", border, colors.Reset, colors.White, colors.Reset, colors.Bold, colors.CyanNeon, baseMatch, colors.Reset, outMBMatch, formatLinesCount(totalValidMatch, true))
	fmt.Printf("  %s║%s  %sFile Còn Lại:%s     %s%s%s%s (%.2f MB | %s dòng)\n", border, colors.Reset, colors.White, colors.Reset, colors.Bold, colors.MangoGold, baseOther, colors.Reset, outMBOther, formatLinesCount(totalValidOther, true))
	fmt.Printf("  %s║%s  %sHuy động phần cứng:%s %s%s%d luồng CPU%s song song (Native Go Engine)\n", border, colors.Reset, colors.White, colors.Reset, colors.Bold, colors.MintGreen, numWorkers, colors.Reset)
	fmt.Printf("  %s║%s  %sTốc độ trung bình:%s %s%s%s CPM%s (%.2f MB/s)\n", border, colors.Reset, colors.White, colors.Reset, colors.Bold, colors.MangoGold, formatLinesCount(avgCPM, true), colors.Reset, speedMB)
	fmt.Printf("  %s║%s  %sThời gian chạy:%s   %s%.2f giây%s\n", border, colors.Reset, colors.White, colors.Reset, colors.MangoAmber, el, colors.Reset)
	fmt.Printf("  %s╚════════════════════════════════════════════════════════════════╝%s\n\n", border, colors.Reset)

	return nil
}


