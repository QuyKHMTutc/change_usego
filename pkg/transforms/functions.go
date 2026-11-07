package transforms

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

var (
	nonAlnumRegex     = regexp.MustCompile(`[^a-zA-Z0-9]`)
	digitsRegex       = regexp.MustCompile(`\d`)
	nonDigitsRegex    = regexp.MustCompile(`\D`)
	nonLettersRegex   = regexp.MustCompile(`[^a-zA-Z]`)
	specialCharsRegex = regexp.MustCompile(`[!@$%&*+=,.'\-#_]`)
	specialABE        = regexp.MustCompile(`[!@$%&*+=,.'\#_:\-;]`)
	syllableRegex     = regexp.MustCompile(`(?i)([bcdfghjklmnpqrstvwxyz]*[aeiouy]+[bcdfghjklmnpqrstvwxyz]*)`)
	chunkRegex        = regexp.MustCompile(`[A-Za-z]+|\d+`)
	dao2nRegex        = regexp.MustCompile(`[!@$%&*+=,'\-#_]`)
	digitsOnlyRegex   = regexp.MustCompile(`\d+`)
)

func splitEmailPass(line string) (string, string, bool) {
	if u, p, ok := ExtractCleanCredentials(line); ok {
		return u, p, true
	}
	if !strings.Contains(line, ":") {
		return "", "", false
	}
	parts := strings.SplitN(line, ":", 2)
	u := strings.TrimSpace(parts[0])
	p := strings.TrimSpace(parts[1])
	if u == "" || p == "" {
		return "", "", false
	}
	return u, p, true
}

func extractDigits(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] >= '0' && s[i] <= '9' {
			b.WriteByte(s[i])
		}
	}
	return b.String()
}

func extractLetters(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		c := s[i]
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') {
			b.WriteByte(c)
		}
	}
	return b.String()
}

func extractNameAndNums(email, pw string) (string, string, bool) {
	uname := email
	if idx := strings.IndexByte(email, '@'); idx >= 0 {
		uname = email[:idx]
	}
	unameLetters := extractLetters(uname)
	if unameLetters == "" {
		return "", "", false
	}
	pwNums := extractDigits(pw)
	if pwNums == "" {
		return "", "", false
	}
	pwLetters := extractLetters(pw)
	var name string
	if pwLetters != "" && strings.EqualFold(pwLetters, unameLetters) {
		name = pwLetters
	} else {
		name = Capitalize(unameLetters)
	}
	return name, pwNums, true
}

// 0. FuncDedup lọc trùng dữ liệu giống EmEditor (giữ nguyên email:pass)
func FuncDedup(line string, _ string) []string {
	line = strings.TrimSpace(line)
	if line == "" {
		return nil
	}
	if u, p, ok := ExtractCleanCredentials(line); ok {
		return []string{u + ":" + p}
	}
	if strings.Contains(line, ":") {
		parts := strings.SplitN(line, ":", 2)
		u := strings.TrimSpace(parts[0])
		p := strings.TrimSpace(parts[1])
		if u != "" && p != "" {
			return []string{u + ":" + p}
		}
	}
	return []string{line}
}

// 1. func_103_beta
func Func103Beta(line string, _ string) []string {
	u, p, ok := splitEmailPass(line)
	if !ok {
		return nil
	}
	p = nonAlnumRegex.ReplaceAllString(p, "")
	if len(p) == 0 {
		return nil
	}
	return []string{u + ":" + Capitalize(p)}
}

// 2. func_2_word
func Func2Word(line string, _ string) []string {
	email, pw, ok := splitEmailPass(line)
	if !ok {
		return nil
	}
	var digitsSym strings.Builder
	for i := 0; i < len(pw); i++ {
		c := pw[i]
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')) {
			digitsSym.WriteByte(c)
		}
	}
	ds := digitsSym.String()
	uname := email
	if idx := strings.IndexByte(email, '@'); idx >= 0 {
		uname = email[:idx]
	}
	if len(ds) == 0 && extractLetters(uname) == uname {
		return nil
	}
	return []string{email + ":" + uname + ds}
}

// 3. func_2nx2
func Func2nx2(line string, _ string) []string {
	email, pw, ok := splitEmailPass(line)
	if !ok {
		return nil
	}
	pw = nonAlnumRegex.ReplaceAllString(pw, "")
	if len(pw) == 0 {
		return nil
	}
	pw = Capitalize(pw)
	nums := extractDigits(pw)
	lets := extractLetters(pw)
	if len(nums) == 0 {
		return nil
	}
	res := lets + nums
	if len(nums) > 1 {
		res += nums[:2]
	}
	return []string{email + ":" + res}
}

// 4. func_number_pass
func FuncNumberPass(line string, _ string) []string {
	email, pw, ok := splitEmailPass(line)
	if !ok {
		return nil
	}
	c := nonAlnumRegex.ReplaceAllString(pw, "")
	nums := extractDigits(c)
	lets := extractLetters(c)
	if len(nums) == 0 && len(lets) == 0 {
		return nil
	}
	if len(lets) > 0 {
		lets = Capitalize(lets)
	}
	return []string{email + ":" + nums + lets}
}

