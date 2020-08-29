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
	FuncName() string
}

type InvocationEvent struct {
	Name string
	Args string
}

func (ie *InvocationEvent) FuncName() string {
	return ie.Name
}

type ReturningEvent struct {
	Name string
}

func (ie *ReturningEvent) FuncName() string {
	return ie.Name
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
		if strings.HasPrefix(row, "Entering function") {
			words := strings.Split(row, " ")
			events = append(events, &InvocationEvent{
				Name: words[2],
				Args: strings.Join(words[3:], " "),
			})
		}

		if strings.HasPrefix(row, "Exiting function") {
			split := strings.Split(row, " ")
			events = append(events, &ReturningEvent{
				Name: split[2],
			})
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return events, nil
}
