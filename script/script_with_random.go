package script

import (
	"fmt"
	"math/rand"
	"poc-go-yaegi/functions"
	"strconv"
	"time"
)

// CalculateOperationStepWithRandom tenta usar math/rand que NÃO está registrada no Yaegi
// Este script é usado para testar que o interpretador falha corretamente quando
// uma biblioteca não registrada é utilizada
func CalculateOperationStepWithRandom(params map[string]string, input functions.Operation) (*functions.Operation, error) {
	fmt.Println("Tentando usar math/rand que não foi registrado no Yaegi...")

	// Validar e extrair quantidade de parcelas
	qtdeParcelasStr, ok := params["qtde_parcelas"]
	if !ok {
		return nil, fmt.Errorf("parâmetro 'qtde_parcelas' não fornecido")
	}

	qtdeParcelas, err := strconv.Atoi(qtdeParcelasStr)
	if err != nil || qtdeParcelas <= 0 {
		return nil, fmt.Errorf("quantidade de parcelas inválida: %s", qtdeParcelasStr)
	}

	// ESTA LINHA VAI CAUSAR ERRO porque math/rand não foi registrado no Yaegi
	randomValue := rand.Float64()
	fmt.Printf("Valor aleatório (não deveria aparecer): %.2f\n", randomValue)

	// Preencher a estrutura da operação
	input.InstallmentsQty = qtdeParcelas
	input.Installments = make([]functions.Installment, qtdeParcelas)

	for i := 0; i < qtdeParcelas; i++ {
		// Usar random na data (também vai falhar)
		randomDays := rand.Intn(30)

		input.Installments[i] = functions.Installment{
			Number:      i + 1,
			Amount:      input.Amount / float64(qtdeParcelas),
			TotalAmount: input.Amount / float64(qtdeParcelas),
			DueDate:     time.Now().AddDate(0, i, randomDays),
		}
	}

	input.TotalAmount = input.Amount

	return &input, nil
}
