package main

import (
	"fmt"
)

func uzbek(d int) {
	fmt.Println(d)
}

func Export_block2281(_ *map[string]any, funcMap *map[string]any) {
	funcsMap := *funcMap
	uzbek(7)
	funcsMap["uzbek"] = uzbek
}
