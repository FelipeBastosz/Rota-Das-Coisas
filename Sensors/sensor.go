package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net"
	"os"
	"strings"
	"time"
)

var ipServidor = "172.16.201.9"

// Estrutura para representar os dados do sensor, com os campos ID, Temperatura, Umidade, Pressao, Ruido e Tempo
type Sensor struct {
	ID          string  `json:"ID"`
	Temperatura float64 `json:"Temperatura"`
	Umidade     float64 `json:"Umidade"`
	Pressao     float64 `json:"Pressao"`
	Ruido       float64 `json:"Ruido"`
	Tempo       string  `json:"Tempo"`
}

func main() {
	//Verifica se o ID do sensor foi passado, caso contrário, exibe uma mensagem de erro e encerra o programa
	if len(os.Args) < 2 {
		fmt.Println("[ERRO] Digite o nome do sensor após o comando! Ex: go run sensor.go SENSOR_01")
		return
	}
	id := strings.ToLower(os.Args[1])

	// Configura o endereço do servidor UDP para enviar os dados dos sensores, e depois inicia a conexão UDP com o servidor
	endereco := fmt.Sprintf("%s:5000", ipServidor)
	addr, err := net.ResolveUDPAddr("udp", endereco)

	if err != nil {
		fmt.Println("[COMUNICAÇÃO UDP] Erro ao se conectar ao servidor")
	}
	conn, _ := net.DialUDP("udp", nil, addr)
	defer conn.Close()

	fmt.Println("===== Sensor", id, "ativo! =====")

	//Gera dados dos sensores de forma aleatória e envia para o servidor a cada 1 segundo,
	//utilizando a função json.Marshal para converter os dados do sensor em formato JSON antes de enviar
	for {
		dadosSensor := Sensor{
			ID:          id,
			Temperatura: 30 + rand.Float64()*60,
			Umidade:     30 + rand.Float64()*50,
			Pressao:     900 + rand.Float64()*200,
			Ruido:       35 + rand.Float64()*45,
			Tempo:       time.Now().Format("15:04:05"),
		}

		// Converte os dados do sensor para JSON e envia para o servidor
		jsonBytes, _ := json.Marshal(dadosSensor)
		_, err := conn.Write(jsonBytes)

		if err != nil {
			fmt.Println("[AVISO] Falha ao enviar os dados ao servidor!")
		} else {
			fmt.Printf("Sensor: %s, enviou os dados: %s\n", dadosSensor.ID, string(jsonBytes))
		}

		time.Sleep(1000 * time.Millisecond) //Manda os dados a cada 1 segundo
	}
}
