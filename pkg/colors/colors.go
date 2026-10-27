package colors

const (
	Reset     = "\033[0m"
	Bold      = "\033[1m"
	Dim       = "\033[2m"
	Italic    = "\033[3m"
	Underline = "\033[4m"

	// Classic
	Green  = "\033[92m"
	Yellow = "\033[93m"
	White  = "\033[97m"
	Cyan   = "\033[96m"

	// Modern Vibrant Cyberpunk & Mango Palette
	CyanNeon    = "\033[38;2;0;240;255m"   // #00f0ff - Điện quang rực rỡ
	MintGreen   = "\033[38;2;0;255;170m"   // #00ffaa - Bạc hà công nghệ
	IceBlue     = "\033[38;2;56;189;248m"  // #38bdf8 - Xanh băng thanh lịch
	SkyBlue     = "\033[38;2;96;165;250m"  // #60a5fa - Xanh lam hiện đại
	Lavender    = "\033[38;2;192;132;252m" // #c084fc - Tím hoa oải hương dịu mắt
	PurpleVIP   = "\033[38;2;216;112;240m" // #d870f0 - Tím hoàng gia VIP
	RoseNeon    = "\033[38;2;251;113;133m" // #fb7185 - Hồng san hô nổi bật
	MangoYellow = "\033[38;2;255;215;0m"   // #ffd700 - Vàng sáng Mango
	MangoGold   = "\033[38;2;255;190;0m"   // #ffbe00 - Vàng hổ phách
	MangoAmber  = "\033[38;2;255;145;0m"   // #ff9100 - Cam hổ phách
	MangoOrange = "\033[38;2;255;110;20m"  // #ff6e14 - Cam rực
	MangoRed    = "\033[38;2;255;75;43m"   // #ff4b2b - Đỏ cam lửa

	// Typography & Borders
	TextWhite   = "\033[38;2;255;255;255m" // #ffffff - Trắng tinh khiết
	TextRegular = "\033[38;2;226;232;240m" // #e2e8f0 - Slate 200 (sáng rõ, dễ đọc)
	TextMuted   = "\033[38;2;148;163;184m" // #94a3b8 - Slate 400 (mô tả dịu mắt)
	BorderCyan  = "\033[38;2;30;75;95m"    // #1e4b5f - Viền xanh thẫm hiện đại
	BorderDim   = "\033[38;2;51;65;85m"    // #334155 - Phân nhóm mờ Slate 700
)
