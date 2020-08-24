# prinTracer
[![Build Status](https://travis-ci.org/DimitarPetrov/printracer.svg?branch=master)](https://travis-ci.org/DimitarPetrov/printracer)
[![Coverage Status](https://coveralls.io/repos/github/DimitarPetrov/printracer/badge.svg?branch=master)](https://coveralls.io/github/DimitarPetrov/printracer?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/DimitarPetrov/printracer)](https://goreportcard.com/report/github.com/DimitarPetrov/printracer)

## Overview

`printracer` is a simple command line tool that instruments all **go** code in the current working directory to print every
 function execution along with its arguments.
 
## Installation

#### Installing from Source
```
go get -u github.com/DimitarPetrov/printracer
```

## Demonstration

Let's say you have a simple `main.go` file in the current working directory with the following contents:
```go
package main

import "fmt"

func test(i int, b bool) int {
	if b {
		return i
	}
	return 0
}

func main() {
	i := test(2, false)
	fmt.Println(i)
}
```

After executing:
```
printracer apply
```

The file will be modified like the following:
```go
package main

import "fmt"

func test(i int, b bool) int {
	fmt.Printf("Entering function test with args (%v) (%v)\n", i, b)
	defer fmt.Printf("Exiting function test\n")
	if b {
		return i
	}
	return 0
}

func main() {
	fmt.Printf("Entering function main\n")
	defer fmt.Printf("Exiting function main\n")
	i := test(2, false)
	fmt.Println(i)
}
```

You can also easily revert all the changes done by `printracer` by just executing:
```
printracer revert
```