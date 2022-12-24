package helper

import "strings"

func EscapeTxt(txt string) string {

	str1 := strings.ReplaceAll(txt, "_", "\\_")
	str2 := strings.ReplaceAll(str1, "*", "\\*")
	str3 := strings.ReplaceAll(str2, "[", "\\[")
	str4 := strings.ReplaceAll(str3, "]", "\\]")
	str5 := strings.ReplaceAll(str4, "(", "\\(")
	str6 := strings.ReplaceAll(str5, ")", "\\)")
	str7 := strings.ReplaceAll(str6, "~", "\\~")
	str8 := strings.ReplaceAll(str7, ">", "\\>")
	str9 := strings.ReplaceAll(str8, "#", "\\#")
	str10 := strings.ReplaceAll(str9, "+", "\\+")
	str11 := strings.ReplaceAll(str10, "-", "\\-")
	str12 := strings.ReplaceAll(str11, "=", "\\=")
	str13 := strings.ReplaceAll(str12, "|", "\\|")
	str14 := strings.ReplaceAll(str13, "{", "\\{")
	str15 := strings.ReplaceAll(str14, "}", "\\}")
	str16 := strings.ReplaceAll(str15, ".", "\\.")
	str17 := strings.ReplaceAll(str16, "!", "\\!")

	return str17

}
