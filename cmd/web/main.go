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

	"tu-proyecto/internal/config"
	"tu-proyecto/internal/handlers"
	"tu-proyecto/internal/middleware"
	"tu-proyecto/internal/repository"
	"tu-proyecto/internal/service"
	"tu-proyecto/internal/utils"

	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"
)

// ============================================
// FUNCIONES PERSONALIZADAS PARA TEMPLATES
// ============================================

var templateFuncs = template.FuncMap{
	"add": func(a, b int) int {
		return a + b
	},
	"sub": func(a, b int) int {
		return a - b
	},
	"mul": func(a, b int) int {
		return a * b
	},
	"div": func(a, b int) int {
		if b == 0 {
			return 0
		}
		return a / b
	},
	"toString": func(v interface{}) string {
		if v == nil {
			return ""
		}
		return fmt.Sprintf("%v", v)
	},
	"default": func(defaultValue string, v interface{}) string {
		if v == nil {
			return defaultValue
		}
		s := fmt.Sprintf("%v", v)
		if s == "" {
			return defaultValue
		}
		return s
	},
	// ✅ AGREGA ESTA FUNCIÓN:
	"iterate": func(count int) []int {
		var result []int
		for i := 0; i < count; i++ {
			result = append(result, i)
		}
		return result
	},
}

type TemplateRenderer struct {
	templates *template.Template
}

func (t *TemplateRenderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	log.Printf("[RENDER] Iniciando renderizado del template: %s", name)

	// Verificar si el template existe
	tmpl := t.templates.Lookup(name)
	if tmpl == nil {
		log.Printf("[RENDER ERROR] Template '%s' no encontrado", name)
		log.Printf("[RENDER ERROR] Templates disponibles:")
		for _, t := range t.templates.Templates() {
			if t.Name() != "" && t.Name() != " " {
				log.Printf("  - %s", t.Name())
			}
		}
		return fmt.Errorf("template '%s' no encontrado", name)
	}

	log.Printf("[RENDER] Template '%s' encontrado, ejecutando...", name)
	err := tmpl.Execute(w, data)
	if err != nil {
		log.Printf("[RENDER ERROR] Error ejecutando template '%s': %v", name, err)
		return err
	}

	log.Printf("[RENDER] Template '%s' renderizado exitosamente", name)
	return nil
}