// 5. func_user_pass
func FuncUserPass(line string, _ string) []string {
	email, pw, ok := splitEmailPass(line)
	if !ok {
		return nil
	}
	pw = specialCharsRegex.ReplaceAllString(pw, "")
	uname := email
	if idx := strings.IndexByte(email, '@'); idx >= 0 {
		uname = email[:idx]
	}
	var chars strings.Builder
	for i := 0; i < len(uname); i++ {
		c := uname[i]
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') {
			chars.WriteByte(c)
		} else if c == '-' || c == '.' || c == '_' || (c >= '0' && c <= '9') {
			break
		}
	}
	pwNums := extractDigits(pw)
	chStr := chars.String()
	if len(chStr) == 0 && len(pwNums) == 0 {
		return nil
	}
	if len(chStr) > 0 {
		chStr = Capitalize(chStr)
	}
	return []string{email + ":" + chStr + pwNums}
}

// 6. func_abe_103r
func FuncAbe103r(line string, _ string) []string {
	email, pw, ok := splitEmailPass(line)
	if !ok {
		return nil
	}
	pw = specialABE.ReplaceAllString(pw, "")
	digits := extractDigits(pw)
	lets := extractLetters(pw)
	if len(digits) < 6 {
		return nil
	}
	res := lets + digits
	uname := email
	if idx := strings.IndexByte(email, '@'); idx >= 0 {
		uname = email[:idx]
	}
	if len(uname) > 0 {
		res += strings.ToUpper(string(uname[0]))
	}
	return []string{email + ":" + res}
}

// 7. func_agian_scoure
func FuncAgianScoure(line string, _ string) []string {
	uname, pw, ok := splitEmailPass(line)
	if !ok {
		return nil
	}
	pwC := specialCharsRegex.ReplaceAllString(pw, "")
	nums := extractDigits(pwC)
	if len(nums) < 6 {
		return nil
	}
	u := uname
	if idx := strings.IndexByte(uname, '@'); idx >= 0 {
		u = uname[:idx]
	}
	for _, sep := range []string{".", "-", "_"} {
		if idx := strings.Index(u, sep); idx >= 0 {
			u = u[:idx]
		}
	}
	if len(u) < 2 {
		return nil
	}
	rev2 := string([]byte{u[1], u[0]})
	return []string{uname + ":" + Capitalize(strings.ToLower(rev2)) + nums}
}

// 8. func_dao_2n
func FuncDao2n(line string, _ string) []string {
	uname, pw, ok := splitEmailPass(line)
	if !ok {
		return nil
	}
	pw = dao2nRegex.ReplaceAllString(pw, "")
	digits := extractDigits(pw)
	if len(digits) == 0 {
		return nil
	}
	revDigits := make([]byte, len(digits))
	for i := 0; i < len(digits); i++ {
		revDigits[i] = digits[len(digits)-1-i]
	}
	var newPw strings.Builder
	dIdx := 0
	for i := 0; i < len(pw); i++ {
		c := pw[i]
		if c >= '0' && c <= '9' {
			newPw.WriteByte(revDigits[dIdx])
			dIdx++
		} else {
			newPw.WriteByte(c)
		}
	}
	res := newPw.String()
	if len(res) > 0 {
		res = Capitalize(res)
	}
	return []string{uname + ":" + res}
}

// 9. func_lap_so_2
func FuncLapSo2(line string, _ string) []string {
	email, pw, ok := splitEmailPass(line)
	if !ok {
		return nil
	}
	nums := extractDigits(pw)
	noNums := digitsRegex.ReplaceAllString(pw, "")
	if len(nums) > 0 {
		if len(nums) <= 2 {
			return []string{email + ":" + noNums + nums + nums}
		}
		return []string{email + ":" + noNums + nums + nums[:2]}
	}
	ne := extractDigits(email)
	if len(ne) > 0 {
		if len(ne) <= 2 {
			return []string{email + ":" + noNums + ne + ne}
		}
		return []string{email + ":" + noNums + ne + ne[:2]}
	}
	return nil
}

// 10. func_3e
func Func3e(line string, _ string) []string {
	uname, pw, ok := splitEmailPass(line)
	if !ok {
		return nil
	}
	pw = specialABE.ReplaceAllString(pw, "")
	if strings.Contains(strings.ToLower(uname), "hotmail") {
		pw = extractDigits(pw)
	}
	if idx := strings.IndexByte(uname, '@'); idx >= 0 {
		uPrefix := uname[:idx]
		if len(uPrefix) > 2 {
			uPrefix = uPrefix[:2]
		}
		pw = strings.ToUpper(uPrefix) + pw
	}
	if len(extractDigits(pw)) < 4 {
		return nil
	}
	return []string{uname + ":" + strings.ToUpper(pw)}
}

