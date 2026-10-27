package transforms

import (
	"testing"
)

func TestExtractCleanCredentials(t *testing.T) {
	// Standard
	u, p, ok := ExtractCleanCredentials("user1@gmail.com:Pass123456")
	if !ok || u != "user1@gmail.com" || p != "Pass123456" {
		t.Fatalf("Failed standard credentials extract: got %s, %s, %v", u, p, ok)
	}

	// Stealer URL
	u, p, ok = ExtractCleanCredentials("https://site.com/login:user2@hotmail.com:Pass789012")
	if !ok || u != "user2@hotmail.com" || p != "Pass789012" {
		t.Fatalf("Failed stealer credentials extract: got %s, %s, %v", u, p, ok)
	}

	// Short pass (< 6)
	_, _, ok = ExtractCleanCredentials("user@domain.com:123")
	if ok {
		t.Fatalf("Expected failure for short pass")
	}
}

func TestFunc14CharReplacement(t *testing.T) {
	// Có chữ a -> thay đổi
	res := Func14CharReplacement("user@domain.com:passWord123", "a")
	if len(res) == 0 || res[0] != "user@domain.com:p@ssWord123" {
		t.Fatalf("Expected replacement, got %v", res)
	}

	// Không có chữ a -> loại bỏ dòng
	res = Func14CharReplacement("user@domain.com:testWord123", "a")
	if len(res) != 0 {
		t.Fatalf("Expected drop (nil), got %v", res)
	}
}

func TestFunc16EmailNChars(t *testing.T) {
	// Đủ 3 chữ
	res := Func16EmailNChars("john123@domain.com:oldPass8899", "3")
	if len(res) == 0 || res[0] != "john123@domain.com:Joh8899" {
		t.Fatalf("Expected Joh8899, got %v", res)
	}

	// Không đủ chữ (chỉ có 2 chữ cái mà đòi 3) -> loại bỏ dòng
	res = Func16EmailNChars("ab123@domain.com:oldPass8899", "3")
	if len(res) != 0 {
		t.Fatalf("Expected drop, got %v", res)
	}
}

func TestCheckDomainMatch(t *testing.T) {
	if !CheckDomainMatch("user@hotmail.com", "1") {
		t.Fatalf("Expected hotmail match")
	}
	if !CheckDomainMatch("user@outlook.com", "1") {
		t.Fatalf("Expected outlook match")
	}
	if CheckDomainMatch("user@gmail.com", "1") {
		t.Fatalf("Expected gmail NOT match 1")
	}
	if !CheckDomainMatch("user@gmail.com", "2") {
		t.Fatalf("Expected gmail match 2")
	}
	if !CheckDomainMatch("user@custom-site.fr", "fr") {
		t.Fatalf("Expected custom domain match")
	}
}

func TestFunc27Rcr(t *testing.T) {
	res := Func27RcrEmailNCharsPos("alexander@domain.com:secret123456", "3D")
	if len(res) != 7 {
		t.Fatalf("Expected 7 variants for Rcr, got %d", len(res))
	}
}

