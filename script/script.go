package script

import (
	"fmt"
)

func CalculateOperationStep(params map[string]string, input interface{}) (interface{}, error) {
	fmt.Println("Iniciando calculo de tabela PRICE")

	fmt.Println("Qtde de Parcelas", params["qtde_parcelas"])

	fmt.Println("Amount", input)

	return nil, nil
}