// 11. func_13_email_nums_to_pw_end
func Func13EmailNumsToPwEnd(line string, mode string) []string {
	email, pw, ok := splitEmailPass(line)
	if !ok {
		return nil
	}
	uname := email
	if idx := strings.IndexByte(email, '@'); idx >= 0 {
		uname = email[:idx]
	}
	digits := extractDigits(uname)
	if len(digits) == 0 {
		return nil
	}
	vAll := email + ":" + pw + digits
	last2 := digits
	if len(digits) >= 2 {
		last2 = digits[len(digits)-2:]
	}
	vLast2 := email + ":" + pw + last2

	modeStr := strings.ToLower(strings.TrimSpace(mode))
	if modeStr == "all" || modeStr == "tatca" {
		return []string{vAll}
	}
	if modeStr == "2" || modeStr == "last2" || modeStr == "2so" {
		return []string{vLast2}
	}
	if len(digits) > 2 {
		return []string{vAll, vLast2}
	}
	return []string{vAll}
}

// 12. func_14_char_replacement
func Func14CharReplacement(line string, charArg string) []string {
	email, pw, ok := splitEmailPass(line)
	if !ok || len(pw) == 0 {
		return nil
	}
	arg := strings.ToLower(strings.TrimSpace(charArg))
	if arg == "" {
		arg = "a"
	}
	var dst string
	var srcList []string
	switch arg {
	case "a":
		dst = "@"
		srcList = []string{"a", "A"}
	case "i":
		dst = "!"
		srcList = []string{"i", "I"}
	case "s":
		dst = "$"
		srcList = []string{"s", "S"}
	default:
		return nil
	}

	newPw := pw
	for _, s := range srcList {
		newPw = strings.ReplaceAll(newPw, s, dst)
	}
	if newPw == pw {
		return nil // Loại bỏ dòng nếu không chứa chữ cần mã hóa
	}
	return []string{email + ":" + newPw}
}

// 13. func_viet_hoa_sau_so_cuoi
func FuncVietHoaSauSoCuoi(line string, _ string) []string {
	email, pw, ok := splitEmailPass(line)
	if !ok {
		return nil
	}
	lastDigitIdx := -1
	for i := len(pw) - 1; i >= 0; i-- {
		if pw[i] >= '0' && pw[i] <= '9' {
			lastDigitIdx = i
			break
		}
	}
	if lastDigitIdx == -1 {
		return nil
	}
	afterDigit := pw[lastDigitIdx+1:]
	if len(afterDigit) == 0 || extractLetters(afterDigit) == "" {
		return nil
	}
	newPw := pw[:lastDigitIdx+1] + strings.ToUpper(afterDigit)
	return []string{email + ":" + newPw}
}

// 14. func_16_email_n_chars
func Func16EmailNChars(line string, nArg string) []string {
	email, pw, ok := splitEmailPass(line)
	if !ok {
		return nil
	}
	n := 2
	if len(nArg) > 0 {
		if v, err := strconv.Atoi(strings.TrimSpace(nArg)); err == nil && v > 0 {
			n = v
		}
	}
	uname := email
	if idx := strings.IndexByte(email, '@'); idx >= 0 {
		uname = email[:idx]
	}
	unameLetters := extractLetters(uname)
	if len(unameLetters) < n {
		return nil // Loại bỏ dòng nếu số chữ cái ít hơn N
	}
	pwNums := extractDigits(pw)
	if len(pwNums) == 0 {
		return nil
	}
	prefix := Capitalize(unameLetters[:n])
	return []string{email + ":" + prefix + pwNums}
}

// 15. func_17_special_before_nums
func Func17SpecialBeforeNums(line string, symbol string) []string {
	email, pw, ok := splitEmailPass(line)
	if !ok {
		return nil
	}
	name, pwNums, ok := extractNameAndNums(email, pw)
	if !ok {
		return nil
	}
	sym := symbol
	if sym == "" {
		sym = "@"
	}
	return []string{email + ":" + name + sym + pwNums}
}

// 16. func_18_special_after_nums
func Func18SpecialAfterNums(line string, symbol string) []string {
	email, pw, ok := splitEmailPass(line)
	if !ok {
		return nil
	}
	name, pwNums, ok := extractNameAndNums(email, pw)
	if !ok {
		return nil
	}
	sym := symbol
	if sym == "" {
		sym = "@"
	}
	return []string{email + ":" + name + pwNums + sym}
}

// 17. func_22_name_and_nums
func Func22NameAndNums(line string, _ string) []string {
	email, pw, ok := splitEmailPass(line)
	if !ok {
		return nil
	}
	uname := email
	if idx := strings.IndexByte(email, '@'); idx >= 0 {
		uname = email[:idx]
	}
	name := extractLetters(uname)
	if name == "" {
		return nil
	}
	name = Capitalize(name)
	pwNums := extractDigits(pw)
	return []string{email + ":" + name + pwNums}
}

