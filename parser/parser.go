package parser

import (
	"bufio"
	"io"
	"strings"
)

//go:generate counterfeiter . Parser
type Parser interface {
	Parse(in io.Reader) ([]FuncEvent, error)
}

type FuncEventType int

const (
	Invocation FuncEventType = iota
	Returning
)

type FuncEvent interface {
	GetCaller() string
	GetCallee() string
	GetCallID() string
}

type InvocationEvent struct {
	Caller string
	Callee string
	CallID string
	Args   string
}

func (ie *InvocationEvent) GetCaller() string {
	return ie.Caller
}

func (ie *InvocationEvent) GetCallee() string {
	return ie.Callee
}

func (ie *InvocationEvent) GetCallID() string {
	return ie.CallID
}

type ReturningEvent struct {
	Caller string
	Callee string
	CallID string
}

func (re *ReturningEvent) GetCaller() string {
	return re.Caller
}

func (re *ReturningEvent) GetCallee() string {
	return re.Callee
}

func (re *ReturningEvent) GetCallID() string {
	return re.CallID
}

type parser struct {
}

func NewParser() Parser {
	return &parser{}
}

func (p *parser) Parse(in io.Reader) ([]FuncEvent, error) {
	var events []FuncEvent
	scanner := bufio.NewScanner(in)
	for scanner.Scan() {
		row := scanner.Text()
		lastSemicolon := strings.LastIndex(row, ";")
		msg := row[:lastSemicolon]
		secondHalf := row[lastSemicolon+1:]
		callID := strings.Split(secondHalf, "=")[1]
		if strings.HasPrefix(msg, "Function") {
			words := strings.Split(msg, " ")
			events = append(events, &InvocationEvent{
				Callee: normalizeFuncName(words[1]),
				Caller: normalizeFuncName(words[4]),
				Args:   strings.Join(words[5:], " "),
				CallID: callID,
			})
		}

		if strings.HasPrefix(msg, "Exiting function") {
			words := strings.Split(msg, " ")
			events = append(events, &ReturningEvent{
				Callee: normalizeFuncName(words[2]),
				Caller: normalizeFuncName(words[5]),
				CallID: callID,
			})
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return events, nil
}

func normalizeFuncName(funcName string) string {
	return funcName[strings.LastIndex(funcName, "/")+1:]
}
