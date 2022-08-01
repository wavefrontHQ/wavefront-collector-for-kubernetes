package main

// Code generated by peg -switch -inline cmd/test-proxy/metric_grammar.peg DO NOT EDIT.

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
)

const endSymbol rune = 1114112

/* The rule types inferred from the grammar are below. */
type pegRule uint8

const (
	ruleUnknown pegRule = iota
	ruleMetricGrammar
	ruleaggregationInterval
	rulehistogramValues
	rulehistogramValue
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
	ruleAction1
	rulePegText
	ruleAction2
	ruleAction3
	ruleAction4
	ruleAction5
	ruleAction6
)

var rul3s = [...]string{
	"Unknown",
	"MetricGrammar",
	"aggregationInterval",
	"histogramValues",
	"histogramValue",
	"tags",
	"tag",
	"tagName",
	"tagValue",
	"timestamp",
	"metricValue",
	"metricName",
	"ALNUM",
	"DIGIT",
	"Action0",
	"Action1",
	"PegText",
	"Action2",
	"Action3",
	"Action4",
	"Action5",
	"Action6",
}

type token32 struct {
	pegRule
	begin, end uint32
}

func (t *token32) String() string {
	return fmt.Sprintf("\x1B[34m%v\x1B[m %v %v", rul3s[t.pegRule], t.begin, t.end)
}

type node32 struct {
	token32
	up, next *node32
}

func (node *node32) print(w io.Writer, pretty bool, buffer string) {
	var print func(node *node32, depth int)
	print = func(node *node32, depth int) {
		for node != nil {
			for c := 0; c < depth; c++ {
				fmt.Fprintf(w, " ")
			}
			rule := rul3s[node.pegRule]
			quote := strconv.Quote(string(([]rune(buffer)[node.begin:node.end])))
			if !pretty {
				fmt.Fprintf(w, "%v %v\n", rule, quote)
			} else {
				fmt.Fprintf(w, "\x1B[36m%v\x1B[m %v\n", rule, quote)
			}
			if node.up != nil {
				print(node.up, depth+1)
			}
			node = node.next
		}
	}
	print(node, 0)
}

func (node *node32) Print(w io.Writer, buffer string) {
	node.print(w, false, buffer)
}

func (node *node32) PrettyPrint(w io.Writer, buffer string) {
	node.print(w, true, buffer)
}

type tokens32 struct {
	tree []token32
}

func (t *tokens32) Trim(length uint32) {
	t.tree = t.tree[:length]
}

func (t *tokens32) Print() {
	for _, token := range t.tree {
		fmt.Println(token.String())
	}
}

func (t *tokens32) AST() *node32 {
	type element struct {
		node *node32
		down *element
	}
	tokens := t.Tokens()
	var stack *element
	for _, token := range tokens {
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
	if stack != nil {
		return stack.node
	}
	return nil
}

func (t *tokens32) PrintSyntaxTree(buffer string) {
	t.AST().Print(os.Stdout, buffer)
}

func (t *tokens32) WriteSyntaxTree(w io.Writer, buffer string) {
	t.AST().Print(w, buffer)
}

func (t *tokens32) PrettyPrintSyntaxTree(buffer string) {
	t.AST().PrettyPrint(os.Stdout, buffer)
}

func (t *tokens32) Add(rule pegRule, begin, end, index uint32) {
	tree, i := t.tree, int(index)
	if i >= len(tree) {
		t.tree = append(tree, token32{pegRule: rule, begin: begin, end: end})
		return
	}
	tree[i] = token32{pegRule: rule, begin: begin, end: end}
}

func (t *tokens32) Tokens() []token32 {
	return t.tree
}

type MetricGrammar struct {
	Histogram  bool
	Name       string
	Value      string
	Timestamp  string
	Tags       map[string]string
	currentTag string

	Buffer string
	buffer []rune
	rules  [22]func() bool
	parse  func(rule ...int) error
	reset  func()
	Pretty bool
	tokens32
}

func (p *MetricGrammar) Parse(rule ...int) error {
	return p.parse(rule...)
}

func (p *MetricGrammar) Reset() {
	p.reset()
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
	tokens, err := []token32{e.max}, "\n"
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
		err += fmt.Sprintf(format,
			rul3s[token.pegRule],
			translations[begin].line, translations[begin].symbol,
			translations[end].line, translations[end].symbol,
			strconv.Quote(string(e.p.buffer[begin:end])))
	}

	return err
}

