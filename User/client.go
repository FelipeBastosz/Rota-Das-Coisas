package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

func main() {
	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		fmt.Println("[Cliente] Erro ao se conectar ao servidor!", err)
	}

	fmt.Printf("[Sistema] Digite o seu nome de usuário:")
	reader := bufio.NewReader(os.Stdin)
	nome, _ := reader.ReadString('\n')
	nome = strings.TrimSpace(nome)

	//Envia o nome cadastrado para o servidor
	fmt.Fprintf(conn, nome+"\n")

	fmt.Printf("\n[Sistema] Seja bem vindo, %s!\n", nome)

	//Função para ficar escutando os dados do servidor
	go func() {
		scanner := bufio.NewScanner(conn)
		for scanner.Scan() {
			fmt.Printf("\r\033[K%s\nInsira o comando para enviar ao servidor: ", scanner.Text())
		}
		fmt.Println("\n[ERRO] Servidor foi encerrado.")
		os.Exit(0)
	}()

	//Função para enviar comandos para o servidor
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Println("======= LISTA DE COMANDOS =======")
	fmt.Println("[1] Receber dados do sensor: 'receber [id]'")
	fmt.Println("[2] Parar de receber dados do sensor: 'parar'")
	fmt.Println("[3] Listar os sensores e atuadores disponíveis: 'listar'")
	fmt.Println("[4] Enviar comando para o atuador: atuar [ID_Atuador] [AÇÃO]")
	fmt.Println("[5] Desconectar do servidor: 'sair'")
	fmt.Println("[6] Limpar terminal: 'limpar'")
	fmt.Println("[7] Ajuda para comandos: 'help'")

	for {
		if !scanner.Scan() {
			break
		}
		comando := scanner.Text()

		//Verifico se a mensagem não está vazia
		if strings.TrimSpace(comando) != "" {
			fmt.Fprintf(conn, comando+"\n")
		}

	}
}
