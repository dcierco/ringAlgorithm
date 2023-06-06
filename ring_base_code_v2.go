// sistemas distribuidos (eleicao em anel)

package main

import (
	"fmt"
	"os"
	"sync"
)

var maxIter int = 30 // número máximo de iterações

type mensagem struct {
	tipo   int    // tipo da mensagem para fazer o controle do que fazer (eleição, confirmacao da eleicao)
	corpo  [3]int // conteudo da mensagem para colocar os ids (usar um tamanho compativel com o numero de processos no anel)
	falho  bool   // indica se o processo que enviou a mensagem está falho ou não
	origem int    // indica o id do processo que enviou a mensagem originalmente
}

var (
	chans = []chan mensagem{ // vetor de canais para formar o anel de eleicao - chan[0], chan[1] and chan[2] ...
		make(chan mensagem),
		make(chan mensagem),
		make(chan mensagem),
		make(chan mensagem),
	}
	controle = make(chan int)
	wg       sync.WaitGroup // wg is used to wait for the program to finish
)

func ElectionControler(in chan int) {
	defer wg.Done()

	var temp mensagem

	// comandos para o anel iniciam aqui

	// iniciar uma eleicao pelo processo 2 - canal de entrada 1 - (defini mensagem tipo 1 pra isto)

	temp.tipo = 1
	temp.corpo[0] = 2           // colocar o id do processo que inicia a eleicao
	temp.falho = false          // indicar que o processo não está falho
	temp.origem = temp.corpo[0] // indicar que o processo é o originador da mensagem
	chans[1] <- temp
	fmt.Printf("Controle: iniciar uma eleicao pelo processo %d\n", temp.corpo[0])

	fmt.Printf("Controle: confirmação %d\n", <-in) // receber e imprimir confirmação

	// mudar o processo 0 - canal de entrada 3 - para falho (defini mensagem tipo 2 pra isto)

	temp.tipo = 2
	temp.falho = true // indicar que o processo está falho
	chans[3] <- temp
	fmt.Printf("Controle: mudar o processo 0 para falho\n")

	fmt.Printf("Controle: confirmação %d\n", <-in) // receber e imprimir confirmação

	// iniciar uma eleicao pelo processo 3 - canal de entrada 2 - (defini mensagem tipo 1 pra isto)

	temp.tipo = 1
	temp.corpo[0] = 3           // colocar o id do processo que inicia a eleicao
	temp.falho = false          // indicar que o processo não está falho
	temp.origem = temp.corpo[0] // indicar que o processo é o originador da mensagem
	chans[2] <- temp
	fmt.Printf("Controle: iniciar uma eleicao pelo processo %d\n", temp.corpo[0])

	fmt.Printf("Controle: confirmação %d\n", <-in) // receber e imprimir confirmação

	// mudar o processo 1 - canal de entrada 0 - para falho (defini mensagem tipo 2 pra isto)

	temp.tipo = 2
	temp.falho = true // indicar que o processo está falho
	chans[0] <- temp
	fmt.Printf("Controle: mudar o processo 1 para falho\n")
	fmt.Printf("Controle: confirmação %d\n", <-in) // receber e imprimir confirmação

	// matar os outrs processos com mensagens não conhecidas (só pra cosumir a leitura)

	temp.tipo = 4
	chans[1] <- temp
	chans[2] <- temp

	fmt.Println("\n   Processo controlador concluído\n")
}

