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

func iniciarServerUDP() {
	addr, _ := net.ResolveUDPAddr("udp", ":5000")
	conn, _ := net.ListenUDP("udp", addr)
	defer conn.Close()

	fmt.Println("[UDP] Escutando porta 5000 (Telemetria)")

	buf := make([]byte, 1024)
	for {
		n, _, _ := conn.ReadFromUDP(buf)

		var dadosSensor Sensor

		err := json.Unmarshal(buf[:n], &dadosSensor)

		if err != nil {
			continue //Só ignora se não for válido
		}

		mu.Lock()
		for monitor, filtro := range clientesInteressados {
			// Envia se o filtro for "todos" ou igual ao ID do sensor
			if filtro == "todos" || filtro == dadosSensor.ID {
				fmt.Fprintf(monitor, "[TELEMETRIA] DADOS RECEBIDOS DO SENSOR: %s\n"+
					"Temperatura: %.2f°C | Pressão: %.2f hPa | Umidade: %.2f%% | Ruído: %.2f dB\n"+
					"Insira um comando:", dadosSensor.ID, dadosSensor.Temperatura, dadosSensor.Pressao,
					dadosSensor.Umidade, dadosSensor.Ruido)
			}
		}
		mu.Unlock()
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
		case "receber":
			id := "todos"
			// Verifica se o usuário passou um ID específico
			if len(partes) > 1 {
				id = partes[1] // Atribuição correta (sem o :)
			}

			mu.Lock()
			clientesInteressados[conn] = id
			mu.Unlock()
			fmt.Fprintf(conn, "[SISTEMA] Agora você recebe dados do sensor: %s\n", id)

		case "parar":
			mu.Lock()
			delete(clientesInteressados, conn)
			mu.Unlock()
			fmt.Fprintf(conn, "[TELEMETRIA] Você parou de receber dados dp sensor.\n")

		case "listar":
			mu.Lock()
			var listaSensores []string
			for sensor := range sensores {
				listaSensores = append(listaSensores, sensor)
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