// 18. func_25_nums_and_email_name
func Func25NumsAndEmailName(line string, _ string) []string {
	email, pw, ok := splitEmailPass(line)
	if !ok {
		return nil
	}
	uname := email
	if idx := strings.IndexByte(email, '@'); idx >= 0 {
		uname = email[:idx]
	}
	name := extractLetters(uname)
	if name == "" {
		return nil
	}
	pwNums := extractDigits(pw)
	if pwNums == "" {
		return nil
	}
	return []string{email + ":" + pwNums + name}
}

// 19. func_26_duplicate_last_char
func Func26DuplicateLastChar(line string, _ string) []string {
	email, pw, ok := splitEmailPass(line)
	if !ok || len(pw) == 0 {
		return nil
	}
	lastChar := pw[len(pw)-1:]
	return []string{email + ":" + pw + lastChar}
}

// 20. func_27_custom_prefix
func Func27CustomPrefix(line string, prefix string) []string {
	email, pw, ok := splitEmailPass(line)
	if !ok {
		return nil
	}
	pwNums := extractDigits(pw)
	if len(pwNums) == 0 {
		return nil
	}
	newPw := prefix + pwNums
	if len(newPw) < 6 {
		return nil
	}
	return []string{email + ":" + newPw}
}

// 21. func_23_capitalize_or_lower_first_letter
func Func23CapitalizeOrLowerFirstLetter(line string, mode string) []string {
	email, pw, ok := splitEmailPass(line)
	if !ok || len(pw) == 0 {
		return nil
	}
	modeStr := strings.ToUpper(strings.TrimSpace(mode))
	isUpper := (modeStr != "T")

	firstLetterIdx := -1
	for i := 0; i < len(pw); i++ {
		c := pw[i]
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') {
			firstLetterIdx = i
			break
		}
	}
	if firstLetterIdx == -1 {
		return []string{email + ":" + pw}
	}
	var newChar byte
	if isUpper {
		newChar = byte(unicode.ToUpper(rune(pw[firstLetterIdx])))
	} else {
		newChar = byte(unicode.ToLower(rune(pw[firstLetterIdx])))
	}
	newPw := pw[:firstLetterIdx] + string(newChar) + pw[firstLetterIdx+1:]
	return []string{email + ":" + newPw}
}

// 22. func_24_filter_pass_min_6
func Func24FilterPassMin6(line string, _ string) []string {
	email, pw, ok := splitEmailPass(line)
	if !ok || len(pw) < 6 {
		return nil
	}
	return []string{email + ":" + pw}
}

// 23. func_24_remove_special_chars
func Func24RemoveSpecialChars(line string, _ string) []string {
	email, pw, ok := splitEmailPass(line)
	if !ok {
		return nil
	}
	newPw := nonAlnumRegex.ReplaceAllString(pw, "")
	if len(newPw) == 0 {
		return nil
	}
	return []string{email + ":" + newPw}
}

// 24. func_25_keep_only_letters
func Func25KeepOnlyLetters(line string, _ string) []string {
	email, pw, ok := splitEmailPass(line)
	if !ok {
		return nil
	}
	letters := nonLettersRegex.ReplaceAllString(pw, "")
	if len(letters) < 6 {
		return nil
	}
	return []string{email + ":" + letters}
}

// 25. func_domain_filter
func FuncDomainFilter(line string, domainArg string) []string {
	u, p, ok := ExtractCleanCredentials(line)
	if !ok || !IsValidEmail(u) || !IsCleanPassword(p) {
		return nil
	}
	if CheckDomainMatch(u, domainArg) {
		return []string{u + ":" + p}
	}
	return nil
}

// 26. func_26_keep_only_numbers
func Func26KeepOnlyNumbers(line string, _ string) []string {
	email, pw, ok := splitEmailPass(line)
	if !ok {
		return nil
	}
	nums := nonDigitsRegex.ReplaceAllString(pw, "")
	if len(nums) < 6 {
		return nil
	}
	return []string{email + ":" + nums}
}

// ParseNAndPos cho chức năng 27
func ParseNAndPos(arg string) (int, string) {
	s := strings.ToUpper(strings.TrimSpace(arg))
	pos := "D"
	if strings.Contains(s, "C") {
		pos = "C"
	}
	m := digitsOnlyRegex.FindString(s)
	n := 3
	if len(m) > 0 {
		if v, err := strconv.Atoi(m); err == nil && v > 0 {
			n = v
		}
	}
	return n, pos
}

