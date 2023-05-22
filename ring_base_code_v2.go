package main

import (
	"fmt"
	"sync"
	"time"
)

type mensagem struct {
	tipo  int    // tipo da mensagem para fazer o controle do que fazer (eleição, confirmação da eleicao, novo coordenador)
	corpo [3]int // conteudo da mensagem para colocar os ids (usar um tamanho compatível com o número de processos no anel)
}

var (
	chans = []chan mensagem{ // vetor de canais para formar o anel de eleição - chan[0], chan[1], chan[2], ...
		make(chan mensagem),
		make(chan mensagem),
		make(chan mensagem),
		make(chan mensagem),
	}
	controle       = make(chan int)
	wg             sync.WaitGroup // wg is used to wait for the program to finish
	numProcesses   = 4            // number of processes in the ring
	coordinatorID  = 0            // initial coordinator ID
	coordinatorSet = false        // flag to indicate if the coordinator is set
)

func ElectionControler(in chan int) {
	defer wg.Done()

	var temp mensagem

	// Mudar o processo 0 - canal de entrada 3 - para falho (defini mensagem tipo 2 pra isso)
	temp.tipo = 2
	chans[3] <- temp
	fmt.Printf("Controle: mudar o processo 0 para falho\n")
	fmt.Printf("Controle: confirmação %d\n", <-in) // receber e imprimir confirmação

	// Mudar o processo 1 - canal de entrada 0 - para falho (defini mensagem tipo 2 pra isso)
	temp.tipo = 2
	chans[0] <- temp
	fmt.Printf("Controle: mudar o processo 1 para falho\n")
	fmt.Printf("Controle: confirmação %d\n", <-in) // receber e imprimir confirmação

	// Matar os outros processos com mensagens não conhecidas (só para consumir a leitura)
	temp.tipo = 4
	chans[1] <- temp
	chans[2] <- temp

	fmt.Println("\n   Processo controlador concluído\n")
}

func ElectionStage(TaskId int, in chan mensagem, out chan mensagem, leader int) {
	defer wg.Done()

	var actualLeader int
	var bFailed bool = false

	actualLeader = leader

	temp := <-in
	fmt.Printf("%2d: recebi mensagem %d, [ %d, %d, %d ]\n", TaskId, temp.tipo, temp.corpo[0], temp.corpo[1], temp.corpo[2])

	switch temp.tipo {
	case 2:
		{
			bFailed = true
			fmt.Printf("%2d: falho %v \n", TaskId, bFailed)
			fmt.Printf("%2d: líder atual %d\n", TaskId, actualLeader)
			controle <- -5
		}
	case 3:
		{
			bFailed = false
			fmt.Printf("%2d: falho %v \n", TaskId, bFailed)
			fmt.Printf("%2d: líder atual %d\n", TaskId, actualLeader)
			controle <- -5
		}
	case 1:
		{
			if bFailed {
				// Process is failed, ignore election message
				fmt.Printf("%2d: processo falho, ignorando mensagem de eleição\n", TaskId)
			} else {
				// Process is not failed, participate in the election
				if temp.corpo[0] > actualLeader {
					actualLeader = temp.corpo[0]
					fmt.Printf("%2d: eleição em andamento, atualizando líder para %d\n", TaskId, actualLeader)
				}
				temp.corpo[TaskId] = TaskId
				out <- temp // Pass the updated election message to the next process in the ring
			}
		}
	default:
		{
			fmt.Printf("%2d: não conheço este tipo de mensagem\n", TaskId)
			fmt.Printf("%2d: líder atual %d\n", TaskId, actualLeader)
		}
	}

	fmt.Printf("%2d: terminei \n", TaskId)
}

func main() {
	wg.Add(numProcesses + 1) // Add a count for each process and the controller

	// Criar os processos do anel de eleição
	for i := 0; i < numProcesses; i++ {
		go ElectionStage(i, chans[i], chans[(i+1)%numProcesses], coordinatorID)
	}

	fmt.Println("\n   Anel de processos criado")

	// Criar o processo controlador
	go ElectionControler(controle)

	fmt.Println("\n   Processo controlador criado\n")

	// Simulação de solicitações externas
	go func() {
		for {
			// Simular uma solicitação externa
			time.Sleep(5 * time.Second)
			fmt.Println("\nSolicitação externa recebida")

			// Verificar se o coordenador está definido
			if !coordinatorSet {
				// Iniciar a eleição
				coordinatorSet = true
				fmt.Println("Iniciando eleição")
				temp := mensagem{
					tipo:  1,
					corpo: [3]int{coordinatorID, coordinatorID, coordinatorID},
				}
				chans[0] <- temp
			}
		}
	}()

	wg.Wait() // Wait for the goroutines to finish
}
