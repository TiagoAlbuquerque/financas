# Controle de Finanças Pessoais (Wails + Go + JS)

Este é um sistema de controle de arrecadação (receitas) e gastos (despesas) pessoais, desenvolvido em Go para o backend e tecnologias web (HTML, CSS, JS e Chart.js) para o frontend. Ele compila em um único executável standalone de alta performance para Windows.

O sistema utiliza um banco de dados local baseado em arquivos **JSON**, organizado em uma estrutura hierárquica inteligente que separa resumos e detalhamentos. Isso permite que a interface apenas carregue e leia os arquivos estritamente necessários à medida que o usuário navega entre os anos e meses, reduzindo o consumo de memória e otimizando a velocidade.

---

## 🏗️ Estrutura de Pastas de Dados (JSON)

Os dados financeiros do usuário são guardados no diretório `data/` localizado na mesma pasta do executável (modo portátil). Caso não haja permissão de escrita nessa pasta (ex: em `C:\Program Files`), o aplicativo automaticamente cria e utiliza a pasta `%APPDATA%/FinancasPersonalApp/data/`.

A estrutura interna segue a lógica abaixo:

```text
data/
├── summary.json                  # Resumo geral (saldo acumulado de todos os anos)
└── [ANO]/                        # Ex: 2026/
    ├── summary.json              # Resumo anual (receitas/despesas acumuladas por mês)
    └── [MÊS]/                    # Ex: 06/
        ├── summary.json          # Resumo mensal (receitas/despesas e divisões por categoria)
        └── details.json          # Detalhamento (lista completa de transações individuais do mês)
```

### Exemplo de fluxo de leitura:
* Ao abrir o aplicativo, a interface lê apenas o arquivo de nível superior `data/summary.json` para exibir o saldo acumulado de todos os anos e montar a árvore de navegação no menu lateral.
* Se o usuário clica em um ano (ex: **2026**), a interface lê `data/2026/summary.json` para montar o gráfico de barras comparativo de Janeiro a Dezembro.
* Se o usuário clica em um mês (ex: **Junho de 2026**), a interface faz a requisição de `data/2026/06/summary.json` (para o gráfico de pizza de categorias) e `data/2026/06/details.json` (para renderizar a tabela de lançamentos).

Quando o usuário insere ou apaga uma transação, o backend em Go reescreve os arquivos do mês correspondente e propaga as alterações para os resumos anual e geral imediatamente.

---

## 🛠️ Como Compilar e Executar

### Pré-requisitos (No ambiente de compilação do Windows):
1. **Go (Golang)** instalado (versão 1.21+ recomendada)
2. **Node.js** instalado (para gerenciar e empacotar o frontend via npm/Vite)
3. **Wails CLI** instalado. Para instalar, abra o terminal no Windows e execute:
   ```bash
   go install github.com/wailsapp/wails/v2/cmd/wails@latest
   ```

### Executando em Modo de Desenvolvimento (Live Reload):
Para abrir o aplicativo em modo de desenvolvimento com recarregamento em tempo real (qualquer mudança no código HTML/CSS/JS ou Go atualizará o aplicativo imediatamente):
```bash
wails dev
```

### Compilando para Windows (Standalone EXE):
Para gerar o executável definitivo e totalmente independente para Windows (sem necessidade de instalar Node.js ou Go nas máquinas dos usuários finais):
```bash
wails build -platform windows/amd64
```
O executável `.exe` será gerado dentro da pasta `build/bin/`.

---

## 🎨 Características do Design e Interface

* **Estética Premium Dark**: Fundo ultra escuro (Slate/Navy) com elementos em glassmorphism (transparências elegantes).
* **Feedback Visual Fluido**: Efeitos de hover brilhantes nos cartões de saldo e itens da barra lateral.
* **Componentização Standalone**: Ícones e fontes integrados localmente, eliminando a dependência de conexões de internet (CDN).
* **Gráficos Dinâmicos (Chart.js)**: Gráficos de barra interativos nas visões geral/anual e gráficos de pizza nas visões mensais.
* **Filtros Avançados**: Filtre transações instantaneamente por palavra-chave ou categoria.
