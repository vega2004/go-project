package main

import (
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"tu-proyecto/internal/config"
	"tu-proyecto/internal/handlers"
	"tu-proyecto/internal/middleware"
	"tu-proyecto/internal/repository"
	"tu-proyecto/internal/service"
	"tu-proyecto/internal/utils"

	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"
)

type TemplateRenderer struct {
	templates *template.Template
}

func (t *TemplateRenderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	tmpl := t.templates.Lookup(name)
	if tmpl == nil {
		return fmt.Errorf("template '%s' no encontrado", name)
	}
	return tmpl.Execute(w, data)
}

func main() {
	// Configuración
	env := config.LoadEnv()
	log.Printf("🚀 Iniciando %s en modo %s", env.AppName, env.AppEnv)

	// Base de datos
	db, err := config.NewDB(env)
	if err != nil {
		log.Fatal("❌ Error conectando a BD:", err)
	}
	defer db.Close()
	log.Println("✅ Conexión a BD establecida")

	// Echo
	e := echo.New()
	e.HideBanner = true

	// Middlewares globales
	e.Use(echomiddleware.Logger())
	e.Use(echomiddleware.Recover())
	e.Use(echomiddleware.Secure())
	e.Use(echomiddleware.Gzip())
	e.Use(echomiddleware.CORSWithConfig(echomiddleware.CORSConfig{
		AllowOrigins: []string{env.BaseURL, "http://localhost:3000", "http://localhost:8080"},
		AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodOptions},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization, "X-CSRF-Token"},
	}))

	// Rate limiting
	e.Use(echomiddleware.RateLimiter(echomiddleware.NewRateLimiterMemoryStore(20)))

	// Templates
	cwd, _ := os.Getwd()
	templatesDir := filepath.Join(cwd, "web", "templates")
	templates, err := template.ParseGlob(filepath.Join(templatesDir, "**/*.html"))
	if err != nil {
		log.Fatal("❌ Error cargando templates:", err)
	}
	e.Renderer = &TemplateRenderer{templates: templates}

	// Archivos estáticos
	e.Static("/static", filepath.Join(cwd, "web", "static"))

	// Repositorios
	authRepo := repository.NewAuthRepository(db)
	perfilRepo := repository.NewPerfilRepository(db)
	moduloRepo := repository.NewModuloRepository(db)
	permisoRepo := repository.NewPermisoRepository(db)
	userRepo := repository.NewUserRepository(db)

	// Servicios
	authService := service.NewAuthService(authRepo, env.RecaptchaSecretKey)
	perfilService := service.NewPerfilService(perfilRepo, authRepo, userRepo)
	moduloService := service.NewModuloService(moduloRepo)
	permisoService := service.NewPermisoService(permisoRepo, perfilRepo, moduloRepo)
	userService := service.NewUserService(userRepo)

	// Session Manager
	sessionManager := utils.NewSessionManager(env.SessionSecret, env.IsProduction())

	// ============================================
	// NUEVO: JWT MANAGER
	// ============================================
	jwtManager := utils.NewJWTManager(env.SessionSecret, 24) // 24 horas de expiración

	// ============================================
	// MIDDLEWARES PERSONALIZADOS (ACTUALIZADOS)
	// ============================================
	// AuthMiddleware con soporte JWT
	authMiddleware := middleware.NewAuthMiddleware(
		sessionManager,
		jwtManager, // ← NUEVO: pasar JWT manager
		authRepo,
		permisoRepo,
		true, // ← HABILITAR JWT (true) o DESHABILITAR (false)
	)

	rbacMiddleware := middleware.NewRBACMiddleware(permisoRepo)
	errorMiddleware := middleware.NewErrorMiddleware(env.IsDevelopment(), "")
	breadcrumbsMiddleware := middleware.NewBreadcrumbMiddleware(db)
	csrfMiddleware := middleware.NewCSRFMiddleware(env.IsProduction())

	// Configurar manejador de errores
	e.HTTPErrorHandler = errorMiddleware.ErrorHandler
	e.Use(errorMiddleware.Recover)
	e.RouteNotFound("/*", errorMiddleware.NotFound)

	// ============================================
	// HANDLERS (ACTUALIZADOS)
	// ============================================
	authHandler := handlers.NewAuthHandler(authService, sessionManager, jwtManager, env) // ← NUEVO: pasar jwtManager
	dashboardHandler := handlers.NewDashboardHandler(db)
	perfilHandler := handlers.NewPerfilHandler(perfilService)
	moduloHandler := handlers.NewModuloHandler(moduloService)
	permisoHandler := handlers.NewPermisoHandler(permisoService, perfilService)
	userHandler := handlers.NewUserHandler(userService, perfilService)
	principalHandler := handlers.NewPrincipalHandler(permisoService)

	// ============================================
	// RUTAS PÚBLICAS
	// ============================================
	e.GET("/", func(c echo.Context) error {
		return c.Redirect(http.StatusSeeOther, "/login")
	})
	e.GET("/login", authHandler.ShowLogin)
	e.POST("/do-login", authHandler.DoLogin, csrfMiddleware.Protect)
	e.GET("/register", authHandler.ShowRegister)
	e.POST("/do-register", authHandler.DoRegister, csrfMiddleware.Protect)
	e.GET("/logout", authHandler.Logout)
	e.GET("/maintenance", authHandler.Maintenance)
	e.GET("/success", authHandler.Success)
	e.GET("/health", func(c echo.Context) error {
		if err := db.Ping(); err != nil {
			return c.String(http.StatusServiceUnavailable, "❌ DB Error")
		}
		return c.String(http.StatusOK, "✅ OK")
	})

	// ============================================
	// RUTAS PROTEGIDAS
	// ============================================
	protected := e.Group("")
	protected.Use(authMiddleware.RequireAuth)
	protected.Use(breadcrumbsMiddleware.Handler)

	// Dashboard
	protected.GET("/dashboard", dashboardHandler.ShowDashboard)

	// Perfil
	protected.GET("/perfil", perfilHandler.ShowPerfil)
	protected.POST("/perfil/update", perfilHandler.UpdatePerfil, csrfMiddleware.Protect)
	protected.POST("/perfil/upload", perfilHandler.UploadFoto, csrfMiddleware.Protect)
	protected.POST("/perfil/change-password", perfilHandler.ChangePassword, csrfMiddleware.Protect)
	protected.GET("/perfil/delete-foto", perfilHandler.DeleteFoto)

	// API Permisos
	protected.GET("/api/permisos", rbacMiddleware.GetUserPermissions)

	// ============================================
	// RUTAS API CON JWT (Ejemplo)
	// ============================================
	api := protected.Group("/api")
	api.GET("/perfil", perfilHandler.ShowPerfilJSON) // Endpoint que devuelve JSON

	// ============================================
	// MÓDULO: PERFILES
	// ============================================
	pg := protected.Group("/seguridad/perfiles")
	pg.Use(rbacMiddleware.CheckPermission("perfiles", "ver"))
	pg.GET("", perfilHandler.Index)
	pg.GET("/nuevo", perfilHandler.CreateForm, rbacMiddleware.CheckPermission("perfiles", "crear"))
	pg.POST("/crear", perfilHandler.Create, rbacMiddleware.CheckPermission("perfiles", "crear"), csrfMiddleware.Protect)
	pg.GET("/editar/:id", perfilHandler.EditForm, rbacMiddleware.CheckPermission("perfiles", "editar"))
	pg.POST("/actualizar/:id", perfilHandler.Update, rbacMiddleware.CheckPermission("perfiles", "editar"), csrfMiddleware.Protect)
	pg.DELETE("/eliminar/:id", perfilHandler.Delete, rbacMiddleware.CheckPermission("perfiles", "eliminar"), csrfMiddleware.Protect)

	// ============================================
	// MÓDULO: MÓDULOS
	// ============================================
	mg := protected.Group("/seguridad/modulos")
	mg.Use(rbacMiddleware.CheckPermission("modulos", "ver"))
	mg.GET("", moduloHandler.Index)
	mg.GET("/nuevo", moduloHandler.CreateForm, rbacMiddleware.CheckPermission("modulos", "crear"))
	mg.POST("/crear", moduloHandler.Create, rbacMiddleware.CheckPermission("modulos", "crear"), csrfMiddleware.Protect)
	mg.GET("/editar/:id", moduloHandler.EditForm, rbacMiddleware.CheckPermission("modulos", "editar"))
	mg.POST("/actualizar/:id", moduloHandler.Update, rbacMiddleware.CheckPermission("modulos", "editar"), csrfMiddleware.Protect)
	mg.DELETE("/eliminar/:id", moduloHandler.Delete, rbacMiddleware.CheckPermission("modulos", "eliminar"), csrfMiddleware.Protect)

	// ============================================
	// MÓDULO: PERMISOS-PERFIL
	// ============================================
	permGroup := protected.Group("/seguridad/permisos-perfil")
	permGroup.Use(rbacMiddleware.RequireAdmin)
	permGroup.GET("", permisoHandler.Index)
	permGroup.POST("/cargar", permisoHandler.LoadPermissions)
	permGroup.POST("/guardar", permisoHandler.SavePermissions, csrfMiddleware.Protect)

	// ============================================
	// MÓDULO: USUARIOS
	// ============================================
	ug := protected.Group("/seguridad/usuarios")
	ug.Use(rbacMiddleware.CheckPermission("usuarios", "ver"))
	ug.GET("", userHandler.Index)
	ug.GET("/nuevo", userHandler.CreateForm, rbacMiddleware.CheckPermission("usuarios", "crear"))
	ug.POST("/crear", userHandler.Create, rbacMiddleware.CheckPermission("usuarios", "crear"), csrfMiddleware.Protect)
	ug.GET("/editar/:id", userHandler.EditForm, rbacMiddleware.CheckPermission("usuarios", "editar"))
	ug.POST("/actualizar/:id", userHandler.Update, rbacMiddleware.CheckPermission("usuarios", "editar"), csrfMiddleware.Protect)
	ug.DELETE("/eliminar/:id", userHandler.Delete, rbacMiddleware.CheckPermission("usuarios", "eliminar"), csrfMiddleware.Protect)
	ug.GET("/detalle/:id", userHandler.Detail, rbacMiddleware.CheckPermission("usuarios", "detalle"))
	ug.POST("/toggle-status/:id", userHandler.ToggleStatus, rbacMiddleware.CheckPermission("usuarios", "editar"), csrfMiddleware.Protect)

	// ============================================
	// MÓDULOS PRINCIPALES
	// ============================================
	protected.GET("/principal/clientes", principalHandler.Principal11, rbacMiddleware.RequireModuleAccess("principal11"))
	protected.GET("/principal/productos", principalHandler.Principal12, rbacMiddleware.RequireModuleAccess("principal12"))
	protected.GET("/principal/facturas", principalHandler.Principal21, rbacMiddleware.RequireModuleAccess("principal21"))
	protected.GET("/principal/proveedores", principalHandler.Principal22, rbacMiddleware.RequireModuleAccess("principal22"))

	// ============================================
	// INICIAR SERVIDOR
	// ============================================
	port := env.AppPort
	log.Printf("🚀 Servidor iniciado en %s:%s", env.BaseURL, port)
	log.Printf("📊 Entorno: %s", env.AppEnv)
	log.Printf("🔐 CSRF: activado")
	log.Printf("🔑 JWT: activado (expiración: 24h)")
	log.Printf("⏱️ Rate Limit: 20 peticiones por segundo")

	if err := e.Start(":" + port); err != nil && err != http.ErrServerClosed {
		log.Fatal("❌ Error iniciando servidor:", err)
	}
}
