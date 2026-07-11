package main // O código demonstra o uso de ponteiros em Go. Ele mostra como passar valores por valor e por referência, permitindo a modificação do valor original de uma variável.

import "fmt" // importa o pacote fmt para usar a função Println

func main() {
	x := 10        // declara uma variável x do tipo int e atribui o valor 10
	fmt.Println(x) // imprime 10

	var p *int = &x // p recebe o endereço de memória de x

	*p = 20        // altera o valor da variável apontada por p, ou seja, x
	fmt.Println(x) // imprime 20

	change(x)      // passa uma cópia do valor de x
	fmt.Println(x) // x continua 20

	changePointer(&x) // passa o endereço de x
	fmt.Println(x)    // x agora é 40
}

func change(i int) {
	i = 30 // altera apenas a cópia local
}

func changePointer(p *int) {
	*p = 40 // altera o valor da variável original
}
