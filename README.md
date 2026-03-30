# 🌐 Rota das Coisas: IoT Message Broker

![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=for-the-badge&logo=go)
![Docker](https://img.shields.io/badge/Docker-Pronto-2496ED?style=for-the-badge&logo=docker)
![Status](https://img.shields.io/badge/Status-Conclu%C3%ADdo-success?style=for-the-badge)
![Concorrência](https://img.shields.io/badge/Carga-3000%2B%20Conex%C3%B5es-orange?style=for-the-badge)

Um Message Broker IoT de alta performance e tolerante a falhas desenvolvido em Go. Este sistema distribuído foi projetado para gerenciar a telemetria de sensores e o controle de atuadores físicos, suportando alta carga de requisições simultâneas sem perda de dados ou vazamento de memória.

Projeto desenvolvido para a disciplina de Redes e Sistemas Distribuídos (PBL).

---

## 🚀 Principais Funcionalidades (Features)

* **Multiplexação de Protocolos:** Arquitetura híbrida que utiliza **UDP** para telemetria de sensores (focando em velocidade e throughput) e **TCP** para comandos críticos de atuadores (garantindo a entrega da mensagem).
* **Concorrência Segura (Thread-Safety):** Implementação rigorosa de `sync.Mutex` para proteger mapas de memória e prevenir *Race Conditions* durante acessos simultâneos.
* **Fila de Requisições FIFO:** Sistema de enfileiramento assíncrono que impede a sobrescrita de comandos. Se múltiplos clientes acionam um dispositivo no mesmo milissegundo, os comandos são processados em ordem de chegada.
* **Tolerância a Falhas & Timeouts:** Clientes não ficam travados caso um atuador perca a conexão física.
* **Graceful Shutdown:** Encerramento seguro do servidor interceptando sinais do SO (`SIGTERM`), notificando clientes e fechando conexões antes da interrupção do processo.

---

## 🏗️ Arquitetura do Sistema

O ecossistema é dividido em 4 componentes principais:

1. **Servidor (Broker):** O coração da rede. Escuta na porta `8080` (TCP) para clientes/atuadores e `5000` (UDP) para sensores.
2. **Atuadores (Hardware):** Dispositivos simulados que recebem comandos do servidor para alterar o ambiente (ex: ligar um motor).
3. **Sensores (Telemetria):** Dispositivos que fazem *broadcast* de dados (Temperatura, Umidade, Pressão, etc.) em tempo real via UDP.
4. **Clientes (Monitores):** Usuários que se conectam ao servidor para ler os dados dos sensores ou enviar comandos aos atuadores.

---

## 🐳 Executando com Docker (Recomendado)

O projeto é *Cloud-Native* e totalmente containerizado. A orquestração dos serviços é feita via Docker Compose, permitindo subir toda a infraestrutura com um único comando.

### 1. Subir a Infraestrutura Completa
Na raiz do projeto, onde está localizado o arquivo `docker-compose.yml`, execute:
```bash
docker-compose up --build
```
Isso iniciará automaticamente o Servidor Broker, instâncias de Sensores transmitindo dados e instâncias de Atuadores prontos para receber comandos, todos comunicando-se em uma rede Docker isolada.
### 2. Encerrar o Sistema
Para desligar todos os containers e limpar a rede:
```bash
docker-compose down
```


##  🛠️ Como Executar (Localmente sem Docker)
Caso prefira rodar os arquivos Go nativamente no seu terminal:

1. Inicie o Servidor:
   ```bash
    go run server.go
   ```

2. Conecte um Atuador:
   ```bash
    go run actuator.go atuador_01
   ```
3. Conecte um Cliente:
   ```bash
   go run cliente.go
   ```
4. Conecte os sensores:
   ```bash
   go run sensor.go sensor_01
   ```
## 💻 Interface de Comandos (Menu do Cliente)
Uma vez conectado como cliente via TCP, você pode usar os seguintes comandos interativos:
| Comando             | Descrição                                     | Exemplo                  |
| ------------------- | --------------------------------------------- | ------------------------ |
| `listar`            | Lista sensores e atuadores ativos             | `listar`                 |
| `receber [id]`      | Escuta dados de um sensor específico ou todos | `receber sensor_01`    |
| `parar`             | Para de escutar os dados recebidos do sensor  | `parar`                  |
| `atuar [id] [ação]` | Envia comando para um atuador (ligar/desligar)| `atuar atuador_01 ligar` |
| `help`              | Mostra menu de ajuda                          | `help`                   |
| `sair`              | Desconecta do servidor                        | `sair`                   |


## 🧪 Teste de Estresse
Para comprovar a resiliência do Scheduler do Go e das nossas filas Mutex, o repositório inclui um script de ataque controlado (stresser.go).

Ele gera milhares de conexões TCP simultâneas simulando clientes reais, aplicando jitter (espalhamento) para evitar o bloqueio de sockets do SO host.

Como executar: 
   ```bash
    go run stresser.go
   ```
Resultado: O servidor suporta e processa em fila cargas massivas de até 3.000 requisições simultâneas distribuídas entre múltiplos atuadores, sem crashes de memória ou perda de concorrência.

## 📚 Referências e Links Úteis

Para a construção da arquitetura e tomada de decisões deste projeto, os seguintes materiais foram consultados:

* [Documentação Oficial do Go (Golang)](https://go.dev/doc/) - Base para a sintaxe, *Goroutines* e gerenciamento de memória.
* [Go by Example: Mutexes](https://gobyexample.com/mutexes) - Referência principal para a implementação de *Thread-Safety* e prevenção de *Race Conditions*.
* [Pacote `net` do Go](https://pkg.go.dev/net) - Documentação oficial utilizada para a criação das conexões via *Sockets* (TCP e UDP) e definição dos *Timeouts* (`SetReadDeadline`).
* [Diferenças entre TCP e UDP (Cloudflare)](https://www.cloudflare.com/pt-br/learning/ddos/glossary/tcp-ip/) - Embasamento teórico para a escolha de protocolos baseada na criticidade dos dados (Telemetria vs. Controle).

## 👨‍💻 Autor
Felipe Bastos - Desenvolvedor Backend & Estudante de Engenharia de Computação - UEFS

---

## ⚖️ Licença

Este projeto está sob a licença MIT. Consulte o arquivo [LICENSE](LICENSE) para mais detalhes.
