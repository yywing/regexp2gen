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
	// Ref, Bol, Eol, Boundary, Nonboundary

	re, err := regexp2.Compile(s, regexp2.RE2)
	require.Nil(t, err)

	g := NewGenerator(s)
	data, err := g.Generate(NewState(true, 3, nil, 0), s)
	require.Nil(t, err)
	result, err := re.MatchString(data)
	require.Nil(t, err)
	require.True(t, result)
}
