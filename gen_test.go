package regexp2gen

import (
	"encoding/hex"
	"fmt"
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

	g := NewGenerator(s)
	data, err := g.Generate(NewState(true, 3, nil, 0), s)
	require.Nil(t, err)
	result, err := re.MatchString(data)
	require.Nil(t, err)
	require.True(t, result)
}

func TestReplace2(t *testing.T) {
	// notoneloop
	s := `^Google\nApple\Z`

	re, err := regexp2.Compile(s, regexp2.Singleline|regexp2.RE2)
	require.Nil(t, err)

	m, err := re.FindStringMatch("Google\nApple\n")
	require.Nil(t, err)
	if m != nil {
		fmt.Println(hex.Dump([]byte(m.String())))
	} else {
		fmt.Println("not match")
	}

}
