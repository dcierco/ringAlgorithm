package main

import (
	"fmt"
	"sync"
)

type mensagem struct {
	tipo  int
	corpo []int
}

var chans []chan mensagem
var coordenadorID int
var controle chan int
var wg sync.WaitGroup

func main() {
	fmt.Println("\n   Anel de processos criado\n")

	// Criar canais de comunicação
	chans = make([]chan mensagem, 4)
	for i := 0; i < 4; i++ {
		chans[i] = make(chan mensagem)
	}

	controle = make(chan int)

	// Criar goroutines para os processos
	wg.Add(4)
	go ElectionStage(0, chans[3], chans[0], controle)
	go ElectionStage(1, chans[0], chans[1], controle)
	go ElectionStage(2, chans[1], chans[2], controle)
	go ElectionStage(3, chans[2], chans[3], controle)

	// Processo controlador
	fmt.Println("\n   Processo controlador criado\n")

	// Mudar o processo 0 para falho
	fmt.Println("Controle: mudar o processo 0 para falho")
	chans[0] <- mensagem{tipo: 2, corpo: []int{}}
	<-controle

	// Mudar o processo 1 para falho
	fmt.Println("Controle: mudar o processo 1 para falho")
	chans[1] <- mensagem{tipo: 2, corpo: []int{}}
	<-controle

	fmt.Println("\n   Processo controlador concluído\n")

	close(chans[3])
	close(chans[2])
	close(chans[1])
	close(chans[0])
	wg.Wait()
}

func ElectionStage(TaskId int, in chan mensagem, out chan mensagem, controle chan int) {
	defer wg.Done()

	var bFailed bool = false // Todos iniciam sem falha

	for {
		select {
		case msg, ok := <-in:
			if !ok {
				// Canal fechado, terminar a goroutine
				fmt.Printf("%2d: terminei \n", TaskId)
				return
			}

			fmt.Printf("%2d: recebi mensagem %d, [ %v ]\n", TaskId, msg.tipo, msg.corpo)

			switch msg.tipo {
			case 2:
				bFailed = true
				fmt.Printf("%2d: falho %v \n", TaskId, bFailed)
				fmt.Printf("%2d: coordenador atual %d\n", TaskId, coordenadorID)
				controle <- -5
			case 3:
				bFailed = false
				fmt.Printf("%2d: falho %v \n", TaskId, bFailed)
				fmt.Printf("%2d: coordenador atual %d\n", TaskId, coordenadorID)
				controle <- -5
			case 1:
				if !bFailed {
					msg.corpo = append(msg.corpo, TaskId) // Inclui o ID do processo na mensagem

					if len(msg.corpo) == len(chans) { // Volta completa, processo iniciador recebe a mensagem
						if TaskId == msg.corpo[0] {
							coordenadorID = TaskId
							fmt.Printf("%2d: sou o coordenador\n", TaskId)
						} else {
							out <- msg // Repassa a mensagem
						}
					} else {
						out <- msg // Repassa a mensagem
					}
				}
			case 4:
				fmt.Printf("%2d: não conheço este tipo de mensagem\n", TaskId)
				fmt.Printf("%2d: coordenador atual %d\n", TaskId, coordenadorID)
			}

		}
	}
}
