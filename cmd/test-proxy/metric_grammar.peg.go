package main

import (
	"fmt"
	"math"
	"sort"
	"strconv"
)

const endSymbol rune = 1114112

/* The rule types inferred from the grammar are below. */
type pegRule uint8

const (
	ruleUnknown pegRule = iota
	ruleMetricGrammar
	ruletags
	ruletag
	ruletagName
	ruletagValue
	ruletimestamp
	rulemetricValue
	rulemetricName
	ruleALNUM
	ruleDIGIT
	ruleAction0
	rulePegText
	ruleAction1
	ruleAction2
	ruleAction3
	ruleAction4
	ruleAction5

	rulePre
	ruleIn
	ruleSuf
)

var rul3s = [...]string{
	"Unknown",
	"MetricGrammar",
	"tags",
	"tag",
	"TagName",
	"TagValue",
	"timestamp",
	"metricValue",
	"MetricName",
	"ALNUM",
	"DIGIT",
	"Action0",
	"PegText",
	"Action1",
	"Action2",
	"Action3",
	"Action4",
	"Action5",

	"Pre_",
	"_In_",
	"_Suf",
}

type node32 struct {
	token32
	up, next *node32
}

func (node *node32) print(depth int, buffer string) {
	for node != nil {
		for c := 0; c < depth; c++ {
			fmt.Printf(" ")
		}
		fmt.Printf("\x1B[34m%v\x1B[m %v\n", rul3s[node.pegRule], strconv.Quote(string(([]rune(buffer)[node.begin:node.end]))))
		if node.up != nil {
			node.up.print(depth+1, buffer)
		}
		node = node.next
	}
}

func (node *node32) Print(buffer string) {
	node.print(0, buffer)
}

type element struct {
	node *node32
	down *element
}

/* ${@} bit structure for abstract syntax tree */
type token32 struct {
	pegRule
	begin, end, next uint32
}

func (t *token32) isZero() bool {
	return t.pegRule == ruleUnknown && t.begin == 0 && t.end == 0 && t.next == 0
}

func (t *token32) isParentOf(u token32) bool {
	return t.begin <= u.begin && t.end >= u.end && t.next > u.next
}

func (t *token32) getToken32() token32 {
	return token32{pegRule: t.pegRule, begin: uint32(t.begin), end: uint32(t.end), next: uint32(t.next)}
}

func (t *token32) String() string {
	return fmt.Sprintf("\x1B[34m%v\x1B[m %v %v %v", rul3s[t.pegRule], t.begin, t.end, t.next)
}

type tokens32 struct {
	tree    []token32
	ordered [][]token32
}

func (t *tokens32) trim(length int) {
	t.tree = t.tree[0:length]
}

func (t *tokens32) Print() {
	for _, token := range t.tree {
		fmt.Println(token.String())
	}
}

func (t *tokens32) Order() [][]token32 {
	if t.ordered != nil {
		return t.ordered
	}

	depths := make([]int32, 1, math.MaxInt16)
	for i, token := range t.tree {
		if token.pegRule == ruleUnknown {
			t.tree = t.tree[:i]
			break
		}
		depth := int(token.next)
		if length := len(depths); depth >= length {
			depths = depths[:depth+1]
		}
		depths[depth]++
	}
	depths = append(depths, 0)

	ordered, pool := make([][]token32, len(depths)), make([]token32, len(t.tree)+len(depths))
	for i, depth := range depths {
		depth++
		ordered[i], pool, depths[i] = pool[:depth], pool[depth:], 0
	}

	for i, token := range t.tree {
		depth := token.next
		token.next = uint32(i)
		ordered[depth][depths[depth]] = token
		depths[depth]++
	}
	t.ordered = ordered
	return ordered
}

type state32 struct {
	token32
	depths []int32
	leaf   bool
}

func (t *tokens32) AST() *node32 {
	tokens := t.Tokens()
	stack := &element{node: &node32{token32: <-tokens}}
	for token := range tokens {
		if token.begin == token.end {
			continue
		}
		node := &node32{token32: token}
		for stack != nil && stack.node.begin >= token.begin && stack.node.end <= token.end {
			stack.node.next = node.up
			node.up = stack.node
			stack = stack.down
		}
		stack = &element{node: node, down: stack}
	}
	return stack.node
}

