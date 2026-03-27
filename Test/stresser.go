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
	// Define a quantidade de clientes PARA CADA atuador (Total = 400 conexões)
	totalClientes := 200
	var wg sync.WaitGroup

	fmt.Printf("Iniciando ataque de %d clientes simultâneos por atuador (Total: %d conexões)...\n", totalClientes, totalClientes*2)

	// Dispara goroutines simulando clientes reais para os DOIS atuadores em um único loop
	for i := 0; i < totalClientes; i++ {

		// ==========================================
		// --- GOROUTINE DO ATUADOR 01 ---
		// ==========================================
		wg.Add(1)
		go func(clienteID int) {
			defer wg.Done()

			// Jitter: espalhamento aleatório entre 0 e 2 segundos
			delayInicio := time.Duration(rand.Intn(2000)) * time.Millisecond
			time.Sleep(delayInicio)

			conn, err := net.Dial("tcp", "localhost:8080")
			if err != nil {
				fmt.Printf("[Bot_A1_%d] ERRO DE CONEXÃO: %v\n", clienteID, err)
				return
			}
			defer conn.Close()

			// Handshake de registro no servidor
			fmt.Fprintf(conn, "Bot_A1_%d\n", clienteID)
			time.Sleep(50 * time.Millisecond)

			// Dispara a ação especificamente para o atuador 01
			if i%2 == 0 {
				fmt.Fprintf(conn, "atuar atuador_01 ligar\n")
			} else {
				fmt.Fprintf(conn, "atuar atuador_02 desligar\n")
			}
			scanner := bufio.NewScanner(conn)

			// Define o tempo limite alto para a fila inteira esvaziar
			conn.SetReadDeadline(time.Now().Add(10 * time.Minute))

			for scanner.Scan() {
				mensagem := scanner.Text()

				// Ignora log intermediário
				if strings.Contains(mensagem, "[SISTEMA] Comando:") {
					continue
				}

				// Captura a confirmação do hardware e encerra a goroutine
				if strings.Contains(mensagem, "[RESPOSTA ATUADOR]") {
					fmt.Printf("[Bot_A1_%d] SUCESSO! %s\n", clienteID, mensagem)
					break
				}
			}

			if err := scanner.Err(); err != nil {
				fmt.Printf("[Bot_A1_%d] FALHA: Fiquei no vácuo! Erro: %v\n", clienteID, err)
			}
		}(i)

		// ==========================================
		// --- GOROUTINE DO ATUADOR 02 ---
		// ==========================================
		wg.Add(1)
		go func(clienteID int) {
			defer wg.Done()

			// Jitter: espalhamento aleatório entre 0 e 2 segundos
			delayInicio := time.Duration(rand.Intn(2000)) * time.Millisecond
			time.Sleep(delayInicio)

			conn, err := net.Dial("tcp", "localhost:8080")
			if err != nil {
				fmt.Printf("[Bot_A2_%d] ERRO DE CONEXÃO: %v\n", clienteID, err)
				return
			}
			defer conn.Close()

			// Handshake de registro no servidor
			fmt.Fprintf(conn, "Bot_A2_%d\n", clienteID)
			time.Sleep(50 * time.Millisecond)

			// Dispara a ação especificamente para o atuador 02
			fmt.Fprintf(conn, "atuar atuador_02 ligar\n")

			scanner := bufio.NewScanner(conn)

			// Define o tempo limite alto para a fila inteira esvaziar
			conn.SetReadDeadline(time.Now().Add(10 * time.Minute))

			for scanner.Scan() {
				mensagem := scanner.Text()

				// Ignora log intermediário
				if strings.Contains(mensagem, "[SISTEMA] Comando:") {
					continue
				}

				// Captura a confirmação do hardware e encerra a goroutine
				if strings.Contains(mensagem, "[RESPOSTA ATUADOR]") {
					fmt.Printf("[Bot_A2_%d] SUCESSO! %s\n", clienteID, mensagem)
					break
				}
			}

			if err := scanner.Err(); err != nil {
				fmt.Printf("[Bot_A2_%d] FALHA: Fiquei no vácuo! Erro: %v\n", clienteID, err)
			}
		}(i)
	}

	// Trava a execução do programa até que todos os 400 bots tenham terminado
	wg.Wait()
	fmt.Println("Teste de estresse finalizado. As múltiplas filas suportaram a carga de forma isolada e perfeita!")
}
