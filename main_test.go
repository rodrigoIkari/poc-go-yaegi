package main

import (
	"go/build"
	"os"
	"poc-go-yaegi/functions"
	"reflect"
	"testing"

	"github.com/traefik/yaegi/interp"
	"github.com/traefik/yaegi/stdlib"
)

// TestScriptExecution_HappyPath testa o cenário feliz de execução do script
func TestScriptExecution_HappyPath(t *testing.T) {
	// Configurar o interpretador Yaegi
	i := interp.New(interp.Options{GoPath: build.Default.GOPATH})

	// Registrar bibliotecas padrão
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

	i.ImportUsed()

	// Carregar o script
	scriptFile, err := os.ReadFile("script/script.go")
	if err != nil {
		t.Fatalf("Erro ao ler arquivo de script: %v", err)
	}

	scriptCode := string(scriptFile)

	// Compilar o script
	_, err = i.Eval(scriptCode)
	if err != nil {
		t.Fatalf("Erro ao compilar script: %v", err)
	}

	// Obter a função do script
	script, err := i.Eval("script.CalculateOperationStep")
	if err != nil {
		t.Fatalf("Erro ao obter função CalculateOperationStep: %v", err)
	}

	// Type assertion
	calculateOperationStep, ok := script.Interface().(func(map[string]string, functions.Operation) (*functions.Operation, error))
	if !ok {
		t.Fatal("Função CalculateOperationStep tem assinatura incorreta")
	}

	// Preparar dados de entrada
	params := map[string]string{
		"qtde_parcelas": "12",
		"taxa_juros":    "0.02",
	}

	inputOperation := functions.Operation{
		Amount: 1000.00,
	}

	// Executar a função
	result, err := calculateOperationStep(params, inputOperation)
	if err != nil {
		t.Fatalf("Erro ao executar script: %v", err)
	}

	// Validações do resultado
	if result == nil {
		t.Fatal("Resultado não deveria ser nil")
	}

	if result.Amount != 1000.00 {
		t.Errorf("Amount incorreto: esperado 1000.00, obtido %.2f", result.Amount)
	}

	if result.InstallmentsQty != 12 {
		t.Errorf("InstallmentsQty incorreto: esperado 12, obtido %d", result.InstallmentsQty)
	}

	// Validar que o total é maior que o principal (devido aos juros)
	if result.TotalAmount <= result.Amount {
		t.Errorf("TotalAmount (%.2f) deveria ser maior que Amount (%.2f)", result.TotalAmount, result.Amount)
	}

	// Validar quantidade de parcelas
	if len(result.Installments) != 12 {
		t.Errorf("Número de parcelas incorreto: esperado 12, obtido %d", len(result.Installments))
	}

	// Validar primeira parcela
	if len(result.Installments) > 0 {
		firstInstallment := result.Installments[0]

		if firstInstallment.Number != 1 {
			t.Errorf("Número da primeira parcela incorreto: esperado 1, obtido %d", firstInstallment.Number)
		}

		if firstInstallment.Amount <= 0 {
			t.Errorf("Valor da parcela deveria ser positivo, obtido %.2f", firstInstallment.Amount)
		}

		// Validar componentes da parcela
		if len(firstInstallment.Components) != 2 {
			t.Errorf("Primeira parcela deveria ter 2 componentes, obtido %d", len(firstInstallment.Components))
		}

		// Verificar se tem componente Principal
		hasPrincipal := false
		hasJuros := false
		for _, comp := range firstInstallment.Components {
			if comp.Name == "Principal" {
				hasPrincipal = true
			}
			if comp.Name == "Juros" {
				hasJuros = true
			}
		}

		if !hasPrincipal {
			t.Error("Primeira parcela deveria ter componente 'Principal'")
		}

		if !hasJuros {
			t.Error("Primeira parcela deveria ter componente 'Juros'")
		}
	}

	// Validar que a soma das parcelas é aproximadamente igual ao total
	// (pode haver pequenas diferenças por arredondamento)
	somaaParcelas := 0.0
	for _, installment := range result.Installments {
		somaaParcelas += installment.Amount
	}

	diferenca := result.TotalAmount - somaaParcelas
	if diferenca > 0.01 || diferenca < -0.01 {
		t.Errorf("Soma das parcelas (%.2f) diverge do total (%.2f) em %.2f",
			somaaParcelas, result.TotalAmount, diferenca)
	}

	// Validar cálculo PRICE esperado
	// Para 1000 reais, 12 parcelas, 2% ao mês
	// Valor esperado por parcela: ~94.56
	expectedInstallment := 94.56
	tolerance := 0.10 // tolerância de 10 centavos

	if result.Installments[0].Amount < expectedInstallment-tolerance ||
	   result.Installments[0].Amount > expectedInstallment+tolerance {
		t.Errorf("Valor da parcela fora do esperado: esperado ~%.2f, obtido %.2f",
			expectedInstallment, result.Installments[0].Amount)
	}
}