func (t *tokens32) PreOrder() (<-chan state32, [][]token32) {
	s, ordered := make(chan state32, 6), t.Order()
	go func() {
		var states [8]state32
		for i := range states {
			states[i].depths = make([]int32, len(ordered))
		}
		depths, state, depth := make([]int32, len(ordered)), 0, 1
		write := func(t token32, leaf bool) {
			S := states[state]
			state, S.pegRule, S.begin, S.end, S.next, S.leaf = (state+1)%8, t.pegRule, t.begin, t.end, uint32(depth), leaf
			copy(S.depths, depths)
			s <- S
		}

		states[state].token32 = ordered[0][0]
		depths[0]++
		state++
		a, b := ordered[depth-1][depths[depth-1]-1], ordered[depth][depths[depth]]
	depthFirstSearch:
		for {
			for {
				if i := depths[depth]; i > 0 {
					if c, j := ordered[depth][i-1], depths[depth-1]; a.isParentOf(c) &&
						(j < 2 || !ordered[depth-1][j-2].isParentOf(c)) {
						if c.end != b.begin {
							write(token32{pegRule: ruleIn, begin: c.end, end: b.begin}, true)
						}
						break
					}
				}

				if a.begin < b.begin {
					write(token32{pegRule: rulePre, begin: a.begin, end: b.begin}, true)
				}
				break
			}

			next := depth + 1
			if c := ordered[next][depths[next]]; c.pegRule != ruleUnknown && b.isParentOf(c) {
				write(b, false)
				depths[depth]++
				depth, a, b = next, b, c
				continue
			}

			write(b, true)
			depths[depth]++
			c, parent := ordered[depth][depths[depth]], true
			for {
				if c.pegRule != ruleUnknown && a.isParentOf(c) {
					b = c
					continue depthFirstSearch
				} else if parent && b.end != a.end {
					write(token32{pegRule: ruleSuf, begin: b.end, end: a.end}, true)
				}

				depth--
				if depth > 0 {
					a, b, c = ordered[depth-1][depths[depth-1]-1], a, ordered[depth][depths[depth]]
					parent = a.isParentOf(b)
					continue
				}

				break depthFirstSearch
			}
		}

		close(s)
	}()
	return s, ordered
}

func (t *tokens32) PrintSyntax() {
	tokens, ordered := t.PreOrder()
	max := -1
	for token := range tokens {
		if !token.leaf {
			fmt.Printf("%v", token.begin)
			for i, leaf, depths := 0, int(token.next), token.depths; i < leaf; i++ {
				fmt.Printf(" \x1B[36m%v\x1B[m", rul3s[ordered[i][depths[i]-1].pegRule])
			}
			fmt.Printf(" \x1B[36m%v\x1B[m\n", rul3s[token.pegRule])
		} else if token.begin == token.end {
			fmt.Printf("%v", token.begin)
			for i, leaf, depths := 0, int(token.next), token.depths; i < leaf; i++ {
				fmt.Printf(" \x1B[31m%v\x1B[m", rul3s[ordered[i][depths[i]-1].pegRule])
			}
			fmt.Printf(" \x1B[31m%v\x1B[m\n", rul3s[token.pegRule])
		} else {
			for c, end := token.begin, token.end; c < end; c++ {
				if i := int(c); max+1 < i {
					for j := max; j < i; j++ {
						fmt.Printf("skip %v %v\n", j, token.String())
					}
					max = i
				} else if i := int(c); i <= max {
					for j := i; j <= max; j++ {
						fmt.Printf("dupe %v %v\n", j, token.String())
					}
				} else {
					max = int(c)
				}
				fmt.Printf("%v", c)
				for i, leaf, depths := 0, int(token.next), token.depths; i < leaf; i++ {
					fmt.Printf(" \x1B[34m%v\x1B[m", rul3s[ordered[i][depths[i]-1].pegRule])
				}
				fmt.Printf(" \x1B[34m%v\x1B[m\n", rul3s[token.pegRule])
			}
			fmt.Printf("\n")
		}
	}
}

