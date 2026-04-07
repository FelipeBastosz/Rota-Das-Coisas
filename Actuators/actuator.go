package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"time"
)

// Estrutura para representar o atuador, com os campos id e ligado
type Atuador struct {
	id     string
	ligado bool
}

// Função para alterar o modo do atuador, alternando entre ligado e desligado
func alterarModo(atuador *Atuador) {
	atuador.ligado = !atuador.ligado
}

var ipServidor = "172.16.201.9"

func main() {
	//Verifica se o atuador tem um nome
	if len(os.Args) < 2 {
		fmt.Println("[ERRO] Digite o nome do atuador após o comando! Ex: go run actuator.go ATUADOR_01")
		return
	}
	id := strings.ToLower(os.Args[1])

	//Iniciando o atuador:
	atuador := Atuador{ligado: false, id: id}

	// Estabelece a conexão TCP com o servidor para receber os comandos
	endereco := fmt.Sprintf("%s:8080", ipServidor)
	conn, err := net.Dial("tcp", endereco)
	if err != nil {
		log.Fatalln("[ERRO] Erro ao conectar o atuador ao Servidor TCP:", err)
	}

	defer conn.Close()

	//Envia para o servidor o nome do atuador
	fmt.Fprintf(conn, "ATUADOR|%s\n", id)
	fmt.Printf("==== Atuador: %s foi conectado ao servidor! ====\n", id)

	//Fica esperando os comandos do servidor, e quando recebe um comando, ele verifica qual é o comando e executa a ação correspondente
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		comando := scanner.Text()
		fmt.Printf("[COMANDO RECEBIDO] Comando enviado pelo servidor: %s\n", comando)
		switch strings.ToLower(comando) {

		//Verifica se o comando é para ligar o atuador
		case "ligar":
			//Se o atuador estiver desligado, ele liga o atuador, e depois manda uma resposta para o servidor informando que o atuador foi ligado,
			// caso contrário, ele manda uma resposta para o servidor avisando que o atuador já está ligado
			if atuador.ligado == false {
				fmt.Println("[ATUADOR] Ligando o atuador...")

				//Simula o tempo de resposta do atuador, onde ele demora 600ms para ligar ou desligar, e depois manda a resposta para o servidor
				time.Sleep(600 * time.Millisecond)
				alterarModo(&atuador)

				fmt.Fprintf(conn, "RESPOSTA|%s foi ligado!\n", atuador.id)
				fmt.Println("[ATUADOR] Está ligado!")
			} else {
				fmt.Fprintf(conn, "RESPOSTA|%s já está ligado\n", atuador.id)
				fmt.Println("[ATUADOR] Aviso: já está ligado!")
			}

		//Verifica se o comando é para desligar o atuador
		case "desligar":
			//Mesma lógica do comando de ligar, só que para desligar o atuador, e mandar a resposta correspondente para o servidor
			if atuador.ligado == false {
				fmt.Fprintf(conn, "RESPOSTA|%s está desligado!\n", atuador.id)
				fmt.Println("[ATUADOR] Já está desligado!")
			} else {
				fmt.Println("[ATUADOR] Desligando atuador...")

				time.Sleep(600 * time.Millisecond)
				alterarModo(&atuador)

				fmt.Println("[ATUADOR] O atuador foi desligado!")
				fmt.Fprintf(conn, "RESPOSTA|%s foi desligado com sucesso!\n", atuador.id)
			}
		default:
			//Avisa de volta ao servidor que o comando não foi reconhecido
			fmt.Fprintf(conn, "RESPOSTA| Comando recebido: %s não foi reconhecido!\n", comando)
			fmt.Println("Comando recebido do servidor não foi reconhecido!")
		}
	}
	fmt.Println("[ATUADOR] Conexão perdida com o servidor!")
}