// TestScriptExecution_InvalidParams testa execução com parâmetros inválidos
func TestScriptExecution_InvalidParams(t *testing.T) {
	i := interp.New(interp.Options{GoPath: build.Default.GOPATH})

	i.Use(interp.Exports{
		"math/math":       stdlib.Symbols["math/math"],
		"fmt/fmt":         stdlib.Symbols["fmt/fmt"],
		"time/time":       stdlib.Symbols["time/time"],
		"strconv/strconv": stdlib.Symbols["strconv/strconv"],
	})

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
		t.Fatalf("Erro ao ler arquivo de script: %v", err)
	}

	_, err = i.Eval(string(scriptFile))
	if err != nil {
		t.Fatalf("Erro ao compilar script: %v", err)
	}

	script, err := i.Eval("script.CalculateOperationStep")
	if err != nil {
		t.Fatalf("Erro ao obter função: %v", err)
	}

	calculateOperationStep, ok := script.Interface().(func(map[string]string, functions.Operation) (*functions.Operation, error))
	if !ok {
		t.Fatal("Type assertion falhou")
	}

	// Teste com parâmetros faltando
	t.Run("MissingQtdeParcelas", func(t *testing.T) {
		params := map[string]string{
			"taxa_juros": "0.02",
		}
		inputOperation := functions.Operation{Amount: 1000.00}

		_, err := calculateOperationStep(params, inputOperation)
		if err == nil {
			t.Error("Deveria retornar erro quando qtde_parcelas está faltando")
		}
	})

	// Teste com parâmetros inválidos
	t.Run("InvalidQtdeParcelas", func(t *testing.T) {
		params := map[string]string{
			"qtde_parcelas": "abc",
			"taxa_juros":    "0.02",
		}
		inputOperation := functions.Operation{Amount: 1000.00}

		_, err := calculateOperationStep(params, inputOperation)
		if err == nil {
			t.Error("Deveria retornar erro quando qtde_parcelas é inválida")
		}
	})
}

// TestScriptWithUnregisteredLibrary testa que o script falha ao usar biblioteca não registrada
func TestScriptWithUnregisteredLibrary(t *testing.T) {
	i := interp.New(interp.Options{GoPath: build.Default.GOPATH})

	// Registrar apenas bibliotecas básicas, SEM math/rand
	i.Use(interp.Exports{
		"fmt/fmt":         stdlib.Symbols["fmt/fmt"],
		"time/time":       stdlib.Symbols["time/time"],
		"strconv/strconv": stdlib.Symbols["strconv/strconv"],
	})

	i.Use(interp.Exports{
		"poc-go-yaegi/functions/functions": map[string]reflect.Value{
			"Operation":   reflect.ValueOf((*functions.Operation)(nil)),
			"Installment": reflect.ValueOf((*functions.Installment)(nil)),
			"Component":   reflect.ValueOf((*functions.Component)(nil)),
		},
	})

	i.ImportUsed()

	// Carregar script que usa math/rand (não registrada)
	scriptFile, err := os.ReadFile("script/script_with_random.go")
	if err != nil {
		t.Fatalf("Erro ao ler script de teste: %v", err)
	}

	// Este script usa math/rand que NÃO foi registrado
	_, err = i.Eval(string(scriptFile))

	// Esperamos que o erro ocorra
	if err == nil {
		t.Fatal("Esperava erro ao usar biblioteca não registrada (math/rand), mas não ocorreu erro")
	}

	// Verificar que a mensagem de erro menciona o problema
	errorMsg := err.Error()
	if errorMsg == "" {
		t.Error("Mensagem de erro está vazia")
	}

	t.Logf("Erro esperado recebido: %v", err)

	// Verificar se a mensagem menciona a biblioteca não encontrada
	// O Yaegi tipicamente retorna erro como "unable to find source related to"
	if len(errorMsg) > 0 {
		// Apenas log do erro, já validamos que ele existe
		t.Logf("Biblioteca não registrada causou erro como esperado: %s", errorMsg)
	}
}
