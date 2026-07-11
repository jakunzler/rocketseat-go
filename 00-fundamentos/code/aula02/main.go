package main

import "fmt"

func main() {
	arr00 := []int{}
	fmt.Println(arr00)

	arr01 := [5]int{1, 2, 3, 4, 5}
	slice01 := arr01[1:4]
	arr01[1] = 100
	slice01[1] = 200
	fmt.Println(slice01)

	slice02 := []int{1, 2, 3, 4, 5}
	slice03 := slice02[:0]
	fmt.Println(slice03, len(slice03), cap(slice03))

	var slice04 []int
	fmt.Println(slice04 == nil)

	var slice05 = []int{}
	fmt.Println(slice05 == nil)
}
