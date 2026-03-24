package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
)

type Atuador struct {
	id     string
	ligado bool
}

func alterarModo(atuador *Atuador) {
	atuador.ligado = !atuador.ligado
}

func main() {
	//Verifica se o atuador tem um nome
	if len(os.Args) < 2 {
		fmt.Println("[ERRO] Digite o nome do atuador após o comando! Ex: go run actuator.go ATUADOR_01")
		return
	}
	id := strings.ToLower(os.Args[1])

	//Iniciando o atuador:
	atuador := Atuador{ligado: false, id: id}

	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		log.Fatalln("[ERRO] Erro ao conectar o atuador ao Servidor TCP:", err)
	}

	defer conn.Close()

	//Envia para o servidor o nome do atuador
	fmt.Fprintf(conn, "ATUADOR|%s\n", id)
	fmt.Printf("==== Atuador: %s foi conectado ao servidor! ====\n", id)

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		comando := scanner.Text()
		fmt.Printf("[COMANDO RECEBIDO] Comando enviado pelo servidor: %s\n", comando)
		switch strings.ToLower(comando) {
		case "ligar":
			if atuador.ligado == false {
				fmt.Fprintf(conn, "RESPOSTA|%s foi ligado!\n", atuador.id)
				alterarModo(&atuador)
				fmt.Println("[ATUADOR] Está ligado!")
			} else {
				fmt.Fprintf(conn, "RESPOSTA|%s já está ligado\n", atuador.id)
			}
		case "desligar":
			if atuador.ligado == false {
				fmt.Fprintf(conn, "RESPOSTA|%s está desligado!\n", atuador.id)
			} else {
				fmt.Fprintf(conn, "RESPOSTA|%s foi desligado com sucesso!\n", atuador.id)
				alterarModo(&atuador)
				fmt.Println("[ATUADOR] Está desligado!")
			}
		default:
			//Avisa de volta ao servidor que o comando não foi reconhecido
			fmt.Fprintf(conn, "RESPOSTA| Comando recebido: %s não foi reconhecido!\n", comando)
			fmt.Println("Comando recebido do servidor não foi reconhecido!")
		}
	}
	fmt.Println("[ATUADOR] Conexão perdida com o servidor!")
}