func main() {
	log.Println("========================================")
	log.Println("INICIANDO APLICACIÓN")
	log.Println("========================================")

	// Configuración
	env := config.LoadEnv()
	log.Printf("🚀 Iniciando %s en modo %s", env.AppName, env.AppEnv)
	log.Printf("📁 Directorio actual: %s", mustGetWd())
	log.Printf("🌐 Base URL: %s", env.BaseURL)
	log.Printf("🔌 Puerto: %s", env.AppPort)

	// Base de datos
	log.Println("[DB] Conectando a base de datos...")
	db, err := config.NewDB(env)
	if err != nil {
		log.Fatal("❌ Error conectando a BD:", err)
	}
	defer db.Close()
	log.Println("[DB] ✅ Conexión a BD establecida")

	// Echo
	e := echo.New()
	e.HideBanner = true

	// Middlewares globales
	log.Println("[MIDDLEWARE] Configurando middlewares globales...")
	e.Use(echomiddleware.Logger())
	e.Use(echomiddleware.Recover())
	e.Use(echomiddleware.Secure())
	e.Use(echomiddleware.Gzip())
	e.Use(echomiddleware.CORSWithConfig(echomiddleware.CORSConfig{
		AllowOrigins: []string{env.BaseURL, "http://localhost:3000", "http://localhost:8080"},
		AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodOptions},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization, "X-CSRF-Token"},
	}))
	e.Use(echomiddleware.RateLimiter(echomiddleware.NewRateLimiterMemoryStore(20)))

	// ============================================
	// TEMPLATES CON HERENCIA CORREGIDA
	// ============================================
	log.Println("[TEMPLATES] Configurando motor de plantillas con herencia...")

	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal("❌ Error obteniendo directorio actual:", err)
	}
	log.Printf("[TEMPLATES] Directorio actual: %s", cwd)

	templatesDir := filepath.Join(cwd, "web", "templates")
	log.Printf("[TEMPLATES] Buscando templates en: %s", templatesDir)

	if _, err := os.Stat(templatesDir); os.IsNotExist(err) {
		log.Fatalf("[TEMPLATES] ❌ Directorio de templates no encontrado: %s", templatesDir)
	}
	log.Printf("[TEMPLATES] ✅ Directorio de templates encontrado")

	// Crear un nuevo template con funciones personalizadas
	tmpl := template.New("").Funcs(templateFuncs)
	log.Println("[TEMPLATES] Funciones personalizadas registradas")

	// Cargar TODOS los templates .html y mantener sus nombres para herencia
	templateCount := 0
	err = filepath.Walk(templatesDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// Solo procesar archivos .html
		if !info.IsDir() && strings.HasSuffix(path, ".html") {
			// Obtener el nombre del template relativo a la carpeta templates
			relPath, err := filepath.Rel(templatesDir, path)
			if err != nil {
				return err
			}
			// Usar la ruta relativa como nombre del template
			templateName := filepath.ToSlash(relPath)

			// Leer el contenido del archivo
			content, err := os.ReadFile(path)
			if err != nil {
				return fmt.Errorf("error leyendo %s: %w", path, err)
			}

			// Parsear el template con su nombre
			_, err = tmpl.New(templateName).Parse(string(content))
			if err != nil {
				return fmt.Errorf("error parseando %s: %w", templateName, err)
			}
			templateCount++
			log.Printf("[TEMPLATES] Cargado: %s", templateName)
		}
		return nil
	})

	if err != nil {
		log.Fatal("❌ Error cargando templates:", err)
	}

	log.Printf("[TEMPLATES] ✅ Total templates cargados: %d", templateCount)
	log.Println("[TEMPLATES] Templates disponibles para herencia:")
	for _, t := range tmpl.Templates() {
		if t.Name() != "" && t.Name() != " " {
			log.Printf("  - %s", t.Name())
		}
	}

	// Configurar el renderer
	e.Renderer = &TemplateRenderer{templates: tmpl}
	log.Println("[TEMPLATES] ✅ Renderer configurado correctamente")

	// Archivos estáticos
	staticDir := filepath.Join(cwd, "web", "static")
	log.Printf("[STATIC] Sirviendo archivos estáticos desde: %s", staticDir)
	e.Static("/static", staticDir)

	// Repositorios
	log.Println("[REPO] Inicializando repositorios...")
	authRepo := repository.NewAuthRepository(db)
	perfilRepo := repository.NewPerfilRepository(db)
	moduloRepo := repository.NewModuloRepository(db)
	permisoRepo := repository.NewPermisoRepository(db)
	userRepo := repository.NewUserRepository(db)
	log.Println("[REPO] ✅ Repositorios inicializados")

	// Servicios
	log.Println("[SERVICE] Inicializando servicios...")
	authService := service.NewAuthService(authRepo, env.RecaptchaSecretKey)
	perfilService := service.NewPerfilService(perfilRepo, authRepo, userRepo)
	moduloService := service.NewModuloService(moduloRepo)
	permisoService := service.NewPermisoService(permisoRepo, perfilRepo, moduloRepo)
	userService := service.NewUserService(userRepo)
	log.Println("[SERVICE] ✅ Servicios inicializados")

	// Session Manager
	log.Println("[SESSION] Inicializando session manager...")
	sessionManager := utils.NewSessionManager(env.SessionSecret, env.IsProduction())
	log.Println("[SESSION] ✅ Session manager inicializado")

	// JWT Manager
	log.Println("[JWT] Inicializando JWT manager...")
	jwtManager := utils.NewJWTManager(env.SessionSecret, 24)
	log.Println("[JWT] ✅ JWT manager inicializado")

	// Middlewares personalizados
	log.Println("[MIDDLEWARE] Inicializando middlewares personalizados...")
	authMiddleware := middleware.NewAuthMiddleware(
		sessionManager,
		jwtManager,
		authRepo,
		permisoRepo,
		true,
	)
	rbacMiddleware := middleware.NewRBACMiddleware(permisoRepo)
	errorMiddleware := middleware.NewErrorMiddleware(env.IsDevelopment(), "")
	breadcrumbsMiddleware := middleware.NewBreadcrumbMiddleware(db)
	csrfMiddleware := middleware.NewCSRFMiddleware(env.IsProduction())
	log.Println("[MIDDLEWARE] ✅ Middlewares personalizados inicializados")

	// Configurar manejador de errores
	e.HTTPErrorHandler = errorMiddleware.ErrorHandler
	e.Use(errorMiddleware.Recover)
	e.RouteNotFound("/*", errorMiddleware.NotFound)

	// Handlers
	log.Println("[HANDLERS] Inicializando handlers...")
	authHandler := handlers.NewAuthHandler(authService, sessionManager, jwtManager, env)
	dashboardHandler := handlers.NewDashboardHandler(db)
	perfilHandler := handlers.NewPerfilHandler(perfilService, env) // ← Agregar env
	moduloHandler := handlers.NewModuloHandler(moduloService, env) // ← Agregar
	permisoHandler := handlers.NewPermisoHandler(permisoService, perfilService, env)
	userHandler := handlers.NewUserHandler(userService, perfilService, env) // ← Agregar env
	principalHandler := handlers.NewPrincipalHandler(permisoService)
	log.Println("[HANDLERS] ✅ Handlers inicializados")

	// ============================================
	// ============================================
	// RUTAS PÚBLICAS
	// ============================================
	log.Println("[ROUTES] Configurando rutas públicas...")

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

	// ✅ Rutas para breadcrumbs - Redirigen a mantenimiento
	e.GET("/seguridad", func(c echo.Context) error {
		return c.Redirect(http.StatusSeeOther, "/maintenance?from=seguridad")
	})

	e.GET("/principal", func(c echo.Context) error {
		return c.Redirect(http.StatusSeeOther, "/maintenance?from=principal")
	})

	e.GET("/health", func(c echo.Context) error {
		if err := db.Ping(); err != nil {
			return c.String(http.StatusServiceUnavailable, "❌ DB Error")
		}
		return c.String(http.StatusOK, "✅ OK")
	})

	log.Println("[ROUTES] ✅ Rutas públicas configuradas")

	// ============================================
	// RUTAS PROTEGIDAS
	// ============================================
	log.Println("[ROUTES] Configurando rutas protegidas...")
	protected := e.Group("")
	protected.Use(authMiddleware.RequireAuth)
	protected.Use(breadcrumbsMiddleware.Handler)

	// Dashboard
	protected.GET("/dashboard", dashboardHandler.ShowDashboard)
	log.Println("[ROUTES]   - GET /dashboard")

	// Perfil
	protected.GET("/perfil", perfilHandler.ShowPerfil)
	protected.POST("/perfil/update", perfilHandler.UpdatePerfil, csrfMiddleware.Protect)
	protected.POST("/perfil/upload", perfilHandler.UploadFoto, csrfMiddleware.Protect)
	protected.POST("/perfil/change-password", perfilHandler.ChangePassword, csrfMiddleware.Protect)
	protected.GET("/perfil/delete-foto", perfilHandler.DeleteFoto)
	log.Println("[ROUTES]   - Perfil routes")

	// API Permisos
	protected.GET("/api/permisos", rbacMiddleware.GetUserPermissions)
	log.Println("[ROUTES]   - GET /api/permisos")

	// API con JWT
	api := protected.Group("/api")
	api.GET("/perfil", perfilHandler.ShowPerfilJSON)
	log.Println("[ROUTES]   - API routes")

	// MÓDULO: PERFILES - Cambiar "perfiles" por "Perfiles"
	pg := protected.Group("/seguridad/perfiles")
	pg.Use(rbacMiddleware.CheckPermission("Perfiles", "ver")) // ← Cambiado
	pg.GET("", perfilHandler.Index)
	pg.GET("/nuevo", perfilHandler.CreateForm, rbacMiddleware.CheckPermission("Perfiles", "crear")) // ← Cambiado
	pg.POST("/crear", perfilHandler.Create, rbacMiddleware.CheckPermission("Perfiles", "crear"), csrfMiddleware.Protect)
	pg.GET("/editar/:id", perfilHandler.EditForm, rbacMiddleware.CheckPermission("Perfiles", "editar")) // ← Cambiado
	pg.POST("/actualizar/:id", perfilHandler.Update, rbacMiddleware.CheckPermission("Perfiles", "editar"), csrfMiddleware.Protect)
	pg.DELETE("/eliminar/:id", perfilHandler.Delete, rbacMiddleware.CheckPermission("Perfiles", "eliminar"), csrfMiddleware.Protect)

	mg := protected.Group("/seguridad/modulos")
	mg.Use(rbacMiddleware.CheckPermission("Módulos", "ver")) // ← Cambiado
	mg.GET("", moduloHandler.Index)
	mg.GET("/nuevo", moduloHandler.CreateForm, rbacMiddleware.CheckPermission("Módulos", "crear")) // ← Cambiado
	mg.GET("/detalle/:id", moduloHandler.Detalle, rbacMiddleware.CheckPermission("Módulos", "ver"))
	mg.POST("/crear", moduloHandler.Create, rbacMiddleware.CheckPermission("Módulos", "crear"), csrfMiddleware.Protect)
	mg.GET("/editar/:id", moduloHandler.EditForm, rbacMiddleware.CheckPermission("Módulos", "editar")) // ← Cambiado
	mg.POST("/actualizar/:id", moduloHandler.Update, rbacMiddleware.CheckPermission("Módulos", "editar"), csrfMiddleware.Protect)
	mg.DELETE("/eliminar/:id", moduloHandler.Delete, rbacMiddleware.CheckPermission("Módulos", "eliminar"), csrfMiddleware.Protect)

	// MÓDULO: PERMISOS-PERFIL
	permGroup := protected.Group("/seguridad/permisos-perfil")
	permGroup.Use(rbacMiddleware.RequireAdmin)
	permGroup.GET("", permisoHandler.Index)
	permGroup.POST("/cargar", permisoHandler.LoadPermissions)
	permGroup.POST("/guardar", permisoHandler.SavePermissions, csrfMiddleware.Protect)
	log.Println("[ROUTES]   - Permisos")

	// MÓDULO: USUARIOS// MÓDULO: MÓDULOS - Cambiar "modulos" por "Módulos" - Cambiar "usuarios" por "Usuarios"
	ug := protected.Group("/seguridad/usuarios")
	ug.Use(rbacMiddleware.CheckPermission("Usuarios", "ver")) // ← Cambiado
	ug.GET("", userHandler.Index)
	ug.GET("/nuevo", userHandler.CreateForm, rbacMiddleware.CheckPermission("Usuarios", "crear")) // ← Cambiado
	ug.POST("/crear", userHandler.Create, rbacMiddleware.CheckPermission("Usuarios", "crear"), csrfMiddleware.Protect)
	ug.GET("/editar/:id", userHandler.EditForm, rbacMiddleware.CheckPermission("Usuarios", "editar")) // ← Cambiado
	ug.POST("/actualizar/:id", userHandler.Update, rbacMiddleware.CheckPermission("Usuarios", "editar"), csrfMiddleware.Protect)
	ug.DELETE("/eliminar/:id", userHandler.Delete, rbacMiddleware.CheckPermission("Usuarios", "eliminar"), csrfMiddleware.Protect)
	ug.GET("/detalle/:id", userHandler.Detail, rbacMiddleware.CheckPermission("Usuarios", "detalle")) // ← Cambiado
	ug.POST("/toggle-status/:id", userHandler.ToggleStatus, rbacMiddleware.CheckPermission("Usuarios", "editar"), csrfMiddleware.Protect)

	// MÓDULOS PRINCIPALES
	// ✅ CORRECTO (coincide con los nombres en BD)
	protected.GET("/principal/clientes", principalHandler.Principal11, rbacMiddleware.RequireModuleAccess("Principal 1.1"))
	protected.GET("/principal/productos", principalHandler.Principal12, rbacMiddleware.RequireModuleAccess("Principal 1.2"))
	protected.GET("/principal/facturas", principalHandler.Principal21, rbacMiddleware.RequireModuleAccess("Principal 2.1"))
	protected.GET("/principal/proveedores", principalHandler.Principal22, rbacMiddleware.RequireModuleAccess("Principal 2.2"))

	log.Println("[ROUTES]   - Módulos principales")

	log.Println("[ROUTES] ✅ Todas las rutas configuradas")

	// ============================================
	// INICIAR SERVIDOR
	// ============================================
	port := env.AppPort
	log.Println("========================================")
	log.Printf("🚀 Servidor iniciado en %s:%s", env.BaseURL, port)
	log.Printf("📊 Entorno: %s", env.AppEnv)
	log.Printf("🔐 CSRF: activado")
	log.Printf("🔑 JWT: activado (expiración: 24h)")
	log.Printf("⏱️ Rate Limit: 20 peticiones por segundo")
	log.Printf("🧪 Ruta de prueba: %s/ping", env.BaseURL)
	log.Printf("🔐 Login: %s/login", env.BaseURL)
	log.Println("========================================")

	if err := e.Start(":" + port); err != nil && err != http.ErrServerClosed {
		log.Fatal("❌ Error iniciando servidor:", err)
	}
}

// mustGetWd obtiene el directorio actual o panic
func mustGetWd() string {
	wd, err := os.Getwd()
	if err != nil {
		return "ERROR: " + err.Error()
	}
	return wd
}
