package regexp2gen

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"math"
	"math/rand"

	"github.com/dlclark/regexp2/syntax"
)

const runeRangeEnd = 0x10ffff
const printableChars = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ\n\r"

var printableCharsNoNL = printableChars[:len(printableChars)-2]

type state struct {
	debug bool

	rand *rand.Rand

	limit int
	// use for .
	chars []rune
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
		debug: debug,
		rand:  r,
		limit: limit,
		chars: chars,
	}
}

type Generator struct{}

func opcodeSize(op syntax.InstOp) int {
	op &= syntax.Mask

	switch op {
	case syntax.Nothing, syntax.Bol, syntax.Eol, syntax.Boundary, syntax.Nonboundary, syntax.ECMABoundary, syntax.NonECMABoundary, syntax.Beginning, syntax.Start, syntax.EndZ,
		syntax.End, syntax.Nullmark, syntax.Setmark, syntax.Getmark, syntax.Setjump, syntax.Backjump, syntax.Forejump, syntax.Stop:
		return 1

	case syntax.One, syntax.Notone, syntax.Multi, syntax.Ref, syntax.Testref, syntax.Goto, syntax.Nullcount, syntax.Setcount, syntax.Lazybranch, syntax.Branchmark, syntax.Lazybranchmark,
		syntax.Prune, syntax.Set:
		return 2

	case syntax.Capturemark, syntax.Branchcount, syntax.Lazybranchcount, syntax.Onerep, syntax.Notonerep, syntax.Oneloop, syntax.Notoneloop, syntax.Onelazy, syntax.Notonelazy,
		syntax.Setlazy, syntax.Setrep, syntax.Setloop:
		return 3

	default:
		panic(fmt.Errorf("Unexpected op code: %v", op))
	}
}

func (g *Generator) printCode(c *syntax.Code) {
	fmt.Println(c.Codes)
	buf := &bytes.Buffer{}
	for i := 0; i < len(c.Codes); i += opcodeSize(syntax.InstOp(c.Codes[i])) {
		fmt.Fprintln(buf, c.OpcodeDescription(i))
	}
	fmt.Println(buf.String())
}

func (g *Generator) Generate(s *state, re string) (string, error) {
	if s.debug {
		fmt.Println(re)
	}

	tree, err := syntax.Parse(re, syntax.RE2)
	if err != nil {
		return "", err
	}
	c, err := syntax.Write(tree)
	if err != nil {
		return "", err
	}

	return g.generate(s, c)
}

func (g *Generator) generate(s *state, c *syntax.Code) (string, error) {
	if s.debug {
		g.printCode(c)
	}

	buf := &bytes.Buffer{}
	index := 0
	for index < len(c.Codes) {
		op := syntax.InstOp(c.Codes[index])
		size := opcodeSize(op)
		switch op {
		case syntax.Onelazy, syntax.Notonelazy, syntax.Setlazy:
			//{2,4}? -> rep(Rep = 2), lazy(Rep = 2)
			//do nothing
		case syntax.One, syntax.Onerep, syntax.Oneloop:
			r := rune(c.Codes[index+1])
			var length int
			if size == 2 {
				// one
				length = 1
			} else {
				length = c.Codes[index+2]
				if length == math.MaxInt32 {
					//{2,4} -> rep(Rep = 2), loop(Rep = 2)
					//add all loop, with 4 char
					//{2,}, this will just write 2 char
					//WARN: a{2,}[a-z] if we get `aaa` will fail to gen.
					length = 0
				}
			}
			result := s.randomRunes([]rune{r}, length)
			for _, j := range result {
				buf.WriteRune(j)
			}
		case syntax.Notone, syntax.Notonerep, syntax.Notoneloop:
			var length int
			if size == 2 {
				// notone
				length = 1
			} else {
				length = c.Codes[index+2]
				if length == math.MaxInt32 {
					//{2,4} -> rep(Rep = 2), loop(Rep = 2)
					//add all loop, with 4 char
					//{2,}, this will just write 2 char
					//WARN: a{2,}[a-z] if we get `aaa` will fail to gen.
					length = 0
				}
			}
			exclude := rune(c.Codes[index+1])
			// get possible chars
			possibleChars := []rune{}
			if exclude == '.' {
				possibleChars = s.chars
			} else {
				for j := 0; j < len(s.chars); j++ {
					c := s.chars[j]
					if c != exclude {
						possibleChars = append(possibleChars, c)
					}
				}
			}
			result := s.randomRunes(possibleChars, length)
			for _, j := range result {
				buf.WriteRune(j)
			}
		case syntax.Set, syntax.Setrep, syntax.Setloop:
			charSet := c.Sets[c.Codes[index+1]]
			// get possible chars
			possibleChars := []rune{}
			for j := 0; j < len(s.chars); j++ {
				c := s.chars[j]
				if charSet.CharIn(c) {
					possibleChars = append(possibleChars, c)
				}
			}
			if len(possibleChars) == 0 {
				return "", fmt.Errorf("code has no suitable chars: %s", c.OpcodeDescription(index))
			}
			var length int
			if size == 2 {
				// set
				length = 1
			} else {
				length = c.Codes[index+2]
				if length == math.MaxInt32 {
					//{2,4} -> rep(Rep = 2), loop(Rep = 2)
					//add all loop, with 4 char
					//{2,}, this will just write 2 char
					//WARN: a{2,}[a-z] if we get `aaa` will fail to gen.
					length = 0
				}
			}
			result := s.randomRunes(possibleChars, length)
			for _, j := range result {
				buf.WriteRune(j)
			}
		case syntax.Multi:
			fmt.Fprintln(buf, string(c.Strings[c.Codes[index+1]]))
		default:
		}
		index += size
	}
	if s.debug {
		fmt.Println(hex.Dump(buf.Bytes()))
	}

	return buf.String(), nil
}

// create a new generator
func NewGenerator(regex string) *Generator {
	return &Generator{}
}
