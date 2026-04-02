package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
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

// Fila de clientes que ficam esperando a resposta do atuador.
var filaAtuadores = make(map[string][]net.Conn)

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
	// Fica esperando o ctrl + c para encerrar o sistema
	encerrarSistema()

	// Vai ficar escutando paralelamente os dados enviados dos sensores
	go iniciarServerUDP()

	//Irá rodar o servidor principal, que vai conectar os usuários aos dados e receber comandos
	iniciarServerTCP()
}

// Inicia o servidor TCP para receber as conexões dos clientes e atuadores, e depois manda para a função identificarConexao para diferenciar os dois tipos de conexões
func iniciarServerTCP() {
	// Configura o servidor TCP para escutar a porta 8080
	ln, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatalln("[ERRO] Erro ao iniciar servidor TCP:", err)
	}

	fmt.Println("[TCP] Servidor TCP está ligado e escutando a porta 8080")

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println("[ERRO] Erro ao conectar o cliente TCP:", err)
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

// Função para tratar a conexão do atuador, onde ele é adicionado na lista de atuadores,
// e depois fica esperando as mensagens do atuador, verificando se são respostas para os clientes, e caso sejam,
// ele manda a resposta para o cliente que está esperando
func tratarAtuador(conn net.Conn, id string, scanner *bufio.Scanner) {
	defer conn.Close()

	mu.Lock()
	atuadores[id] = conn
	mu.Unlock()
	fmt.Println("[ATUADOR] Atuador conectado com o ID:", id)

	for scanner.Scan() {
		//Recebe a mensagem do atuador
		mensagem := scanner.Text()
		//Verifico se a mensagem recebida do atuador é uma resposta
		if strings.HasPrefix(mensagem, "RESPOSTA") {
			fmt.Printf("[RESPOSTA ATUADOR] Atuador %s, respondeu: %s\n", id, mensagem)
			mu.Lock()
			fila := filaAtuadores[id]

			if len(fila) > 0 {
				//Respondo o primeiro da fila
				connCliente := fila[0]
				//Atualiza a fila para o próximo elemento da lista
				filaAtuadores[id] = fila[1:]
				mu.Unlock()

				// Manda a resposta do atuador para o cliente que está esperando
				respostaAtuador := strings.Replace(mensagem, "RESPOSTA|", "", 1)
				fmt.Fprintf(connCliente, "[RESPOSTA ATUADOR] O atuador respondeu %s\n", respostaAtuador)

			} else {
				mu.Unlock()
			}
		}
	}

	//Quando for parar de ser utilizado ele é retirado da lista de atuadores
	mu.Lock()
	delete(atuadores, id)
	mu.Unlock()
	fmt.Printf("[ATUADOR] Atuador com o ID: %s, foi desconectado!\n", id)
}

