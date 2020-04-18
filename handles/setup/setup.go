package setup

import (
	"strings"
	"unicode"
	"weagent/gfunc"
)

var (
	coefficient = []int32{7, 9, 10, 5, 8, 4, 2, 1, 6, 3, 7, 9, 10, 5, 8, 4, 2}
	code        = []byte{'1', '0', 'X', '9', '8', '7', '6', '5', '4', '3', '2'}
)

// checkCardCodeValid 校验一个身份证是否是合法的身份证
func checkCardCodeValid(cardcode string) bool {
	if len(cardcode) != 18 {
		return false
	}

	idByte := []byte(strings.ToUpper(cardcode))

	sum := int32(0)
	for i := 0; i < 17; i++ {
		sum += int32(byte(idByte[i])-byte('0')) * coefficient[i]
	}
	return code[sum%11] == idByte[17]
}

func checkNameValid(name string) bool {
	// 检测是否都是中文
	hancount := 0
	for _, v := range name {
		ishan := unicode.Is(unicode.Han, v)
		if !ishan {
			return false
		}

		hancount++
	}

	// 检查汉字个数对不对
	if hancount < 2 || hancount > 8 {
		return false
	}

	// 检测有没有敏感词
	tmpname := name
	gfunc.ReplaceSensitiveWord(tmpname)
	if tmpname != name {
		return false
	}

	return true
}
