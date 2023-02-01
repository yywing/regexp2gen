package regexp2gen

import "math/rand"

const printableChars = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ-_ \\\n\r"

var printableCharsNoNL = printableChars[:len(printableChars)-2]

var defaultBoundary = ' '

type state struct {
	debug bool

	rand *rand.Rand

	limit int
	// use for .
	chars []rune

	boundary rune
}

func (s *state) randomRunes(chars []rune, length int) []rune {
	result := []rune{}
	for j := 0; j < length; j++ {
		r := chars[s.rand.Intn(len(chars))]
		result = append(result, r)
	}
	return result
}

func NewState(debug bool, limit int, chars []rune, seed int64) *state {
	r := rand.New(rand.NewSource(seed))

	if chars == nil {
		chars = []rune(printableCharsNoNL)
	}

	return &state{
		debug:    debug,
		rand:     r,
		limit:    limit,
		chars:    chars,
		boundary: defaultBoundary,
	}
}
