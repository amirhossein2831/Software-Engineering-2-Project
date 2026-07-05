package main

import (
	"fmt"
	"test-service-go/calculater"
)

func main() {
	fmt.Println("2 + 3 =", calculator.Add(2, 3))
	fmt.Println("10 - 4 =", calculator.Subtract(10, 4))
	fmt.Println("3 * 4 =", calculator.Multiply(3, 4))

	result, err := calculator.Divide(10, 2)
	if err != nil {
		fmt.Println("error:", err)
	} else {
		fmt.Println("10 / 2 =", result)
	}
}
