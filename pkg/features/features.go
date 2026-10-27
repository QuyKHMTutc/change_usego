package features

import (
	"strconv"
	"strings"
	"tools2/pkg/transforms"
)

type Feature struct {
	ID          int
	Name        string
	Desc        string
	HandlerName string
	HasArg      bool
	Handler     transforms.TransformFunc
}

type Task struct {
	Feat Feature
	Arg  string
}

var Features = []Feature{
	{1, "Làm sạch ký tự đặc biệt & Viết hoa chữ đầu (103 BETA)", "Chỉ giữ chữ + số, viết hoa chữ cái đầu tiên của mật khẩu.", "func_103_beta", false, transforms.Func103Beta},
	{2, "Chèn Username của Email vào đầu Mật khẩu (2_word)", "Xóa chữ cái trong pass, lấy toàn bộ Username của email dán lên đầu pass.", "func_2_word", false, transforms.Func2Word},
	{3, "Đưa số ra sau + Nhân bản 2 số đầu ra đuôi pass (2nX2)", "Làm sạch, viết hoa đầu, dồn số ra sau, lấy 2 số đầu dán thêm vào cuối.", "func_2nx2", false, transforms.Func2nx2},
	{4, "Đưa số lên đầu + Chữ viết hoa ra sau (number_pass)", "Tách số và chữ, đưa toàn bộ số lên trước, chữ viết hoa theo sau.", "func_number_pass", false, transforms.FuncNumberPass},
	{5, "Chữ Username + Số trong Mật khẩu (USER_PASS)", "Lấy chữ từ username (trước dấu & số) ghép với toàn bộ số của pass.", "func_user_pass", false, transforms.FuncUserPass},
	{6, "Đưa số ra sau + Chèn chữ đầu Username vào cuối (abe 103R)", "Chuyển số ra sau, chèn chữ đầu username in hoa vào cuối (pass >=6 số).", "func_abe_103r", false, transforms.FuncAbe103r},
	{7, "Đảo 2 chữ đầu Username + Giữ số Mật khẩu (agian SCOURE)", "Lấy 2 chữ đầu username đảo ngược viết hoa + giữ lại số pass (>=6 số).", "func_agian_scoure", false, transforms.FuncAgianScoure},
	{8, "Đảo ngược thứ tự các con số trong Mật khẩu (dao_2N)", "Đảo ngược các con số trong pass nhưng giữ nguyên vị trí chữ cái.", "func_dao_2n", false, transforms.FuncDao2n},
	{9, "Lặp / Nhân đôi số cuối mật khẩu (lap_so_2)", "Lặp lại 1-2 số cuối; nếu >2 số thì lặp 2 số đầu; nếu không có số lấy từ email.", "func_lap_so_2", false, transforms.FuncLapSo2},
	{10, "Hotmail: 2 chữ cái Username + Viết hoa toàn bộ Pass (3E)", "Lọc hotmail, lấy 2 chữ cái đầu username ghép pass, viết hoa toàn bộ pass (>=4 số).", "func_3e", false, transforms.Func3e},
	{11, "Lấy số từ Email thêm vào cuối Pass (tất cả số & 2 số cuối)", "Lấy số trong username email ghép vào đuôi pass. Hỗ trợ: 11 (cả hai), 11.all, 11.2", "func_13_email_nums_to_pw_end", true, transforms.Func13EmailNumsToPwEnd},
	{12, "Mã hóa ký tự trong Pass (12 a, 12 i, 12 s)", "Thay ký tự: 12 a (a->@), 12 i (i->!), 12 s (s->$).", "func_14_char_replacement", true, transforms.Func14CharReplacement},
	{13, "Viết hoa chữ cái sau Số cuối cùng", "Sau số ở vị trí cuối, chữ sẽ viết hoa. Không đúng điều kiện sẽ bị loại bỏ.", "func_viet_hoa_sau_so_cuoi", false, transforms.FuncVietHoaSauSoCuoi},
	{14, "Lấy N chữ đầu Email + Số (Gõ 14.<N> tùy ý)", "Lấy N chữ đầu email (14.2, 14.3, 14.5...) ghép số pass cũ.", "func_16_email_n_chars", true, transforms.Func16EmailNChars},
	{15, "Ký tự đặc biệt TRƯỚC SỐ (15 <ký tự tùy ý>)", "Username in hoa + [ký tự tùy ý] + số pass (Ví dụ: 15 @, 15 $, 15 !...).", "func_17_special_before_nums", true, transforms.Func17SpecialBeforeNums},
	{16, "Ký tự đặc biệt SAU SỐ (16 <ký tự tùy ý>)", "Username in hoa + số pass + [ký tự tùy ý] (Ví dụ: 16 @, 16 #, 16 !...).", "func_18_special_after_nums", true, transforms.Func18SpecialAfterNums},
	{17, "Tên Mail viết hoa đầu + Số mật khẩu", "Lấy tên email ghép số mật khẩu, không có số chỉ lấy tên.", "func_22_name_and_nums", false, transforms.Func22NameAndNums},
	{18, "Số mật khẩu + Tên Mail", "Số mật khẩu đứng trước, Tên email (chữ cái) đứng sau.", "func_25_nums_and_email_name", false, transforms.Func25NumsAndEmailName},
	{19, "Nhân đôi ký tự/số cuối mật khẩu", "Ví dụ: 12345 -> 123455, pass123! -> pass123!!, abc -> abcc.", "func_26_duplicate_last_char", false, transforms.Func26DuplicateLastChar},
	{20, "Chữ bất kỳ từ bàn phím + Số mật khẩu cũ", "Ghép chữ nhập từ bàn phím với số của pass cũ (>=6 ký tự). Hỗ trợ: 20 Aa | 20 Ab", "func_27_custom_prefix", true, transforms.Func27CustomPrefix},
	{21, "Viết Hoa / Thường chữ cái đầu Pass (21 H / 21 T)", "21 H: Viết hoa chữ cái đầu tiên | 21 T: Viết thường chữ cái đầu tiên.", "func_23_capitalize_or_lower_first_letter", true, transforms.Func23CapitalizeOrLowerFirstLetter},
	{22, "Lọc Mật khẩu >= 6 ký tự (Xóa < 6 & thiếu pass)", "Xóa dòng chỉ có mail không pass, xóa dòng có mật khẩu dưới 6 ký tự.", "func_24_filter_pass_min_6", false, transforms.Func24FilterPassMin6},
	{23, "Xóa ký tự đặc biệt trong Pass (Giữ chữ & số)", "Chỉ xóa ký tự đặc biệt (!@#...), giữ nguyên chữ cái hoa/thường và số.", "func_24_remove_special_chars", false, transforms.Func24RemoveSpecialChars},
	{24, "Chỉ lấy chữ trong Pass (Xóa số & ký tự, giữ >= 6 chữ)", "Chỉ lấy chữ cái trong pass (xóa số & ký tự). Nếu pass sau khi lọc < 6 chữ thì xóa dòng.", "func_25_keep_only_letters", false, transforms.Func25KeepOnlyLetters},
	{25, "Tách file Domain & Phần còn lại (Hotmail/Gmail/Domain...)", "Tách đôi 2 file song song: File Domain (25_hotmail/gmail) VÀ File Còn lại (25_other.txt).", "func_domain_filter", true, transforms.FuncDomainFilter},
	{26, "Chỉ lấy số trong Pass (Xóa chữ & ký tự, giữ >= 6 số)", "Chỉ lấy chữ số trong pass (xóa chữ và ký tự). Nếu pass sau khi lọc < 6 số thì xóa dòng.", "func_26_keep_only_numbers", false, transforms.Func26KeepOnlyNumbers},
	{27, "N Chữ đầu Email + Số Pass (27 3D / 27 3C)", "Kế thừa Rcr.py: Ghép ĐẦU (3D) / CUỐI (3C) và tự gộp đủ 7 biến thể (!, @, $, *, #, @sau_2chữ).", "func_27_rcr_email_n_chars_pos", true, transforms.Func27RcrEmailNCharsPos},
	{28, "Tên Email làm Mật khẩu mới (>=6 & <=15 chữ)", "Lấy chữ từ tên mail (6-15 chữ) viết hoa đầu làm pass mới (email:Tenmail).", "func_28_get_email_name", false, transforms.Func28GetEmailName},
	{29, "Giữ User + Pass 2 âm tiết quan trọng & Giữ số", "Giữ nguyên User & Số trong Pass, trích xuất 2 âm tiết quan trọng của chữ trong Pass.", "func_29_important_syllables", false, transforms.Func29ImportantSyllables},
}

