// test_templates.go
package main

import (
	"fmt"
	"html/template"
	"log"
	"path/filepath"
)

func main() {
	funcMap := template.FuncMap{
		"sub": func(a, b int) int { return a - b },
		"add": func(a, b int) int { return a + b },
	}

	tmpl := template.New("").Funcs(funcMap)
	_, err := tmpl.ParseGlob(filepath.Join("web", "templates", "**/*.html"))
	if err != nil {
		log.Fatal("Error:", err)
	}
	fmt.Println("Templates cargados correctamente")
}
