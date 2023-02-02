package regexp2gen

import (
	"testing"

	"github.com/dlclark/regexp2"
	"github.com/stretchr/testify/require"
)

func TestReplace(t *testing.T) {
	// notoneloop
	s := "[^0]{1,3}"
	s += "[^0]{1,}"
	// oneloop
	s += "0{1,3}"
	s += "0{1,}"
	// setloop
	s += "[a-z]{1,3}"
	s += "[a-z]{1,}"
	// onelazy
	s += "A{2,4}?"
	// notonelazy
	s += ".{2,4}?"
	// setlazy
	s += "[a-z]{2,4}?"
	// one
	s += "b"
	// notone
	s += "[^a]"
	// set
	s += "[a-z]"
	// multi
	s += "test"
	// setmark, capturemark, ref
	s += `(a(a))\1`

	re, err := regexp2.Compile(s, regexp2.RE2)
	require.Nil(t, err)

	g := NewGenerator()
	data, err := g.Generate(NewState(true, 3, nil, 0), s, regexp2.RE2)
	require.Nil(t, err)
	result, err := re.MatchString(data)
	require.Nil(t, err)
	require.True(t, result)
}

func TestRequire(t *testing.T) {
	/*
		000003 *Setjump()
		000004 *Setmark()
		000005  Multi-Rtl(String = gg)
		000007 *Getmark()
		000008 *Forejump()
	*/
	cases := []string{
		`(?:gg)aa`,
		`(?:gg)*aa`,
		`aa(?=gg)`,
		`(?<=gg)aa`,
		`(?<a>gg)aa`,
	}

	for _, s := range cases {
		re, err := regexp2.Compile(s, regexp2.RE2)
		require.Nil(t, err)

		g := NewGenerator()
		data, err := g.Generate(NewState(true, 3, nil, 0), s, regexp2.RE2)
		require.Nil(t, err)
		result, err := re.MatchString(data)
		require.Nil(t, err)
		require.True(t, result)
	}
}

func TestPrevent(t *testing.T) {
	/*
		000003 *Setjump()
		000004 *Lazybranch(Addr = 9)
		000006  Multi-Rtl(String = gg)
		000008 *Backjump()
		000009 *Forejump()
	*/
	cases := []string{
		`(?!gg)aa`,
		// `(?<!gg)aa`,
	}

	for _, s := range cases {
		re, err := regexp2.Compile(s, regexp2.RE2)
		require.Nil(t, err)

		g := NewGenerator()
		data, err := g.Generate(NewState(true, 3, nil, 0), s, regexp2.RE2)
		require.Nil(t, err)
		result, err := re.MatchString(data)
		require.Nil(t, err)
		require.True(t, result)
	}

}

func TestGoto(t *testing.T) {
	s := "^(3C|C0)$"

	re, err := regexp2.Compile(s, regexp2.RE2)
	require.Nil(t, err)

	g := NewGenerator()
	data, err := g.Generate(NewState(true, 3, nil, 0), s, regexp2.RE2)
	require.Nil(t, err)
	result, err := re.MatchString(data)
	require.Nil(t, err)
	require.True(t, result)
}

func TestCharSet(t *testing.T) {
	// 测试字符不在 state 给定的范围，能否找到合适的字符
	s := `^[\x80-\xff][\x00\x03]$`

	re, err := regexp2.Compile(s, regexp2.RE2)
	require.Nil(t, err)

	g := NewGenerator()
	data, err := g.Generate(NewState(true, 3, nil, 0), s, regexp2.RE2)
	require.Nil(t, err)
	result, err := re.MatchString(data)
	require.Nil(t, err)
	require.True(t, result)

	// 测试 recover 是否生效
	s = `^[\d]$`

	_, err = g.Generate(NewState(true, 3, []rune{}, 0), s, regexp2.RE2)
	require.NotNil(t, err)
}

func TestBranchCount(t *testing.T) {
	s := `^(a){5}(b){1}(c)?$`

	re, err := regexp2.Compile(s, regexp2.RE2)
	require.Nil(t, err)

	g := NewGenerator()
	data, err := g.Generate(NewState(true, 3, nil, 0), s, regexp2.RE2)
	require.Nil(t, err)
	result, err := re.MatchString(data)
	require.Nil(t, err)
	require.True(t, result)
}

