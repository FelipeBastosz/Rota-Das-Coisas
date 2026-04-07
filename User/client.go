package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"os/exec"
	"strings"
)

var ipServidor = "172,16,201.9"

func main() {
	//Estabelece a conexão TCP com o servidor para enviar os comandos e receber as respostas
	endereco := fmt.Sprintf("%s:8080", ipServidor)
	conn, err := net.Dial("tcp", endereco)
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
			mensagem := scanner.Text()
			fmt.Printf("\r\033[K%s\n", mensagem)
			//Se a mensagem recebida for a de desconectar, ela é interceptada e encerramos o cliente
			if strings.Contains(mensagem, "se desconectou!") {
				os.Exit(0)
			}
			fmt.Printf("Insira o comando para enviar ao servidor: ")
		}
		fmt.Println("\n[ERRO] Conexão com o servidor foi perdida!")
		os.Exit(0)
	}()

	//Função para enviar comandos para o servidor
	scanner := bufio.NewScanner(os.Stdin)

	//Imprime o menu de opções
	imprimirMenu()
	fmt.Printf("Insira o comando para enviar ao servidor: ")

	//Loop para enviar comandos ao servidor
	for {
		if !scanner.Scan() {
			break
		}
		comando := scanner.Text()

		//Verifica logo aqui se o usuário quer limpar o terminal, sem precisar enviar para o servidor tratar
		if strings.ToLower(strings.TrimSpace(comando)) == "limpar" {
			//Tenta limpar o terminal no Linux
			cmd := exec.Command("clear")
			cmd.Stdout = os.Stdout
			err := cmd.Run()

			//Se não for linux, ele vai tentar executar o comando do Windows
			if err != nil {
				cmd = exec.Command("cmd", "/c", "cls")
				cmd.Stdout = os.Stdout
				cmd.Run()
			}
			imprimirMenu()
			fmt.Printf("Insira o comando para enviar ao servidor: ")
			continue
		}

		//Verifico se a mensagem não está vazia e envio o comando ao servidor
		if strings.TrimSpace(comando) != "" {
			fmt.Fprintf(conn, comando+"\n")
		}

	}
}

// Função para imprimir o menu de opções para o usuário
func imprimirMenu() {
	fmt.Println("======= LISTA DE COMANDOS =======")
	fmt.Println("[1] Receber dados do sensor: 'receber [id]'")
	fmt.Println("[2] Parar de receber dados do sensor: 'parar'")
	fmt.Println("[3] Listar os sensores e atuadores disponíveis: 'listar'")
	fmt.Println("[4] Enviar comando para o atuador: atuar [ID_Atuador] [AÇÃO]")
	fmt.Println("[5] Desconectar do servidor: 'sair'")
	fmt.Println("[6] Limpar terminal: 'limpar'")
	fmt.Println("[7] Ajuda para comandos: 'help'")
	fmt.Println("=================================")
}