// 27. func_27_rcr_email_n_chars_pos
func Func27RcrEmailNCharsPos(line string, arg string) []string {
	email, pw, ok := splitEmailPass(line)
	if !ok {
		return nil
	}
	n, pos := ParseNAndPos(arg)
	uname := email
	if idx := strings.IndexByte(email, '@'); idx >= 0 {
		uname = email[:idx]
	}
	unameLetters := extractLetters(uname)
	var userChars string
	if len(unameLetters) >= n {
		userChars = Capitalize(unameLetters[:n])
	} else if len(uname) >= n {
		userChars = Capitalize(uname[:n])
	} else {
		return nil
	}
	pwNums := extractDigits(pw)
	if len(pwNums) < 6 {
		return nil
	}
	var basePw string
	if pos == "D" {
		basePw = userChars + pwNums
	} else {
		basePw = pwNums + userChars
	}
	vars := []string{
		email + ":" + basePw,
		email + ":" + basePw + "!",
		email + ":" + basePw + "@",
		email + ":" + basePw + "$",
		email + ":" + basePw + "*",
		email + ":" + basePw + "#",
	}
	if len(basePw) >= 2 {
		vars = append(vars, email + ":" + basePw[:2] + "@" + basePw[2:])
	}
	return vars
}

// 28. func_28_get_email_name
func Func28GetEmailName(line string, _ string) []string {
	email, _, ok := splitEmailPass(line)
	if !ok {
		return nil
	}
	uname := email
	if idx := strings.IndexByte(email, '@'); idx >= 0 {
		uname = email[:idx]
	}
	letters := extractLetters(uname)
	if len(letters) >= 6 && len(letters) <= 15 {
		return []string{email + ":" + Capitalize(letters)}
	}
	return nil
}

// 29. func_29_important_syllables
func ExtractImportantSyllables(word string) string {
	w := strings.ToLower(word)
	if len(w) <= 4 {
		return w
	}
	matches := syllableRegex.FindAllString(w, -1)
	if len(matches) > 0 {
		var selected []string
		if len(matches) > 2 {
			selected = matches[:2]
		} else {
			selected = matches
		}
		var shortened strings.Builder
		for _, s := range selected {
			if len(s) > 3 {
				shortened.WriteString(s[:3])
			} else {
				shortened.WriteString(s)
			}
		}
		res := shortened.String()
		if len(res) > 0 {
			return res
		}
		return w[:4]
	}
	return w[:4]
}

// 29. Func29InsertCharPos chèn ký tự đặc biệt ở vị trí bất kỳ, đầu hoặc cuối trong mật khẩu
// Hỗ trợ:
// - Chèn ở CUỐI: 29 C,, 29 C@, 29 C!, 29 ,, 29 @... (password -> password, hoặc password@)
// - Chèn ở ĐẦU: 29 D,, 29 D@, 29 D!... (password -> ,password)
// - Chèn theo VỊ TRÍ: 29 2@, 29 3$, 29 2,... (password -> p@assword)
func Func29InsertCharPos(line string, arg string) []string {
	u, p, ok := splitEmailPass(line)
	if !ok || len(p) == 0 {
		return nil
	}

	arg = strings.TrimSpace(arg)
	if arg == "" {
		arg = "C@"
	}

	upperArg := strings.ToUpper(arg)

	// Trường hợp 1: Chèn vào CUỐI rõ ràng bằng C hoặc c (ví dụ: "C,", "C@", "C!", "C")
	if strings.HasPrefix(upperArg, "C") {
		insertChar := strings.TrimSpace(arg[1:])
		if insertChar == "" {
			insertChar = "@"
		}
		return []string{u + ":" + p + insertChar}
	}

	// Trường hợp 2: Chèn vào ĐẦU rõ ràng bằng D hoặc d (ví dụ: "D,", "D@", "D")
	if strings.HasPrefix(upperArg, "D") {
		insertChar := strings.TrimSpace(arg[1:])
		if insertChar == "" {
			insertChar = "@"
		}
		return []string{u + ":" + insertChar + p}
	}

	// Trường hợp 3: Kiểm tra xem có số vị trí ở đầu không (ví dụ: "2@", "3$", "2,")
	i := 0
	for i < len(arg) && arg[i] >= '0' && arg[i] <= '9' {
		i++
	}

	// Nếu KHÔNG có số vị trí ở đầu (ví dụ người dùng chỉ gõ ",", "@", "#", "!" hoặc "29 ,", "29,")
	// thì mặc định hiểu là người dùng muốn thêm ký tự đó vào CUỐI mật khẩu
	if i == 0 {
		insertChar := arg
		if insertChar == "" {
			insertChar = "@"
		}
		return []string{u + ":" + p + insertChar}
	}

	// Nếu CÓ số vị trí cụ thể (ví dụ 2@, 3$, 2,)
	pos := 2
	if parsed, err := strconv.Atoi(arg[:i]); err == nil && parsed > 0 {
		pos = parsed
	}

	insertChar := strings.TrimSpace(arg[i:])
	if insertChar == "" {
		insertChar = "@"
	}

	// 1-based indexing: vị trí 2 -> index 1 (chèn sau ký tự thứ 1)
	idx := pos - 1
	if idx > len(p) {
		return nil // Không đủ ký tự trong mật khẩu để chèn
	}
	if idx < 0 {
		idx = 0
	}

	newPw := p[:idx] + insertChar + p[idx:]
	return []string{u + ":" + newPw}
}

