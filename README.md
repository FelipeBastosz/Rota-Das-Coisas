# 🌐 Rota das Coisas: IoT Message Broker

![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=for-the-badge&logo=go)
![Docker](https://img.shields.io/badge/Docker-Pronto-2496ED?style=for-the-badge&logo=docker)
![Status](https://img.shields.io/badge/Status-Conclu%C3%ADdo-success?style=for-the-badge)
![Concorrência](https://img.shields.io/badge/Carga-3000%2B%20Conex%C3%B5es-orange?style=for-the-badge)
![Protocolos](https://img.shields.io/badge/Protocolos-UDP_TCP-purple?style=for-the-badge)


## 📌 Sobre o Projeto
Este projeto implementa um **Message Broker IoT concorrente em Golang**, capaz de gerenciar sensores e atuadores em uma arquitetura distribuída, suportando centenas de conexões simultâneas através de **goroutines e controle de concorrência com Mutex**.

Projeto desenvolvido para a disciplina de Redes e Sistemas Distribuídos (PBL).

---
## 🎯 Objetivo do Projeto

Este projeto tem como objetivo simular um **ambiente IoT distribuído**, onde sensores enviam dados de telemetria e clientes podem monitorar ou controlar dispositivos atuadores através de um **Message Broker central**.

O sistema foi projetado para explorar conceitos de:

- Concorrência em Golang
- Comunicação via sockets TCP e UDP
- Arquiteturas distribuídas
- Containerização com Docker
- Testes de carga e tolerância a falhas
  
---

## 🚀 Principais Funcionalidades (Features)

* **Multiplexação de Protocolos:** Arquitetura híbrida que utiliza **UDP** para telemetria de sensores (focando em velocidade e throughput) e **TCP** para comandos críticos de atuadores (garantindo a entrega da mensagem).
* **Concorrência Segura (Thread-Safety):** Implementação rigorosa de `sync.Mutex` para proteger mapas de memória e prevenir *Race Conditions* durante acessos simultâneos.
* **Fila de Requisições FIFO:** Sistema de enfileiramento assíncrono que impede a sobrescrita de comandos. Se múltiplos clientes acionam um dispositivo no mesmo milissegundo, os comandos são processados em ordem de chegada.
* **Tolerância a Falhas & Timeouts:** Clientes não ficam travados caso um atuador perca a conexão física.
* **Graceful Shutdown:** Encerramento seguro do servidor interceptando sinais do SO (`SIGTERM`), notificando clientes e fechando conexões antes da interrupção do processo.

---

## 📡 Especificação do Protocolo

Para garantir o isolamento e roteamento correto, o sistema implementa um protocolo rigoroso de *Handshake* e encapsulamento:

* **Sensores (UDP):** Transmitem dados periódicos encapsulados em JSON contendo: `ID`, `Temperatura`, `Umidade`, `Pressao`, `Ruido` e `Tempo`.
* **Atuadores (TCP):** Realizam o *handshake* enviando o prefixo obrigatório `ATUADOR|[ID_DO_ATUADOR]`. Ao executarem uma ação, devolvem uma resposta ao servidor com o prefixo `RESPOSTA|`.
* **Clientes (TCP):** Identificam-se apenas com o nome de usuário e utilizam comandos (ex: `receber`, `atuar`).

---

## 🏗️ Arquitetura do Sistema

O sistema é dividido em 4 componentes principais:

1. **Servidor (Broker):** Responsável por toda lógica do sistema. Escuta na porta `8080` (TCP) para clientes/atuadores e `5000` (UDP) para sensores.
2. **Atuadores:** Dispositivos simulados que recebem comandos do servidor para alterar o ambiente (ex: ligar/desligar).
3. **Sensores (Telemetria):** Dispositivos que enviam dados (Temperatura, Umidade, Pressão, etc.) em tempo real via UDP.
4. **Clientes:** Usuários que se conectam ao servidor para ler os dados dos sensores ou enviar comandos aos atuadores.


<p align="center">
  <img width="710" height="287" alt="image" src="https://github.com/user-attachments/assets/973adeda-6a3e-4e1a-b1d1-fd0abaa34ec7" />
  <br>
  <em>Figura 1 – Arquitetura distribuída do Message Broker IoT.</em>
</p>

## 📂 Estrutura do Projeto
```
Rota-Das-Coisas
│
├── Server/          # Broker principal
│   ├── server.go
│   └── Dockerfile
│
├── Sensors/         # Simuladores de sensores (telemetria UDP)
│   ├── sensor.go
│   └── Dockerfile
│
├── Actuators/       # Dispositivos atuadores controlados pelo broker
│   ├── actuator.go
│   └── Dockerfile
│
├── User/            # Cliente interativo TCP
│   ├── client.go
│   └── Dockerfile
│
├── Test/            # Testes de estresse do sistema
│   ├── stresser.go
│   └── Dockerfile
│
├── docker-compose.yml
├── go.mod
├── README.md
└── LICENSE
```

---
## 🐳 Executando com Docker (Recomendado)

A organização dos serviços é feita via Docker Compose, permitindo subir toda a arquitetura com um único comando.

### 1. Subir a Infraestrutura Completa
Na raiz do projeto `Rota-Das-Coisas`, onde está localizado o arquivo `docker-compose.yml`, execute:
```bash
docker-compose up --build
```
Isso iniciará automaticamente o Servidor, três instâncias de Sensores transmitindo dados e duas instâncias de Atuadores prontos para receber comandos, todos comunicando-se em uma rede Docker isolada.
### 2. Encerrar o Sistema
Para desligar todos os containers e limpar a rede:
```bash
docker-compose down
```
---
### 3. Comandos Úteis de Monitoramento e Interação
Enquanto o sistema estiver rodando, abra um novo terminal. Esses são os principais comandos para interagir com o sistema:

#### Listar todos os containers ativos:
Para ver o nome exato dos containers, o status e as portas alocadas:

```bash
docker ps
```
---

#### Acompanhar os logs de tudo que está acontecendo em tempo real:
Ver tudo que está acontecendo na rede
```bash
docker-compose logs -f
```
---

#### Acompanhar logs de um container específico:
Visualizar apenas os logs de um container específico, exemplo: O que o servidor está recebendo, os estados dos atuadores, etc.
```bash
docker logs -f <nome_do_container>
```
---
#### Executar um novo Cliente interativo:
Para criar um cliente que faça a interação com o Servidor, basta executar:
```bash
docker-compose run --rm client
```

---
#### Criar um novo container de atuador ou sensor:
Criar um novo sensor:
```bash
docker-compose run -d --rm sensor-01 ./sensor <nome_sensor>
```

Criar um novo atuador:
```bash
docker-compose run -d --rm atuador-01 ./actuator <nome_atuador>
```
---
#### Testar a Tolerância a Falhas (Parar um container):
Caso deseje testar se o Broker detecta a queda de um sensor após 20 segundos, derrube o container dele propositalmente:

```bash
docker stop <nome_do_container_sensor>
```
---

#### Reiniciar um container específico:
Para ligar novamente um dispositivo que foi derrubado:

```bash
docker restart <nome_do_container>
```
---

## 🌐 Executando em Múltiplos Computadores (Rede Distribuída)
Para testar a verdadeira natureza distribuída do sistema (ex: Servidor em um computador e Cliente/Atuadores em outros PCs conectados na mesma rede local), siga os passos abaixo:

1. Configurar o IP do Servidor:
Abra os arquivos fonte (client.go, sensor.go, actuator.go, stresser.go) e altere a variável ipServidor no topo do código para o IP físico da máquina que rodará o Servidor (Ex: 172.16.201.9). Salve os arquivos.

2. No Computador 1 (Servidor):
Levante apenas a infraestrutura do sistema:
```bash
docker-compose up -d --build
```

3. Nos outros computadores:
Com os arquivos atualizados com o IP correto, compile as novas imagens e execute os serviços de forma isolada utilizando a flag --no-deps (isso impede que o Docker tente recriar o servidor na máquina local):

* Para rodar um cliente que interaja com o servidor faça:
  ```bash
   docker-compose build client
   docker-compose run --rm --no-deps client
  ```

##  🛠️ Como Executar (Localmente sem Docker)
Caso prefira rodar os programas Go diretamente da raiz do projeto no terminal:

1. Inicie o Servidor:
   ```bash
    go run ./Server/server.go
   ```
2. Defina a variável ipServidor dos clientes, atuadores, sensores e stresser para:
   ```bash
    ipServidor = "localhost"
   ```
3. Conecte um Atuador:
   ```bash
    go run ./Actuators/actuator.go atuador_01
   ```
4. Conecte um Cliente:
   ```bash
    go run ./User/client.go
   ```
5. Conecte um sensor:
   ```bash
   go run ./Sensors/sensor.go sensor_01
   ```
## 💻 Interface de Comandos (Menu do Cliente)
Uma vez conectado como cliente via TCP, você pode usar os seguintes comandos interativos:
| Comando             | Descrição                                     | Exemplo                  |
| ------------------- | --------------------------------------------- | ------------------------ |
| `listar`            | Lista sensores e atuadores ativos             | `listar`                 |
| `receber [id]`      | Escuta dados de um sensor específico ou todos | `receber sensor_01`      |
| `parar`             | Para de escutar os dados recebidos do sensor  | `parar`                  |
| `atuar [id] [ação]` | Envia comando para um atuador (ligar/desligar)| `atuar atuador_01 ligar` |
| `help`              | Mostra menu de ajuda                          | `help`                   |
| `sair`              | Desconecta do servidor                        | `sair`                   |


## 🧪 Teste de Estresse
Para comprovar a resistência do servidor e das filas de requisição Mutex, o repositório inclui um script de ataque controlado (stresser.go).

Ele gera centenas de conexões TCP simultâneas simulando clientes reais, aplicando jitter (espalhamento) para evitar o bloqueio de sockets do SO host.

### **⚙️ Capacidade e Configuração:**
Dentro do código, é possível ajustar a variável que define a quantidade de "bots" simultâneos, basta colocar a quantidade que deseja na variável totalClientes. 
```go
   totalClientes := <numero_que_deseja>
```

A arquitetura foi testada com até 6000 requisições simultâneas, lembrando que a quantidade de bots criados são o dobro do definido no totalClientes, já que executamos dois tipos de testes simultaneamente. Um fica alterando os estados do atuador 1 e 2 dependendo se o ID do bot for par ou ímpar. Já o outro, fica ligando exaustivamente o atuador 2.

Como executar:
### Cenário 1: Se tudo estiver no mesmo computador:
#### Docker (Recomendado): 
Com a infraestrutura já rodando, abra um novo terminal e execute o teste com o comando:
```bash
docker-compose run --rm test
```

#### Localmente:
Certifique-se de ter o Servidor e dois atuadores (nomeados obrigatoriamente como atuador_01 e atuador_02) rodando e execute na pasta Test:
   ```bash
    go run stresser.go
   ```

### Cenário 2: Arquitetura Distribuída (Servidor em outro PC)
Caso você queira rodar o servidor em uma máquina e executar o teste de estresse a partir de outra na mesma rede local, siga esta ordem:

Abra o arquivo stresser.go e altere a variável ipServidor no topo do código para o IP real da máquina onde o servidor está hospedado (ex: 172.16.201.9).

No terminal do PC cliente, force a reconstrução da imagem Docker para que ela absorva o novo IP:

```bash
docker-compose build test
```

Agora rode a instância do teste:
```bash
docker-compose run --rm --no-deps test
```

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
