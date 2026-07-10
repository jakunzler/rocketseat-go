package filereader

import (
	"encoding/json"
	"fmt"
	"os"
)

type ExchangeRates struct {
	Base  string             `json:"base"`
	Date  string             `json:"date"`
	Rates map[string]float64 `json:"rates"`
}

func ReadFile() (ExchangeRates, error) {
	data, err := os.ReadFile("./assets/rates.json")
	if err != nil {
		fmt.Println("Error reading file:", err)
		return ExchangeRates{}, err
	}

	var dataMap map[string]any
	err = json.Unmarshal(data, &dataMap)
	if err != nil {
		fmt.Println("Error parsing JSON:", err)
		return ExchangeRates{}, err
	}

	base := dataMap["base"].(string)
	fmt.Println("Base Currency:", base)

	date := dataMap["date"].(string)
	fmt.Println("Date:", date)

	rates, ok := dataMap["rates"].(map[string]any)
	if !ok {
		fmt.Println("Error: 'rates' key not found or is not a map")
		return ExchangeRates{}, fmt.Errorf("error: 'rates' key not found or is not a map")
	}

	// fmt.Println("==========================================")
	// fmt.Println("Criando uma struct para armazenar os dados do JSON")
	// Criando uma struct para armazenar os dados do JSON
	var exchangeRates ExchangeRates
	exchangeRates.Base = base
	exchangeRates.Date = date
	exchangeRates.Rates = make(map[string]float64)
	for currency, rate := range rates {
		exchangeRates.Rates[currency] = rate.(float64)
	}

	// fmt.Println("Struct ExchangeRates:")
	// fmt.Printf("%+v\n", exchangeRates)

	return exchangeRates, nil
}
