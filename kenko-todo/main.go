package main

import kenkorest "github.com/akagiyui/go-together/kenko-rest"

func main() {
	println("Hello World")
	println(kenkorest.Hello())
	println(kenkorest.Add(1, 2))
}
