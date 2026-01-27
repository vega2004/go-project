package main

import (
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"tu-proyecto/config"
	"tu-proyecto/handler"
	"tu-proyecto/repository"
	"tu-proyecto/service"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// TemplateRenderer para usar templates HTML
type TemplateRenderer struct {
	templates *template.Template
}

// Render implementa la interfaz de Echo para renderizar templates
func (t *TemplateRenderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	// En producción, menos logs (puedes dejarlos para debug)
	if os.Getenv("RAILWAY_ENVIRONMENT") == "" {
		fmt.Printf("\n🎨 [RENDERER] Renderizando template: '%s'\n", name)
	}

	// Buscar el template
	tmpl := t.templates.Lookup(name)
	if tmpl == nil {
		fmt.Printf("❌ ERROR: Template '%s' no encontrado!\n", name)
		return fmt.Errorf("template '%s' not found", name)
	}

	// Crear un buffer para capturar la salida
	buf := new(strings.Builder)
	err := tmpl.Execute(buf, data)
	if err != nil {
		fmt.Printf("❌ ERROR ejecutando template: %v\n", err)
		return err
	}

	// Escribir al writer real
	_, err = w.Write([]byte(buf.String()))
	return err
}

func main() {
	fmt.Println("=== INICIANDO SISTEMA ===")

	// Obtener puerto de Railway (IMPORTANTE PARA PRODUCCIÓN)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Default para desarrollo local
	}
	fmt.Printf("🚀 Puerto: %s\n", port)

	// Conexión a la base de datos
	fmt.Println("1. Conectando a PostgreSQL...")
	db, err := config.ConnectDB()
	if err != nil {
		log.Fatal("❌ Error conectando a la base de datos:", err)
	}
	defer db.Close()
	fmt.Println("✅ Conexión a DB exitosa")

	// Inicializar Echo
	fmt.Println("\n2. Inicializando Echo...")
	e := echo.New()

	// Middleware básicos para producción
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: `${time_rfc3339} ${status} ${method} ${host}${path} ${latency_human}` + "\n",
	}))
	e.Use(middleware.RecoverWithConfig(middleware.RecoverConfig{
		StackSize: 1 << 10, // 1 KB
	}))
	e.Use(middleware.Secure()) // Middleware de seguridad para producción
	e.Use(middleware.Gzip())   // Compresión GZIP para mejor performance
	fmt.Println("✅ Middlewares configurados")

	// *** CONFIGURAR TEMPLATE RENDERER ***
	fmt.Println("\n3. Configurando templates...")

	// Usar path relativo que funcione en cualquier entorno
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal("❌ Error obteniendo directorio actual:", err)
	}

	// Para debug: mostrar estructura de directorios
	if os.Getenv("RAILWAY_ENVIRONMENT") == "" {
		fmt.Printf("   Directorio actual: %s\n", cwd)
		files, _ := os.ReadDir(".")
		fmt.Println("   Archivos en directorio raíz:")
		for _, f := range files {
			fmt.Printf("   - %s\n", f.Name())
		}
	}

	templatesDir := filepath.Join(cwd, "templates")
	fmt.Printf("   Buscando templates en: %s\n", templatesDir)

	// Verificar si existe la carpeta templates
	if _, err := os.Stat(templatesDir); os.IsNotExist(err) {
		log.Fatalf("❌ Carpeta 'templates' no encontrada en: %s", templatesDir)
	}

	// Cargar templates
	templates, err := template.ParseGlob(filepath.Join(templatesDir, "*.html"))
	if err != nil {
		log.Fatal("❌ Error cargando templates:", err)
	}

	renderer := &TemplateRenderer{
		templates: templates,
	}
	e.Renderer = renderer

	// Verificar templates cargados
	fmt.Printf("   ✅ Templates cargados: %d\n", len(renderer.templates.Templates()))
	for i, t := range renderer.templates.Templates() {
		fmt.Printf("   %d. %s\n", i+1, t.Name())
	}

	// Servir archivos estáticos - IMPORTANTE para Railway
	staticDir := filepath.Join(cwd, "static")
	fmt.Printf("   📁 Serviendo archivos estáticos desde: %s\n", staticDir)

	// Verificar si existe la carpeta static
	if _, err := os.Stat(staticDir); os.IsNotExist(err) {
		fmt.Printf("   ⚠️  Carpeta 'static' no encontrada, creando...\n")
		os.MkdirAll(staticDir, 0755)
	}

	e.Static("/static", staticDir)
	fmt.Println("✅ Archivos estáticos configurados")

	// *** MIDDLEWARE DE BREADCRUMBS ***
	fmt.Println("\n4. Configurando middleware de breadcrumbs...")
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			breadcrumbs := []map[string]string{
				{"name": "🏠 Inicio", "url": "/"},
			}

			currentPath := c.Path()

			// Solo log en desarrollo
			if os.Getenv("RAILWAY_ENVIRONMENT") == "" {
				fmt.Printf("   🍞 [MIDDLEWARE] Path actual: %s\n", currentPath)
			}

			if currentPath == "/form" {
				breadcrumbs = append(breadcrumbs, map[string]string{
					"name": "📝 Registro",
					"url":  "/form",
				})
			} else if currentPath == "/success" {
				breadcrumbs = append(breadcrumbs,
					map[string]string{"name": "📝 Registro", "url": "/form"},
					map[string]string{"name": "✅ Éxito", "url": "/success"},
				)
			} else if currentPath == "/maintenance" {
				breadcrumbs = append(breadcrumbs, map[string]string{
					"name": "🔧 Mantenimiento",
					"url":  "/maintenance",
				})
			}

			c.Set("breadcrumbs", breadcrumbs)
			return next(c)
		}
	})

	// Inicializar capas
	fmt.Println("\n5. Inicializando capas de aplicación...")
	userRepo := repository.NewUserRepository(db)
	userService := service.NewUserService(userRepo)
	userHandler := handler.NewUserHandler(userService)
	fmt.Println("✅ Capas inicializadas")

	// *** RUTAS ***
	fmt.Println("\n6. Configurando rutas...")

	// Ruta de inicio - Usando el handler Welcome
	e.GET("/", userHandler.Welcome)

	// Ruta de debug (solo en desarrollo)
	if os.Getenv("RAILWAY_ENVIRONMENT") == "" {
		e.GET("/debug", func(c echo.Context) error {
			html := "<h1>Debug Info</h1><pre>"
			html += fmt.Sprintf("Entorno: %s\n", os.Getenv("RAILWAY_ENVIRONMENT"))
			html += fmt.Sprintf("Puerto: %s\n", port)
			html += fmt.Sprintf("Templates cargados: %d\n", len(renderer.templates.Templates()))
			for i, t := range renderer.templates.Templates() {
				html += fmt.Sprintf("%d. %s\n", i+1, t.Name())
			}
			html += "</pre>"
			html += `<p><a href="/">Volver</a> | <a href="/form">Formulario</a></p>`
			return c.HTML(http.StatusOK, html)
		})
	}

	// Otras rutas
	e.GET("/form", userHandler.ShowForm)
	e.POST("/submit", userHandler.SubmitForm)
	e.GET("/success", userHandler.ShowSuccess)
	e.GET("/maintenance", userHandler.Maintenance)

	fmt.Println("✅ Rutas configuradas")

	// Health check para Railway
	e.GET("/health", func(c echo.Context) error {
		// Verificar conexión a base de datos
		err := db.Ping()
		if err != nil {
			return c.String(http.StatusServiceUnavailable, "Database connection failed")
		}
		return c.String(http.StatusOK, "OK")
	})

	// Iniciar servidor
	fmt.Println("\n" + strings.Repeat("=", 50))
	fmt.Printf("🚀 Servidor iniciado en puerto %s\n", port)
	if os.Getenv("RAILWAY_ENVIRONMENT") != "" {
		fmt.Println("✅ Entorno: Railway (Producción)")
	} else {
		fmt.Println("✅ Entorno: Desarrollo Local")
	}
	fmt.Println("📌 Presiona Ctrl+C para detener el servidor")
	fmt.Println(strings.Repeat("=", 50))

	// IMPORTANTE: Usar el puerto de Railway
	serverAddr := ":" + port
	fmt.Printf("🎯 Escuchando en: %s\n", serverAddr)

	if err := e.Start(serverAddr); err != nil && err != http.ErrServerClosed {
		log.Fatal("❌ Error iniciando servidor:", err)
	}
}
