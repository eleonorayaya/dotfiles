package main

import "fmt"

func main() {
}

func other() {
	main()
	if true {
		main()
	}
	fmt.Println("hi")

}

func test() (string, string) {
	return "hi", "bye"
}
