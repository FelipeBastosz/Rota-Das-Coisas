// package main
//
// import (
//
//	"bufio"
//	"fmt"
//	"net"
//	"strings"
//	"sync"
//	"time"
//
// )
//
//	func main() {
//		totalClientes := 100
//		var wg sync.WaitGroup
//
//		fmt.Printf("Iniciando ataque de %d clientes simultâneos...\n", totalClientes)
//
//		// Dispara 100 goroutines simulando clientes reais
//		for i := 0; i < totalClientes; i++ {
//			wg.Add(1)
//			go func(clienteID int) {
//				defer wg.Done()
//
//				conn, err := net.Dial("tcp", "localhost:8080")
//				if err != nil {
//					fmt.Printf("[Bot_%d] ERRO DE CONEXÃO: Servidor offline?\n", clienteID)
//					return
//				}
//				defer conn.Close()
//
//				// 1. Handshake: Nome do cliente
//				fmt.Fprintf(conn, "Bot_%d\n", clienteID)
//				time.Sleep(50 * time.Millisecond) // Pequeno delay pro servidor registrar o nome
//
//				// 2. Manda o comando para o atuador
//				fmt.Fprintf(conn, "atuar atuador_01 ligar\n")
//
//				// 3. Fica escutando as respostas do servidor pacientemente usando Scanner
//				scanner := bufio.NewScanner(conn)
//
//				// Damos um tempo limite de 5 segundos
//				conn.SetReadDeadline(time.Now().Add(5 * time.Second))
//
//				for scanner.Scan() {
//					mensagem := scanner.Text()
//
//					// Se for só o aviso de entrada ou de sistema intermediário, apenas ignoramos
//					if strings.Contains(mensagem, "[SISTEMA] Comando:") {
//						continue
//					}
//
//					// Se a mensagem contiver a resposta final do atuador, imprimimos e encerramos a escuta!
//					if strings.Contains(mensagem, "[RESPOSTA ATUADOR]") {
//						fmt.Printf("[Bot_%d] SUCESSO! %s\n", clienteID, mensagem)
//						break // Sai do loop e finaliza a goroutine deste bot
//					}
//				}
//
//				if err := scanner.Err(); err != nil {
//					fmt.Printf("[Bot_%d] FALHA: Fiquei no vácuo por muito tempo! Erro: %v\n", clienteID, err)
//				}
//
//			}(i)
//		}
//
//		wg.Wait()
//		fmt.Println("Teste de estresse finalizado. A fila suportou a carga!")
//	}
package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"net"
	"strings"
	"sync"
	"time"
)

func main() {
	totalClientes := 3000
	var wg sync.WaitGroup

	fmt.Printf("Iniciando ataque de %d clientes simultâneos...\n", totalClientes)

	// Dispara 500 goroutines simulando clientes reais
	for i := 0; i < totalClientes; i++ {
		wg.Add(1)
		go func(clienteID int) {
			defer wg.Done()

			// 1. O SEGREDO DO TESTE EM LOTE: "Espalhar" as conexões iniciais
			// Cada bot vai esperar um tempo aleatório entre 0 e 2 segundos antes de tentar conectar
			delayInicio := time.Duration(rand.Intn(2000)) * time.Millisecond
			time.Sleep(delayInicio)

			conn, err := net.Dial("tcp", "localhost:8080")
			if err != nil {
				fmt.Printf("[Bot_%d] ERRO DE CONEXÃO: %v\n", clienteID, err)
				return
			}
			defer conn.Close()

			// ... resto do código igualzinho ao anterior ...
			fmt.Fprintf(conn, "Bot_%d\n", clienteID)
			time.Sleep(50 * time.Millisecond)

			fmt.Fprintf(conn, "atuar atuador_01 ligar\n")

			scanner := bufio.NewScanner(conn)

			// ATENÇÃO: Aumente o Deadline! Com 500 caras na fila e um delay de 600ms no atuador,
			// o último cara da fila vai demorar no mínimo 5 minutos para receber a resposta!
			// Para o teste não quebrar por timeout, coloque 10 minutos.
			conn.SetReadDeadline(time.Now().Add(10 * time.Minute))

			for scanner.Scan() {
				mensagem := scanner.Text()

				if strings.Contains(mensagem, "[SISTEMA] Comando:") {
					continue
				}

				if strings.Contains(mensagem, "[RESPOSTA ATUADOR]") {
					fmt.Printf("[Bot_%d] SUCESSO! %s\n", clienteID, mensagem)
					break
				}
			}

			if err := scanner.Err(); err != nil {
				fmt.Printf("[Bot_%d] FALHA: Fiquei no vácuo por muito tempo! Erro: %v\n", clienteID, err)
			}

		}(i)
	}

	wg.Wait()
	fmt.Println("Teste de estresse finalizado. A fila suportou a carga colossal!")
}
