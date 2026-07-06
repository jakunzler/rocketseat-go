package main

import "fmt"

func main() {
	fmt.Println("Hello, World!")

	checkAdult(17)
}

func checkAdult(age int) {
	if age < 18 {
		fmt.Println("You are a child.")
	} else if age == 18 {
		fmt.Println("A fresh new adult!")
	} else {
		fmt.Println("You are an adult.")
	}
}
