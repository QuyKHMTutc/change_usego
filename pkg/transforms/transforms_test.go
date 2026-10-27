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


func TestFunc29InsertCharPos(t *testing.T) {
	// Test 29 2@: password -> p@assword
	res := Func29InsertCharPos("user@example.com:password", "2@")
	if len(res) == 0 || res[0] != "user@example.com:p@assword" {
		t.Fatalf("Expected p@assword, got %v", res)
	}

	// Test 29 3$: password -> pa$ssword
	res = Func29InsertCharPos("user@example.com:password", "3$")
	if len(res) == 0 || res[0] != "user@example.com:pa$ssword" {
		t.Fatalf("Expected pa$ssword, got %v", res)
	}

	// Test 29 1#: password -> #password
	res = Func29InsertCharPos("user@example.com:password", "1#")
	if len(res) == 0 || res[0] != "user@example.com:#password" {
		t.Fatalf("Expected #password, got %v", res)
	}

	// Test password too short for pos 10
	res = Func29InsertCharPos("user@example.com:pass", "10@")
	if len(res) != 0 {
		t.Fatalf("Expected nil for pass shorter than pos, got %v", res)
	}
}

func TestFunc30PassLast2DigitsToYear(t *testing.T) {
	// password7679 -> password1979
	res := Func30PassLast2DigitsToYear("user@test.com:password7679", "")
	if len(res) == 0 || res[0] != "user@test.com:password1979" {
		t.Fatalf("Expected password1979, got %v", res)
	}

	// password724401 -> password2001
	res = Func30PassLast2DigitsToYear("user@test.com:password724401", "")
	if len(res) == 0 || res[0] != "user@test.com:password2001" {
		t.Fatalf("Expected password2001, got %v", res)
	}

	// Out of range (35) -> drop
	res = Func30PassLast2DigitsToYear("user@test.com:password35", "")
	if len(res) != 0 {
		t.Fatalf("Expected nil for 35, got %v", res)
	}

	// Out of range (00) -> drop
	res = Func30PassLast2DigitsToYear("user@test.com:password00", "")
	if len(res) != 0 {
		t.Fatalf("Expected nil for 00, got %v", res)
	}

	// No digits -> drop
	res = Func30PassLast2DigitsToYear("user@test.com:password", "")
	if len(res) != 0 {
		t.Fatalf("Expected nil for no digits, got %v", res)
	}
}