// Inicia o servidor UDP para receber os dados dos sensores, e depois manda os dados para os clientes interessados
func iniciarServerUDP() {
	// Configura o servidor UDP para escutar a porta 5000
	addr, _ := net.ResolveUDPAddr("udp", ":5000")
	conn, _ := net.ListenUDP("udp", addr)
	defer conn.Close()

	fmt.Println("[UDP] Servidor UDP está ligado e escutando porta 5000")

	// Buffer para receber os dados dos sensores
	buf := make([]byte, 1024)

	// Loop para receber os dados dos sensores, e depois enviar para os clientes interessados
	for {
		n, _, _ := conn.ReadFromUDP(buf)

		var dadosSensor Sensor

		// Ele tenta transformar os dados recebidos no formato JSON em um objeto do tipo Sensor, por meio do unmarshal, e se não conseguir, ele ignora o dado recebido e continua esperando
		// os próximos dados dos sensores, para evitar que o servidor trave por causa de um dado mal formatado
		err := json.Unmarshal(buf[:n], &dadosSensor)

		if err != nil {
			continue //Só ignora se não for válido
		}

		mu.Lock()
		sensores[dadosSensor.ID] = time.Now() //Digo qual foi a última vez que o sensor recebeu um dado

		//Percorre a lista de clientes interessados, e verifica se o filtro do cliente é "todos" ou se é igual ao ID do sensor, para então enviar os dados para ele
		for cliente, filtro := range clientesInteressados {
			// Envia se o filtro for "todos" ou igual ao ID do sensor
			if filtro == "todos" || filtro == dadosSensor.ID {
				fmt.Fprintf(cliente, "[TELEMETRIA] DADOS RECEBIDOS DO SENSOR: %s\n"+
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
		//Junta tudo para verificar o comando inteiro, e depois separa para pegar apenas o comando
		comandoInteiro := strings.ToLower(strings.TrimSpace(scanner.Text()))
		partes := strings.Split(comandoInteiro, " ")
		comando := partes[0]

		fmt.Println("[Sistema] Cliente:", nome, "executou o comando:", comandoInteiro)

		switch comando {

		// Comando para receber os dados dos sensores, ele verifica se o cliente quer receber de um sensor específico ou de todos,
		// e depois adiciona ele na lista de interessados
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

		// Comando para parar de receber os dados dos sensores, ele verifica se o cliente está escutando algum sensor, e se estiver, ele para de escutar
		//  e remove ele da lista de interessados
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

		// Comando utilizado para enviar um comando de ligar ou desligar para um atuador específico, verificando se o atuador existe e está ativo,
		// e depois guardando a conexão do cliente em uma fila de espera para receber a resposta do atuador
		case "atuar":
			if len(partes) < 3 {
				fmt.Fprintf(conn, "[ERRO] Comando inserido de maneira errada! Maneira correta: atuar <ID_Atuador> <ação>\n")
				continue
			}

			idAtuador := strings.ToLower(partes[1])
			acao := strings.ToLower(partes[2])

			//Verifico se o atuador existe
			mu.Lock()
			connAtuador, existe := atuadores[idAtuador]
			mu.Unlock()

			if existe == false {
				fmt.Fprintf(conn, "[ERRO] O atuador com ID: %s não existe ou está offline!\n", idAtuador)
				continue
			}

			//Guardo a conexão do cliente com base no ID do atuador
			mu.Lock()
			filaAtuadores[idAtuador] = append(filaAtuadores[idAtuador], conn)
			mu.Unlock()

			//Enviando o comando para o atuador
			fmt.Fprintf(connAtuador, "%s\n", acao)
			fmt.Fprintf(conn, "[SISTEMA] Comando: %s foi enviado para o atuador: %s\n", acao, idAtuador)

		// Comando para listar os sensores e atuadores disponíveis, verificando se eles estão ativos ou não, e mostrando para o cliente
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
			//Verifica se há atuadores ativos no momento
			if len(listaAtuadores) == 0 {
				fmt.Fprintf(conn, "[SISTEMA] Nenhum atuador detectado na rede no momento.\n")
			} else {
				fmt.Fprintf(conn, "[SISTEMA] Atuadores disponíveis: %s\n", strings.Join(listaAtuadores, ", "))
			}

		// Encerra a conexão do cliente com o servidor, removendo ele da lista de clientes e da lista de interessados, caso ele esteja escutando algum sensor
		case "sair":
			fmt.Fprintf(conn, "[SISTEMA] %s se desconectou! Até logo.\n", nome)
			fmt.Println("[USUÁRIO]", nome, "se desconectou")
			return

		// Comando para ajudar o cliente com os comandos disponíveis, explicando cada um deles
		case "help":
			fmt.Fprintf(conn, "[1] Receber dados do sensor: receber [ID], o ID deve ser um dos sensores disponíveis ao digitar listar\n"+
				"[2] Parar de receber dados: 'parar', ele vai ser responsável por parada de receber os dados dos sensores\n"+
				"[3] Listar os sensores e atuadores disponíveis: 'listar', é o comando utilizado para listar todos os sensores e atuadores disponíveis na rede\n"+
				"[4] Enviar comando para o atuador: atuar [ID_Atuador] - nome do atuador a executar a ação [AÇÃO] ação a ser executada: ligar ou desligar\n"+
				"[5] Desconectar do Servidor: 'sair', será o responsável por desconectar o cliente do servidor\n"+
				"[6] Limpar terminal: 'limpar', vai limpar o terminal para o usuário\n"+
				"[7] Ajuda: 'help', mostra um menu de ajuda mostrando como usar os comandos para o usuário\n")

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

// Caso o usuário digite CTRL + C no servidor, é essa função que vai fazer o tratamento de erro
// Implementa a ideia do graceful shutdown, onde o servidor avisa os clientes que ele está desligando e desconecta eles antes de finalizar o processo
func encerrarSistema() {
	sc := make(chan os.Signal, 1)
	//É quem fica escutando caso o usuário dê CTRL + C, aí finaliza o processo
	signal.Notify(sc, os.Interrupt, syscall.SIGTERM)

	go func() {
		// Aqui é onde o servidor avisa os clientes que ele está desligando, e depois desconecta eles, para então finalizar o processo
		<-sc
		fmt.Println("[SISTEMA] Finalizando o sistema...")
		fmt.Println("[SISTEMA] Desconectando os clientes do servidor...")
		mu.Lock()

		//Percorre a lista de clientes e vai desconectando eles
		for conn, _ := range clientes {
			fmt.Fprintf(conn, "\n[SISTEMA] O servidor está sendo desligado. Você será desconectado!\n")
			time.Sleep(200 * time.Millisecond)
			//Finaliza a conexão do cliente com o servidor
			conn.Close()
		}

		//Mesma coisa acontece para os atuadores, percorre o map e desconecta eles
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
