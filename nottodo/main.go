package main

import "github.com/akagiyui/go-together/rest"

func main() {
	println("Hello World")
	println(rest.Hello())
	println(rest.Add(1, 2))
}
