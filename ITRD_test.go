package main

import "fmt"

func main() {
	m := map[string]struct{}{
		"yinuo": {},
		"love": {},
	}
	for k, v := range m {
		fmt.Println(k, v)
	}
}