func (t *tokens32) PrintSyntaxTree(buffer string) {
	tokens, _ := t.PreOrder()
	for token := range tokens {
		for c := 0; c < int(token.next); c++ {
			fmt.Printf(" ")
		}
		fmt.Printf("\x1B[34m%v\x1B[m %v\n", rul3s[token.pegRule], strconv.Quote(string(([]rune(buffer)[token.begin:token.end]))))
	}
}

func (t *tokens32) Add(rule pegRule, begin, end, depth uint32, index int) {
	t.tree[index] = token32{pegRule: rule, begin: uint32(begin), end: uint32(end), next: uint32(depth)}
}

func (t *tokens32) Tokens() <-chan token32 {
	s := make(chan token32, 16)
	go func() {
		for _, v := range t.tree {
			s <- v.getToken32()
		}
		close(s)
	}()
	return s
}

func (t *tokens32) Error() []token32 {
	ordered := t.Order()
	length := len(ordered)
	tokens, length := make([]token32, length), length-1
	for i := range tokens {
		o := ordered[length-i]
		if len(o) > 1 {
			tokens[i] = o[len(o)-2].getToken32()
		}
	}
	return tokens
}

func (t *tokens32) Expand(index int) {
	tree := t.tree
	if index >= len(tree) {
		expanded := make([]token32, 2*len(tree))
		copy(expanded, tree)
		t.tree = expanded
	}
}

type MetricGrammar struct {
	Name       string
	Value      string
	Timestamp  string
	Tags       map[string]string
	currentTag string

	Buffer string
	buffer []rune
	rules  [18]func() bool
	Parse  func(rule ...int) error
	Reset  func()
	Pretty bool
	tokens32
}

type textPosition struct {
	line, symbol int
}

type textPositionMap map[int]textPosition

func translatePositions(buffer []rune, positions []int) textPositionMap {
	length, translations, j, line, symbol := len(positions), make(textPositionMap, len(positions)), 0, 1, 0
	sort.Ints(positions)

search:
	for i, c := range buffer {
		if c == '\n' {
			line, symbol = line+1, 0
		} else {
			symbol++
		}
		if i == positions[j] {
			translations[positions[j]] = textPosition{line, symbol}
			for j++; j < length; j++ {
				if i != positions[j] {
					continue search
				}
			}
			break search
		}
	}

	return translations
}

type parseError struct {
	p   *MetricGrammar
	max token32
}

func (e *parseError) Error() string {
	tokens, error := []token32{e.max}, "\n"
	positions, p := make([]int, 2*len(tokens)), 0
	for _, token := range tokens {
		positions[p], p = int(token.begin), p+1
		positions[p], p = int(token.end), p+1
	}
	translations := translatePositions(e.p.buffer, positions)
	format := "parse error near %v (line %v symbol %v - line %v symbol %v):\n%v\n"
	if e.p.Pretty {
		format = "parse error near \x1B[34m%v\x1B[m (line %v symbol %v - line %v symbol %v):\n%v\n"
	}
	for _, token := range tokens {
		begin, end := int(token.begin), int(token.end)
		error += fmt.Sprintf(format,
			rul3s[token.pegRule],
			translations[begin].line, translations[begin].symbol,
			translations[end].line, translations[end].symbol,
			strconv.Quote(string(e.p.buffer[begin:end])))
	}

	return error
}

func (p *MetricGrammar) PrintSyntaxTree() {
	p.tokens32.PrintSyntaxTree(p.Buffer)
}

func (p *MetricGrammar) Highlighter() {
	p.PrintSyntax()
}

func (p *MetricGrammar) Execute() {
	buffer, _buffer, text, begin, end := p.Buffer, p.buffer, "", 0, 0
	for token := range p.Tokens() {
		switch token.pegRule {

		case rulePegText:
			begin, end = int(token.begin), int(token.end)
			text = string(_buffer[begin:end])

		case ruleAction0:
			p.Tags = map[string]string{}
		case ruleAction1:
			p.currentTag = text
		case ruleAction2:
			p.Tags[p.currentTag] = text
		case ruleAction3:
			p.Timestamp = text
		case ruleAction4:
			p.Value = text
		case ruleAction5:
			p.Name = text

		}
	}
	_, _, _, _, _ = buffer, _buffer, text, begin, end
}

