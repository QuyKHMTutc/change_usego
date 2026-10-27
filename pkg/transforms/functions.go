package transforms

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

var (
	nonAlnumRegex = regexp.MustCompile(`[^a-zA-Z0-9]`)
	digitsRegex   = regexp.MustCompile(`\d`)
	nonDigitsRegex = regexp.MustCompile(`\D`)
	nonLettersRegex = regexp.MustCompile(`[^a-zA-Z]`)
	specialCharsRegex = regexp.MustCompile(`[!@$%&*+=,.'\-#_]`)
	specialABE = regexp.MustCompile(`[!@$%&*+=,.'\#_:\-;]`)
	syllableRegex = regexp.MustCompile(`(?i)([bcdfghjklmnpqrstvwxyz]*[aeiouy]+[bcdfghjklmnpqrstvwxyz]*)`)
	chunkRegex = regexp.MustCompile(`[A-Za-z]+|\d+`)
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
	pw = regexp.MustCompile(`[!@$%&*+=,'\-#_]`).ReplaceAllString(pw, "")
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
	re := regexp.MustCompile(`\d+`)
	m := re.FindString(s)
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

// 29. Func29InsertCharPos chèn ký tự đặc biệt ở vị trí bất kỳ trong mật khẩu
// Ví dụ: arg = "2@" -> chèn ở vị trí 2 (sau chữ 1): password -> p@assword
// Ví dụ: arg = "3$" -> chèn ở vị trí 3 (sau chữ 2): password -> pa$ssword
func Func29InsertCharPos(line string, arg string) []string {
	u, p, ok := splitEmailPass(line)
	if !ok || len(p) == 0 {
		return nil
	}

	arg = strings.TrimSpace(arg)
	if arg == "" {
		arg = "2@"
	}

	// Tách số vị trí ở đầu và ký tự đặc biệt
	i := 0
	for i < len(arg) && arg[i] >= '0' && arg[i] <= '9' {
		i++
	}

	pos := 2
	if i > 0 {
		if parsed, err := strconv.Atoi(arg[:i]); err == nil && parsed > 0 {
			pos = parsed
		}
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

var TransformsMap = map[string]TransformFunc{
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
}