// TransformsMap ánh xạ tên hàm sang TransformFunc

// 30. Func30PassLast2DigitsToYear lấy 2 số cuối mật khẩu ghép thành năm:
// - 50 đến 99 -> ghép thành 1950 - 1999 (ví dụ: password7679 -> password1979)
// - 01 đến 26 -> ghép thành 2001 - 2026 (ví dụ: password724401 -> password2001)
// - Ngoài phạm vi này hoặc mật khẩu không có số -> loại bỏ dòng
func Func30PassLast2DigitsToYear(line string, _ string) []string {
	u, p, ok := splitEmailPass(line)
	if !ok || len(p) == 0 {
		return nil
	}

	digits := extractDigits(p)
	if len(digits) < 2 {
		return nil
	}

	last2 := digits[len(digits)-2:]
	val, err := strconv.Atoi(last2)
	if err != nil {
		return nil
	}

	var year string
	if val >= 50 && val <= 99 {
		year = "19" + last2
	} else if val >= 1 && val <= 26 {
		year = fmt.Sprintf("20%02d", val)
	} else {
		return nil
	}

	textPart := strings.TrimRight(p, "0123456789")
	if textPart == "" {
		return nil
	}

	newPw := textPart + year
	if len(newPw) < 6 {
		return nil
	}

	return []string{u + ":" + newPw}
}

// 31. Func31DeleteCharPos xóa ký tự thứ N của mật khẩu:
// - 31 2: xóa ký tự thứ 2 (1-based: index 1). Ví dụ: password -> pssword
// - 31 1: xóa ký tự thứ 1 (1-based: index 0). Ví dụ: password -> assword
// - 31 C / 31 c: xóa ký tự cuối cùng (index len(p)-1). Ví dụ: password -> passwor
func Func31DeleteCharPos(line string, arg string) []string {
	u, p, ok := splitEmailPass(line)
	if !ok || len(p) == 0 {
		return nil
	}

	arg = strings.TrimSpace(arg)
	if arg == "" {
		arg = "C"
	}

	var newPw string
	if strings.EqualFold(arg, "C") || strings.EqualFold(arg, "CUOI") {
		// Xóa ký tự cuối cùng
		if len(p) <= 1 {
			return nil
		}
		newPw = p[:len(p)-1]
	} else {
		// Xóa ký tự thứ N (1-based index)
		n, err := strconv.Atoi(arg)
		if err != nil || n <= 0 || n > len(p) {
			return nil
		}
		idx := n - 1
		newPw = p[:idx] + p[idx+1:]
	}

	if len(newPw) == 0 {
		return nil
	}

	return []string{u + ":" + newPw}
}

// 32. Func32ExtractNLetters lấy N chữ cái mật khẩu (chỉ lấy mỗi chữ, không lọc < 6 ký tự):
// - Ví dụ: password123 với N=4 -> pass
// - Ví dụ: p@ssw0rd với N=2 -> ps
// - Không lọc điều kiện < 6 ký tự, giữ nguyên chữ cái hoa/thường ban đầu
func Func32ExtractNLetters(line string, arg string) []string {
	email, pw, ok := splitEmailPass(line)
	if !ok || len(pw) == 0 {
		return nil
	}

	n := 4
	arg = strings.TrimSpace(arg)
	if arg != "" {
		if v, err := strconv.Atoi(arg); err == nil && v > 0 {
		n = v
		}
	}

	letters := nonLettersRegex.ReplaceAllString(pw, "")
	if len(letters) < n {
		return nil // Loại bỏ dòng nếu tổng số chữ cái ít hơn N
	}

	newPw := Capitalize(letters[:n])
	return []string{email + ":" + newPw}
}

