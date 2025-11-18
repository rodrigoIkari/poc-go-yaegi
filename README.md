# POC Go + Yaegi

Prova de conceito demonstrando interpretação dinâmica de código Go usando [Yaegi](https://github.com/traefik/yaegi).

## Visão Geral

Este projeto explora a capacidade de executar código Go dinamicamente em runtime, sem necessidade de recompilação. O caso de uso demonstrado é um sistema de cálculo de operações financeiras usando a **Tabela PRICE**, onde a lógica de cálculo pode ser modificada através de scripts interpretados.

### O que é Yaegi?

Yaegi é um interpretador Go escrito em Go. Ele permite executar código Go de forma dinâmica, tornando possível:
- Modificar lógica de negócio sem recompilar a aplicação
- Criar sistemas de plugins dinâmicos
- Implementar engines de regras configuráveis
- Desenvolver ambientes de scripting em Go

## Estrutura do Projeto

```
.
├── functions/           # Tipos de domínio (estruturas de dados)
│   └── operation.go     # Operation, Installment, Component
├── script/              # Scripts interpretados dinamicamente
│   └── script.go        # Lógica de cálculo PRICE
├── main.go              # Ponto de entrada e configuração do Yaegi
├── go.mod
└── README.md
```

## Funcionalidades

### Cálculo da Tabela PRICE

A Tabela PRICE é um sistema de amortização com parcelas fixas, amplamente utilizado em financiamentos. A fórmula implementada é:

```
PMT = PV × [i × (1+i)^n] / [(1+i)^n - 1]
```

Onde:
- **PMT**: Valor da parcela
- **PV**: Valor presente (principal)
- **i**: Taxa de juros
- **n**: Número de parcelas

### Características do POC

- ✅ Interpretação dinâmica de código Go
- ✅ Passagem de tipos customizados entre código compilado e interpretado
- ✅ Validação de parâmetros e tratamento de erros
- ✅ Cálculo completo de amortização com detalhamento de juros e principal
- ✅ Type-safe com tipos específicos ao invés de `interface{}`

## Como Executar

### Pré-requisitos

- Go 1.22 ou superior

### Instalação

```bash
# Clone o repositório
git clone <url-do-repo>
cd poc-go-yaegi

# Baixe as dependências
go mod download
```

### Execução

```bash
go run main.go
```

### Saída Esperada

```
Compilando script...
Script compilado com sucesso
Executando script...
Iniciando cálculo de tabela PRICE
Valor do financiamento: R$ 1000.00
Quantidade de parcelas: 12
Taxa de juros mensal: 2.00%
Cálculo concluído! Total a pagar: R$ 1134.72

=== Resultado do Cálculo ===
Valor Financiado: R$ 1000.00
Quantidade de Parcelas: 12
Valor Total a Pagar: R$ 1134.72
Valor de Cada Parcela: R$ 94.56

Script finalizado com sucesso!
```

## Como Funciona

### 1. Inicialização do Interpretador (main.go)

```go
i := interp.New(interp.Options{GoPath: build.Default.GOPATH})

// Registrar bibliotecas padrão disponíveis para o script
i.Use(interp.Exports{
    "math/math":       stdlib.Symbols["math/math"],
    "fmt/fmt":         stdlib.Symbols["fmt/fmt"],
    "time/time":       stdlib.Symbols["time/time"],
    "strconv/strconv": stdlib.Symbols["strconv/strconv"],
})

// Registrar tipos customizados
i.Use(interp.Exports{
    "poc-go-yaegi/functions/functions": map[string]reflect.Value{
        "Operation":   reflect.ValueOf((*functions.Operation)(nil)),
        "Installment": reflect.ValueOf((*functions.Installment)(nil)),
        "Component":   reflect.ValueOf((*functions.Component)(nil)),
    },
})
```

### 2. Carregamento e Execução do Script

```go
// Ler arquivo do script
scriptFile, err := os.ReadFile("script/script.go")
scriptCode := string(scriptFile)

// Compilar o script
_, err = i.Eval(scriptCode)

// Obter referência para a função
script, err := i.Eval("script.CalculateOperationStep")

// Type assertion segura
calculateOperationStep, ok := script.Interface().(func(map[string]string, functions.Operation) (*functions.Operation, error))
```

### 3. Execução com Dados Reais

```go
params := map[string]string{
    "qtde_parcelas": "12",
    "taxa_juros":    "0.02",  // 2% ao mês
}

inputOperation := functions.Operation{
    Amount: 1000.00,
}

result, err := calculateOperationStep(params, inputOperation)
```

## Modificando o Script

O grande diferencial deste POC é que você pode modificar `script/script.go` **sem recompilar** o programa principal. Basta:

1. Editar `script/script.go`
2. Executar `go run main.go` novamente
3. A nova lógica será interpretada em runtime

### Exemplo: Adicionar IOF

```go
// Em script/script.go, adicione:
iof := input.Amount * 0.0038 // 0.38% de IOF

components := []functions.Component{
    {Name: "Principal", Amount: amortizacao},
    {Name: "Juros", Amount: juros},
    {Name: "IOF", Amount: iof / float64(qtdeParcelas)}, // IOF dividido nas parcelas
}
```

## Tipos de Dados

### Operation

```go
type Operation struct {
    Amount          float64       // Valor principal financiado
    TotalAmount     float64       // Valor total a pagar
    InstallmentsQty int           // Quantidade de parcelas
    Installments    []Installment // Lista de parcelas
}
```

### Installment

```go
type Installment struct {
    Number      int         // Número da parcela
    Amount      float64     // Valor da parcela
    TotalAmount float64     // Valor total acumulado
    DueDate     time.Time   // Data de vencimento
    Components  []Component // Componentes (juros, principal, etc)
}
```

### Component

```go
type Component struct {
    Name   string  // Nome do componente
    Amount float64 // Valor do componente
}
```

## Limitações do Yaegi

Ao trabalhar com Yaegi, esteja ciente de algumas limitações:

1. **Performance**: Código interpretado é mais lento que código compilado
2. **Debugging**: Mensagens de erro podem ser menos claras
3. **Imports**: Todas as bibliotecas usadas no script devem ser registradas explicitamente no `Use()`
4. **Reflexão**: Algumas operações avançadas de reflexão podem não funcionar
5. **Generics**: Suporte limitado a generics do Go 1.18+

## Casos de Uso

Este padrão é útil para:

- **Sistemas de regras de negócio**: Permitir que usuários avançados modifiquem lógica sem redeploy
- **Motores de cálculo**: Diferentes fórmulas de cálculo que podem mudar frequentemente
- **Plugins**: Sistema de plugins sem necessidade de compilação externa
- **Prototipagem**: Testar lógicas rapidamente sem recompilar
- **Configuração avançada**: Scripts como forma de configuração programática

## Segurança

⚠️ **ATENÇÃO**: Executar código arbitrário pode ser perigoso. Em produção:

1. Valide e sanitize scripts antes de executar
2. Implemente sandboxing apropriado
3. Limite bibliotecas disponíveis (evite `os`, `exec`, `net`)
4. Implemente timeouts para evitar loops infinitos
5. Considere code review para scripts em produção

## Melhorias Futuras

- [ ] Adicionar testes unitários
- [ ] Implementar outros sistemas de amortização (SAC, SAM)
- [ ] Cache de scripts compilados
- [ ] Hot reload de scripts em runtime
- [ ] Interface web para edição de scripts
- [ ] Métricas e observabilidade
- [ ] Validação de scripts com linting

## Referências

- [Yaegi GitHub](https://github.com/traefik/yaegi)
- [Yaegi Documentation](https://pkg.go.dev/github.com/traefik/yaegi)
- [Tabela PRICE - Wikipédia](https://pt.wikipedia.org/wiki/Tabela_Price)

## Licença

Este é um projeto de prova de conceito para fins educacionais.

## Autor

Rodrigo Ikari
