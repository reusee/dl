package main

import (
	"fmt"

	"github.com/reusee/e/v2"
)

var (
	pt     = fmt.Printf
	me     = e.Default.WithStack()
	ce, he = e.New(me)
)
