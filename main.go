package main

import (
	"fmt"
	"go/build"

	"github.com/traefik/yaegi/interp"
	"github.com/traefik/yaegi/stdlib"
)

func main() {

	i := interp.New(interp.Options{GoPath: build.Default.GOPATH})

	fmt.Println("Executando script...")

	i.Use(interp.Exports{
		"math/math": stdlib.Symbols["math/math"],
		"fmt/fmt":   stdlib.Symbols["fmt/fmt"],
	})
	// i.Use(stdlib.Symbols)

	_, err := i.Eval(`

	package main

	import (
	"fmt"
	"math"
	)

	func main() {
		fmt.Println("Hello, World!")
	}
	
	
	`)

	if err != nil {
		fmt.Println("Erro ao executar script: ", err)
		return
	}

	fmt.Println("Script finalizado com sucesso!")

}
