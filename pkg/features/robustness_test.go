package features

import (
	"testing"
)

func TestAllFeaturesRobustness(t *testing.T) {
	testLines := []string{
		"",
		":",
		":::",
		"admin",
		"admin:1",
		"admin:12",
		"admin:12345",
		"admin:123456",
		"admin@gmail.com:password",
		"admin@hotmail.com:12345678",
		"test@domain.com:!@#$%^&*()",
		"longusernameherefortestingpurposes@somedomain.org:P@ssw0rd2024",
		"invalid-email-no-at-sign:validpassword123",
		"http://stealer.log/site:user@test.com:pass123456",
		"a@b.c:123456",
		"user@gmail.com:my:secret:pass:123456",
		"user@gmail.com:password7679",
		"user@gmail.com:password724401",
		"user@gmail.com:password00",
		"user@gmail.com:password27",
		"user@gmail.com:password50",
		"user@gmail.com:password99",
	}

	testArgs := map[int][]string{
		11: {"", "all", "2"},
		12: {"", "a", "i", "s", "z"},
		14: {"", "2", "3", "5", "99"},
		15: {"", "@", "$", "!"},
		16: {"", "@", "#", "!"},
		20: {"", "Aa", "Prefix123"},
		21: {"", "H", "T"},
		25: {"", "1", "2", "hotmail", "gmail", "yahoo"},
		27: {"", "3D", "3C", "5D"},
		29: {"", "2@", "3$", "1#", "99!"},
		30: {""},
	}

	for _, feat := range Features {
		args := []string{""}
		if customArgs, exists := testArgs[feat.ID]; exists {
			args = customArgs
		}

		for _, arg := range args {
			for _, line := range testLines {
				func() {
					defer func() {
						if r := recover(); r != nil {
							t.Fatalf("PANIC on Feature %d (%s) with arg '%s', line '%s': %v", feat.ID, feat.Name, arg, line, r)
						}
					}()
					_ = feat.Handler(line, arg)
				}()
			}
		}
	}
}