func TestBranchMark(t *testing.T) {
	s := `^a(?:[\w/+=])+a(?:[\w/+=])+a$`

	re, err := regexp2.Compile(s, regexp2.RE2)
	require.Nil(t, err)

	g := NewGenerator()
	data, err := g.Generate(NewState(true, 3, nil, 0), s, regexp2.RE2)
	require.Nil(t, err)
	result, err := re.MatchString(data)
	require.Nil(t, err)
	require.True(t, result)
}

func TestLazybranchmark(t *testing.T) {
	s := `^a(?:b)*?aaa(?:b)*?aaa`

	re, err := regexp2.Compile(s, regexp2.RE2)
	require.Nil(t, err)

	g := NewGenerator()
	data, err := g.Generate(NewState(true, 3, nil, 0), s, regexp2.RE2)
	require.Nil(t, err)
	result, err := re.MatchString(data)
	require.Nil(t, err)
	require.True(t, result)
}

// func TestAll(t *testing.T) {
// 	/*
// 		000003 *Setjump()
// 		000004 *Setmark()
// 		000005  Multi-Rtl(String = gg)
// 		000007 *Getmark()
// 		000008 *Forejump()
// 	*/
// 	cases := []string{
// 		`abc`,
// 		`abc`,
// 		`abc`,
// 		`ab*c`,
// 		`ab*bc`,
// 		`ab*bc`,
// 		`ab*bc`,
// 		`.{1}`,
// 		`.{3,4}`,
// 		`ab{0,}bc`,
// 		`ab+bc`,
// 		`ab+bc`,
// 		`ab{1,}bc`,
// 		`ab{1,3}bc`,
// 		`ab{3,4}bc`,
// 		`ab?bc`,
// 		`ab?bc`,
// 		`ab{0,1}bc`,
// 		`ab?c`,
// 		`ab{0,1}c`,
// 		`^abc$`,
// 		`^abc`,
// 		`abc$`,
// 		`^`,
// 		`$`,
// 		`a.c`,
// 		`a.c`,
// 		`a.*c`,
// 		`a[bc]d`,
// 		`a[b-d]e`,
// 		`a[b-d]`,
// 		`a[-b]`,
// 		`a[b-]`,
// 		`a]`,
// 		`a[]]b`,
// 		`a[^bc]d`,
// 		`a[^-b]c`,
// 		`a[^]b]c`,
// 		`\ba\b`,
// 		`\ba\b`,
// 		`\ba\b`,
// 		`\By\b`,
// 		`\by\B`,
// 		`\By\B`,
// 		`\w`,
// 		`\W`,
// 		`a\sb`,
// 		`a\Sb`,
// 		`\d`,
// 		`\D`,
// 		`[\w]`,
// 		`[\W]`,
// 		`a[\s]b`,
// 		`a[\S]b`,
// 		`[\d]`,
// 		`[\D]`,
// 		`ab|cd`,
// 		`ab|cd`,
// 		`()ef`,
// 		`a\(b`,
// 		`a\(*b`,
// 		`a\(*b`,
// 		`a\\b`,
// 		`((a))`,
// 		`(a)b(c)`,
// 		`a+b+c`,
// 		`a{1,}b{1,}c`,
// 		`a.+?c`,
// 		`(a+|b)*`,
// 		`(a+|b){0,}`,
// 		`(a+|b)+`,
// 		`(a+|b){1,}`,
// 		`(a+|b)?`,
// 		`(a+|b){0,1}`,
// 		`[^ab]*`,
// 		`a*`,
// 		`([abc])*d`,
// 		`([abc])*bcd`,
// 		`a|b|c|d|e`,
// 		`(a|b|c|d|e)f`,
// 		`abcd*efg`,
// 		`ab*`,
// 		`ab*`,
// 		`(ab|cd)e`,
// 		`[abhgefdc]ij`,
// 		`(abc|)ef`,
// 		`(a|b)c*d`,
// 		`(ab|ab*)bc`,
// 		`a([bc]*)c*`,
// 		`a([bc]*)(c*d)`,
// 		`a([bc]+)(c*d)`,
// 		`a([bc]*)(c+d)`,
// 		`a[bcd]*dcdcde`,
// 		`(ab|a)b*c`,
// 		`((a)(b)c)(d)`,
// 		`[a-zA-Z_][a-zA-Z0-9_]*`,
// 		`^a(bc+|b[eh])g|.h$`,
// 		`(bc+d$|ef*g.|h?i(j|k))`,
// 		`(bc+d$|ef*g.|h?i(j|k))`,
// 		`(bc+d$|ef*g.|h?i(j|k))`,
// 		`((((((((((a))))))))))`,
// 		`((((((((((a))))))))))\10`,
// 		`((((((((((a))))))))))!`,
// 		`(((((((((a)))))))))`,
// 		`multiple words`,
// 		`(.*)c(.*)`,
// 		`\((.*), (.*)\)`,
// 		`abcd`,
// 		`a(bc)d`,
// 		`a[-]?c`,
// 		`(abc)\1`,
// 		`([a-c]*)\1`,
// 		`(a)|\1`,
// 		`(([a-c])b*?\2)*`,
// 		`(([a-c])b*?\2){3}`,
// 		`((\3|b)\2(a)x)+`,
// 		`((\3|b)\2(a)){2,}`,
// 		`abc`,
// 		`abc`,
// 		`abc`,
// 		`ab*c`,
// 		`ab*bc`,
// 		`ab*bc`,
// 		`ab*?bc`,
// 		`ab{0,}?bc`,
// 		`ab+?bc`,
// 		`ab+bc`,
// 		`ab{1,}?bc`,
// 		`ab{1,3}?bc`,
// 		`ab{3,4}?bc`,
// 		`ab??bc`,
// 		`ab??bc`,
// 		`ab{0,1}?bc`,
// 		`ab??c`,
// 		`ab{0,1}?c`,
// 		`^abc$`,
// 		`^abc`,
// 		`abc$`,
// 		`^`,
// 		`$`,
// 		`a.c`,
// 		`a.c`,
// 		`a.*?c`,
// 		`a[bc]d`,
// 		`a[b-d]e`,
// 		`a[b-d]`,
// 		`a[-b]`,
// 		`a[b-]`,
// 		`a]`,
// 		`a[]]b`,
// 		`a[^bc]d`,
// 		`a[^-b]c`,
// 		`a[^]b]c`,
// 		`ab|cd`,
// 		`ab|cd`,
// 		`()ef`,
// 		`a\(b`,
// 		`a\(*b`,
// 		`a\(*b`,
// 		`a\\b`,
// 		`((a))`,
// 		`(a)b(c)`,
// 		`a+b+c`,
// 		`a{1,}b{1,}c`,
// 		`a.+?c`,
// 		`a.*?c`,
// 		`a.{0,5}?c`,
// 		`(a+|b)*`,
// 		`(a+|b){0,}`,
// 		`(a+|b)+`,
// 		`(a+|b){1,}`,
// 		`(a+|b)?`,
// 		`(a+|b){0,1}`,
// 		`(a+|b){0,1}?`,
// 		`[^ab]*`,
// 		`a*`,
// 		`([abc])*d`,
// 		`([abc])*bcd`,
// 		`a|b|c|d|e`,
// 		`(a|b|c|d|e)f`,
// 		`abcd*efg`,
// 		`ab*`,
// 		`ab*`,
// 		`(ab|cd)e`,
// 		`[abhgefdc]ij`,
// 		`(abc|)ef`,
// 		`(a|b)c*d`,
// 		`(ab|ab*)bc`,
// 		`a([bc]*)c*`,
// 		`a([bc]*)(c*d)`,
// 		`a([bc]+)(c*d)`,
// 		`a([bc]*)(c+d)`,
// 		`a[bcd]*dcdcde`,
// 		`(ab|a)b*c`,
// 		`((a)(b)c)(d)`,
// 		`[a-zA-Z_][a-zA-Z0-9_]*`,
// 		`^a(bc+|b[eh])g|.h$`,
// 		`(bc+d$|ef*g.|h?i(j|k))`,
// 		`(bc+d$|ef*g.|h?i(j|k))`,
// 		`(bc+d$|ef*g.|h?i(j|k))`,
// 		`((((((((((a))))))))))`,
// 		`((((((((((a))))))))))\10`,
// 		`((((((((((a))))))))))!`,
// 		`(((((((((a)))))))))`,
// 		`(?:(?:(?:(?:(?:(?:(?:(?:(?:(a))))))))))`,
// 		`(?:(?:(?:(?:(?:(?:(?:(?:(?:(a|b|c))))))))))`,
// 		`multiple words`,
// 		`(.*)c(.*)`,
// 		`\((.*), (.*)\)`,
// 		`abcd`,
// 		`a(bc)d`,
// 		`a[-]?c`,
// 		`(abc)\1`,
// 		`([a-c]*)\1`,
// 		`a(?!b).`,
// 		`a(?=d).`,
// 		`a(?=c|d).`,
// 		`a(?:b|c|d)(.)`,
// 		`a(?:b|c|d)*(.)`,
// 		`a(?:b|c|d)+?(.)`,
// 		`a(?:b|c|d)+?(.)`,
// 		`a(?:b|c|d)+(.)`,
// 		`a(?:b|c|d){2}(.)`,
// 		`a(?:b|c|d){4,5}(.)`,
// 		`a(?:b|c|d){4,5}?(.)`,
// 		`((foo)|(bar))*`,
// 		`a(?:b|c|d){6,7}(.)`,
// 		`a(?:b|c|d){6,7}?(.)`,
// 		`a(?:b|c|d){5,6}(.)`,
// 		`a(?:b|c|d){5,6}?(.)`,
// 		`a(?:b|c|d){5,7}(.)`,
// 		`a(?:b|c|d){5,7}?(.)`,
// 		`a(?:b|(c|e){1,2}?|d)+?(.)`,
// 		`^(.+)?B`,
// 		`^([^a-z])|(\^)$`,
// 		`^[<>]&`,
// 		`^(a\1?){4}$`,
// 		`^(a(?(1)\1)){4}$`,
// 		`((a{4})+)`,
// 		`(((aa){2})+)`,
// 		`(((a{2}){2})+)`,
// 		`(?:(f)(o)(o)|(b)(a)(r))*`,
// 		`(?<=a)b`,
// 		`(?<!c)b`,
// 		`(?<!c)b`,
// 		`(?<!c)b`,
// 		`(?:..)*a`,
// 		`(?:..)*?a`,
// 		`^(?:b|a(?=(.)))*\1`,
// 		`^(){3,5}`,
// 		`^(a+)*ax`,
// 		`^((a|b)+)*ax`,
// 		`^((a|bc)+)*ax`,
// 		`(a|x)*ab`,
// 		`(a)*ab`,
// 		`(?:(?i)a)b`,
// 		`((?i)a)b`,
// 		`(?:(?i)a)b`,
// 		`((?i)a)b`,
// 		`(?i:a)b`,
// 		`((?i:a))b`,
// 		`(?i:a)b`,
// 		`((?i:a))b`,
// 		`(?:(?-i)a)b`,
// 		`((?-i)a)b`,
// 		`(?:(?-i)a)b`,
// 		`((?-i)a)b`,
// 		`(?:(?-i)a)b`,
// 		`((?-i)a)b`,
// 		`(?-i:a)b`,
// 		`((?-i:a))b`,
// 		`(?-i:a)b`,
// 		`((?-i:a))b`,
// 		`(?-i:a)b`,
// 		`((?-i:a))b`,
// 		`((?s-i:a.))b`,
// 		`(?:c|d)(?:)(?:a(?:)(?:b)(?:b(?:))(?:b(?:)(?:b)))`,
// 		`(?:c|d)(?:)(?:aaaaaaaa(?:)(?:bbbbbbbb)(?:bbbbbbbb(?:))(?:bbbbbbbb(?:)(?:bbbbbbbb)))`,
// 		`(ab)\d\1`,
// 		`(ab)\d\1`,
// 		`foo\w*\d{4}baz`,
// 		`x(~~)*(?:(?:F)?)?`,
// 		`^a(?#xxx){3}c`,
// 		`(?<![cd])[ab]`,
// 		`(?<!(c|d))[ab]`,
// 		`(?<!cd)[ab]`,
// 		`((?s)^a(.))((?m)^b$)`,
// 		`((?m)^b$)`,
// 		`(?m)^b`,
// 		`(?m)^(b)`,
// 		`((?m)^b)`,
// 		`\n((?m)^b)`,
// 		`((?s).)c(?!.)`,
// 		`((?s).)c(?!.)`,
// 		`((?s)b.)c(?!.)`,
// 		`((?s)b.)c(?!.)`,
// 		`((?m)^b)`,
// 		`(x)?(?(1)b|a)`,
// 		`()?(?(1)b|a)`,
// 		`()?(?(1)a|b)`,
// 		`^(\()?blah(?(1)(\)))$`,
// 		`^(\()?blah(?(1)(\)))$`,
// 		`^(\(+)?blah(?(1)(\)))$`,
// 		`^(\(+)?blah(?(1)(\)))$`,
// 		`(?(?!a)b|a)`,
// 		`(?(?=a)a|b)`,
// 		`(?=(a+?))(\1ab)`,
// 		`(\w+:)+`,
// 		`$(?<=^(a))`,
// 		`(?=(a+?))(\1ab)`,
// 		`([\w:]+::)?(\w+)$`,
// 		`([\w:]+::)?(\w+)$`,
// 		`^[^bcd]*(c+)`,
// 		`(a*)b+`,
// 		`([\w:]+::)?(\w+)$`,
// 		`([\w:]+::)?(\w+)$`,
// 		`^[^bcd]*(c+)`,
// 		`(?>a+)b`,
// 		`([[:]+)`,
// 		`([[=]+)`,
// 		`([[.]+)`,
// 		`[a[:]b[:c]`,
// 		`[a[:]b[:c]`,
// 		`((?>a+)b)`,
// 		`(?>(a+))b`,
// 		`((?>[^()]+)|\([^()]*\))+`,
// 		`(?<=x+)`,
// 		`\Z`,
// 		`\z`,
// 		`$`,
// 		`\Z`,
// 		`\z`,
// 		`$`,
// 		`\Z`,
// 		`\z`,
// 		`$`,
// 		`\Z`,
// 		`\z`,
// 		`$`,
// 		`\Z`,
// 		`\z`,
// 		`$`,
// 		`\Z`,
// 		`\z`,
// 		`$`,
// 		`a\Z`,
// 		`a$`,
// 		`a\Z`,
// 		`a\z`,
// 		`a$`,
// 		`a$`,
// 		`a\Z`,
// 		`a$`,
// 		`a\Z`,
// 		`a\z`,
// 		`a$`,
// 		`aa\Z`,
// 		`aa$`,
// 		`aa\Z`,
// 		`aa\z`,
// 		`aa$`,
// 		`aa$`,
// 		`aa\Z`,
// 		`aa$`,
// 		`aa\Z`,
// 		`aa\z`,
// 		`aa$`,
// 		`ab\Z`,
// 		`ab$`,
// 		`ab\Z`,
// 		`ab\z`,
// 		`ab$`,
// 		`ab$`,
// 		`ab\Z`,
// 		`ab$`,
// 		`ab\Z`,
// 		`ab\z`,
// 		`ab$`,
// 		`abb\Z`,
// 		`abb$`,
// 		`abb\Z`,
// 		`abb\z`,
// 		`abb$`,
// 		`abb$`,
// 		`abb\Z`,
// 		`abb$`,
// 		`abb\Z`,
// 		`abb\z`,
// 		`abb$`,
// 		`(^|x)(c)`,
// 		`round\(((?>[^()]+))\)`,
// 		`foo.bart`,
// 		`^d[x][x][x]`,
// 		`.X(.+)+X`,
// 		`.X(.+)+XX`,
// 		`.XX(.+)+X`,
// 		`.X(.+)+[X]`,
// 		`.X(.+)+[X][X]`,
// 		`.XX(.+)+[X]`,
// 		`.[X](.+)+[X]`,
// 		`.[X](.+)+[X][X]`,
// 		`.[X][X](.+)+[X]`,
// 		`tt+$`,
// 		`([\d-z]+)`,
// 		`([\d-\s]+)`,
// 		`(\d+\.\d+)`,
// 		`(\ba.{0,10}br)`,
// 		`\.c(pp|xx|c)?$`,
// 		`(\.c(pp|xx|c)?$)`,
// 		`^\S\s+aa$`,
// 		`(^|a)b`,
// 		`^([ab]*?)(b)?(c)$`,
// 		`^(?:.,){2}c`,
// 		`^(.,){2}c`,
// 		`^(?:[^,]*,){2}c`,
// 		`^([^,]*,){2}c`,
// 		`^([^,]*,){3}d`,
// 		`^([^,]*,){3,}d`,
// 		`^([^,]*,){0,3}d`,
// 		`^([^,]{1,3},){3}d`,
// 		`^([^,]{1,3},){3,}d`,
// 		`^([^,]{1,3},){0,3}d`,
// 		`^([^,]{1,},){3}d`,
// 		`^([^,]{1,},){3,}d`,
// 		`^([^,]{1,},){0,3}d`,
// 		`^([^,]{0,3},){3}d`,
// 		`^([^,]{0,3},){3,}d`,
// 		`^([^,]{0,3},){0,3}d`,
// 		`(?i)`,
// 		`(?!\A)x`,
// 		`^(a(b)?)+$`,
// 		`^(aa(bb)?)+$`,
// 		`^.{9}abc.*\n`,
// 		`^(a)?a$`,
// 		`^(a\1?)(a\1?)(a\2?)(a\3?)$`,
// 		`^(a\1?){4}$`,
// 		`^(0+)?(?:x(1))?`,
// 		`^([0-9a-fA-F]+)(?:x([0-9a-fA-F]+)?)(?:x([0-9a-fA-F]+))?`,
// 		`^(b+?|a){1,2}c`,
// 		`^(b+?|a){1,2}c`,
// 		`\((\w\. \w+)\)`,
// 		`((?:aaaa|bbbb)cccc)?`,
// 		`((?:aaaa|bbbb)cccc)?`,
// 		`^(foo)|(bar)$`,
// 		`^(foo)|(bar)$`,
// 		`b`,
// 		`bab`,
// 		`abb`,
// 		`b$`,
// 		`^a`,
// 		`^aaab`,
// 		`abb{2}`,
// 		`abb{1,2}`,
// 		`abb{1,2}`,
// 		`\Ab`,
// 		`\Abab$`,
// 		`b\Z`,
// 		`b\z`,
// 		`a\G`,
// 		`\Abaaa\G`,
// 		`\bc`,
// 		`\bc`,
// 		`\bc`,
// 		`\bc`,
// 		`\Bc`,
// 		`\Bc`,
// 		`\Bc`,
// 		`b(a?)b`,
// 		`b{4}`,
// 		`b\1aa(.)`,
// 		`^(a\1?){4}$`,
// 		`^([0-9a-fA-F]+)(?:x([0-9a-fA-F]+)?)(?:x([0-9a-fA-F]+))?`,
// 		`^(b+?|a){1,2}c`,
// 		`\((\w\. \w+)\)`,
// 		`((?:aaaa|bbbb)cccc)?`,
// 		`((?:aaaa|bbbb)cccc)?`,
// 		`(?<=a)b`,
// 		`(?<!c)b`,
// 		`(?<!c)b`,
// 		`(?<!c)b`,
// 		`a(?=d).`,
// 		`a(?=c|d).`,
// 		`ab*c`,
// 		`ab*bc`,
// 		`ab*bc`,
// 		`ab*bc`,
// 		`.{1}`,
// 		`.{3,4}`,
// 		`ab{0,}bc`,
// 		`ab+bc`,
// 		`ab+bc`,
// 		`ab{1,}bc`,
// 		`ab{1,3}bc`,
// 		`ab{3,4}bc`,
// 		`ab?bc`,
// 		`ab?bc`,
// 		`ab{0,1}bc`,
// 		`ab?c`,
// 		`ab{0,1}c`,
// 		`^abc$`,
// 		`^abc`,
// 		`abc$`,
// 		`^`,
// 		`$`,
// 		`a.c`,
// 		`a.c`,
// 		`a.*c`,
// 		`a[bc]d`,
// 		`a[b-d]e`,
// 		`a[b-d]`,
// 		`a[-b]`,
// 		`a[b-]`,
// 		`a]`,
// 		`a[]]b`,
// 		`a[^bc]d`,
// 		`a[^-b]c`,
// 		`a[^]b]c`,
// 		`\ba\b`,
// 		`\ba\b`,
// 		`\ba\b`,
// 		`\By\b`,
// 		`\by\B`,
// 		`\By\B`,
// 		`\w`,
// 		`\W`,
// 		`a\sb`,
// 		`a\Sb`,
// 		`\d`,
// 		`\D`,
// 		`[\w]`,
// 		`[\W]`,
// 		`a[\s]b`,
// 		`a[\S]b`,
// 		`[\d]`,
// 		`[\D]`,
// 		`ab|cd`,
// 		`ab|cd`,
// 		`()ef`,
// 		`a\(b`,
// 		`a\(*b`,
// 		`a\(*b`,
// 		`a\\b`,
// 		`((a))`,
// 		`(a)b(c)`,
// 		`a+b+c`,
// 		`a{1,}b{1,}c`,
// 		`a.+?c`,
// 		`(a+|b)*`,
// 		`(a+|b){0,}`,
// 		`(a+|b)+`,
// 		`(a+|b){1,}`,
// 		`(a+|b)?`,
// 		`(a+|b){0,1}`,
// 		`[^ab]*`,
// 		`a*`,
// 		`([abc])*d`,
// 		`([abc])*bcd`,
// 		`a|b|c|d|e`,
// 		`(a|b|c|d|e)f`,
// 		`abcd*efg`,
// 		`ab*`,
// 		`ab*`,
// 		`(ab|cd)e`,
// 		`[abhgefdc]ij`,
// 		`(abc|)ef`,
// 		`(a|b)c*d`,
// 		`(ab|ab*)bc`,
// 		`a([bc]*)c*`,
// 		`a([bc]*)(c*d)`,
// 		`a([bc]+)(c*d)`,
// 		`a([bc]*)(c+d)`,
// 		`a[bcd]*dcdcde`,
// 		`(ab|a)b*c`,
// 		`((a)(b)c)(d)`,
// 		`[a-zA-Z_][a-zA-Z0-9_]*`,
// 		`^a(bc+|b[eh])g|.h$`,
// 		`(bc+d$|ef*g.|h?i(j|k))`,
// 		`(bc+d$|ef*g.|h?i(j|k))`,
// 		`(bc+d$|ef*g.|h?i(j|k))`,
// 		`((((((((((a))))))))))`,
// 		`\10((((((((((a))))))))))`,
// 		`((((((((((a))))))))))!`,
// 		`(((((((((a)))))))))`,
// 		`multiple words`,
// 		`(.*)c(.*)`,
// 		`\((.*), (.*)\)`,
// 		`abcd`,
// 		`a(bc)d`,
// 		`a[-]?c`,
// 		`\1(abc)`,
// 		`\1([a-c]*)`,
// 		`(a)|\1`,
// 		`(([a-c])b*?\2)*`,
// 		`\((?>[^()]+|\((?<depth>)|\)(?<-depth>))*(?(depth)(?!))\)`,
// 		`^\((?>[^()]+|\((?<depth>)|\)(?<-depth>))*(?(depth)(?!))\)$`,
// 		`(((?<foo>\()[^()]*)+((?<bar-foo>\))[^()]*)+)+(?(foo)(?!))`,
// 		`^(((?<foo>\()[^()]*)+((?<bar-foo>\))[^()]*)+)+(?(foo)(?!))$`,
// 		`(((?<foo>\()[^()]*)+((?<bar-foo>\))[^()]*)+)+(?(foo)(?!))`,
// 		`(((?<foo>\()[^()]*)+((?<bar-foo>\))[^()]*)+)+(?(foo)(?!))`,
// 		`b`,
// 		`^((\[(?<NAME>[^\]]+)\])|(?<NAME>[^\.\[\]]+))$`,
// 		`^((\[(?<NAME>[^\]]+)\])|(?<NAME>[^\.\[\]]+))$`,
// 		`^((\[(?<NAME>[^\]]+)\])|(?<NAME>[^\.\[\]]+))$`,
// 		`^((\[(?<NAME>[^\]]+)\])|(?<NAME>[^\.\[\]]+))$`,
// 		`^((\[(?<NAME>[^\]]+)\])|(?<NAME>[^\.\[\]]+))$`,
// 		`^((\[(?<SCHEMA>[^\]]+)\])|(?<SCHEMA>[^\.\[\]]+))\s*\.\s*((\[(?<NAME>[^\]]+)\])|(?<NAME>[^\.\[\]]+))$`,
// 		`^((\[(?<SCHEMA>[^\]]+)\])|(?<SCHEMA>[^\.\[\]]+))\s*\.\s*((\[(?<NAME>[^\]]+)\])|(?<NAME>[^\.\[\]]+))$`,
// 		`^((\[(?<SCHEMA>[^\]]+)\])|(?<SCHEMA>[^\.\[\]]+))\s*\.\s*((\[(?<NAME>[^\]]+)\])|(?<NAME>[^\.\[\]]+))$`,
// 		`^((\[(?<SCHEMA>[^\]]+)\])|(?<SCHEMA>[^\.\[\]]+))\s*\.\s*((\[(?<NAME>[^\]]+)\])|(?<NAME>[^\.\[\]]+))$`,
// 		`^((\[(?<SCHEMA>[^\]]+)\])|(?<SCHEMA>[^\.\[\]]+))\s*\.\s*((\[(?<NAME>[^\]]+)\])|(?<NAME>[^\.\[\]]+))$`,
// 		`^((\[(?<SCHEMA>[^\]]+)\])|(?<SCHEMA>[^\.\[\]]+))\s*\.\s*((\[(?<NAME>[^\]]+)\])|(?<NAME>[^\.\[\]]+))$`,
// 		`^((\[(?<CATALOG>[^\]]+)\])|(?<CATALOG>[^\.\[\]]+))\s*\.\s*((\[(?<SCHEMA>[^\]]+)\])|(?<SCHEMA>[^\.\[\]]+))\s*\.\s*((\[(?<NAME>[^\]]+)\])|(?<NAME>[^\.\[\]]+))$`,
// 		`^((\[(?<CATALOG>[^\]]+)\])|(?<CATALOG>[^\.\[\]]+))\s*\.\s*((\[(?<SCHEMA>[^\]]+)\])|(?<SCHEMA>[^\.\[\]]+))\s*\.\s*((\[(?<NAME>[^\]]+)\])|(?<NAME>[^\.\[\]]+))$`,
// 		`^((\[(?<CATALOG>[^\]]+)\])|(?<CATALOG>[^\.\[\]]+))\s*\.\s*((\[(?<SCHEMA>[^\]]+)\])|(?<SCHEMA>[^\.\[\]]+))\s*\.\s*((\[(?<NAME>[^\]]+)\])|(?<NAME>[^\.\[\]]+))$`,
// 		`^((\[(?<CATALOG>[^\]]+)\])|(?<CATALOG>[^\.\[\]]+))\s*\.\s*((\[(?<NAME>[^\]]+)\])|(?<NAME>[^\.\[\]]+))$`,
// 		`^((\[(?<CATALOG>[^\]]+)\])|(?<CATALOG>[^\.\[\]]+))\s*\.\s*((\[(?<NAME>[^\]]+)\])|(?<NAME>[^\.\[\]]+))$`,
// 		`^((\[(?<SCHEMA>[^\]]+)\])|(?<SCHEMA>[^\.\[\]]+))\s*\.\s*((\[(?<CATALOG>[^\]]+)\])|(?<CATALOG>[^\.\[\]]+))\s*\.\s*((\[(?<NAME>[^\]]+)\])|(?<NAME>[^\.\[\]]+))$`,
// 		`^((\[(?<SCHEMA>[^\]]+)\])|(?<SCHEMA>[^\.\[\]]+))\s*\.\s*((\[(?<CATALOG>[^\]]+)\])|(?<CATALOG>[^\.\[\]]+))\s*\.\s*((\[(?<NAME>[^\]]+)\])|(?<NAME>[^\.\[\]]+))$`,
// 		`^((\[(?<ColName>.+)\])|(?<ColName>\S+))([ ]+(?<Order>ASC|DESC))?$`,
// 		`a{1,2147483647}`,
// 		`^((\[(?<NAME>[^\]]+)\])|(?<NAME>[^\.\[\]]+))$`,
// 	}

// 	for _, s := range cases {
// 		re, err := regexp2.Compile(s, regexp2.RE2)
// 		require.Nil(t, err)

// 		g := NewGenerator()
// 		data, err := g.Generate(NewState(true, 3, nil, 0), s, regexp2.RE2)
// 		require.Nil(t, err)
// 		result, err := re.MatchString(data)
// 		require.Nil(t, err)
// 		require.True(t, result)
// 	}
// }
