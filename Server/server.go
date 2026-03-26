package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"strings"
	"sync"
	"time"
)

type Sensor struct {
	ID          string  `json:"ID"`
	Temperatura float64 `json:"Temperatura"`
	Umidade     float64 `json:"Umidade"`
	Pressao     float64 `json:"Pressao"`
	Ruido       float64 `json:"Ruido"`
	Tempo       string  `json:"Tempo"`
}

// Mapa de sensores ativos funcionando
var sensores = make(map[string]time.Time)

// Mapa que guarda as conexões dos usuários indexados pelos seus nomes
var clientes = make(map[net.Conn]string)

// Clientes interessados em escutar os dados dos sensores
var clientesInteressados = make(map[net.Conn]string)

// Mapa que guarda os atuadores
var atuadores = make(map[string]net.Conn)

var mu sync.Mutex //Vai ser utilizado para proteger o mapa de clientes, impedindo adicionar e retirar clientes

func main() {
	// Vai ficar escutando paralelamente os dados enviados dos sensores
	go iniciarServerUDP()

	//Irá rodar o servidor principal, que vai conectar os usuários aos dados e receber comandos
	iniciarServerTCP()
}

func iniciarServerTCP() {
	ln, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatalln("Erro ao iniciar servidor TCP:", err)

	}

	fmt.Println("[TCP] Servidor TCP está ligado e escutando a porta 8080")

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println("Erro ao conectar o cliente TCP:", err)
		}
		//Verifica se quem entrou é um atuador ou cliente
		go identificarConexao(conn)
	}
}

// Diferencia o atuador de um cliente
func identificarConexao(conn net.Conn) {
	scanner := bufio.NewScanner(conn)
	if !scanner.Scan() {
		conn.Close()
		return
	}
	//Junta tudo para verificar se é atuador ou cliente
	primeiraLinha := strings.TrimSpace(scanner.Text())

	//Separa e verifica se o primeiro nome é atuador
	primeiroNome := strings.Split(primeiraLinha, "|")
	if primeiroNome[0] == "ATUADOR" {
		if len(primeiroNome) == 2 {
			tratarAtuador(conn, primeiroNome[1], scanner)
		} else {
			conn.Close()
		}
		return
	}
	//Se não for atuador, verifico que é cliente e mando para o clienteHandler fazer o tratamento dele
	clienteHandler(conn, primeiroNome[0], scanner)
}

func tratarAtuador(conn net.Conn, id string, scanner *bufio.Scanner) {
	defer conn.Close()

	mu.Lock()
	atuadores[id] = conn
	mu.Unlock()
	fmt.Println("[ATUADOR] Atuador conectado com o ID:", id)
	for scanner.Scan() {
		mensagem := scanner.Text()
		if strings.HasPrefix(mensagem, "RESPOSTA") {
			fmt.Printf("[RESPOSTA ATUADOR] Atuador %s, respondeu: %s\n", id, mensagem)
		}
	}

	//Quando for parar de ser utilizado ele é retirado da lista de atuadores
	mu.Lock()
	delete(atuadores, id)
	mu.Unlock()
	fmt.Printf("[ATUADOR] Atuador com o ID: %s, foi desconectado!\n", id)
}

func iniciarServerUDP() {
	addr, _ := net.ResolveUDPAddr("udp", ":5000")
	conn, _ := net.ListenUDP("udp", addr)
	defer conn.Close()

	fmt.Println("[UDP] Servido UDP está ligado e escutando porta 5000")

	buf := make([]byte, 1024)
	for {
		n, _, _ := conn.ReadFromUDP(buf)

		var dadosSensor Sensor

		err := json.Unmarshal(buf[:n], &dadosSensor)

		if err != nil {
			continue //Só ignora se não for válido
		}

		mu.Lock()
		sensores[dadosSensor.ID] = time.Now() //Digo qual foi a última vez que o sensor recebeu um dado

		for monitor, filtro := range clientesInteressados {
			// Envia se o filtro for "todos" ou igual ao ID do sensor
			if filtro == "todos" || filtro == dadosSensor.ID {
				fmt.Fprintf(monitor, "[TELEMETRIA] DADOS RECEBIDOS DO SENSOR: %s\n"+
					"Temperatura: %.2f°C | Pressão: %.2f hPa | Umidade: %.2f%% | Ruído: %.2f dB\n", dadosSensor.ID, dadosSensor.Temperatura, dadosSensor.Pressao,
					dadosSensor.Umidade, dadosSensor.Ruido)
			}
		}
		mu.Unlock()
	}
}

