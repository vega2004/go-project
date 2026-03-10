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
	fmt.Println("=== INICIANDO SISTEMA VERSIÓN 2.0 CON MEJORAS ===")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	fmt.Printf("🔌 Puerto: %s\n", port)

	// ============================================
	// 1. CONEXIÓN A BASE DE DATOS
	// ============================================
	fmt.Println("\n📦 1. Conectando a PostgreSQL...")
	db, err := config.ConnectDB()
	if err != nil {
		log.Fatal("❌ Error conectando a la base de datos:", err)
	}
	defer db.Close()
	fmt.Println("✅ Conexión a DB exitosa")

	// ============================================
	// 2. INICIALIZAR ECHO
	// ============================================
	fmt.Println("\n🚀 2. Inicializando Echo...")
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
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodOptions},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization},
	}))
	e.Use(handler.ErrorHandler)
	fmt.Println("✅ Middlewares configurados")

	// ============================================
	// 3. CONFIGURAR TEMPLATES
	// ============================================
	fmt.Println("\n📄 3. Configurando templates...")

	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal("❌ Error obteniendo directorio actual:", err)
	}

	templatesDir := filepath.Join(cwd, "templates")
	fmt.Printf("   Buscando templates en: %s\n", templatesDir)

	if _, err := os.Stat(templatesDir); os.IsNotExist(err) {
		log.Fatalf("❌ Carpeta 'templates' no encontrada en: %s", templatesDir)
	}

	// Parsear todos los templates HTML
	templates, err := template.ParseGlob(filepath.Join(templatesDir, "*.html"))
	if err != nil {
		log.Fatal("❌ Error cargando templates:", err)
	}

	renderer := &TemplateRenderer{
		templates: templates,
	}
	e.Renderer = renderer

	fmt.Printf("   ✅ Templates cargados: %d\n", len(renderer.templates.Templates()))
	for i, t := range renderer.templates.Templates() {
		fmt.Printf("      %d. %s\n", i+1, t.Name())
	}

	// ============================================
	// 4. CONFIGURAR ARCHIVOS ESTÁTICOS
	// ============================================
	fmt.Println("\n🎨 4. Configurando archivos estáticos...")

	staticDir := filepath.Join(cwd, "static")
	fmt.Printf("   Sirviendo archivos estáticos desde: %s\n", staticDir)

	// Crear directorios necesarios
	dirsToCreate := []string{
		staticDir,
		filepath.Join(staticDir, "css"),
		filepath.Join(staticDir, "js"),
		filepath.Join(staticDir, "images"),
		filepath.Join(staticDir, "uploads"),
		filepath.Join(staticDir, "uploads", "carrusel"),
		filepath.Join(staticDir, "uploads", "perfil"),
		"logs",
	}

	for _, dir := range dirsToCreate {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			os.MkdirAll(dir, 0755)
			fmt.Printf("   📁 Creado directorio: %s\n", dir)
		}
	}

	// Servir archivos estáticos - SOLO ESTO ES SUFICIENTE
	e.Static("/static", staticDir)

	fmt.Println("✅ Archivos estáticos configurados")

	// ============================================
	// 5. MIDDLEWARE DE BREADCRUMBS
	// ============================================
	fmt.Println("\n🍞 5. Configurando middleware de breadcrumbs...")
	e.Use(handler.BreadcrumbMiddleware)
	fmt.Println("✅ Breadcrumbs configurados")

	// ============================================
	// 6. INICIALIZAR CAPAS DE LA APLICACIÓN
	// ============================================
	fmt.Println("\n🏗️ 6. Inicializando capas de aplicación...")

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

	fmt.Println("✅ Capas inicializadas correctamente")

	// ============================================
	// 7. CONFIGURAR RUTAS
	// ============================================
	fmt.Println("\n🛣️ 7. Configurando rutas...")

	// ============================================
	// 7.1 RUTAS PÚBLICAS (sin autenticación)
	// ============================================
	fmt.Println("   📍 Rutas públicas:")

	// Redirigir raíz a login
	e.GET("/", func(c echo.Context) error {
		return c.Redirect(http.StatusSeeOther, "/login")
	})
	fmt.Println("      - GET  / → /login")

	// Autenticación
	e.GET("/login", authHandler.ShowLogin)
	fmt.Println("      - GET  /login")
	e.POST("/do-login", authHandler.DoLogin)
	fmt.Println("      - POST /do-login")
	e.GET("/register", authHandler.ShowRegister)
	fmt.Println("      - GET  /register")
	e.POST("/do-register", authHandler.DoRegister)
	fmt.Println("      - POST /do-register")
	e.GET("/logout", authHandler.Logout)
	fmt.Println("      - GET  /logout")

	// Páginas públicas
	e.GET("/maintenance", authHandler.Maintenance)
	fmt.Println("      - GET  /maintenance")
	e.GET("/success", authHandler.ShowSuccess)
	fmt.Println("      - GET  /success")

	// Health check
	e.GET("/health", func(c echo.Context) error {
		err := db.Ping()
		if err != nil {
			return c.String(http.StatusServiceUnavailable, "❌ Database connection failed")
		}
		return c.String(http.StatusOK, "✅ OK")
	})
	fmt.Println("      - GET  /health")

	// ============================================
	// 7.2 RUTAS PROTEGIDAS (requieren autenticación)
	// ============================================
	fmt.Println("\n   🔒 Rutas protegidas:")
	protected := e.Group("")
	protected.Use(handler.AuthMiddleware(sessionManager, authRepo))

	// Dashboard
	protected.GET("/dashboard", dashboardHandler.ShowDashboard)
	fmt.Println("      - GET  /dashboard")

	// CRUD de Personas
	protected.GET("/crud", crudHandler.ShowCrud)
	fmt.Println("      - GET  /crud")
	protected.POST("/crud/create", crudHandler.Create)
	fmt.Println("      - POST /crud/create")
	protected.POST("/crud/update/:id", crudHandler.Update)
	fmt.Println("      - POST /crud/update/:id")
	protected.GET("/crud/delete/:id", crudHandler.Delete)
	fmt.Println("      - GET  /crud/delete/:id")
	protected.GET("/crud/list", crudHandler.List)
	fmt.Println("      - GET  /crud/list")
	protected.POST("/crud/filter", crudHandler.Filter)
	fmt.Println("      - POST /crud/filter")

	// Carrusel de Imágenes
	protected.GET("/carrusel", imagenHandler.ShowCarrusel)
	fmt.Println("      - GET  /carrusel")
	protected.POST("/carrusel/upload", imagenHandler.Upload)
	fmt.Println("      - POST /carrusel/upload")
	protected.GET("/carrusel/delete/:id", imagenHandler.Delete)
	fmt.Println("      - GET  /carrusel/delete/:id")
	protected.POST("/carrusel/reorder", imagenHandler.Reorder)
	fmt.Println("      - POST /carrusel/reorder")
	protected.GET("/carrusel/api/list", imagenHandler.GetCarruselJSON)
	fmt.Println("      - GET  /carrusel/api/list")

	// Perfil de Usuario
	protected.GET("/perfil", perfilHandler.ShowPerfil)
	fmt.Println("      - GET  /perfil")
	protected.POST("/perfil/update", perfilHandler.UpdatePerfil)
	fmt.Println("      - POST /perfil/update")
	protected.POST("/perfil/upload", perfilHandler.UploadFoto)
	fmt.Println("      - POST /perfil/upload")
	protected.POST("/perfil/change-password", perfilHandler.ChangePassword)
	fmt.Println("      - POST /perfil/change-password")
	protected.GET("/perfil/delete-foto", perfilHandler.DeleteFoto)
	fmt.Println("      - GET  /perfil/delete-foto")
	protected.GET("/perfil/json", perfilHandler.GetPerfilJSON)
	fmt.Println("      - GET  /perfil/json")

	// ============================================
	// 7.3 RUTAS DE ADMINISTRACIÓN
	// ============================================
	fmt.Println("\n   👑 Rutas de administración:")
	adminGroup := protected.Group("/admin")
	adminGroup.Use(handler.AdminMiddleware)

	adminGroup.GET("/users", adminHandler.ShowUsers)
	fmt.Println("      - GET  /admin/users")
	adminGroup.GET("/users/create", adminHandler.CreateUserForm)
	fmt.Println("      - GET  /admin/users/create")
	adminGroup.POST("/users/create", adminHandler.CreateUser)
	fmt.Println("      - POST /admin/users/create")
	adminGroup.GET("/users/edit/:id", adminHandler.EditUserForm)
	fmt.Println("      - GET  /admin/users/edit/:id")
	adminGroup.POST("/users/update/:id", adminHandler.UpdateUser)
	fmt.Println("      - POST /admin/users/update/:id")
	adminGroup.DELETE("/users/:id", adminHandler.DeleteUser)
	fmt.Println("      - DELETE /admin/users/:id")

	// ============================================
	// 7.4 RUTAS API (para Fetch API)
	// ============================================
	fmt.Println("\n   🔌 Rutas API:")
	apiGroup := protected.Group("/api")

	// API para carrusel
	apiGroup.GET("/carrusel", imagenHandler.GetCarruselJSON)
	fmt.Println("      - GET  /api/carrusel")

	// API para usuarios (admin)
	apiAdminGroup := apiGroup.Group("/admin")
	apiAdminGroup.Use(handler.AdminMiddleware)
	// Aquí irían las rutas API adicionales

	// ============================================
	// 8. MIDDLEWARE DE MANEJO DE ERRORES
	// ============================================
	fmt.Println("\n⚠️ 8. Configurando manejo de errores...")
	e.HTTPErrorHandler = handler.CustomHTTPErrorHandler
	fmt.Println("✅ Manejador de errores configurado")

	// ============================================
	// 9. VERIFICACIÓN FINAL
	// ============================================
	fmt.Println("\n🔍 9. Verificando configuración...")

	// Verificar conexión a BD
	if err := db.Ping(); err != nil {
		log.Fatal("❌ Error en conexión a BD:", err)
	}
	fmt.Println("   ✅ Conexión a BD verificada")

	// Verificar archivos JavaScript
	jsFiles := []string{"api.js", "carousel.js", "validations.js", "search-filters.js"}
	for _, jsFile := range jsFiles {
		jsPath := filepath.Join(staticDir, "js", jsFile)
		if _, err := os.Stat(jsPath); os.IsNotExist(err) {
			fmt.Printf("   ⚠️ Archivo JS no encontrado: %s\n", jsFile)
		} else {
			fmt.Printf("   ✅ Archivo JS encontrado: %s\n", jsFile)
		}
	}

	// ============================================
	// 10. INICIAR SERVIDOR
	// ============================================
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("🎯 SERVIDOR LISTO PARA INICIAR")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Printf("📌 Puerto: %s\n", port)
	if os.Getenv("RAILWAY_ENVIRONMENT") != "" {
		fmt.Println("🌍 Entorno: Railway (Producción)")
	} else {
		fmt.Println("💻 Entorno: Desarrollo Local")
	}
	fmt.Printf("📁 Directorio actual: %s\n", cwd)
	fmt.Printf("📊 Templates cargados: %d\n", len(renderer.templates.Templates()))
	fmt.Printf("🔐 Sesiones: Activas (duración: 24h)\n")
	fmt.Printf("🖼️  Uploads: /static/uploads/\n")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("⏳ Presiona Ctrl+C para detener el servidor")
	fmt.Println(strings.Repeat("=", 60))

	serverAddr := ":" + port
	fmt.Printf("\n🚀 Servidor escuchando en: http://localhost%s\n", serverAddr)
	fmt.Printf("📝 Logs disponibles en: %s\n", filepath.Join(cwd, "logs"))

	// Iniciar servidor
	if err := e.Start(serverAddr); err != nil && err != http.ErrServerClosed {
		log.Fatal("❌ Error iniciando servidor:", err)
	}
}
