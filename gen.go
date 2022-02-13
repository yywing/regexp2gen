package regexp2gen

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"math"

	"github.com/dlclark/regexp2"
	"github.com/dlclark/regexp2/syntax"
)

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
		panic(fmt.Sprintf("Unexpected op code: %v", op))
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

func (g *Generator) Generate(s *state, re string, op regexp2.RegexOptions) (string, error) {
	if s.debug {
		fmt.Println(re)
	}

	reg, err := regexp2.Compile(re, op)
	if err != nil {
		return "", err
	}

	tree, err := syntax.Parse(re, syntax.RegexOptions(op))
	if err != nil {
		return "", err
	}
	c, err := syntax.Write(tree)
	if err != nil {
		return "", err
	}

	result, err := g.generate(s, c)
	if err != nil {
		return "", err
	}

	ok, err := reg.MatchString(result)
	if err != nil {
		return "", err
	}
	if !ok {
		return "", errors.New("generate string fail")
	}

	return result, nil
}

/*
	TODO： 这里只实现了简单的罗列，没有考虑一些非匹配和匹配之间相互影响的问题
*/
func (g *Generator) generate(s *state, c *syntax.Code) (string, error) {
	if s.debug {
		g.printCode(c)
	}

	buf := NewBuffer()
	index := 0

	for index < len(c.Codes) {
		op := syntax.InstOp(c.Codes[index])
		size := opcodeSize(op)
		op &= syntax.Mask

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
			for _, r := range c.Strings[c.Codes[index+1]] {
				buf.WriteRune(r)
			}
		case syntax.Ref:
			refIndex := c.Codes[index+1]
			groupBuffer, ok := buf.Getmark(refIndex)
			if !ok {
				return "", fmt.Errorf("ref get index err: %d", refIndex)
			}
			_, err := buf.WriteAll(groupBuffer.Bytes())
			if err != nil {
				return "", err
			}

		/*
			input: end sends endure lender
			\bend\b -> (end)
			\Bend\B -> s(end)s, l(end)er
		*/
		case syntax.Boundary:
			// TODO: 无脑写
			buf.WriteRune(s.boundary)
		case syntax.Nonboundary:
			// TODO: 无脑写
			buf.WriteRune(s.randomRunes(s.chars, 1)[0])
		case syntax.ECMABoundary:
			// TODO: 无脑写
			buf.WriteRune(s.boundary)
		case syntax.NonECMABoundary:
			// TODO: 无脑写
			buf.WriteRune(s.randomRunes(s.chars, 1)[0])

		/*
			^ Matches the beginning of a line.

			$ Matches the end of a line.

			\A Matches the beginning of the string.

			\z Matches the end of the string.

			\Z Matches the end of the string unless the string ends with a "\n", in which case it matches just before the "\n".

			DOTALL: ^ = \A, $= \Z
			input: Google\nApple
			^Google\nApple$   -> Google\nApple
			\AGoogle\nApple\z -> Google\nApple
			\AGoogle\nApple\Z -> Google\nApple
			input: Google\nApple\n
			^Google\nApple$   -> Google\nApple
			\AGoogle\nApple\z ->
			\AGoogle\nApple\Z -> Google\nApple
			input: Google\nApple\n\n
			^Google\nApple$   ->
			\AGoogle\nApple\z ->
			\AGoogle\nApple\Z ->
			MULTILINE:
			^Google\nApple$   -> Google\nApple
			\AGoogle\nApple\z -> Google\nApple
			\AGoogle\nApple\Z -> Google\nApple
			input: Google\nApple\n
			^Google\nApple$   -> Google\nApple
			\AGoogle\nApple\z ->
			\AGoogle\nApple\Z -> Google\nApple
			input: Google\nApple\n\n
			^Google\nApple$   -> Google\nApple
			\AGoogle\nApple\z ->
			\AGoogle\nApple\Z ->
		*/
		case syntax.Bol:
		case syntax.Eol:
		case syntax.Beginning:
		case syntax.Start:
		case syntax.EndZ:
		case syntax.End:
		case syntax.Nothing:

		case syntax.Setmark:
			buf.Setmark()
		case syntax.Capturemark:
			refIndex := c.Codes[index+1]
			// TODO: 这里还有一个参数, 不知道是用来干啥的， unidex？？ 非捕获么？
			err := buf.Backmark(true, refIndex)
			if err != nil {
				return "", err
			}
		case syntax.Getmark:
			err := buf.Backmark(false, -1)
			if err != nil {
				return "", err
			}

		case syntax.Setjump:
		case syntax.Forejump:
		case syntax.Backjump:
			/*
				TODO: 这个地方还没实现。
				一旦出现这个证明出现了 ?!， 需要生成一个不符合其中正则的内容
				需要把上面能产生实际内容的字符串生成的 case 都写一个否定逻辑然后在这里使用一下
			*/

		case syntax.Lazybranch:
		case syntax.Branchmark:
		case syntax.Lazybranchmark:
		case syntax.Nullcount:
		case syntax.Setcount:
		case syntax.Branchcount:
		case syntax.Lazybranchcount:
		case syntax.Nullmark:
		case syntax.Testref:
		case syntax.Goto:

		case syntax.Prune:
		case syntax.Stop:
		default:
			return "", fmt.Errorf("unknown code %d", op)
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