// Recebe a conexão, o nome do usuário e o scanner que já estava sendo utilizado
func clienteHandler(conn net.Conn, nome string, scanner *bufio.Scanner) {
	defer conn.Close()

	//Bloqueia para que possa inserir um cliente no map de clientes, sem ter o risco de concorrência
	mu.Lock()
	clientes[conn] = nome
	mu.Unlock()
	fmt.Println("[Sistema] Cliente:", nome, "conectado")

	// Loop de comandos
	for scanner.Scan() {
		comandoInteiro := strings.ToLower(strings.TrimSpace(scanner.Text()))
		partes := strings.Split(comandoInteiro, " ")
		comando := partes[0]

		fmt.Println("[Sistema] Cliente:", nome, "executou o comando:", comandoInteiro)

		switch comando {
		case "receber":
			id := "todos"
			// Verifica se o usuário passou um ID específico
			if len(partes) > 1 {
				id = partes[1]
			}

			mu.Lock()
			// Verifica se existe ou está funcionando o sensor passado pelo usuário
			if id != "todos" {
				ultimaVez, existe := sensores[id]
				//Verifica se o sensor está ativo há mais de 20 segundos e se ele existe, para então escutar ele
				isAtivo := existe && time.Since(ultimaVez) < (20*time.Second)

				if !isAtivo {
					mu.Unlock()
					// CORREÇÃO: Passando o 'id' para preencher o %s
					fmt.Fprintf(conn, "[ERRO] O sensor de ID: '%s' não existe ou está offline!\n", id)
					continue
				}
			}

			// Se for "todos" ou se o sensor específico estiver ativo, ele salva aqui
			clientesInteressados[conn] = id
			mu.Unlock()
			fmt.Fprintf(conn, "[SISTEMA] Agora você recebe dados do sensor: %s\n", id)

		case "parar":
			mu.Lock()
			//Verifica se o cliente está escutando algum sensor, antes de desconectar ele efetivamente
			idSensor, encontrado := clientesInteressados[conn]

			//Se não encontrou, simplesmente retorna um aviso informando que ele não está escutando nenhum sensor
			if encontrado == false {
				mu.Unlock()
				fmt.Fprintf(conn, "[AVISO] Você não está escutando nenhum sensor!\n")
				continue
			}

			//Se estiver escutando, ele deleta aqui o cliente da lista de interessados
			delete(clientesInteressados, conn)
			mu.Unlock()
			fmt.Fprintf(conn, "[TELEMETRIA] Você parou de receber dados do sensor %s!\n", idSensor)

		case "atuar":
			if len(partes) < 3 {
				fmt.Fprintf(conn, "[ERRO] Comando inserido de maneira errada! Maneira correta: atuar <ID_Atuador> <ação>\n")
			}

			idAtuador := strings.ToLower(partes[1])
			acao := strings.ToLower(partes[2])

			mu.Lock()
			connAtuador, existe := atuadores[idAtuador]
			mu.Unlock()

			if existe == false {
				fmt.Fprintf(conn, "[ERRO] O atuador com ID: %s não existe ou está offline!\n", idAtuador)
				continue
			}

			//Enviando o comando para o atuador
			fmt.Fprintf(connAtuador, "%s\n", acao)
			fmt.Fprintf(conn, "[SISTEMA] Comando: %s foi enviado para o atuador: %s\n", acao, idAtuador)

		case "listar":
			mu.Lock()
			var listaAtuadores []string
			for idAtuador := range atuadores {
				listaAtuadores = append(listaAtuadores, idAtuador)
			}

			var listaSensores []string
			for sensor, ultimaVez := range sensores {
				//Deixo apenas os sensores que responderam nos últimos 20 segundos
				if time.Since(ultimaVez) < 20*time.Second {
					listaSensores = append(listaSensores, sensor)
				}
			}
			mu.Unlock()

			//Verifica se há sensores ativos no momento
			if len(listaSensores) == 0 {
				fmt.Fprintf(conn, "[SISTEMA] Nenhum sensor detectado na rede no momento.\n")
			} else {
				//Lista os sensores disponíveis
				fmt.Fprintf(conn, "[SISTEMA] Sensores disponíveis: %s\n", strings.Join(listaSensores, ", "))
			}

			if len(listaAtuadores) == 0 {
				fmt.Fprintf(conn, "[SISTEMA] Nenhum atuador detectado na rede no momento.\n")
			} else {
				fmt.Fprintf(conn, "[SISTEMA] Sensores disponíveis %s\n", strings.Join(listaAtuadores, ", "))
			}

		case "sair":
			fmt.Fprintf(conn, "%s se desconectou! Até logo.\n", nome)
			return
		case "help":
			fmt.Fprintf(conn, "[1] Receber dados do sensor: receber [ID], o ID deve ser um dos sensores disponíveis ao digitar listar\n"+
				"[2] Parar de receber dados: 'parar', ele vai ser responsável por parada de receber os dados dos sensores\n"+
				"[3] Listar os sensores e atuadores disponíveis: 'listar', é o comando utilizado para listar todos os sensores e atuadores disponíveis na rede\n"+
				"[4] Enviar comando para o atuador: atuar [ID_Atuador] - nome do atuador a executar a ação [AÇÃO] ação a ser executada: ligar ou desligar\n"+
				"[5] Desconectar do Servidor: 'sair', será o responsável por desconectar o cliente do servidor\n")

		//case "limpar":
		//	limparTerminal(runtime.GOOS)
		default:
			fmt.Fprintf(conn, "[Sistema] Comando inválido!\n")
		}
	}

	// Se sair do loop, remove o cliente:
	mu.Lock()
	delete(clientes, conn)
	delete(clientesInteressados, conn)
	mu.Unlock()
}

func tratamentoDeDesligamento() {
	c := make(chan os.Signal, 1)
	//É quem fica escutando caso o usuário dê CTRL + C, aí finaliza o processo
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		fmt.Println("[SISTEMA] Finalizando o sistema...")
		fmt.Println("[SISTEMA] Desconectando os clientes do servidor...")
		mu.Lock()
		for conn, _ := range clientes {
			fmt.Fprintf(conn, "\n[SISTEMA] O servidor está sendo desligado. Você será desconectado!\n")
			time.Sleep(200 * time.Millisecond)
			//Finaliza a conexão do cliente com o servidor
			conn.Close()
		}

		for _, conn := range atuadores {
			fmt.Fprintf(conn, "\n[SISTEMA] O servidor está sendo desligado. Você será desconectado!\n")
			//Finaliza a conexão do atuador com o servidor
			time.Sleep(200 * time.Millisecond)
			conn.Close()
		}
		mu.Unlock()
		fmt.Println("[SISTEMA] Servidor foi encerrado com sucesso!")
		os.Exit(0)
	}()
}