// 33. Func33EmailNameParts lấy đầu và cuối tên Email trước và sau ký tự đặc biệt
// - Ví dụ: dorina_ranga@hotmail.com -> Dorina và Ranga (email:Dorina và email:Ranga)
// - Gõ 33 hoặc không có tham số: Lấy cả phần trước và phần sau (> 3 ký tự), gộp vào 1 file
// - Gõ 33 D: Chỉ lấy phần trước ký tự đặc biệt (> 3 ký tự)
// - Gõ 33 C: Chỉ lấy phần sau ký tự đặc biệt (> 3 ký tự)
// - Nếu tên email không có ký tự đặc biệt phân cách: Lấy 6 ký tự đầu của username (viết hoa chữ đầu), đảm bảo > 3 ký tự
// - Tất cả kết quả đều viết hoa chữ cái đầu tiên, các chữ còn lại viết thường
func Func33EmailNameParts(line string, arg string) []string {
	email, _, ok := splitEmailPass(line)
	if !ok {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "@") {
			email = line
			if idx := strings.IndexByte(email, ':'); idx != -1 {
				email = email[:idx]
			}
		} else {
			return nil
		}
	}

	atIdx := strings.IndexByte(email, '@')
	if atIdx <= 0 {
		return nil
	}
	uname := email[:atIdx]

	// Kiểm tra xem username có ký tự đặc biệt phân cách hay không
	hasSpecial := false
	for _, r := range uname {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
		hasSpecial = true
			break
		}
	}

	argUpper := strings.ToUpper(strings.TrimSpace(arg))
	var results []string

	if hasSpecial {
		// Tách username bởi tất cả các ký tự đặc biệt (hỗ trợ đầu, giữa, cuối)
		rawParts := strings.FieldsFunc(uname, func(r rune) bool {
			return !unicode.IsLetter(r) && !unicode.IsDigit(r)
		})

		var validParts []string
		for _, rp := range rawParts {
			letters := extractLetters(rp)
			// Lấy >= 3 ký tự
			if len(letters) >= 3 {
				validParts = append(validParts, Capitalize(strings.ToLower(letters)))
			}
		}

		if len(validParts) == 0 {
			return nil
		}

		var targetParts []string
		if argUpper == "D" {
			targetParts = append(targetParts, validParts[0])
		} else if argUpper == "C" {
			targetParts = append(targetParts, validParts[len(validParts)-1])
		} else {
			targetParts = validParts
		}

		seen := make(map[string]bool)
		for _, p := range targetParts {
			if !seen[p] {
				seen[p] = true
				results = append(results, email+":"+p)
			}
		}
	} else {
		// Tên email không có ký tự đặc biệt: lấy 6 ký tự đầu
		letters := extractLetters(uname)
		if len(letters) >= 6 {
			letters = letters[:6]
		}
		// Đảm bảo lấy >= 3 ký tự
		if len(letters) >= 3 {
			name := Capitalize(strings.ToLower(letters))
			results = append(results, email+":"+name)
		}
	}

	if len(results) == 0 {
		return nil
	}
	return results
}

// 34. Func34EmailNamePartsAndPwNums lấy các phần tên Email (đầu, giữa, cuối) ghép thêm số mật khẩu cũ
// - Ví dụ: world.traveler_carolyn@hotmail.com:Pass123 -> World123, Traveler123, Carolyn123
// - Gõ 34 hoặc không có tham số: Lấy tất cả các phần (đầu, giữa, cuối >= 3 ký tự) ghép số mật khẩu cũ, gộp vào 1 file
// - Gõ 34 D: Chỉ lấy phần đầu tiên (>= 3 ký tự) ghép số mật khẩu cũ
// - Gõ 34 C: Chỉ lấy phần cuối cùng (>= 3 ký tự) ghép số mật khẩu cũ
// - Nếu tên email không có ký tự đặc biệt phân cách: Lấy 6 ký tự đầu của username (viết hoa chữ đầu) ghép số mật khẩu cũ, đảm bảo >= 3 ký tự
// - Tất cả kết quả đều viết hoa chữ cái đầu tiên của phần tên, các chữ còn lại viết thường
func Func34EmailNamePartsAndPwNums(line string, arg string) []string {
	email, pw, ok := splitEmailPass(line)
	if !ok {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "@") {
			parts := strings.SplitN(line, ":", 2)
			email = parts[0]
			if len(parts) > 1 {
				pw = parts[1]
			}
		} else {
			return nil
		}
	}

	atIdx := strings.IndexByte(email, '@')
	if atIdx <= 0 {
		return nil
	}
	uname := email[:atIdx]

	// Trích xuất toàn bộ chữ số trong mật khẩu cũ
	pwNums := extractDigits(pw)
	if len(pwNums) == 0 {
		return nil // Mật khẩu cũ không có số thì loại bỏ dòng
	}

	// Kiểm tra xem username có ký tự đặc biệt phân cách hay không
	hasSpecial := false
	for _, r := range uname {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			hasSpecial = true
			break
		}
	}

	argUpper := strings.ToUpper(strings.TrimSpace(arg))
	var results []string

	if hasSpecial {
		// Tách username bởi tất cả các ký tự đặc biệt (hỗ trợ đầu, giữa, cuối)
		rawParts := strings.FieldsFunc(uname, func(r rune) bool {
			return !unicode.IsLetter(r) && !unicode.IsDigit(r)
		})

		var validParts []string
		for _, rp := range rawParts {
			letters := extractLetters(rp)
			// Lấy >= 3 ký tự
			if len(letters) >= 3 {
				validParts = append(validParts, Capitalize(strings.ToLower(letters)))
			}
		}

		if len(validParts) == 0 {
			return nil
		}

		var targetParts []string
		if argUpper == "D" {
			targetParts = append(targetParts, validParts[0])
		} else if argUpper == "C" {
			targetParts = append(targetParts, validParts[len(validParts)-1])
		} else {
			targetParts = validParts
		}

		seen := make(map[string]bool)
		for _, p := range targetParts {
			newPw := p + pwNums
			if !seen[newPw] {
				seen[newPw] = true
				results = append(results, email+":"+newPw)
			}
		}
	} else {
		// Tên email không có ký tự đặc biệt: lấy 6 ký tự đầu
		letters := extractLetters(uname)
		if len(letters) >= 6 {
			letters = letters[:6]
		}
		// Đảm bảo lấy >= 3 ký tự
		if len(letters) >= 3 {
			name := Capitalize(strings.ToLower(letters))
			results = append(results, email+":"+name+pwNums)
		}
	}

	if len(results) == 0 {
		return nil
	}
	return results
}

