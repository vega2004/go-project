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
	"tu-proyecto/utils"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// TemplateRenderer para usar templates HTML
type TemplateRenderer struct {
	templates *template.Template
}

// Render implementa la interfaz de Echo para renderizar templates
func (t *TemplateRenderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	if os.Getenv("RAILWAY_ENVIRONMENT") == "" {
		fmt.Printf("\n [RENDERER] Renderizando template: '%s'\n", name)
	}

	tmpl := t.templates.Lookup(name)
	if tmpl == nil {
		fmt.Printf(" ERROR: Template '%s' no encontrado!\n", name)
		return fmt.Errorf("template '%s' not found", name)
	}

	buf := new(strings.Builder)
	err := tmpl.Execute(buf, data)
	if err != nil {
		fmt.Printf(" ERROR ejecutando template: %v\n", err)
		return err
	}

	_, err = w.Write([]byte(buf.String()))
	return err
}

func main() {
	fmt.Println("=== INICIANDO SISTEMA ===")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	fmt.Printf(" Puerto: %s\n", port)

	fmt.Println("1. Conectando a PostgreSQL...")
	db, err := config.ConnectDB()
	if err != nil {
		log.Fatal(" Error conectando a la base de datos:", err)
	}
	defer db.Close()
	fmt.Println(" Conexión a DB exitosa")

	fmt.Println("\n2. Inicializando Echo...")
	e := echo.New()

	// Middlewares globales
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: `${time_rfc3339} ${status} ${method} ${host}${path} ${latency_human}` + "\n",
	}))
	e.Use(middleware.RecoverWithConfig(middleware.RecoverConfig{
		StackSize: 1 << 10,
	}))
	e.Use(middleware.Secure())
	e.Use(middleware.Gzip())
	e.Use(handler.ErrorHandler) // Middleware de errores global
	fmt.Println(" Middlewares configurados")

	// *** CONFIGURAR TEMPLATE RENDERER ***
	fmt.Println("\n3. Configurando templates...")

	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal(" Error obteniendo directorio actual:", err)
	}

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

	if _, err := os.Stat(templatesDir); os.IsNotExist(err) {
		log.Fatalf(" Carpeta 'templates' no encontrada en: %s", templatesDir)
	}

	templates, err := template.ParseGlob(filepath.Join(templatesDir, "*.html"))
	if err != nil {
		log.Fatal(" Error cargando templates:", err)
	}

	renderer := &TemplateRenderer{
		templates: templates,
	}
	e.Renderer = renderer

	fmt.Printf("    Templates cargados: %d\n", len(renderer.templates.Templates()))
	for i, t := range renderer.templates.Templates() {
		fmt.Printf("   %d. %s\n", i+1, t.Name())
	}

	staticDir := filepath.Join(cwd, "static")
	fmt.Printf("    Serviendo archivos estáticos desde: %s\n", staticDir)

	if _, err := os.Stat(staticDir); os.IsNotExist(err) {
		fmt.Printf("     Carpeta 'static' no encontrada, creando...\n")
		os.MkdirAll(staticDir, 0755)
	}

	e.Static("/static", staticDir)
	fmt.Println(" Archivos estáticos configurados")

	// *** MIDDLEWARE DE BREADCRUMBS ***
	fmt.Println("\n4. Configurando middleware de breadcrumbs...")
	e.Use(handler.BreadcrumbMiddleware)

	// *** INICIALIZAR CAPAS ***
	fmt.Println("\n5. Inicializando capas de aplicación...")

	// Repositorios
	userRepo := repository.NewUserRepository(db)
	authRepo := repository.NewAuthRepository(db)
	crudRepo := repository.NewCrudRepository(db)
	imagenRepo := repository.NewImagenRepository(db)
	perfilRepo := repository.NewPerfilRepository(db) // ← NUEVO

	// Servicios
	userService := service.NewUserService(userRepo, authRepo)
	authService := service.NewAuthService(authRepo)
	crudService := service.NewCrudService(crudRepo)
	imagenService := service.NewImagenService(imagenRepo)
	perfilService := service.NewPerfilService(perfilRepo, authRepo) // ← NUEVO

	// Session Manager
	sessionManager := utils.NewSessionManager()

	// Handlers
	userHandler := handler.NewUserHandler(userService)
	authHandler := handler.NewAuthHandler(authService, sessionManager)
	dashboardHandler := handler.NewDashboardHandler()
	crudHandler := handler.NewCrudHandler(crudService)
	imagenHandler := handler.NewImagenHandler(imagenService)
	adminHandler := handler.NewAdminHandler(userService)
	perfilHandler := handler.NewPerfilHandler(perfilService, sessionManager) // ← NUEVO

	fmt.Println(" Capas inicializadas")

	// *** RUTAS PÚBLICAS (sin autenticación) ***
	fmt.Println("\n6. Configurando rutas...")

	e.GET("/", userHandler.Welcome)
	e.GET("/form", userHandler.ShowForm)
	e.POST("/submit", userHandler.SubmitForm)
	e.GET("/success", userHandler.ShowSuccess)
	e.GET("/maintenance", userHandler.Maintenance)

	e.GET("/login", authHandler.ShowLogin)
	e.POST("/do-login", authHandler.DoLogin)
	e.GET("/register", authHandler.ShowRegister)
	e.POST("/do-register", authHandler.DoRegister)
	e.GET("/logout", authHandler.Logout)

	// *** RUTAS PROTEGIDAS (requieren autenticación) ***
	protected := e.Group("")
	protected.Use(handler.AuthMiddleware(sessionManager))

	// Dashboard
	protected.GET("/dashboard", dashboardHandler.ShowDashboard)

	// CRUD de Personas
	protected.GET("/crud", crudHandler.ShowCrud)
	protected.POST("/crud/create", crudHandler.Create)
	protected.POST("/crud/update/:id", crudHandler.Update)
	protected.GET("/crud/delete/:id", crudHandler.Delete)
	protected.GET("/crud/list", crudHandler.List)
	protected.POST("/crud/filter", crudHandler.Filter)

	// Carrusel de Imágenes
	protected.GET("/carrusel", imagenHandler.ShowCarrusel)
	protected.POST("/carrusel/upload", imagenHandler.Upload)
	protected.GET("/carrusel/delete/:id", imagenHandler.Delete)
	protected.POST("/carrusel/reorder", imagenHandler.Reorder)
	protected.GET("/carrusel/api/list", imagenHandler.GetCarruselJSON)

	// *** PERFIL DE USUARIO ***
	protected.GET("/perfil", perfilHandler.ShowPerfil)
	protected.POST("/perfil/update", perfilHandler.UpdatePerfil)
	protected.POST("/perfil/upload", perfilHandler.UploadFoto)
	protected.POST("/perfil/change-password", perfilHandler.ChangePassword)
	protected.GET("/perfil/delete-foto", perfilHandler.DeleteFoto) // Opcional
	protected.GET("/perfil/json", perfilHandler.GetPerfilJSON)     // Para AJAX

	// *** RUTAS DE ADMIN (requieren rol de administrador) ***
	adminGroup := protected.Group("/admin")
	adminGroup.Use(handler.AdminMiddleware) // Middleware que verifica role_id=1

	adminGroup.GET("/users", adminHandler.ShowUsers)
	adminGroup.GET("/users/create", adminHandler.CreateUserForm)
	adminGroup.POST("/users/create", adminHandler.CreateUser)
	adminGroup.GET("/users/edit/:id", adminHandler.EditUserForm)
	adminGroup.POST("/users/update/:id", adminHandler.UpdateUser)
	adminGroup.DELETE("/users/:id", adminHandler.DeleteUser)

	// *** RUTA DE DEBUG (solo desarrollo) ***
	if os.Getenv("RAILWAY_ENVIRONMENT") == "" {
		protected.GET("/debug", func(c echo.Context) error {
			userID := c.Get("user_id")
			userName := c.Get("user_name")
			userRole := c.Get("user_role")

			html := "<h1>Debug Info (Autenticado)</h1><pre>"
			html += fmt.Sprintf("User ID: %v\n", userID)
			html += fmt.Sprintf("User Name: %v\n", userName)
			html += fmt.Sprintf("User Role: %v\n", userRole)
			html += fmt.Sprintf("Entorno: %s\n", os.Getenv("RAILWAY_ENVIRONMENT"))
			html += fmt.Sprintf("Puerto: %s\n", port)
			html += fmt.Sprintf("Templates cargados: %d\n", len(renderer.templates.Templates()))
			for i, t := range renderer.templates.Templates() {
				html += fmt.Sprintf("%d. %s\n", i+1, t.Name())
			}
			html += "</pre>"
			html += `<p><a href="/dashboard">Dashboard</a> | <a href="/logout">Cerrar Sesión</a> | <a href="/perfil">Perfil</a></p>`
			return c.HTML(http.StatusOK, html)
		})
	}

	fmt.Println(" Rutas configuradas")

	// *** HEALTH CHECK ***
	e.GET("/health", func(c echo.Context) error {
		err := db.Ping()
		if err != nil {
			return c.String(http.StatusServiceUnavailable, "Database connection failed")
		}
		return c.String(http.StatusOK, "OK")
	})

	// *** INICIAR SERVIDOR ***
	fmt.Println("\n" + strings.Repeat("=", 50))
	fmt.Printf(" Servidor iniciado en puerto %s\n", port)
	if os.Getenv("RAILWAY_ENVIRONMENT") != "" {
		fmt.Println(" Entorno: Railway (Producción)")
	} else {
		fmt.Println(" Entorno: Desarrollo Local")
	}
	fmt.Println(" Presiona Ctrl+C para detener el servidor")
	fmt.Println(strings.Repeat("=", 50))

	serverAddr := ":" + port
	fmt.Printf(" Escuchando en: %s\n", serverAddr)

	if err := e.Start(serverAddr); err != nil && err != http.ErrServerClosed {
		log.Fatal(" Error iniciando servidor:", err)
	}
}