func ElectionStage(TaskId int, in chan mensagem, out chan mensagem, leader int) {
	defer wg.Done()

	// variaveis locais que indicam se este processo é o lider e se esta ativo

	var actualLeader int
	var bFailed bool = false   // todos inciam sem falha
	var bElection bool = false // indica se o processo iniciou uma eleicao

	actualLeader = leader // indicação do lider veio por parâmetro

	for { // loop infinito para receber mensagens continuamente
		temp := <-in // ler mensagem
		fmt.Printf("%2d: recebi mensagem %d, [ %d, %d, %d ], falho: %v, origem: %d\n", TaskId, temp.tipo, temp.corpo[0], temp.corpo[1], temp.corpo[2], temp.falho, temp.origem)

		switch temp.tipo {
		case 0:
			{
				if !bFailed {
					actualLeader = temp.corpo[0] // atualizar o lider com o id na mensagem
					fmt.Printf("%2d: novo lider %d\n", TaskId, actualLeader)
					bElection = false           // encerrar a eleicao
					if TaskId != actualLeader { // se nao for o lider, enviar a mensagem para o proximo no anel
						out <- temp
					} else { // se for o lider, enviar uma confirmacao para o controlador
						controle <- actualLeader
					}
				} else { // se estiver falho, ignorar a mensagem ou repassar para o próximo no anel se for de origem diferente
					if temp.origem != TaskId {
						fmt.Printf("%2d: repassando mensagem %d\n", TaskId, temp.tipo)
						out <- temp
					} else {
						fmt.Printf("%2d: ignorando mensagem %d\n", TaskId, temp.tipo)
					}
				}
			}
		case 1:
			{
				if !bFailed {
					if temp.corpo[0] == TaskId { // se for o processo que iniciou a eleicao
						bElection = false              // encerrar a eleicao
						actualLeader = max(temp.corpo) // escolher o maior id na mensagem como o novo lider
						fmt.Printf("%2d: novo lider %d\n", TaskId, actualLeader)
						temp.tipo = 1 // mudar o tipo da mensagem para anunciar o novo lider
						out <- temp   // enviar a mensagem para o proximo no anel
					} else { // se nao for o processo que iniciou a eleicao
						if bElection { // se ja estiver em uma eleicao
							if TaskId > temp.corpo[2] { // se seu id for maior que o ultimo na mensagem
								temp.corpo[2] = temp.corpo[1] // deslocar os ids na mensagem
								temp.corpo[1] = temp.corpo[0]
								temp.corpo[0] = TaskId // adicionar seu id na mensagem
							}
							out <- temp // enviar a mensagem para o proximo no anel
						} else { // se nao estiver em uma eleicao
							bElection = true            // iniciar uma eleicao
							if TaskId > temp.corpo[0] { // se seu id for maior que o primeiro na mensagem
								temp.corpo[2] = temp.corpo[1] // deslocar os ids na mensagem
								temp.corpo[1] = temp.corpo[0]
								temp.corpo[0] = TaskId // adicionar seu id na mensagem
							}
							out <- temp // enviar a mensagem para o proximo no anel
						}
					}
				} else { // se estiver falho, ignorar a mensagem ou repassar para o próximo no anel se for de origem diferente
					if temp.origem != TaskId {
						fmt.Printf("%2d: repassando mensagem %d\n", TaskId, temp.tipo)
						out <- temp
					} else {
						fmt.Printf("%2d: ignorando mensagem %d\n", TaskId, temp.tipo)
					}
				}
			}
		case 2:
			{
				bFailed = true
				fmt.Printf("%2d: falho %v \n", TaskId, bFailed)
				fmt.Printf("%2d: lider atual %d\n", TaskId, actualLeader)
				controle <- -5
			}
		case 3:
			{
				bFailed = false
				fmt.Printf("%2d: falho %v \n", TaskId, bFailed)
				fmt.Printf("%2d: lider atual %d\n", TaskId, actualLeader)
				controle <- -5
			}
		default:
			{
				fmt.Printf("%2d: não conheço este tipo de mensagem\n", TaskId)
				fmt.Printf("%2d: lider atual %d\n", TaskId, actualLeader)
			}
		}

		fmt.Printf("%2d: terminei \n", TaskId)
		maxIter--         // decrementar o contador
		if maxIter == 0 { // verificar se atingiu o limite
			os.Exit(0) // encerrar o programa
		}
	}
}

func max(arr [3]int) int {
	max := arr[0]
	for _, v := range arr {
		if v > max {
			max = v
		}
	}
	return max
}

func main() {

	wg.Add(5) // Add a count of four, one for each goroutine

	// criar os processo do anel de eleicao

	go ElectionStage(0, chans[3], chans[1], 0) // este é o lider
	go ElectionStage(1, chans[1], chans[2], 0) // não é lider, é o processo 0
	go ElectionStage(2, chans[1], chans[2], 0) // não é lider, é o processo 0
	go ElectionStage(3, chans[2], chans[3], 0) // não é lider, é o processo 0

	fmt.Println("\n   Anel de processos criado")

	// criar o processo controlador

	go ElectionControler(controle)

	fmt.Println("\n   Processo controlador criado\n")

	wg.Wait() // Wait for the goroutines to finish\
}
