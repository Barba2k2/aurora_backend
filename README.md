# 💇‍♀️ Sistema de Agendamentos para Profissionais de Beleza 💅

![Versão](https://img.shields.io/badge/versão-1.0.0-blue)
![Licença](https://img.shields.io/badge/licença-MIT-green)

Sistema completo para gerenciamento de agendamentos no setor de beleza, permitindo que profissionais administrem seus horários e clientes realizem agendamentos de forma prática e eficiente.

## 📋 Índice

- [💇‍♀️ Sistema de Agendamentos para Profissionais de Beleza 💅](#️-sistema-de-agendamentos-para-profissionais-de-beleza-)
  - [📋 Índice](#-índice)
  - [🔍 Visão Geral](#-visão-geral)
  - [🛠️ Tecnologias](#️-tecnologias)
  - [🏛️ Arquitetura](#️-arquitetura)
  - [✨ Funcionalidades Principais](#-funcionalidades-principais)
    - [👩‍💼 Para Profissionais](#-para-profissionais)
    - [👩‍🦰 Para Clientes](#-para-clientes)
  - [🚀 Instalação e Configuração](#-instalação-e-configuração)
    - [Pré-requisitos](#pré-requisitos)
    - [Passos para Instalação](#passos-para-instalação)
  - [🔌 Estrutura da API](#-estrutura-da-api)
    - [🔹 API do Cliente (`/api/v1/client/...`)](#-api-do-cliente-apiv1client)
    - [🔹 API do Profissional (`/api/v1/professional/...`)](#-api-do-profissional-apiv1professional)
  - [💳 Integração com Pagamentos](#-integração-com-pagamentos)
  - [🔔 Sistema de Notificações](#-sistema-de-notificações)
  - [📝 Licença](#-licença)

## 🔍 Visão Geral

O Sistema de Agendamentos para Profissionais de Beleza é uma plataforma completa desenvolvida para atender às necessidades específicas de salões de beleza, barbearias, estúdios de manicure/pedicure e profissionais autônomos.

A plataforma permite que clientes visualizem disponibilidade em tempo real, agendem serviços e recebam lembretes automáticos. Para profissionais, oferece uma interface de gerenciamento completa, incluindo controle de agenda, clientes e relatórios.

## 🛠️ Tecnologias

- **Backend:**
  - 🔹 GoLang (API RESTful)
  - 🔹 PostgreSQL (Banco de Dados)
  - 🔹 Docker (Containerização)
  - 🔹 JWT (Autenticação)

- **Integrações:**
  - 💳 Stripe (Gateway de pagamento principal)
  - 💰 MercadoPago (Gateway alternativo)
  - 📱 APIs de notificação (Email, SMS, WhatsApp, Push)
  - 🗺️ Serviços de Geocodificação

## 🏛️ Arquitetura

O sistema utiliza uma **Arquitetura Monolítica Modular**:

- 📦 Monolito bem estruturado com módulos internos de baixo acoplamento
- 🔄 Comunicação entre módulos via interfaces Go e injeção de dependências
- 🚀 Design pensado para facilitar eventual migração para microsserviços
- 🔐 Estratégia de soft delete para preservação de dados

```txt
┌─────────────────────────────────────────────────────────┐
│                  API Gateway Layer                      │
├─────────────────────────┬───────────────────────────────┤
│   API do Cliente        │     API do Profissional       │
├─────────────────────────┴───────────────────────────────┤
│                  Controller Layer                       │
├─────────────────────────────────────────────────────────┤
│                   Service Layer                         │
├─────────┬───────────────┬────────────┬────────────┬─────┤
│ Agendas │ Notificaçções │ Pagamentos │ Fidelidade │ ... │
├─────────┴───────────────┴────────────┴────────────┴─────┤
│                   Repository Layer                      │
├─────────────────────────────────────────────────────────┤
│                     PostgreSQL                          │
└─────────────────────────────────────────────────────────┘
```

## ✨ Funcionalidades Principais

### 👩‍💼 Para Profissionais

- 📅 Gestão completa de agenda e disponibilidade
- 👥 Cadastro e histórico de clientes
- 🧾 Cadastro e gestão de serviços
- 📊 Dashboard com métricas e relatórios
- 🔔 Configuração de notificações e lembretes automáticos
- 🏆 Programa de fidelidade personalizável
- 💸 Integração com gateways de pagamento

### 👩‍🦰 Para Clientes

- 🔍 Busca de profissionais e serviços
- 📲 Agendamento online com confirmação instantânea
- 🕒 Visualização de histórico de agendamentos
- ⭐ Sistema de avaliações
- 🎁 Acúmulo e resgate de pontos de fidelidade
- 📱 Recebimento de lembretes por múltiplos canais

## 🚀 Instalação e Configuração

### Pré-requisitos

- Docker e Docker Compose
- Go 1.18+
- PostgreSQL 13+

### Passos para Instalação

1. Clone o repositório:

```bash
git clone https://github.com/barba2k2/aurora_backend.git
cd aurora_backend
```

2. Configure as variáveis de ambiente:

```bash
cp .env.example .env
# Edite o arquivo .env com as configurações necessárias
```

3. Inicie os containers com Docker Compose:

```bash
docker-compose up -d
```

4. Execute as migrações do banco de dados:

```bash
make migrate-up
```

5. Acesse a API:

```bash
API Cliente: http://localhost:8080/api/v1/client
API Profissional: http://localhost:8080/api/v1/professional
```

## 🔌 Estrutura da API

O sistema possui duas APIs principais com versionamento explícito:

### 🔹 API do Cliente (`/api/v1/client/...`)

- Autenticação e gestão de perfil
- Busca de profissionais e serviços
- Gestão de agendamentos
- Avaliações e programa de fidelidade

### 🔹 API do Profissional (`/api/v1/professional/...`)

- Gestão de estabelecimento e serviços
- Controle de agenda e disponibilidade
- Gestão de clientes e histórico
- Relatórios e dashboard
- Configuração de módulos e notificações

## 💳 Integração com Pagamentos

O sistema integra dois gateways de pagamento:

- **Stripe (Principal)**:
  - Checkout embarcado
  - Salvamento seguro de cartões
  - Assinaturas recorrentes
  
- **MercadoPago (Secundário)**:
  - Suporte a métodos locais (PIX, boleto)
  - Checkout transparente

## 🔔 Sistema de Notificações

Sistema multicanal de notificações e lembretes:

- 📧 Email (MailGun)
- 📱 SMS (Twilio)
- 💬 WhatsApp (Evolution API)
- 🌐 Web Push Notifications (W3C)

Características:

- Templates personalizáveis
- Lembretes automáticos configuráveis
- Fallback inteligente entre canais
- Rastreamento de entregas

## 📝 Licença

Este projeto está licenciado sob a MIT License - veja o arquivo [LICENSE](LICENSE) para mais detalhes.

---

Desenvolvido com ❤️ por [Barba Tech](https://barbatech.solutions)
