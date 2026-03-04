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
	fmt.Println("=== INICIANDO SISTEMA VERSIÓN 2.0 ===")

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
	e.Use(handler.ErrorHandler)
	fmt.Println(" Middlewares configurados")

	// *** CONFIGURAR TEMPLATE RENDERER ***
	fmt.Println("\n3. Configurando templates...")

	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal(" Error obteniendo directorio actual:", err)
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
	perfilRepo := repository.NewPerfilRepository(db)

	// Servicios
	userService := service.NewUserService(userRepo, authRepo)
	authService := service.NewAuthService(authRepo)
	crudService := service.NewCrudService(crudRepo)
	imagenService := service.NewImagenService(imagenRepo)
	perfilService := service.NewPerfilService(perfilRepo, authRepo)

	// Session Manager
	sessionManager := utils.NewSessionManager()

	// Handlers
	authHandler := handler.NewAuthHandler(authService, sessionManager)
	dashboardHandler := handler.NewDashboardHandler()
	crudHandler := handler.NewCrudHandler(crudService)
	imagenHandler := handler.NewImagenHandler(imagenService)
	adminHandler := handler.NewAdminHandler(userService)
	perfilHandler := handler.NewPerfilHandler(perfilService, sessionManager)

	fmt.Println(" Capas inicializadas")

	// *** RUTAS PÚBLICAS (sin autenticación) ***
	fmt.Println("\n6. Configurando rutas...")

	// Redirigir raíz a login
	e.GET("/", func(c echo.Context) error {
		return c.Redirect(http.StatusSeeOther, "/login")
	})

	// Autenticación
	e.GET("/login", authHandler.ShowLogin)
	e.POST("/do-login", authHandler.DoLogin)
	e.GET("/register", authHandler.ShowRegister)
	e.POST("/do-register", authHandler.DoRegister)
	e.GET("/logout", authHandler.Logout)

	// Páginas públicas
	e.GET("/maintenance", authHandler.Maintenance)
	e.GET("/success", authHandler.ShowSuccess)

	// *** RUTAS PROTEGIDAS (requieren autenticación) ***
	protected := e.Group("")
	protected.Use(handler.AuthMiddleware(sessionManager, authRepo))

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

	// Perfil de Usuario
	protected.GET("/perfil", perfilHandler.ShowPerfil)
	protected.POST("/perfil/update", perfilHandler.UpdatePerfil)
	protected.POST("/perfil/upload", perfilHandler.UploadFoto)
	protected.POST("/perfil/change-password", perfilHandler.ChangePassword)
	protected.GET("/perfil/delete-foto", perfilHandler.DeleteFoto)
	protected.GET("/perfil/json", perfilHandler.GetPerfilJSON)

	// *** RUTAS DE ADMIN ***
	adminGroup := protected.Group("/admin")
	adminGroup.Use(handler.AdminMiddleware)

	adminGroup.GET("/users", adminHandler.ShowUsers)
	adminGroup.GET("/users/create", adminHandler.CreateUserForm)
	adminGroup.POST("/users/create", adminHandler.CreateUser)
	adminGroup.GET("/users/edit/:id", adminHandler.EditUserForm)
	adminGroup.POST("/users/update/:id", adminHandler.UpdateUser)
	adminGroup.DELETE("/users/:id", adminHandler.DeleteUser)

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
