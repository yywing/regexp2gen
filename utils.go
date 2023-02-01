package regexp2gen

import (
	"errors"

	"github.com/dlclark/regexp2/syntax"
)

// 尝试寻找一个能满足的匹配项
// TODO：因为 charSet 没有提供相应的属性或者方法出来，先简单搞一下
func resolveCharSet(set *syntax.CharSet) (r rune, err error) {
	defer func() {
		if e := recover(); e != nil {
			err = errors.New("cat get char set")
		}
	}()
	if set.IsNegated() {
		return r, errors.New("cat get char set")
	}

	r = set.SingletonChar()
	return r, nil
}
