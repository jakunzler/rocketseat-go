package rateconvert

import "fmt"

func ConvertRate(amount float64, rate float64) float64 {
	fmt.Println("------------------------------------------")
	fmt.Println("I'm inside convert package.")
	fmt.Println("Amount:", amount)
	fmt.Println("Rate:", rate)
	fmt.Printf("Converted amount: %.2f\n", amount*rate)
	return amount * rate
}