// 35. Func35SplitLastDigitsAndRemaining tách số cuối và phần số còn lại trong mật khẩu:
// Ví dụ: nitathomas27@hotmail.com:Blessed2018 -> Blessed18 và Blessed20
// Ví dụ: marco_noboa@hotmail.com:Quchufos19742 -> Quchufos42 và Quchufos197
func Func35SplitLastDigitsAndRemaining(line string, arg string) []string {
	u, p, ok := splitEmailPass(line)
	if !ok || len(p) == 0 {
		return nil
	}

	numCut := 2
	arg = strings.TrimSpace(arg)
	if arg != "" {
		if v, err := strconv.Atoi(arg); err == nil && v > 0 {
			numCut = v
		}
	}

	// Tìm vị trí của khối số cuối cùng trong mật khẩu
	lastDigit := -1
	for i := len(p) - 1; i >= 0; i-- {
		if p[i] >= '0' && p[i] <= '9' {
			lastDigit = i
			break
		}
	}
	if lastDigit == -1 {
		return nil // Không có chữ số nào trong mật khẩu
	}

	firstDigit := lastDigit
	for firstDigit >= 0 && p[firstDigit] >= '0' && p[firstDigit] <= '9' {
		firstDigit--
	}
	firstDigit++

	digits := p[firstDigit : lastDigit+1]
	if len(digits) <= numCut {
		return nil // Cần nhiều chữ số hơn numCut để có cả số cuối và phần còn lại
	}

	prefix := p[:firstDigit]
	suffix := p[lastDigit+1:]

	lastN := digits[len(digits)-numCut:]
	remaining := digits[:len(digits)-numCut]

	pw1 := prefix + lastN + suffix
	pw2 := prefix + remaining + suffix

	var results []string
	if len(pw1) >= 6 {
		results = append(results, u+":"+pw1)
	}
	if len(pw2) >= 6 && pw2 != pw1 {
		results = append(results, u+":"+pw2)
	}

	if len(results) == 0 {
		return nil
	}

	return results
}

var TransformsMap = map[string]TransformFunc{
	"func_dedup":                               FuncDedup,
	"func_103_beta":                            Func103Beta,
	"func_2_word":                              Func2Word,
	"func_2nx2":                                Func2nx2,
	"func_number_pass":                         FuncNumberPass,
	"func_user_pass":                           FuncUserPass,
	"func_abe_103r":                            FuncAbe103r,
	"func_agian_scoure":                        FuncAgianScoure,
	"func_dao_2n":                              FuncDao2n,
	"func_lap_so_2":                            FuncLapSo2,
	"func_3e":                                  Func3e,
	"func_13_email_nums_to_pw_end":             Func13EmailNumsToPwEnd,
	"func_14_char_replacement":                 Func14CharReplacement,
	"func_viet_hoa_sau_so_cuoi":                FuncVietHoaSauSoCuoi,
	"func_16_email_n_chars":                    Func16EmailNChars,
	"func_17_special_before_nums":              Func17SpecialBeforeNums,
	"func_18_special_after_nums":               Func18SpecialAfterNums,
	"func_22_name_and_nums":                    Func22NameAndNums,
	"func_25_nums_and_email_name":              Func25NumsAndEmailName,
	"func_26_duplicate_last_char":              Func26DuplicateLastChar,
	"func_27_custom_prefix":                    Func27CustomPrefix,
	"func_23_capitalize_or_lower_first_letter": Func23CapitalizeOrLowerFirstLetter,
	"func_24_filter_pass_min_6":                Func24FilterPassMin6,
	"func_24_remove_special_chars":             Func24RemoveSpecialChars,
	"func_25_keep_only_letters":                Func25KeepOnlyLetters,
	"func_domain_filter":                       FuncDomainFilter,
	"func_26_keep_only_numbers":                Func26KeepOnlyNumbers,
	"func_27_rcr_email_n_chars_pos":            Func27RcrEmailNCharsPos,
	"func_28_get_email_name":                   Func28GetEmailName,
	"func_29_insert_special_char_pos":          Func29InsertCharPos,
	"func_30_pass_last_2_digits_to_year":      Func30PassLast2DigitsToYear,
	"func_31_delete_char_pos":                 Func31DeleteCharPos,
	"func_32_extract_n_letters":               Func32ExtractNLetters,
	"func_33_email_name_parts":                Func33EmailNameParts,
	"func_34_email_name_parts_and_pw_nums":   Func34EmailNamePartsAndPwNums,
	"func_35_split_last_digits_and_remaining": Func35SplitLastDigitsAndRemaining,
}

