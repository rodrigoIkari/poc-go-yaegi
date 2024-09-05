package main

import (
	"fmt"
	"go/build"
	"os"

	"github.com/traefik/yaegi/interp"
	"github.com/traefik/yaegi/stdlib"
)

func main() {

	i := interp.New(interp.Options{GoPath: build.Default.GOPATH})

	fmt.Println("Compilando script...")

	// importFunctions()

	i.Use(interp.Exports{
		"math/math": stdlib.Symbols["math/math"],
		"fmt/fmt":   stdlib.Symbols["fmt/fmt"],
		"time/time": stdlib.Symbols["time/time"],
		// "poc-go-yaegi/functions": stdlib.Symbols["poc-go-yaegi/functions"],
	})

	i.ImportUsed()

	script_file, err := os.ReadFile("script/script.go")
	if err != nil {
		fmt.Println("Erro ao abrir script: ", err)
		return
	}

	script_code := string(script_file)

	_, err = i.Eval(script_code)

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

	calculateOperationStep := script.Interface().(func(map[string]string))

	params := map[string]string{
		"qtde_parcelas": "12",
	}

	// inputOperation := functions.Operation{
	// 	Amount: 1000.00,
	// }

	calculateOperationStep(params)
	if err != nil {
		fmt.Println("Erro ao executar script: ", err)
		return
	}

	// fmt.Println("Valor Total da Operação: ", op.TotalAmount)

	fmt.Println("Script finalizado com sucesso!")

}

// func importFunctions() {
// 	stdlib.Symbols["github.com/rodrigoIkari/poc-go-yaegi/functions"] = map[string]reflect.Value{
// 		// function, constant and variable definitions
// 		"Operation":   reflect.ValueOf(functions.Operation{}),
// 		"Installment": reflect.ValueOf(functions.Installment{}),
// 		"Component":   reflect.ValueOf(functions.Component{}),
// 	}
// }
