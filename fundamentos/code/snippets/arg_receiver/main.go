package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Uso: go run main.go <nome>")
		return
	}

	for i, arg := range os.Args {
		fmt.Printf("Argumento %d: %s\n", i, arg)
	}

	nome := os.Args[1]
	fmt.Printf("Olá, %s!\n", nome)
}