func GetFeatureByID(id int) *Feature {
	for i := range Features {
		if Features[i].ID == id {
			return &Features[i]
		}
	}
	return nil
}

// ParseChoiceString phân tích chuỗi người dùng nhập (hỗ trợ ghép nhiều tác vụ bằng dấu |)
func ParseChoiceString(choiceStr string) ([]Task, string) {
	rawParts := strings.Split(choiceStr, "|")
	var tasks []Task

	for _, part := range rawParts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		fields := strings.Fields(part)
		if len(fields) == 0 {
			continue
		}

		firstTok := fields[0]
		var arg string
		if len(fields) > 1 {
			arg = strings.TrimSpace(fields[1])
		}

		// Cú pháp 11.x (ví dụ 11.all, 11.2)
		if strings.HasPrefix(firstTok, "11.") {
			subArg := strings.TrimPrefix(firstTok, "11.")
			if f := GetFeatureByID(11); f != nil {
				tasks = append(tasks, Task{*f, subArg})
				continue
			}
		}

		// Cú pháp 14.x (ví dụ 14.2, 14.3, 14.5)
		if strings.HasPrefix(firstTok, "14.") {
			subArg := strings.TrimPrefix(firstTok, "14.")
			if f := GetFeatureByID(14); f != nil {
				if subArg == "" {
					subArg = "2"
				}
				tasks = append(tasks, Task{*f, subArg})
				continue
			}
		}

		// Cú pháp 25.x (ví dụ 25.1, 25.2, 25.hotmail)
		if strings.HasPrefix(firstTok, "25.") {
			subArg := strings.TrimPrefix(firstTok, "25.")
			if f := GetFeatureByID(25); f != nil {
				if subArg == "" {
					subArg = "1"
				}
				tasks = append(tasks, Task{*f, subArg})
				continue
			}
		}

		// Cú pháp 21.H / 21.T
		if strings.HasPrefix(firstTok, "21.") {
			subArg := strings.TrimPrefix(firstTok, "21.")
			if f := GetFeatureByID(21); f != nil {
				if subArg == "" {
					subArg = "H"
				}
				tasks = append(tasks, Task{*f, subArg})
				continue
			}
		}

		// Cú pháp số thông thường: 1, 2, 25, 27
		id, err := strconv.Atoi(firstTok)
		if err == nil {
			if f := GetFeatureByID(id); f != nil {
				tasks = append(tasks, Task{*f, arg})
				continue
			}
		}

		// Cú pháp dính liền ví dụ: 12a, 12i, 12s, 15@, 15$, 16@, 16#, 21h, 21t, 273d, 273c
		if len(firstTok) >= 2 {
			prefix2 := firstTok[:2]
			suffix := firstTok[2:]
			if pID, err := strconv.Atoi(prefix2); err == nil {
				if f := GetFeatureByID(pID); f != nil {
					tasks = append(tasks, Task{*f, suffix})
					continue
				}
			}
		}
	}

	if len(tasks) == 0 {
		return nil, "Không nhận diện được chức năng hợp lệ!"
	}

	return tasks, ""
}

