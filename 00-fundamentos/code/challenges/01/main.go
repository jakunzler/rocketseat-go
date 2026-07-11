package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	filereader "challenges/01/fileReader"
	rateconverter "challenges/01/rateConverter"
)

var (
	valor_em_brl  float64
	moeda_destino string
	convertOnly   bool
	err           error
)

func main() {
	fmt.Println("==========================================")
	fmt.Println("Hello, World! This is my first Go program.")
	fmt.Println("==========================================")
	if len(os.Args) < 2 {
		fmt.Println("Uso: ./convertido <valor_em_brl> <moeda_destino> <convertOption>")
		fmt.Println("Exemplo: go run main.go 100 BRL false")
		fmt.Println("Assumindo valores padrão: valor_em_brl=100, moeda_destino=USD, convertOption=false")
		valor_em_brl = 100
		moeda_destino = "USD"
		convertOnly = false
	} else if len(os.Args) < 3 {
		fmt.Println("Uso: go run main.go <valor_em_brl> <moeda_destino> <convertOption>")
		fmt.Println("Exemplo: go run main.go 100 BRL false")
		fmt.Println("Assumindo valores padrão: moeda_destino=USD, convertOption=false")
		valor_em_brl, err = strconv.ParseFloat(os.Args[1], 64)
		if err != nil {
			fmt.Println("Error parsing valor_em_brl:", err)
			return
		}
		moeda_destino = "USD"
		convertOnly = false
	} else if len(os.Args) < 4 {
		fmt.Println("Uso: go run main.go <valor_em_brl> <moeda_destino> <convertOption>")
		fmt.Println("Exemplo: go run main.go 100 BRL false")
		fmt.Println("Assumindo valor padrão: convertOption=false")
		valor_em_brl, err = strconv.ParseFloat(os.Args[1], 64)
		if err != nil {
			fmt.Println("Error parsing valor_em_brl:", err)
			return
		}
		moeda_destino = strings.ToUpper(os.Args[2])
		convertOnly = false
	} else {
		valor_em_brl, err = strconv.ParseFloat(os.Args[1], 64)
		if err != nil {
			fmt.Println("Error parsing valor_em_brl:", err)
			return
		}
		moeda_destino = strings.ToUpper(os.Args[2])
		convertOnly, err = strconv.ParseBool(os.Args[3])
		if err != nil {
			fmt.Println("Error parsing convertOnly:", err)
			return
		}
	}

	if convertOnly == true {
		fmt.Println("------------------------------------------")
		fmt.Println("Conversion Only Mode Enabled")
		fmt.Println("------------------------------------------")
		exchangeRates, err := filereader.ReadFile()
		if err != nil {
			fmt.Println("Error reading file:", err)
			return
		}
		fmt.Println("Target Currency:", moeda_destino)

		if moeda_destino == "BRL" {
			fmt.Println("------------------------------------------")
			fmt.Printf("%.2f\n", valor_em_brl)
			return
		}
		actualRate, ok := exchangeRates.Rates[moeda_destino]
		// fmt.Println("Actual Rate:", actualRate)
		if !ok {
			fmt.Printf("Error: Rate for currency %s not found\n", moeda_destino)
			return
		}

		amountConverted := rateconverter.ConvertRate(valor_em_brl, actualRate)
		fmt.Println("------------------------------------------")
		fmt.Printf("%.2f\n", amountConverted)

	} else {
		fmt.Println("------------------------------------------")
		fmt.Println("Full Mode Enabled")
		fmt.Println("------------------------------------------")
		exchangeRates, err := filereader.ReadFile()
		if err != nil {
			fmt.Println("Error reading file:", err)
			return
		}
		fmt.Println("Target Currency:", moeda_destino)

		if moeda_destino == "BRL" {
			fmt.Println("------------------------------------------")
			fmt.Println("Valor convertido:", valor_em_brl, "BRL para", moeda_destino, "é:", valor_em_brl)
			return
		}
		actualRate, ok := exchangeRates.Rates[moeda_destino]
		fmt.Println("Actual Rate:", actualRate)
		if !ok {
			fmt.Printf("Error: Rate for currency %s not found\n", moeda_destino)
			return
		}

		fmt.Println("Exchange Rates:")
		for currency, rate := range exchangeRates.Rates {
			fmt.Printf("%s: %.2f\n", currency, rate)
		}

		amountConverted := rateconverter.ConvertRate(valor_em_brl, exchangeRates.Rates[moeda_destino])
		fmt.Println("------------------------------------------")
		fmt.Printf("Valor convertido: %.2f BRL para %s é: %.2f\n", valor_em_brl, moeda_destino, amountConverted)
	}
}
