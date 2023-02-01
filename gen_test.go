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
	// TODO: 这个还没有实现
	t.Skip()
	cases := []string{
		`(?!gg)aa`,
		`(?<!gg)aa`,
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
