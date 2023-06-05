package main

import (
	"fmt"
	"sync"
)

type mensagem struct {
	tipo  int    // tipo da mensagem para fazer o controle do que fazer (eleição, confirmação da eleição, novo coordenador)
	corpo []int  // conteúdo da mensagem para armazenar os IDs dos processos no anel
}

var (
	chans         = []chan mensagem{ // vetor de canais para formar o anel de eleição - chan[0], chan[1], chan[2], ...
		make(chan mensagem),
		make(chan mensagem),
		make(chan mensagem),
		make(chan mensagem),
	}
	controle      = make(chan int)
	coordenadorID = 0
	wg            sync.WaitGroup // wg is used to wait for the program to finish
)

func ElectionControler(in chan int) {
	defer wg.Done()

	var temp mensagem

	// Simular a detecção de que o coordenador não está mais ativo (por exemplo, receber uma mensagem externa)
	// Neste exemplo, o processo 0 é definido como falho (mensagem tipo 2)

	temp.tipo = 2
	chans[0] <- temp
	fmt.Printf("Controle: mudar o processo 0 para falho\n")

	<-in // Aguardar confirmação

	// Simular a falha do processo 1 (mensagem tipo 2)
	temp.tipo = 2
	chans[1] <- temp
	fmt.Printf("Controle: mudar o processo 1 para falho\n")
	<-in // Aguardar confirmação

	// Simular a falha de outros processos com mensagens desconhecidas (só para consumir a leitura)
	temp.tipo = 4
	chans[2] <- temp
	chans[3] <- temp

	fmt.Println("\n   Processo controlador concluído\n")
}

func ElectionStage(TaskId int, in chan mensagem, out chan mensagem) {
	defer wg.Done()

	var bFailed bool = false // Todos iniciam sem falha

	temp := <-in // ler mensagem
	fmt.Printf("%2d: recebi mensagem %d, [ %v ]\n", TaskId, temp.tipo, temp.corpo)

	switch temp.tipo {
	case 2:
		{
			bFailed = true
			fmt.Printf("%2d: falho %v \n", TaskId, bFailed)
			fmt.Printf("%2d: coordenador atual %d\n", TaskId, coordenadorID)
			controle <- -5
		}
	case 3:
		{
			bFailed = false
			fmt.Printf("%2d: falho %v \n", TaskId, bFailed)
			fmt.Printf("%2d: coordenador atual %d\n", TaskId, coordenadorID)
			controle <- -5
		}
	default:
		{
			// Recebeu uma mensagem de eleição
			if !bFailed {
				temp.corpo = append(temp.corpo, TaskId) // Inclui o ID do processo na mensagem

				if len(temp.corpo) == len(chans) { // A mensagem completou uma volta no anel
					if temp.corpo[0] == TaskId { // O processo que iniciou a eleição recebeu a mensagem de volta
						// Escolher o novo coordenador com base no maior ID
						maxID := -1
						for _, id := range temp.corpo {
							if id > maxID {
								maxID = id
							}
						}
						coordenadorID = maxID
						fmt.Printf("%2d: eleição concluída, novo coordenador: %d\n", TaskId, coordenadorID)

						// Enviar mensagem informando o novo coordenador para todos os processos no anel
						temp.tipo = 5
						temp.corpo = []int{coordenadorID}
						out <- temp
						fmt.Printf("%2d: repassou mensagem de novo coordenador [ %v ]\n", TaskId, temp.corpo)
					} else {
						out <- temp // Repassar a mensagem para o próximo processo no anel
						fmt.Printf("%2d: repassou mensagem de eleição [ %v ]\n", TaskId, temp.corpo)
					}
				} else {
					out <- temp // Repassar a mensagem para o próximo processo no anel
					fmt.Printf("%2d: repassou mensagem de eleição [ %v ]\n", TaskId, temp.corpo)
				}
			}
		}
	}

	fmt.Printf("%2d: terminei \n", TaskId)
}

func main() {

	wg.Add(5) // Adiciona uma contagem de cinco, uma para cada goroutine

	// Criar os processos do anel de eleição

	go ElectionStage(0, chans[3], chans[0])
	go ElectionStage(1, chans[0], chans[1])
	go ElectionStage(2, chans[1], chans[2])
	go ElectionStage(3, chans[2], chans[3])

	fmt.Println("\n   Anel de processos criado")

	// Criar o processo controlador

	go ElectionControler(controle)

	fmt.Println("\n   Processo controlador criado\n")

	// Simular uma nova eleição (por exemplo, receber uma mensagem externa)

	temp := mensagem{
		tipo:  1,
		corpo: []int{0}, // O processo 0 inicia a eleição
	}
	chans[0] <- temp // Enviar mensagem de eleição para o processo 0

	// Aguardar o término das goroutines
	wg.Wait()
}