func (p *MetricGrammar) PrintSyntaxTree() {
	if p.Pretty {
		p.tokens32.PrettyPrintSyntaxTree(p.Buffer)
	} else {
		p.tokens32.PrintSyntaxTree(p.Buffer)
	}
}

func (p *MetricGrammar) WriteSyntaxTree(w io.Writer) {
	p.tokens32.WriteSyntaxTree(w, p.Buffer)
}

func (p *MetricGrammar) SprintSyntaxTree() string {
	var bldr strings.Builder
	p.WriteSyntaxTree(&bldr)
	return bldr.String()
}

func (p *MetricGrammar) Execute() {
	buffer, _buffer, text, begin, end := p.Buffer, p.buffer, "", 0, 0
	for _, token := range p.Tokens() {
		switch token.pegRule {

		case rulePegText:
			begin, end = int(token.begin), int(token.end)
			text = string(_buffer[begin:end])

		case ruleAction0:
			p.Histogram = true
		case ruleAction1:
			p.Tags = map[string]string{}
		case ruleAction2:
			p.currentTag = text
		case ruleAction3:
			p.Tags[p.currentTag] = text
		case ruleAction4:
			p.Timestamp = text
		case ruleAction5:
			p.Value = text
		case ruleAction6:
			p.Name = text

		}
	}
	_, _, _, _, _ = buffer, _buffer, text, begin, end
}

func Pretty(pretty bool) func(*MetricGrammar) error {
	return func(p *MetricGrammar) error {
		p.Pretty = pretty
		return nil
	}
}

