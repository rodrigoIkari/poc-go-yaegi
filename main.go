package main

import (
	"fmt"
	"go/build"
	"os"
	"poc-go-yaegi/functions"
	"reflect"

	"github.com/traefik/yaegi/interp"
	"github.com/traefik/yaegi/stdlib"
)

func main() {

	i := interp.New(interp.Options{GoPath: build.Default.GOPATH})

	fmt.Println("Compilando script...")

	// Registrar bibliotecas padrão disponíveis para o script
	i.Use(interp.Exports{
		"math/math":       stdlib.Symbols["math/math"],
		"fmt/fmt":         stdlib.Symbols["fmt/fmt"],
		"time/time":       stdlib.Symbols["time/time"],
		"strconv/strconv": stdlib.Symbols["strconv/strconv"],
	})

	// Registrar o pacote functions com seus tipos
	i.Use(interp.Exports{
		"poc-go-yaegi/functions/functions": map[string]reflect.Value{
			"Operation":   reflect.ValueOf((*functions.Operation)(nil)),
			"Installment": reflect.ValueOf((*functions.Installment)(nil)),
			"Component":   reflect.ValueOf((*functions.Component)(nil)),
		},
	})

	i.ImportUsed()

	scriptFile, err := os.ReadFile("script/script.go")
	if err != nil {
		fmt.Println("Erro ao abrir script: ", err)
		return
	}

	scriptCode := string(scriptFile)

	_, err = i.Eval(scriptCode)

	if err != nil {
		fmt.Println("Erro ao interpretar arquivo de script: ", err)
		return
	}

	script, err := i.Eval("script.CalculateOperationStep")

	if err != nil {
		fmt.Println("Erro ao interpretar função CalculateOperationStep: ", err)
		return
	}
	fmt.Println("Script compilado com sucesso")
	fmt.Println("Executando script...")

	// Type assertion segura
	calculateOperationStep, ok := script.Interface().(func(map[string]string, functions.Operation) (*functions.Operation, error))
	if !ok {
		fmt.Println("Erro: função CalculateOperationStep tem assinatura incorreta")
		return
	}

	params := map[string]string{
		"qtde_parcelas": "10",
		"taxa_juros":    "0.02",
	}

	inputOperation := functions.Operation{
		Amount: 1500.00,
	}

	// Validar entrada
	if inputOperation.Amount <= 0 {
		fmt.Println("Erro: valor da operação deve ser positivo")
		return
	}

	result, err := calculateOperationStep(params, inputOperation)
	if err != nil {
		fmt.Println("Erro ao executar script: ", err)
		return
	}

	fmt.Printf("\n=== Resultado do Cálculo ===\n")
	fmt.Printf("Valor Financiado: R$ %.2f\n", result.Amount)
	fmt.Printf("Quantidade de Parcelas: %d\n", result.InstallmentsQty)
	fmt.Printf("Valor Total a Pagar: R$ %.2f\n", result.TotalAmount)
	fmt.Printf("Valor de Cada Parcela: R$ %.2f\n\n", result.Installments[0].Amount)

	fmt.Println("Script finalizado com sucesso!")

}
