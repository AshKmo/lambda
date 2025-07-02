package main

import (
	"fmt"
	"strings"
	"os"
	"io"
)

type ElementType int

const (
	LambdaElement ElementType = iota
	ApplicationElement = iota
	NameElement = iota
	BracketElement = iota
	BackslashElement = iota
	ClosureElement = iota
	InvalidElement = iota
)

type Element interface {
	Type() ElementType
}

type Name string

func (_ Name) Type() ElementType {
	return NameElement
}

type Lambda struct {
	Parameter Name
	Expression Element
}

func (_ Lambda) Type() ElementType {
	return LambdaElement
}

type Scope struct {
	Parent *Scope
	Variable Name
	Value Element
}

func (s Scope) Get(n Name) Element {
	if s.Variable == n {
		return s.Value
	}

	if s.Parent == nil {
		panic(fmt.Sprintf("variable not found: %q", n))
	}

	return (*s.Parent).Get(n)
}

type Closure struct {
	Enclosure Scope
	Parameter Name
	Expression Element
}

func (_ Closure) Type() ElementType {
	return ClosureElement
}

type Application struct {
	A Element
	B Element
}

func (_ Application) Type() ElementType {
	return ApplicationElement
}

type Bracket bool

func (_ Bracket) Type() ElementType {
	return BracketElement
}

type Backslash struct{}

func (_ Backslash) Type() ElementType {
	return BackslashElement
}

func tokenise(script string) []Element {
	var tokens []Element

	oldType := InvalidElement
	newType := InvalidElement

	var currentToken strings.Builder

	for i := 0; i <= len(script); i++ {
		var c byte

		if i == len(script) {
			c = '\n'
		} else {
			c = script[i]
		}

		switch c {
		case ' ', '\n', '\r', '\t':
			newType = InvalidElement
		case '(', ')':
			newType = BracketElement
		case '\\':
			newType = BackslashElement
		default:
			newType = NameElement
		}

		if oldType != newType && currentToken.Len() > 0 {
			switch oldType {
			case BackslashElement:
				tokens = append(tokens, Backslash{})
			case BracketElement:
				tokens = append(tokens, Bracket(currentToken.String()[0] == '('))
			case NameElement:
				tokens = append(tokens, Name(currentToken.String()))
			}

			currentToken.Reset()
		}

		oldType = newType

		if oldType != InvalidElement {
			currentToken.WriteByte(c)
		}
	}

	return tokens
}

func treeify(tokens []Element) Element {
	var i int
	return treeifyExpression(&i, tokens)
}

func treeifyExpression(i *int, tokens []Element) Element {
	var result Element

	var branch Element

	for ; *i < len(tokens); *i++ {
		e := tokens[*i]

		switch v := e.(type) {
		case Bracket:
			if !v {
				return result
			}

			*i++
			branch = treeifyExpression(i, tokens)
		case Backslash:
			*i++
			parameter := tokens[*i].(Name)
			*i++
			branch = Lambda{parameter, treeifyExpression(i, tokens)}
			*i--
		default:
			branch = e
		}

		if result == nil {
			result = branch
		} else {
			result = Application{result, branch}
		}
	}

	return result
}

func evaluate(e Element, scope Scope) Element {
	switch v := e.(type) {
	case Name:
		return scope.Get(v)
	case Lambda:
		return Closure{scope, v.Parameter, v.Expression}
	case Application:
		a := evaluate(v.A, scope).(Closure)
		b := evaluate(v.B, scope)

		newScope := Scope{&a.Enclosure, a.Parameter, b}
		return evaluate(a.Expression, newScope)
	}

	return nil
}

func main() {
	file, e := os.Open("script.txt")
	if e != nil {
		panic("can't open script file")
	}

	script, e := io.ReadAll(file)
	if e != nil {
		panic("can't read script file")
	}

	file.Close()

	tokens := tokenise(string(script))
	fmt.Printf("tokens: %#v\n\n", tokens)

	ast := treeify(tokens)
	fmt.Printf("ast: %#v\n\n", ast)

	result := evaluate(ast, Scope{})
	fmt.Printf("result: %#v\n\n", result)
}