func Size(size int) func(*MetricGrammar) error {
	return func(p *MetricGrammar) error {
		p.tokens32 = tokens32{tree: make([]token32, 0, size)}
		return nil
	}
}
func (p *MetricGrammar) Init(options ...func(*MetricGrammar) error) error {
	var (
		max                  token32
		position, tokenIndex uint32
		buffer               []rune
	)
	for _, option := range options {
		err := option(p)
		if err != nil {
			return err
		}
	}
	p.reset = func() {
		max = token32{}
		position, tokenIndex = 0, 0

		p.buffer = []rune(p.Buffer)
		if len(p.buffer) == 0 || p.buffer[len(p.buffer)-1] != endSymbol {
			p.buffer = append(p.buffer, endSymbol)
		}
		buffer = p.buffer
	}
	p.reset()

	_rules := p.rules
	tree := p.tokens32
	p.parse = func(rule ...int) error {
		r := 1
		if len(rule) > 0 {
			r = rule[0]
		}
		matches := p.rules[r]()
		p.tokens32 = tree
		if matches {
			p.Trim(tokenIndex)
			return nil
		}
		return &parseError{p, max}
	}

	add := func(rule pegRule, begin uint32) {
		tree.Add(rule, begin, position, tokenIndex)
		tokenIndex++
		if begin != position && position > max.end {
			max = token32{rule, begin, position}
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
		/* 0 MetricGrammar <- <((aggregationInterval (' ' timestamp)? ' ' histogramValues ' ' metricName ' ' tags !.) / (metricName ' ' metricValue (' ' timestamp)? ' ' tags !.))> */
		func() bool {
			position0, tokenIndex0 := position, tokenIndex
			{
				position1 := position
				{
					position2, tokenIndex2 := position, tokenIndex
					{
						position4 := position
						if buffer[position] != rune('!') {
							goto l3
						}
						position++
						{
							switch buffer[position] {
							case 'D':
								if buffer[position] != rune('D') {
									goto l3
								}
								position++
							case 'H':
								if buffer[position] != rune('H') {
									goto l3
								}
								position++
							default:
								if buffer[position] != rune('M') {
									goto l3
								}
								position++
							}
						}

						{
							add(ruleAction0, position)
						}
						add(ruleaggregationInterval, position4)
					}
					{
						position7, tokenIndex7 := position, tokenIndex
						if buffer[position] != rune(' ') {
							goto l7
						}
						position++
						if !_rules[ruletimestamp]() {
							goto l7
						}
						goto l8
					l7:
						position, tokenIndex = position7, tokenIndex7
					}
				l8:
					if buffer[position] != rune(' ') {
						goto l3
					}
					position++
					{
						position9 := position
						if !_rules[rulehistogramValue]() {
							goto l3
						}
					l10:
						{
							position11, tokenIndex11 := position, tokenIndex
							if buffer[position] != rune(' ') {
								goto l11
							}
							position++
							if !_rules[rulehistogramValue]() {
								goto l11
							}
							goto l10
						l11:
							position, tokenIndex = position11, tokenIndex11
						}
						add(rulehistogramValues, position9)
					}
					if buffer[position] != rune(' ') {
						goto l3
					}
					position++
					if !_rules[rulemetricName]() {
						goto l3
					}
					if buffer[position] != rune(' ') {
						goto l3
					}
					position++
					if !_rules[ruletags]() {
						goto l3
					}
					{
						position12, tokenIndex12 := position, tokenIndex
						if !matchDot() {
							goto l12
						}
						goto l3
					l12:
						position, tokenIndex = position12, tokenIndex12
					}
					goto l2
				l3:
					position, tokenIndex = position2, tokenIndex2
					if !_rules[rulemetricName]() {
						goto l0
					}
					if buffer[position] != rune(' ') {
						goto l0
					}
					position++
					if !_rules[rulemetricValue]() {
						goto l0
					}
					{
						position13, tokenIndex13 := position, tokenIndex
						if buffer[position] != rune(' ') {
							goto l13
						}
						position++
						if !_rules[ruletimestamp]() {
							goto l13
						}
						goto l14
					l13:
						position, tokenIndex = position13, tokenIndex13
					}
				l14:
					if buffer[position] != rune(' ') {
						goto l0
					}
					position++
					if !_rules[ruletags]() {
						goto l0
					}
					{
						position15, tokenIndex15 := position, tokenIndex
						if !matchDot() {
							goto l15
						}
						goto l0
					l15:
						position, tokenIndex = position15, tokenIndex15
					}
				}
			l2:
				add(ruleMetricGrammar, position1)
			}
			return true
		l0:
			position, tokenIndex = position0, tokenIndex0
			return false
		},
		/* 1 aggregationInterval <- <('!' ((&('D') 'D') | (&('H') 'H') | (&('M') 'M')) Action0)> */
		nil,
		/* 2 histogramValues <- <(histogramValue (' ' histogramValue)*)> */
		nil,
		/* 3 histogramValue <- <('#' DIGIT+ ' ' metricValue)> */
		func() bool {
			position18, tokenIndex18 := position, tokenIndex
			{
				position19 := position
				if buffer[position] != rune('#') {
					goto l18
				}
				position++
				if !_rules[ruleDIGIT]() {
					goto l18
				}
			l20:
				{
					position21, tokenIndex21 := position, tokenIndex
					if !_rules[ruleDIGIT]() {
						goto l21
					}
					goto l20
				l21:
					position, tokenIndex = position21, tokenIndex21
				}
				if buffer[position] != rune(' ') {
					goto l18
				}
				position++
				if !_rules[rulemetricValue]() {
					goto l18
				}
				add(rulehistogramValue, position19)
			}
			return true
		l18:
			position, tokenIndex = position18, tokenIndex18
			return false
		},
		/* 4 tags <- <(Action1 tag (' ' tag)*)> */
		func() bool {
			position22, tokenIndex22 := position, tokenIndex
			{
				position23 := position
				{
					add(ruleAction1, position)
				}
				if !_rules[ruletag]() {
					goto l22
				}
			l25:
				{
					position26, tokenIndex26 := position, tokenIndex
					if buffer[position] != rune(' ') {
						goto l26
					}
					position++
					if !_rules[ruletag]() {
						goto l26
					}
					goto l25
				l26:
					position, tokenIndex = position26, tokenIndex26
				}
				add(ruletags, position23)
			}
			return true
		l22:
			position, tokenIndex = position22, tokenIndex22
			return false
		},
		/* 5 tag <- <(tagName '=' tagValue)> */
		func() bool {
			position27, tokenIndex27 := position, tokenIndex
			{
				position28 := position
				{
					position29 := position
					{
						position30, tokenIndex30 := position, tokenIndex
						if buffer[position] != rune('"') {
							goto l31
						}
						position++
						{
							position32 := position
							{
								position35, tokenIndex35 := position, tokenIndex
								if !_rules[ruleALNUM]() {
									goto l36
								}
								goto l35
							l36:
								position, tokenIndex = position35, tokenIndex35
								{
									switch buffer[position] {
									case '.':
										if buffer[position] != rune('.') {
											goto l31
										}
										position++
									case '_':
										if buffer[position] != rune('_') {
											goto l31
										}
										position++
									case '-':
										if buffer[position] != rune('-') {
											goto l31
										}
										position++
									default:
										if buffer[position] != rune('/') {
											goto l31
										}
										position++
									}
								}

							}
						l35:
						l33:
							{
								position34, tokenIndex34 := position, tokenIndex
								{
									position38, tokenIndex38 := position, tokenIndex
									if !_rules[ruleALNUM]() {
										goto l39
									}
									goto l38
								l39:
									position, tokenIndex = position38, tokenIndex38
									{
										switch buffer[position] {
										case '.':
											if buffer[position] != rune('.') {
												goto l34
											}
											position++
										case '_':
											if buffer[position] != rune('_') {
												goto l34
											}
											position++
										case '-':
											if buffer[position] != rune('-') {
												goto l34
											}
											position++
										default:
											if buffer[position] != rune('/') {
												goto l34
											}
											position++
										}
									}

								}
							l38:
								goto l33
							l34:
								position, tokenIndex = position34, tokenIndex34
							}
							add(rulePegText, position32)
						}
						if buffer[position] != rune('"') {
							goto l31
						}
						position++
						goto l30
					l31:
						position, tokenIndex = position30, tokenIndex30
						{
							position41 := position
							{
								position44, tokenIndex44 := position, tokenIndex
								if !_rules[ruleALNUM]() {
									goto l45
								}
								goto l44
							l45:
								position, tokenIndex = position44, tokenIndex44
								{
									switch buffer[position] {
									case '.':
										if buffer[position] != rune('.') {
											goto l27
										}
										position++
									case '_':
										if buffer[position] != rune('_') {
											goto l27
										}
										position++
									default:
										if buffer[position] != rune('-') {
											goto l27
										}
										position++
									}
								}

							}
						l44:
						l42:
							{
								position43, tokenIndex43 := position, tokenIndex
								{
									position47, tokenIndex47 := position, tokenIndex
									if !_rules[ruleALNUM]() {
										goto l48
									}
									goto l47
								l48:
									position, tokenIndex = position47, tokenIndex47
									{
										switch buffer[position] {
										case '.':
											if buffer[position] != rune('.') {
												goto l43
											}
											position++
										case '_':
											if buffer[position] != rune('_') {
												goto l43
											}
											position++
										default:
											if buffer[position] != rune('-') {
												goto l43
											}
											position++
										}
									}

								}
							l47:
								goto l42
							l43:
								position, tokenIndex = position43, tokenIndex43
							}
							add(rulePegText, position41)
						}
					}
				l30:
					{
						add(ruleAction2, position)
					}
					add(ruletagName, position29)
				}
				if buffer[position] != rune('=') {
					goto l27
				}
				position++
				{
					position51 := position
					if buffer[position] != rune('"') {
						goto l27
					}
					position++
					{
						position52 := position
						{
							position55, tokenIndex55 := position, tokenIndex
							if buffer[position] != rune('\\') {
								goto l56
							}
							position++
							if buffer[position] != rune('"') {
								goto l56
							}
							position++
							goto l55
						l56:
							position, tokenIndex = position55, tokenIndex55
							{
								position57, tokenIndex57 := position, tokenIndex
								if buffer[position] != rune('"') {
									goto l57
								}
								position++
								goto l27
							l57:
								position, tokenIndex = position57, tokenIndex57
							}
							if !matchDot() {
								goto l27
							}
						}
					l55:
					l53:
						{
							position54, tokenIndex54 := position, tokenIndex
							{
								position58, tokenIndex58 := position, tokenIndex
								if buffer[position] != rune('\\') {
									goto l59
								}
								position++
								if buffer[position] != rune('"') {
									goto l59
								}
								position++
								goto l58
							l59:
								position, tokenIndex = position58, tokenIndex58
								{
									position60, tokenIndex60 := position, tokenIndex
									if buffer[position] != rune('"') {
										goto l60
									}
									position++
									goto l54
								l60:
									position, tokenIndex = position60, tokenIndex60
								}
								if !matchDot() {
									goto l54
								}
							}
						l58:
							goto l53
						l54:
							position, tokenIndex = position54, tokenIndex54
						}
						add(rulePegText, position52)
					}
					if buffer[position] != rune('"') {
						goto l27
					}
					position++
					{
						add(ruleAction3, position)
					}
					add(ruletagValue, position51)
				}
				add(ruletag, position28)
			}
			return true
		l27:
			position, tokenIndex = position27, tokenIndex27
			return false
		},
		/* 6 tagName <- <((('"' <(ALNUM / ((&('.') '.') | (&('_') '_') | (&('-') '-') | (&('/') '/')))+> '"') / <(ALNUM / ((&('.') '.') | (&('_') '_') | (&('-') '-')))+>) Action2)> */
		nil,
		/* 7 tagValue <- <('"' <(('\\' '"') / (!'"' .))+> '"' Action3)> */
		nil,
		/* 8 timestamp <- <(<(DIGIT DIGIT DIGIT DIGIT DIGIT DIGIT DIGIT DIGIT DIGIT DIGIT (DIGIT DIGIT DIGIT)? (DIGIT DIGIT DIGIT)?)> Action4)> */
		func() bool {
			position64, tokenIndex64 := position, tokenIndex
			{
				position65 := position
				{
					position66 := position
					if !_rules[ruleDIGIT]() {
						goto l64
					}
					if !_rules[ruleDIGIT]() {
						goto l64
					}
					if !_rules[ruleDIGIT]() {
						goto l64
					}
					if !_rules[ruleDIGIT]() {
						goto l64
					}
					if !_rules[ruleDIGIT]() {
						goto l64
					}
					if !_rules[ruleDIGIT]() {
						goto l64
					}
					if !_rules[ruleDIGIT]() {
						goto l64
					}
					if !_rules[ruleDIGIT]() {
						goto l64
					}
					if !_rules[ruleDIGIT]() {
						goto l64
					}
					if !_rules[ruleDIGIT]() {
						goto l64
					}
					{
						position67, tokenIndex67 := position, tokenIndex
						if !_rules[ruleDIGIT]() {
							goto l67
						}
						if !_rules[ruleDIGIT]() {
							goto l67
						}
						if !_rules[ruleDIGIT]() {
							goto l67
						}
						goto l68
					l67:
						position, tokenIndex = position67, tokenIndex67
					}
				l68:
					{
						position69, tokenIndex69 := position, tokenIndex
						if !_rules[ruleDIGIT]() {
							goto l69
						}
						if !_rules[ruleDIGIT]() {
							goto l69
						}
						if !_rules[ruleDIGIT]() {
							goto l69
						}
						goto l70
					l69:
						position, tokenIndex = position69, tokenIndex69
					}
				l70:
					add(rulePegText, position66)
				}
				{
					add(ruleAction4, position)
				}
				add(ruletimestamp, position65)
			}
			return true
		l64:
			position, tokenIndex = position64, tokenIndex64
			return false
		},
		/* 9 metricValue <- <(<('-'? DIGIT+ ('.' DIGIT+)?)> Action5)> */
		func() bool {
			position72, tokenIndex72 := position, tokenIndex
			{
				position73 := position
				{
					position74 := position
					{
						position75, tokenIndex75 := position, tokenIndex
						if buffer[position] != rune('-') {
							goto l75
						}
						position++
						goto l76
					l75:
						position, tokenIndex = position75, tokenIndex75
					}
				l76:
					if !_rules[ruleDIGIT]() {
						goto l72
					}
				l77:
					{
						position78, tokenIndex78 := position, tokenIndex
						if !_rules[ruleDIGIT]() {
							goto l78
						}
						goto l77
					l78:
						position, tokenIndex = position78, tokenIndex78
					}
					{
						position79, tokenIndex79 := position, tokenIndex
						if buffer[position] != rune('.') {
							goto l79
						}
						position++
						if !_rules[ruleDIGIT]() {
							goto l79
						}
					l81:
						{
							position82, tokenIndex82 := position, tokenIndex
							if !_rules[ruleDIGIT]() {
								goto l82
							}
							goto l81
						l82:
							position, tokenIndex = position82, tokenIndex82
						}
						goto l80
					l79:
						position, tokenIndex = position79, tokenIndex79
					}
				l80:
					add(rulePegText, position74)
				}
				{
					add(ruleAction5, position)
				}
				add(rulemetricValue, position73)
			}
			return true
		l72:
			position, tokenIndex = position72, tokenIndex72
			return false
		},
		/* 10 metricName <- <((('"' <('∆'? '~'? (ALNUM / ((&(',') ',') | (&('/') '/') | (&('.') '.') | (&('_') '_') | (&('-') '-') | (&('~') '~')))+)> '"') / <('∆'? '~'? (ALNUM / ((&(',') ',') | (&('/') '/') | (&('.') '.') | (&('_') '_') | (&('-') '-') | (&('~') '~')))+)>) Action6)> */
		func() bool {
			position84, tokenIndex84 := position, tokenIndex
			{
				position85 := position
				{
					position86, tokenIndex86 := position, tokenIndex
					if buffer[position] != rune('"') {
						goto l87
					}
					position++
					{
						position88 := position
						{
							position89, tokenIndex89 := position, tokenIndex
							if buffer[position] != rune('∆') {
								goto l89
							}
							position++
							goto l90
						l89:
							position, tokenIndex = position89, tokenIndex89
						}
					l90:
						{
							position91, tokenIndex91 := position, tokenIndex
							if buffer[position] != rune('~') {
								goto l91
							}
							position++
							goto l92
						l91:
							position, tokenIndex = position91, tokenIndex91
						}
					l92:
						{
							position95, tokenIndex95 := position, tokenIndex
							if !_rules[ruleALNUM]() {
								goto l96
							}
							goto l95
						l96:
							position, tokenIndex = position95, tokenIndex95
							{
								switch buffer[position] {
								case ',':
									if buffer[position] != rune(',') {
										goto l87
									}
									position++
								case '/':
									if buffer[position] != rune('/') {
										goto l87
									}
									position++
								case '.':
									if buffer[position] != rune('.') {
										goto l87
									}
									position++
								case '_':
									if buffer[position] != rune('_') {
										goto l87
									}
									position++
								case '-':
									if buffer[position] != rune('-') {
										goto l87
									}
									position++
								default:
									if buffer[position] != rune('~') {
										goto l87
									}
									position++
								}
							}

						}
					l95:
					l93:
						{
							position94, tokenIndex94 := position, tokenIndex
							{
								position98, tokenIndex98 := position, tokenIndex
								if !_rules[ruleALNUM]() {
									goto l99
								}
								goto l98
							l99:
								position, tokenIndex = position98, tokenIndex98
								{
									switch buffer[position] {
									case ',':
										if buffer[position] != rune(',') {
											goto l94
										}
										position++
									case '/':
										if buffer[position] != rune('/') {
											goto l94
										}
										position++
									case '.':
										if buffer[position] != rune('.') {
											goto l94
										}
										position++
									case '_':
										if buffer[position] != rune('_') {
											goto l94
										}
										position++
									case '-':
										if buffer[position] != rune('-') {
											goto l94
										}
										position++
									default:
										if buffer[position] != rune('~') {
											goto l94
										}
										position++
									}
								}

							}
						l98:
							goto l93
						l94:
							position, tokenIndex = position94, tokenIndex94
						}
						add(rulePegText, position88)
					}
					if buffer[position] != rune('"') {
						goto l87
					}
					position++
					goto l86
				l87:
					position, tokenIndex = position86, tokenIndex86
					{
						position101 := position
						{
							position102, tokenIndex102 := position, tokenIndex
							if buffer[position] != rune('∆') {
								goto l102
							}
							position++
							goto l103
						l102:
							position, tokenIndex = position102, tokenIndex102
						}
					l103:
						{
							position104, tokenIndex104 := position, tokenIndex
							if buffer[position] != rune('~') {
								goto l104
							}
							position++
							goto l105
						l104:
							position, tokenIndex = position104, tokenIndex104
						}
					l105:
						{
							position108, tokenIndex108 := position, tokenIndex
							if !_rules[ruleALNUM]() {
								goto l109
							}
							goto l108
						l109:
							position, tokenIndex = position108, tokenIndex108
							{
								switch buffer[position] {
								case ',':
									if buffer[position] != rune(',') {
										goto l84
									}
									position++
								case '/':
									if buffer[position] != rune('/') {
										goto l84
									}
									position++
								case '.':
									if buffer[position] != rune('.') {
										goto l84
									}
									position++
								case '_':
									if buffer[position] != rune('_') {
										goto l84
									}
									position++
								case '-':
									if buffer[position] != rune('-') {
										goto l84
									}
									position++
								default:
									if buffer[position] != rune('~') {
										goto l84
									}
									position++
								}
							}

						}
					l108:
					l106:
						{
							position107, tokenIndex107 := position, tokenIndex
							{
								position111, tokenIndex111 := position, tokenIndex
								if !_rules[ruleALNUM]() {
									goto l112
								}
								goto l111
							l112:
								position, tokenIndex = position111, tokenIndex111
								{
									switch buffer[position] {
									case ',':
										if buffer[position] != rune(',') {
											goto l107
										}
										position++
									case '/':
										if buffer[position] != rune('/') {
											goto l107
										}
										position++
									case '.':
										if buffer[position] != rune('.') {
											goto l107
										}
										position++
									case '_':
										if buffer[position] != rune('_') {
											goto l107
										}
										position++
									case '-':
										if buffer[position] != rune('-') {
											goto l107
										}
										position++
									default:
										if buffer[position] != rune('~') {
											goto l107
										}
										position++
									}
								}

							}
						l111:
							goto l106
						l107:
							position, tokenIndex = position107, tokenIndex107
						}
						add(rulePegText, position101)
					}
				}
			l86:
				{
					add(ruleAction6, position)
				}
				add(rulemetricName, position85)
			}
			return true
		l84:
			position, tokenIndex = position84, tokenIndex84
			return false
		},
		/* 11 ALNUM <- <((&('0' | '1' | '2' | '3' | '4' | '5' | '6' | '7' | '8' | '9') [0-9]) | (&('A' | 'B' | 'C' | 'D' | 'E' | 'F' | 'G' | 'H' | 'I' | 'J' | 'K' | 'L' | 'M' | 'N' | 'O' | 'P' | 'Q' | 'R' | 'S' | 'T' | 'U' | 'V' | 'W' | 'X' | 'Y' | 'Z') [A-Z]) | (&('a' | 'b' | 'c' | 'd' | 'e' | 'f' | 'g' | 'h' | 'i' | 'j' | 'k' | 'l' | 'm' | 'n' | 'o' | 'p' | 'q' | 'r' | 's' | 't' | 'u' | 'v' | 'w' | 'x' | 'y' | 'z') [a-z]))> */
		func() bool {
			position115, tokenIndex115 := position, tokenIndex
			{
				position116 := position
				{
					switch buffer[position] {
					case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l115
						}
						position++
					case 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z':
						if c := buffer[position]; c < rune('A') || c > rune('Z') {
							goto l115
						}
						position++
					default:
						if c := buffer[position]; c < rune('a') || c > rune('z') {
							goto l115
						}
						position++
					}
				}

				add(ruleALNUM, position116)
			}
			return true
		l115:
			position, tokenIndex = position115, tokenIndex115
			return false
		},
		/* 12 DIGIT <- <[0-9]> */
		func() bool {
			position118, tokenIndex118 := position, tokenIndex
			{
				position119 := position
				if c := buffer[position]; c < rune('0') || c > rune('9') {
					goto l118
				}
				position++
				add(ruleDIGIT, position119)
			}
			return true
		l118:
			position, tokenIndex = position118, tokenIndex118
			return false
		},
		/* 14 Action0 <- <{ p.Histogram = true }> */
		nil,
		/* 15 Action1 <- <{ p.Tags = map[string]string{} }> */
		nil,
		nil,
		/* 17 Action2 <- <{ p.currentTag = text }> */
		nil,
		/* 18 Action3 <- <{ p.Tags[p.currentTag] = text }> */
		nil,
		/* 19 Action4 <- <{ p.Timestamp = text }> */
		nil,
		/* 20 Action5 <- <{ p.Value = text }> */
		nil,
		/* 21 Action6 <- <{ p.Name = text }> */
		nil,
	}
	p.rules = _rules
	return nil
}
