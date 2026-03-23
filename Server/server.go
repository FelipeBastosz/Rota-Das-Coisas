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
		go clienteHandler(conn)
	}
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

func clienteHandler(conn net.Conn) {
	defer conn.Close()

	// Recebe os dados enviados do cliente
	scanner := bufio.NewScanner(conn)

	// Faz o cadastro do cliente, salvando ele em um map indexado pelo seu nome
	if !scanner.Scan() {
		return
	}
	nome := strings.TrimSpace(scanner.Text())

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

		fmt.Println("[Sistema] Cliente:", nome, "executou o comando:", comando)

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

		case "listar":
			mu.Lock()
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

		case "sair":
			fmt.Fprintf(conn, "%s se desconectou! Até logo.\n", nome)
			return

		default:
			fmt.Fprintf(conn, "[Sistema] Comando inválido!\n")
		}
	}

	// Se sair do loop, remove o cliente
	mu.Lock()
	delete(clientes, conn)
	delete(clientesInteressados, conn)
	mu.Unlock()
}