func (p *MetricGrammar) Init() {
	p.buffer = []rune(p.Buffer)
	if len(p.buffer) == 0 || p.buffer[len(p.buffer)-1] != endSymbol {
		p.buffer = append(p.buffer, endSymbol)
	}

	tree := tokens32{tree: make([]token32, math.MaxInt16)}
	var max token32
	position, depth, tokenIndex, buffer, _rules := uint32(0), uint32(0), 0, p.buffer, p.rules

	p.Parse = func(rule ...int) error {
		r := 1
		if len(rule) > 0 {
			r = rule[0]
		}
		matches := p.rules[r]()
		p.tokens32 = tree
		if matches {
			p.trim(tokenIndex)
			return nil
		}
		return &parseError{p, max}
	}

	p.Reset = func() {
		position, tokenIndex, depth = 0, 0, 0
	}

	add := func(rule pegRule, begin uint32) {
		tree.Expand(tokenIndex)
		tree.Add(rule, begin, position, depth, tokenIndex)
		tokenIndex++
		if begin != position && position > max.end {
			max = token32{rule, begin, position, depth}
		}
	}

	matchDot := func() bool {
		if buffer[position] != endSymbol {
			position++
			return true
		}
		return false
	}

	/*matchChar := func(c byte) bool {
		if buffer[position] == c {
			position++
			return true
		}
		return false
	}*/

	/*matchRange := func(lower byte, upper byte) bool {
		if c := buffer[position]; c >= lower && c <= upper {
			position++
			return true
		}
		return false
	}*/

	_rules = [...]func() bool{
		nil,
		/* 0 MetricGrammar <- <(MetricName ' ' metricValue (' ' timestamp)? ' ' tags !.)> */
		func() bool {
			position0, tokenIndex0, depth0 := position, tokenIndex, depth
			{
				position1 := position
				depth++
				{
					position2 := position
					depth++
					{
						position3, tokenIndex3, depth3 := position, tokenIndex, depth
						if buffer[position] != rune('"') {
							goto l4
						}
						position++
						{
							position5 := position
							depth++
							{
								position6, tokenIndex6, depth6 := position, tokenIndex, depth
								if buffer[position] != rune('~') {
									goto l6
								}
								position++
								goto l7
							l6:
								position, tokenIndex, depth = position6, tokenIndex6, depth6
							}
						l7:
							{
								position10, tokenIndex10, depth10 := position, tokenIndex, depth
								if !_rules[ruleALNUM]() {
									goto l11
								}
								goto l10
							l11:
								position, tokenIndex, depth = position10, tokenIndex10, depth10
								{
									switch buffer[position] {
									case ',':
										if buffer[position] != rune(',') {
											goto l4
										}
										position++
										break
									case '/':
										if buffer[position] != rune('/') {
											goto l4
										}
										position++
										break
									case '.':
										if buffer[position] != rune('.') {
											goto l4
										}
										position++
										break
									case '_':
										if buffer[position] != rune('_') {
											goto l4
										}
										position++
										break
									case '-':
										if buffer[position] != rune('-') {
											goto l4
										}
										position++
										break
									default:
										if buffer[position] != rune('~') {
											goto l4
										}
										position++
										break
									}
								}

							}
						l10:
						l8:
							{
								position9, tokenIndex9, depth9 := position, tokenIndex, depth
								{
									position13, tokenIndex13, depth13 := position, tokenIndex, depth
									if !_rules[ruleALNUM]() {
										goto l14
									}
									goto l13
								l14:
									position, tokenIndex, depth = position13, tokenIndex13, depth13
									{
										switch buffer[position] {
										case ',':
											if buffer[position] != rune(',') {
												goto l9
											}
											position++
											break
										case '/':
											if buffer[position] != rune('/') {
												goto l9
											}
											position++
											break
										case '.':
											if buffer[position] != rune('.') {
												goto l9
											}
											position++
											break
										case '_':
											if buffer[position] != rune('_') {
												goto l9
											}
											position++
											break
										case '-':
											if buffer[position] != rune('-') {
												goto l9
											}
											position++
											break
										default:
											if buffer[position] != rune('~') {
												goto l9
											}
											position++
											break
										}
									}

								}
							l13:
								goto l8
							l9:
								position, tokenIndex, depth = position9, tokenIndex9, depth9
							}
							depth--
							add(rulePegText, position5)
						}
						if buffer[position] != rune('"') {
							goto l4
						}
						position++
						goto l3
					l4:
						position, tokenIndex, depth = position3, tokenIndex3, depth3
						{
							position16 := position
							depth++
							{
								position17, tokenIndex17, depth17 := position, tokenIndex, depth
								if buffer[position] != rune('~') {
									goto l17
								}
								position++
								goto l18
							l17:
								position, tokenIndex, depth = position17, tokenIndex17, depth17
							}
						l18:
							{
								position21, tokenIndex21, depth21 := position, tokenIndex, depth
								if !_rules[ruleALNUM]() {
									goto l22
								}
								goto l21
							l22:
								position, tokenIndex, depth = position21, tokenIndex21, depth21
								{
									switch buffer[position] {
									case ',':
										if buffer[position] != rune(',') {
											goto l0
										}
										position++
										break
									case '/':
										if buffer[position] != rune('/') {
											goto l0
										}
										position++
										break
									case '.':
										if buffer[position] != rune('.') {
											goto l0
										}
										position++
										break
									case '_':
										if buffer[position] != rune('_') {
											goto l0
										}
										position++
										break
									case '-':
										if buffer[position] != rune('-') {
											goto l0
										}
										position++
										break
									default:
										if buffer[position] != rune('~') {
											goto l0
										}
										position++
										break
									}
								}

							}
						l21:
						l19:
							{
								position20, tokenIndex20, depth20 := position, tokenIndex, depth
								{
									position24, tokenIndex24, depth24 := position, tokenIndex, depth
									if !_rules[ruleALNUM]() {
										goto l25
									}
									goto l24
								l25:
									position, tokenIndex, depth = position24, tokenIndex24, depth24
									{
										switch buffer[position] {
										case ',':
											if buffer[position] != rune(',') {
												goto l20
											}
											position++
											break
										case '/':
											if buffer[position] != rune('/') {
												goto l20
											}
											position++
											break
										case '.':
											if buffer[position] != rune('.') {
												goto l20
											}
											position++
											break
										case '_':
											if buffer[position] != rune('_') {
												goto l20
											}
											position++
											break
										case '-':
											if buffer[position] != rune('-') {
												goto l20
											}
											position++
											break
										default:
											if buffer[position] != rune('~') {
												goto l20
											}
											position++
											break
										}
									}

								}
							l24:
								goto l19
							l20:
								position, tokenIndex, depth = position20, tokenIndex20, depth20
							}
							depth--
							add(rulePegText, position16)
						}
					}
				l3:
					{
						add(ruleAction5, position)
					}
					depth--
					add(rulemetricName, position2)
				}
				if buffer[position] != rune(' ') {
					goto l0
				}
				position++
				{
					position28 := position
					depth++
					{
						position29 := position
						depth++
						{
							position30, tokenIndex30, depth30 := position, tokenIndex, depth
							if buffer[position] != rune('-') {
								goto l30
							}
							position++
							goto l31
						l30:
							position, tokenIndex, depth = position30, tokenIndex30, depth30
						}
					l31:
						if !_rules[ruleDIGIT]() {
							goto l0
						}
					l32:
						{
							position33, tokenIndex33, depth33 := position, tokenIndex, depth
							if !_rules[ruleDIGIT]() {
								goto l33
							}
							goto l32
						l33:
							position, tokenIndex, depth = position33, tokenIndex33, depth33
						}
						{
							position34, tokenIndex34, depth34 := position, tokenIndex, depth
							if buffer[position] != rune('.') {
								goto l34
							}
							position++
							if !_rules[ruleDIGIT]() {
								goto l34
							}
						l36:
							{
								position37, tokenIndex37, depth37 := position, tokenIndex, depth
								if !_rules[ruleDIGIT]() {
									goto l37
								}
								goto l36
							l37:
								position, tokenIndex, depth = position37, tokenIndex37, depth37
							}
							goto l35
						l34:
							position, tokenIndex, depth = position34, tokenIndex34, depth34
						}
					l35:
						depth--
						add(rulePegText, position29)
					}
					{
						add(ruleAction4, position)
					}
					depth--
					add(rulemetricValue, position28)
				}
				{
					position39, tokenIndex39, depth39 := position, tokenIndex, depth
					if buffer[position] != rune(' ') {
						goto l39
					}
					position++
					{
						position41 := position
						depth++
						{
							position42 := position
							depth++
							if !_rules[ruleDIGIT]() {
								goto l39
							}
							if !_rules[ruleDIGIT]() {
								goto l39
							}
							if !_rules[ruleDIGIT]() {
								goto l39
							}
							if !_rules[ruleDIGIT]() {
								goto l39
							}
							if !_rules[ruleDIGIT]() {
								goto l39
							}
							if !_rules[ruleDIGIT]() {
								goto l39
							}
							if !_rules[ruleDIGIT]() {
								goto l39
							}
							if !_rules[ruleDIGIT]() {
								goto l39
							}
							if !_rules[ruleDIGIT]() {
								goto l39
							}
							if !_rules[ruleDIGIT]() {
								goto l39
							}
							{
								position43, tokenIndex43, depth43 := position, tokenIndex, depth
								if !_rules[ruleDIGIT]() {
									goto l43
								}
								if !_rules[ruleDIGIT]() {
									goto l43
								}
								if !_rules[ruleDIGIT]() {
									goto l43
								}
								goto l44
							l43:
								position, tokenIndex, depth = position43, tokenIndex43, depth43
							}
						l44:
							{
								position45, tokenIndex45, depth45 := position, tokenIndex, depth
								if !_rules[ruleDIGIT]() {
									goto l45
								}
								if !_rules[ruleDIGIT]() {
									goto l45
								}
								if !_rules[ruleDIGIT]() {
									goto l45
								}
								goto l46
							l45:
								position, tokenIndex, depth = position45, tokenIndex45, depth45
							}
						l46:
							depth--
							add(rulePegText, position42)
						}
						{
							add(ruleAction3, position)
						}
						depth--
						add(ruletimestamp, position41)
					}
					goto l40
				l39:
					position, tokenIndex, depth = position39, tokenIndex39, depth39
				}
			l40:
				if buffer[position] != rune(' ') {
					goto l0
				}
				position++
				{
					position48 := position
					depth++
					{
						add(ruleAction0, position)
					}
					if !_rules[ruletag]() {
						goto l0
					}
				l50:
					{
						position51, tokenIndex51, depth51 := position, tokenIndex, depth
						if buffer[position] != rune(' ') {
							goto l51
						}
						position++
						if !_rules[ruletag]() {
							goto l51
						}
						goto l50
					l51:
						position, tokenIndex, depth = position51, tokenIndex51, depth51
					}
					depth--
					add(ruletags, position48)
				}
				{
					position52, tokenIndex52, depth52 := position, tokenIndex, depth
					if !matchDot() {
						goto l52
					}
					goto l0
				l52:
					position, tokenIndex, depth = position52, tokenIndex52, depth52
				}
				depth--
				add(ruleMetricGrammar, position1)
			}
			return true
		l0:
			position, tokenIndex, depth = position0, tokenIndex0, depth0
			return false
		},
		/* 1 tags <- <(Action0 tag (' ' tag)*)> */
		nil,
		/* 2 tag <- <(TagName '=' TagValue)> */
		func() bool {
			position54, tokenIndex54, depth54 := position, tokenIndex, depth
			{
				position55 := position
				depth++
				{
					position56 := position
					depth++
					{
						position57, tokenIndex57, depth57 := position, tokenIndex, depth
						if buffer[position] != rune('"') {
							goto l58
						}
						position++
						{
							position59 := position
							depth++
							{
								position62, tokenIndex62, depth62 := position, tokenIndex, depth
								if !_rules[ruleALNUM]() {
									goto l63
								}
								goto l62
							l63:
								position, tokenIndex, depth = position62, tokenIndex62, depth62
								{
									switch buffer[position] {
									case '.':
										if buffer[position] != rune('.') {
											goto l58
										}
										position++
										break
									case '_':
										if buffer[position] != rune('_') {
											goto l58
										}
										position++
										break
									case '-':
										if buffer[position] != rune('-') {
											goto l58
										}
										position++
										break
									default:
										if buffer[position] != rune('/') {
											goto l58
										}
										position++
										break
									}
								}

							}
						l62:
						l60:
							{
								position61, tokenIndex61, depth61 := position, tokenIndex, depth
								{
									position65, tokenIndex65, depth65 := position, tokenIndex, depth
									if !_rules[ruleALNUM]() {
										goto l66
									}
									goto l65
								l66:
									position, tokenIndex, depth = position65, tokenIndex65, depth65
									{
										switch buffer[position] {
										case '.':
											if buffer[position] != rune('.') {
												goto l61
											}
											position++
											break
										case '_':
											if buffer[position] != rune('_') {
												goto l61
											}
											position++
											break
										case '-':
											if buffer[position] != rune('-') {
												goto l61
											}
											position++
											break
										default:
											if buffer[position] != rune('/') {
												goto l61
											}
											position++
											break
										}
									}

								}
							l65:
								goto l60
							l61:
								position, tokenIndex, depth = position61, tokenIndex61, depth61
							}
							depth--
							add(rulePegText, position59)
						}
						if buffer[position] != rune('"') {
							goto l58
						}
						position++
						goto l57
					l58:
						position, tokenIndex, depth = position57, tokenIndex57, depth57
						{
							position68 := position
							depth++
							{
								position71, tokenIndex71, depth71 := position, tokenIndex, depth
								if !_rules[ruleALNUM]() {
									goto l72
								}
								goto l71
							l72:
								position, tokenIndex, depth = position71, tokenIndex71, depth71
								{
									switch buffer[position] {
									case '.':
										if buffer[position] != rune('.') {
											goto l54
										}
										position++
										break
									case '_':
										if buffer[position] != rune('_') {
											goto l54
										}
										position++
										break
									default:
										if buffer[position] != rune('-') {
											goto l54
										}
										position++
										break
									}
								}

							}
						l71:
						l69:
							{
								position70, tokenIndex70, depth70 := position, tokenIndex, depth
								{
									position74, tokenIndex74, depth74 := position, tokenIndex, depth
									if !_rules[ruleALNUM]() {
										goto l75
									}
									goto l74
								l75:
									position, tokenIndex, depth = position74, tokenIndex74, depth74
									{
										switch buffer[position] {
										case '.':
											if buffer[position] != rune('.') {
												goto l70
											}
											position++
											break
										case '_':
											if buffer[position] != rune('_') {
												goto l70
											}
											position++
											break
										default:
											if buffer[position] != rune('-') {
												goto l70
											}
											position++
											break
										}
									}

								}
							l74:
								goto l69
							l70:
								position, tokenIndex, depth = position70, tokenIndex70, depth70
							}
							depth--
							add(rulePegText, position68)
						}
					}
				l57:
					{
						add(ruleAction1, position)
					}
					depth--
					add(ruletagName, position56)
				}
				if buffer[position] != rune('=') {
					goto l54
				}
				position++
				{
					position78 := position
					depth++
					if buffer[position] != rune('"') {
						goto l54
					}
					position++
					{
						position79 := position
						depth++
						{
							position82, tokenIndex82, depth82 := position, tokenIndex, depth
							if buffer[position] != rune('\\') {
								goto l83
							}
							position++
							if buffer[position] != rune('"') {
								goto l83
							}
							position++
							goto l82
						l83:
							position, tokenIndex, depth = position82, tokenIndex82, depth82
							{
								position84, tokenIndex84, depth84 := position, tokenIndex, depth
								if buffer[position] != rune('"') {
									goto l84
								}
								position++
								goto l54
							l84:
								position, tokenIndex, depth = position84, tokenIndex84, depth84
							}
							if !matchDot() {
								goto l54
							}
						}
					l82:
					l80:
						{
							position81, tokenIndex81, depth81 := position, tokenIndex, depth
							{
								position85, tokenIndex85, depth85 := position, tokenIndex, depth
								if buffer[position] != rune('\\') {
									goto l86
								}
								position++
								if buffer[position] != rune('"') {
									goto l86
								}
								position++
								goto l85
							l86:
								position, tokenIndex, depth = position85, tokenIndex85, depth85
								{
									position87, tokenIndex87, depth87 := position, tokenIndex, depth
									if buffer[position] != rune('"') {
										goto l87
									}
									position++
									goto l81
								l87:
									position, tokenIndex, depth = position87, tokenIndex87, depth87
								}
								if !matchDot() {
									goto l81
								}
							}
						l85:
							goto l80
						l81:
							position, tokenIndex, depth = position81, tokenIndex81, depth81
						}
						depth--
						add(rulePegText, position79)
					}
					if buffer[position] != rune('"') {
						goto l54
					}
					position++
					{
						add(ruleAction2, position)
					}
					depth--
					add(ruletagValue, position78)
				}
				depth--
				add(ruletag, position55)
			}
			return true
		l54:
			position, tokenIndex, depth = position54, tokenIndex54, depth54
			return false
		},
		/* 3 TagName <- <((('"' <(ALNUM / ((&('.') '.') | (&('_') '_') | (&('-') '-') | (&('/') '/')))+> '"') / <(ALNUM / ((&('.') '.') | (&('_') '_') | (&('-') '-')))+>) Action1)> */
		nil,
		/* 4 TagValue <- <('"' <(('\\' '"') / (!'"' .))+> '"' Action2)> */
		nil,
		/* 5 timestamp <- <(<(DIGIT DIGIT DIGIT DIGIT DIGIT DIGIT DIGIT DIGIT DIGIT DIGIT (DIGIT DIGIT DIGIT)? (DIGIT DIGIT DIGIT)?)> Action3)> */
		nil,
		/* 6 metricValue <- <(<('-'? DIGIT+ ('.' DIGIT+)?)> Action4)> */
		nil,
		/* 7 MetricName <- <((('"' <('~'? (ALNUM / ((&(',') ',') | (&('/') '/') | (&('.') '.') | (&('_') '_') | (&('-') '-') | (&('~') '~')))+)> '"') / <('~'? (ALNUM / ((&(',') ',') | (&('/') '/') | (&('.') '.') | (&('_') '_') | (&('-') '-') | (&('~') '~')))+)>) Action5)> */
		nil,
		/* 8 ALNUM <- <((&('0' | '1' | '2' | '3' | '4' | '5' | '6' | '7' | '8' | '9') [0-9]) | (&('A' | 'B' | 'C' | 'D' | 'E' | 'F' | 'G' | 'H' | 'I' | 'J' | 'K' | 'L' | 'M' | 'N' | 'O' | 'P' | 'Q' | 'R' | 'S' | 'T' | 'U' | 'V' | 'W' | 'X' | 'Y' | 'Z') [A-Z]) | (&('a' | 'b' | 'c' | 'd' | 'e' | 'f' | 'g' | 'h' | 'i' | 'j' | 'k' | 'l' | 'm' | 'n' | 'o' | 'p' | 'q' | 'r' | 's' | 't' | 'u' | 'v' | 'w' | 'x' | 'y' | 'z') [a-z]))> */
		func() bool {
			position94, tokenIndex94, depth94 := position, tokenIndex, depth
			{
				position95 := position
				depth++
				{
					switch buffer[position] {
					case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l94
						}
						position++
						break
					case 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l94
						}
						position++
						break
					default:
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l94
						}
						position++
						break
					}
				}

				depth--
				add(ruleALNUM, position95)
			}
			return true
		l94:
			position, tokenIndex, depth = position94, tokenIndex94, depth94
			return false
		},
		/* 9 DIGIT <- <[0-9]> */
		func() bool {
			position97, tokenIndex97, depth97 := position, tokenIndex, depth
			{
				position98 := position
				depth++
				if c := buffer[position]; c < rune('0') || c > rune('9') {
					goto l97
				}
				position++
				depth--
				add(ruleDIGIT, position98)
			}
			return true
		l97:
			position, tokenIndex, depth = position97, tokenIndex97, depth97
			return false
		},
		/* 11 Action0 <- <{ p.Tags = map[string]string{} }> */
		nil,
		nil,
		/* 13 Action1 <- <{ p.currentTag = text }> */
		nil,
		/* 14 Action2 <- <{ p.Tags[p.currentTag] = text }> */
		nil,
		/* 15 Action3 <- <{ p.Timestamp = text }> */
		nil,
		/* 16 Action4 <- <{ p.Value = text }> */
		nil,
		/* 17 Action5 <- <{ p.Name = text }> */
		nil,
	}
	p.rules = _rules
}
