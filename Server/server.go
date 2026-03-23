package main

import (
	"fmt"
	"log"
	"net"
	"strings"
)

type Sensor struct {
	ID          string  `json:"ID"`
	Temperatura float64 `json:"Temperatura"`
	Umidade     float64 `json:"Umidade"`
	Pressao     float64 `json:"Pressao"`
	Ruido       float64 `json:"Ruido"`
	Tempo       string  `json:"Tempo"`
}

// Mapa que guarda as conexões dos usuários indexados pelos seus nomes
var clientes = make(map[net.Conn]string)

func main() {
	iniciarServerTCP()
}

func iniciarServerTCP() {
	ln, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatalln("Erro ao iniciar servidor TCP:", err)

	}

	fmt.Println("Servidor TCP está ligado e escutando a porta 8080")

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println("Erro ao conectar o cliente TCP:", err)
		}
		go clienteHandler(conn)
	}
}

func clienteHandler(conn net.Conn) {
	defer conn.Close()

	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		log.Println("Erro ao receber mensagem do cliente TCP:", err)
	}

	//Cadastro do usuário
	nome := strings.TrimSpace(string(buf[:n]))
	clientes[conn] = nome

	fmt.Println("[Sistema] Cliente:", nome, "conectado")

	for {
		n, err = conn.Read(buf)
		if err != nil {
			delete(clientes, conn)
			fmt.Println("Erro ao receber mensagem do cliente TCP:", err)
			return
		}

		comando := strings.ToLower(strings.TrimSpace(string(buf[:n])))
		fmt.Println("[Sistema] Cliente:", nome, "executou o comando:", comando)

		switch comando {
		case "sair":
			fmt.Fprintf(conn, "%s saiu do servidor.", nome)
			delete(clientes, conn)
			conn.Close()
			return
			//Adicionar os outros comandos
		default:
			fmt.Println("[Sistema] Comando inválido!")
		}

	}

}
