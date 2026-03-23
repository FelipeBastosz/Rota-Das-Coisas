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

type Sensor struct {
	ID          string  `json:"ID"`
	Temperatura float64 `json:"Temperatura"`
	Umidade     float64 `json:"Umidade"`
	Pressao     float64 `json:"Pressao"`
	Ruido       float64 `json:"Ruido"`
	Tempo       string  `json:"Tempo"`
}

func main() {
	//Recebe o ID do servidor pelo terminal
	if len(os.Args) < 2 {
		fmt.Println("[ERRO] Digite o nome do sensor após o comando! Ex: go run sensor.go SENSOR_01")
		return
	}
	id := strings.ToLower(os.Args[1])

	addr, err := net.ResolveUDPAddr("udp", ":5000")
	if err != nil {
		fmt.Println("[COMUNICAÇÃO UDP] Erro ao se conectar ao servidor")
	}

	conn, _ := net.DialUDP("udp", nil, addr)
	defer conn.Close()

	fmt.Println("===== Sensor", id, "ativo! =====")

	for {
		dadosSensor := Sensor{
			ID:          id,
			Temperatura: 30 + rand.Float64()*60,
			Umidade:     30 + rand.Float64()*50,
			Pressao:     900 + rand.Float64()*200,
			Ruido:       35 + rand.Float64()*45,
			Tempo:       time.Now().Format("15:04:05"),
		}

		jsonBytes, _ := json.Marshal(dadosSensor)
		conn.Write(jsonBytes)

		fmt.Printf("Sensor: %s, enviou os dados: %s", dadosSensor.ID, string(jsonBytes))
		time.Sleep(1000 * time.Millisecond) //Manda os dados a cada 300ms
	}
}
