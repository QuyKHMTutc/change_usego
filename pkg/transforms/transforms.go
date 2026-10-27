package transforms

import (
	"strings"
	"unicode"
)

// TransformFunc đại diện cho hàm xử lý dòng dữ liệu.
// Trả về slice rỗng hoặc nil nếu dòng bị loại bỏ.
type TransformFunc func(line string, arg string) []string

// IsValidEmail kiểm tra định dạng email hợp lệ (ASCII sạch, có @, domain có dấu chấm, không chứa control chars)
func IsValidEmail(u string) bool {
	if len(u) < 5 || len(u) > 128 {
		return false
	}
	atIdx := strings.IndexByte(u, '@')
	if atIdx <= 0 || atIdx != strings.LastIndexByte(u, '@') {
		return false
	}
	user := u[:atIdx]
	dom := u[atIdx+1:]
	if len(user) == 0 || len(dom) < 3 || strings.HasPrefix(dom, ".") || strings.HasSuffix(dom, ".") || strings.Contains(dom, "..") {
		return false
	}
	if !strings.Contains(dom, ".") {
		return false
	}
	for i := 0; i < len(dom); i++ {
		c := dom[i]
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '.' || c == '-') {
			return false
		}
	}
	for i := 0; i < len(user); i++ {
		c := user[i]
		if c <= 32 || c > 126 {
			return false
		}
	}
	return true
}

// IsCleanPassword kiểm tra mật khẩu hợp lệ (>= 6 ký tự, không chứa ký tự điều khiển hỏng/C1 control codes)
func IsCleanPassword(pw string) bool {
	if len(pw) < 6 {
		return false
	}
	for i := 0; i < len(pw); i++ {
		c := pw[i]
		if c < 32 || (c >= 127 && c <= 159) {
			return false
		}
	}
	return true
}

// ExtractCleanCredentials bóc tách và làm sạch dòng dữ liệu (chống ReDoS, hỗ trợ cả URL stealer log)
func ExtractCleanCredentials(line string) (string, string, bool) {
	line = strings.TrimSpace(line)
	if len(line) < 8 || len(line) > 1000 || !strings.Contains(line, ":") {
		return "", "", false
	}

	// TH1: Combo chuẩn chỉ có đúng 1 dấu ':'
	if strings.Count(line, ":") == 1 {
		parts := strings.SplitN(line, ":", 2)
		u := strings.TrimSpace(parts[0])
		p := strings.TrimSpace(parts[1])
		if len(u) > 0 && len(p) >= 6 && !strings.Contains(u, " ") &&
			!strings.HasPrefix(u, "http://") && !strings.HasPrefix(u, "https://") && !strings.HasPrefix(u, "android://") {
			return u, p, true
		}
		return "", "", false
	}

	// TH2: Stealer log có nhiều dấu ':' (ví dụ: https://site.com/login:user@email.com:password)
	parts := strings.Split(line, ":")
	if len(parts) >= 3 {
		pCand := strings.TrimSpace(parts[len(parts)-1])
		uCand := strings.TrimSpace(parts[len(parts)-2])
		if strings.Contains(uCand, "@") && strings.Contains(uCand, ".") && !strings.Contains(uCand, " ") && len(pCand) >= 6 {
			return uCand, pCand, true
		}
		if strings.Contains(parts[0], "@") && strings.Contains(parts[0], ".") && !strings.Contains(parts[0], " ") {
			uCand = strings.TrimSpace(parts[0])
			pCand = strings.TrimSpace(strings.Join(parts[1:], ":"))
			if len(pCand) >= 6 && !strings.HasPrefix(uCand, "http://") && !strings.HasPrefix(uCand, "https://") {
				return uCand, pCand, true
			}
		}
	}

	return "", "", false
}

// Capitalize viết hoa chữ cái đầu tiên
func Capitalize(s string) string {
	if len(s) == 0 {
		return s
	}
	r := []rune(s)
	r[0] = unicode.ToUpper(r[0])
	return string(r)
}

// CheckDomainMatch kiểm tra domain của email
func CheckDomainMatch(email string, domainArg string) bool {
	if len(email) == 0 || !strings.Contains(email, "@") || strings.Contains(email, " ") {
		return false
	}
	emailLower := strings.ToLower(email)
	parts := strings.SplitN(emailLower, "@", 2)
	if len(parts) != 2 {
		return false
	}
	domainPart := parts[1]

	target := "1"
	if len(domainArg) > 0 {
		target = strings.ToLower(strings.TrimSpace(domainArg))
	}
	cleanTarget := strings.Trim(target, ".@")

	if cleanTarget == "1" || cleanTarget == "hotmail" || cleanTarget == "outlook" || cleanTarget == "live" || cleanTarget == "msn" || cleanTarget == "windowslive" || cleanTarget == "passport" {
		msDoms := []string{"hotmail.", "outlook.", "live.", "msn.", "windowslive.", "passport."}
		for _, d := range msDoms {
			if strings.Contains(domainPart, d) {
				return true
			}
		}
		msPrefixes := []string{"hotmail", "outlook", "live", "msn"}
		for _, p := range msPrefixes {
			if strings.HasPrefix(domainPart, p) {
				return true
			}
		}
		return false
	}

	if cleanTarget == "2" || cleanTarget == "gmail" || cleanTarget == "googlemail" || cleanTarget == "google" {
		return strings.Contains(domainPart, "gmail.") || strings.Contains(domainPart, "googlemail.") ||
			strings.HasPrefix(domainPart, "gmail") || strings.HasPrefix(domainPart, "googlemail")
	}

	if len(cleanTarget) > 0 && strings.Contains(domainPart, cleanTarget) {
		return true
	}

	return false
}

