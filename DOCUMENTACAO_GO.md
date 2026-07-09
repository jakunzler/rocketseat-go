# Documentacao Completa da Linguagem Go

Go, também chamada de Golang, é uma linguagem de programação criada pelo Google com foco em simplicidade, produtividade, concorrência e alta performance. Ela é compilada, estaticamente tipada e muito usada para construir APIs, CLIs, sistemas distríbuidos, ferramentas de infraestrutura, servicos web e aplicações de backend.

## Sumario

- [Visão geral](#visao-geral)
- [Instalação e ferramentas](#instalacao-e-ferramentas)
- [Estrutura de um programa Go](#estrutura-de-um-programa-go)
- [Pacotes e imports](#pacotes-e-imports)
- [Variáveis e constantes](#variaveis-e-constantes)
- [Tipos de dados](#tipos-de-dados)
- [Operadores](#operadores)
- [Controle de fluxo](#controle-de-fluxo)
- [Funções](#funcoes)
- [Ponteiros](#ponteiros)
- [Arrays, slices e maps](#arrays-slices-e-maps)
- [Structs](#structs)
- [Métodos](#metodos)
- [Interfaces](#interfaces)
- [Tratamento de erros](#tratamento-de-erros)
- [Concorrencia](#concorrencia)
- [Generics](#generics)
- [Módulos e dependências](#modulos-e-dependencias)
- [Testes](#testes)
- [Formatação e convenções](#formatacao-e-convencoes)
- [Boas práticas](#boas-praticas)
- [Exemplo de API HTTP](#exemplo-de-api-http)
- [Comandos úteis](#comandos-uteis)

## Visão geral

Principais características da linguagem:

- Compilada: o código Go é compilado para binários nativos.
- Tipagem estática: os tipos são verificados em tempo de compilação.
- Sintaxe simples: a linguagem evita recursos excessivamente complexos.
- Garbage collector: a memória é gerenciada automaticamente.
- Concorrência nativa: goroutines e channels fazem parte da linguagem.
- Biblioteca padrão forte: inclui pacotes para HTTP, JSON, testes, arquivos, criptografia e muito mais.
- Build rápido: o processo de compilação costuma ser muito veloz.

Um programa Go normalmente é organizado em pacotes. O pacote `main` representa um programa executável.

## Instalação e ferramentas

Verifique se o Go esta instalado:

```bash
go version
```

Crie um modulo:

```bash
go mod init github.com/usuario/projeto
```

Execute um arquivo:

```bash
go run main.go
```

Compile o projeto:

```bash
go build
```

Rode testes:

```bash
go test ./...
```

Formate o codigo:

```bash
go fmt ./...
```

## Estrutura de um programa Go

Exemplo basico:

```go
package main

import "fmt"

func main() {
	fmt.Println("Ola, Go!")
}
```

Explicacao:

- `package main`: define o pacote do arquivo.
- `import "fmt"`: importa o pacote de formatacao.
- `func main()`: ponto de entrada de um programa executavel.
- `fmt.Println`: imprime uma mensagem no terminal.

## Pacotes e imports

Todo arquivo Go pertence a um pacote:

```go
package main
```

Imports simples:

```go
import "fmt"
```

Imports agrupados:

```go
import (
	"fmt"
	"time"
)
```

Pacotes da biblioteca padrao sao importados pelo nome, como `fmt`, `time`, `net/http`, `encoding/json` e `os`.

Pacotes de terceiros sao adicionados ao projeto com:

```bash
go get github.com/exemplo/pacote
```

## Variaveis e constantes

Declaracao com tipo explicito:

```go
var nome string = "Ana"
var idade int = 30
```

Declaracao com inferencia de tipo:

```go
var cidade = "Sao Paulo"
```

Declaracao curta, usada dentro de funcoes:

```go
curso := "Go"
```

Multiplas variaveis:

```go
var (
	nome  = "Ana"
	idade = 30
)
```

Constantes:

```go
const pi = 3.14159
const linguagem = "Go"
```

Constantes nao podem ter o valor alterado depois de declaradas.

## Tipos de dados

### Tipos numericos

Inteiros:

```go
var a int = 10
var b int8 = 127
var c int64 = 1000000
```

Inteiros sem sinal:

```go
var idade uint = 25
```

Ponto flutuante:

```go
var preco float64 = 19.99
```

Complexos:

```go
var numero complex64 = 1 + 2i
```

### Texto

Strings sao sequencias imutaveis de bytes:

```go
mensagem := "Ola"
```

Caracteres Unicode podem ser representados com `rune`:

```go
var letra rune = 'A'
```

### Booleanos

```go
ativo := true
finalizado := false
```

### Valor zero

Quando uma variavel e declarada sem valor inicial, Go atribui um valor zero:

- `int`: `0`
- `float64`: `0`
- `bool`: `false`
- `string`: `""`
- ponteiros, slices, maps, channels, interfaces e funcoes: `nil`

## Operadores

Aritmeticos:

```go
soma := 10 + 5
subtracao := 10 - 5
multiplicacao := 10 * 5
divisao := 10 / 5
resto := 10 % 3
```

Comparacao:

```go
10 == 10
10 != 5
10 > 5
10 >= 10
5 < 10
5 <= 5
```

Logicos:

```go
true && false
true || false
!true
```

Atribuicao:

```go
x := 10
x += 5
x -= 2
x *= 3
x /= 2
```

## Controle de fluxo

### if, else if e else

```go
idade := 18

if idade >= 18 {
	fmt.Println("Maior de idade")
} else {
	fmt.Println("Menor de idade")
}
```

Com inicializacao curta:

```go
if nota := 8; nota >= 7 {
	fmt.Println("Aprovado")
}
```

### switch

```go
dia := "segunda"

switch dia {
case "segunda":
	fmt.Println("Inicio da semana")
case "sexta":
	fmt.Println("Quase fim de semana")
default:
	fmt.Println("Outro dia")
}
```

Switch sem expressao:

```go
idade := 20

switch {
case idade < 18:
	fmt.Println("Menor")
case idade >= 18:
	fmt.Println("Adulto")
}
```

### for

Go possui apenas o laco `for`.

Formato tradicional:

```go
for i := 0; i < 5; i++ {
	fmt.Println(i)
}
```

Como `while`:

```go
contador := 0

for contador < 5 {
	fmt.Println(contador)
	contador++
}
```

Laco infinito:

```go
for {
	fmt.Println("executando")
	break
}
```

Com `range`:

```go
nomes := []string{"Ana", "Joao", "Maria"}

for indice, nome := range nomes {
	fmt.Println(indice, nome)
}
```

## Funcoes

Declaracao basica:

```go
func somar(a int, b int) int {
	return a + b
}
```

Parametros do mesmo tipo podem ser agrupados:

```go
func somar(a, b int) int {
	return a + b
}
```

Multiplos retornos:

```go
func dividir(a, b int) (int, error) {
	if b == 0 {
		return 0, fmt.Errorf("divisao por zero")
	}

	return a / b, nil
}
```

Retornos nomeados:

```go
func coordenadas() (x int, y int) {
	x = 10
	y = 20
	return
}
```

Funcoes variadicas:

```go
func somarTudo(numeros ...int) int {
	total := 0

	for _, numero := range numeros {
		total += numero
	}

	return total
}
```

Funcoes anonimas:

```go
dobro := func(numero int) int {
	return numero * 2
}

fmt.Println(dobro(5))
```

## Ponteiros

Um ponteiro guarda o endereco de memoria de uma variavel.

```go
numero := 10
ponteiro := &numero

fmt.Println(numero)   // valor
fmt.Println(ponteiro) // endereco
fmt.Println(*ponteiro) // valor apontado
```

Alterando um valor por ponteiro:

```go
func incrementar(numero *int) {
	*numero++
}

func main() {
	valor := 10
	incrementar(&valor)
	fmt.Println(valor) // 11
}
```

Use ponteiros quando precisar alterar um valor original, evitar copias grandes ou representar ausencia com `nil`.

## Arrays, slices e maps

### Arrays

Arrays possuem tamanho fixo:

```go
var numeros [3]int
numeros[0] = 10
numeros[1] = 20
numeros[2] = 30
```

Inicializacao direta:

```go
numeros := [3]int{10, 20, 30}
```

### Slices

Slices sao mais usados que arrays porque possuem tamanho dinamico.

```go
nomes := []string{"Ana", "Joao"}
nomes = append(nomes, "Maria")
```

Criando com `make`:

```go
numeros := make([]int, 0, 10)
numeros = append(numeros, 1)
```

Fatiamento:

```go
valores := []int{10, 20, 30, 40}
parte := valores[1:3] // 20, 30
```

Percorrendo:

```go
for indice, valor := range valores {
	fmt.Println(indice, valor)
}
```

### Maps

Maps armazenam pares chave-valor:

```go
idades := map[string]int{
	"Ana":  30,
	"Joao": 25,
}
```

Adicionar ou alterar:

```go
idades["Maria"] = 28
```

Ler com verificacao:

```go
idade, existe := idades["Ana"]
if existe {
	fmt.Println(idade)
}
```

Remover:

```go
delete(idades, "Joao")
```

## Structs

Structs agrupam campos relacionados:

```go
type Usuario struct {
	Nome  string
	Email string
	Idade int
}
```

Criando uma struct:

```go
usuario := Usuario{
	Nome:  "Ana",
	Email: "ana@email.com",
	Idade: 30,
}
```

Acessando campos:

```go
fmt.Println(usuario.Nome)
```

Struct anonima:

```go
produto := struct {
	Nome  string
	Preco float64
}{
	Nome:  "Teclado",
	Preco: 150.0,
}
```

Tags em structs, muito usadas com JSON:

```go
type Produto struct {
	ID    int     `json:"id"`
	Nome  string  `json:"nome"`
	Preco float64 `json:"preco"`
}
```

## Metodos

Metodos sao funcoes associadas a um tipo.

```go
type Usuario struct {
	Nome string
}

func (u Usuario) Saudacao() string {
	return "Ola, " + u.Nome
}
```

Uso:

```go
usuario := Usuario{Nome: "Ana"}
fmt.Println(usuario.Saudacao())
```

Receiver por valor:

```go
func (u Usuario) Saudacao() string {
	return "Ola, " + u.Nome
}
```

Receiver por ponteiro:

```go
func (u *Usuario) AlterarNome(nome string) {
	u.Nome = nome
}
```

Use receiver por ponteiro quando o metodo precisa alterar o valor original ou quando a struct e grande.

## Interfaces

Interfaces definem comportamentos. Um tipo implementa uma interface implicitamente quando possui todos os metodos exigidos.

```go
type Notificador interface {
	Notificar(mensagem string) error
}
```

Implementacao:

```go
type Email struct{}

func (e Email) Notificar(mensagem string) error {
	fmt.Println("Enviando email:", mensagem)
	return nil
}
```

Uso:

```go
func enviar(n Notificador) {
	err := n.Notificar("Bem-vindo")
	if err != nil {
		fmt.Println(err)
	}
}
```

A interface vazia `any` aceita qualquer tipo:

```go
func imprimir(valor any) {
	fmt.Println(valor)
}
```

Type assertion:

```go
valor := any("Go")
texto, ok := valor.(string)
if ok {
	fmt.Println(texto)
}
```

Type switch:

```go
func descrever(valor any) {
	switch v := valor.(type) {
	case string:
		fmt.Println("string:", v)
	case int:
		fmt.Println("int:", v)
	default:
		fmt.Println("tipo desconhecido")
	}
}
```

## Tratamento de erros

Go usa valores de erro explicitos, normalmente como ultimo retorno de uma funcao.

```go
func buscarUsuario(id int) (string, error) {
	if id <= 0 {
		return "", fmt.Errorf("id invalido")
	}

	return "Ana", nil
}
```

Uso:

```go
usuario, err := buscarUsuario(1)
if err != nil {
	fmt.Println("erro:", err)
	return
}

fmt.Println(usuario)
```

Criando erros:

```go
import "errors"

var ErrNaoEncontrado = errors.New("nao encontrado")
```

Envolvendo erros:

```go
return fmt.Errorf("buscar usuario: %w", err)
```

Verificando erros:

```go
if errors.Is(err, ErrNaoEncontrado) {
	fmt.Println("usuario nao encontrado")
}
```

Extraindo tipos de erro:

```go
var pathErr *os.PathError
if errors.As(err, &pathErr) {
	fmt.Println(pathErr.Path)
}
```

## Concorrencia

Go tem suporte nativo a concorrencia por meio de goroutines e channels.

### Goroutines

Uma goroutine e uma funcao executando concorrentemente.

```go
go func() {
	fmt.Println("executando em paralelo")
}()
```

Exemplo com `sync.WaitGroup`:

```go
var wg sync.WaitGroup

wg.Add(1)
go func() {
	defer wg.Done()
	fmt.Println("tarefa")
}()

wg.Wait()
```

### Channels

Channels permitem comunicacao entre goroutines.

```go
canal := make(chan string)

go func() {
	canal <- "mensagem"
}()

mensagem := <-canal
fmt.Println(mensagem)
```

Channel com buffer:

```go
canal := make(chan int, 2)
canal <- 1
canal <- 2
```

Fechando channel:

```go
close(canal)
```

Percorrendo channel:

```go
for valor := range canal {
	fmt.Println(valor)
}
```

### select

`select` espera por operacoes em multiplos channels.

```go
select {
case msg := <-canalA:
	fmt.Println(msg)
case msg := <-canalB:
	fmt.Println(msg)
case <-time.After(time.Second):
	fmt.Println("timeout")
}
```

### Mutex

Use `sync.Mutex` para proteger dados compartilhados.

```go
var mu sync.Mutex
contador := 0

mu.Lock()
contador++
mu.Unlock()
```

Tambem e comum usar `defer`:

```go
mu.Lock()
defer mu.Unlock()
```

## Generics

Generics permitem escrever funcoes e tipos que trabalham com diferentes tipos sem perder seguranca de tipo.

Funcao generica:

```go
func Primeiro[T any](valores []T) T {
	return valores[0]
}
```

Uso:

```go
numeros := []int{1, 2, 3}
fmt.Println(Primeiro(numeros))
```

Constraints:

```go
type Numero interface {
	~int | ~int64 | ~float64
}

func Somar[T Numero](a, b T) T {
	return a + b
}
```

Use generics quando eles reduzirem duplicacao real. Para muitos casos, interfaces simples continuam sendo a melhor opcao.

## Modulos e dependencias

Um modulo Go e definido pelo arquivo `go.mod`.

Criar modulo:

```bash
go mod init github.com/usuario/projeto
```

Adicionar dependencia:

```bash
go get github.com/gin-gonic/gin
```

Organizar dependencias:

```bash
go mod tidy
```

Exemplo de `go.mod`:

```go
module github.com/usuario/projeto

go 1.22

require github.com/gin-gonic/gin v1.10.0
```

O arquivo `go.sum` registra hashes das dependencias para garantir reproducibilidade.

## Testes

Arquivos de teste terminam com `_test.go`.

Exemplo:

```go
package calculadora

import "testing"

func Somar(a, b int) int {
	return a + b
}

func TestSomar(t *testing.T) {
	resultado := Somar(2, 3)
	esperado := 5

	if resultado != esperado {
		t.Fatalf("esperado %d, recebeu %d", esperado, resultado)
	}
}
```

Rodar testes:

```bash
go test ./...
```

Com cobertura:

```bash
go test -cover ./...
```

Benchmarks:

```go
func BenchmarkSomar(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Somar(2, 3)
	}
}
```

Rodar benchmarks:

```bash
go test -bench=. ./...
```

Testes em tabela:

```go
func TestSomarTabela(t *testing.T) {
	casos := []struct {
		nome     string
		a        int
		b        int
		esperado int
	}{
		{"positivos", 2, 3, 5},
		{"negativos", -2, -3, -5},
	}

	for _, caso := range casos {
		t.Run(caso.nome, func(t *testing.T) {
			resultado := Somar(caso.a, caso.b)
			if resultado != caso.esperado {
				t.Fatalf("esperado %d, recebeu %d", caso.esperado, resultado)
			}
		})
	}
}
```

## Formatacao e convencoes

Go possui formatacao padronizada com `gofmt` ou `go fmt`.

```bash
go fmt ./...
```

Convencoes comuns:

- Nomes exportados comecam com letra maiuscula: `Usuario`, `CriarPedido`.
- Nomes nao exportados comecam com letra minuscula: `usuario`, `criarPedido`.
- Pacotes costumam ter nomes curtos e em minusculo: `user`, `http`, `store`.
- Evite nomes genericos como `utils` quando houver um nome de dominio mais claro.
- Erros devem ser tratados explicitamente.
- O codigo deve ser simples de ler antes de ser sofisticado.

Comentarios de simbolos exportados devem comecar com o nome do simbolo:

```go
// Usuario representa uma pessoa cadastrada no sistema.
type Usuario struct {
	Nome string
}
```

## Boas praticas

- Trate erros perto de onde eles acontecem.
- Retorne erros com contexto usando `fmt.Errorf`.
- Prefira composicao a heranca. Go nao possui heranca tradicional.
- Mantenha interfaces pequenas.
- Declare interfaces no pacote consumidor quando fizer sentido.
- Evite criar abstracoes antes de haver necessidade real.
- Use `context.Context` para cancelamento, timeout e valores de escopo de requisicao.
- Nao compartilhe memoria desnecessariamente entre goroutines.
- Quando houver dados compartilhados, use channels, mutexes ou outra estrategia clara de sincronizacao.
- Escreva testes para regras de negocio, casos de erro e integracoes importantes.
- Use `go mod tidy` para manter dependencias limpas.
- Rode `go fmt` antes de versionar alteracoes.

## Exemplo de API HTTP

Exemplo usando apenas a biblioteca padrao:

```go
package main

import (
	"encoding/json"
	"log"
	"net/http"
)

type Resposta struct {
	Mensagem string `json:"mensagem"`
}

func saudacaoHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	resposta := Resposta{Mensagem: "Ola, Go!"}

	if err := json.NewEncoder(w).Encode(resposta); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func main() {
	http.HandleFunc("/saudacao", saudacaoHandler)

	log.Println("Servidor rodando em http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
```

Execute:

```bash
go run main.go
```

Acesse:

```text
http://localhost:8080/saudacao
```

## Comandos uteis

Criar modulo:

```bash
go mod init nome-do-modulo
```

Executar:

```bash
go run .
```

Compilar:

```bash
go build
```

Instalar dependencias:

```bash
go get pacote
```

Limpar dependencias:

```bash
go mod tidy
```

Formatar:

```bash
go fmt ./...
```

Analisar problemas comuns:

```bash
go vet ./...
```

Testar:

```bash
go test ./...
```

Ver documentacao de um pacote:

```bash
go doc fmt
```

Listar variaveis de ambiente do Go:

```bash
go env
```

## Glossario rapido

- `package`: unidade de organizacao de codigo.
- `module`: conjunto versionado de pacotes definido por `go.mod`.
- `goroutine`: execucao concorrente leve.
- `channel`: mecanismo de comunicacao entre goroutines.
- `interface`: contrato de comportamento.
- `struct`: tipo composto por campos.
- `slice`: sequencia dinamica baseada em array.
- `map`: estrutura chave-valor.
- `defer`: adia a execucao de uma chamada ate o fim da funcao.
- `panic`: interrompe o fluxo normal por erro grave.
- `recover`: permite recuperar uma goroutine em panic dentro de uma funcao deferida.

## Proximos estudos recomendados

Depois dos fundamentos, vale estudar:

- `context.Context`
- APIs REST com `net/http`
- Manipulacao de JSON
- Banco de dados com `database/sql`
- Migrations
- Autenticacao e autorizacao
- Logs estruturados
- Testes de integracao
- Concorrencia aplicada
- Clean Architecture em Go
- Observabilidade com metricas, traces e logs
