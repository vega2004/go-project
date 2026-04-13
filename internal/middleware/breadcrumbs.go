package middleware

import (
	"database/sql"
	"log"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/labstack/echo/v4"
)

// ============================================
// ESTRUCTURAS
// ============================================

type Breadcrumb struct {
	Name string `json:"name"`
	URL  string `json:"url"`
	Icon string `json:"icon,omitempty"`
}

type BreadcrumbConfig struct {
	Pattern     string
	Name        string
	Parent      string
	Icon        string
	GetNameFunc func(c echo.Context, db *sql.DB) string
}

type BreadcrumbMiddleware struct {
	db          *sql.DB
	routes      []BreadcrumbConfig
	routesCache map[string]*BreadcrumbConfig
	cacheMutex  sync.RWMutex
}

// NewBreadcrumbMiddleware - Constructor (requiere db)
func NewBreadcrumbMiddleware(db *sql.DB) *BreadcrumbMiddleware {
	bm := &BreadcrumbMiddleware{
		db:          db,
		routesCache: make(map[string]*BreadcrumbConfig),
	}
	bm.initRoutes()
	return bm
}

// initRoutes - Inicializa la configuración de rutas
func (bm *BreadcrumbMiddleware) initRoutes() {
	bm.routes = []BreadcrumbConfig{
		// Dashboard
		{Pattern: "/dashboard", Name: "Dashboard", Parent: "/", Icon: "bi-speedometer2"},

		// Autenticación
		{Pattern: "/login", Name: "Iniciar Sesión", Parent: "/", Icon: "bi-box-arrow-in-right"},
		{Pattern: "/register", Name: "Registro", Parent: "/", Icon: "bi-person-plus"},

		// Perfil
		{Pattern: "/perfil", Name: "Mi Perfil", Parent: "/dashboard", Icon: "bi-person-circle"},

		// Seguridad - Perfiles
		{Pattern: "/seguridad/perfiles", Name: "Perfiles", Parent: "/dashboard", Icon: "bi-person-badge"},
		{Pattern: "/seguridad/perfiles/nuevo", Name: "Nuevo Perfil", Parent: "/seguridad/perfiles", Icon: "bi-plus-circle"},
		{Pattern: "/seguridad/perfiles/editar/:id", Name: "Editar Perfil", Parent: "/seguridad/perfiles", Icon: "bi-pencil",
			GetNameFunc: bm.getPerfilName},

		// Seguridad - Módulos
		{Pattern: "/seguridad/modulos", Name: "Módulos", Parent: "/dashboard", Icon: "bi-grid-3x3"},
		{Pattern: "/seguridad/modulos/nuevo", Name: "Nuevo Módulo", Parent: "/seguridad/modulos", Icon: "bi-plus-circle"},
		{Pattern: "/seguridad/modulos/editar/:id", Name: "Editar Módulo", Parent: "/seguridad/modulos", Icon: "bi-pencil",
			GetNameFunc: bm.getModuloName},

		// Seguridad - Permisos
		{Pattern: "/seguridad/permisos-perfil", Name: "Permisos por Perfil", Parent: "/dashboard", Icon: "bi-shield-lock"},

		// Seguridad - Usuarios
		{Pattern: "/seguridad/usuarios", Name: "Usuarios", Parent: "/dashboard", Icon: "bi-people"},
		{Pattern: "/seguridad/usuarios/nuevo", Name: "Nuevo Usuario", Parent: "/seguridad/usuarios", Icon: "bi-plus-circle"},
		{Pattern: "/seguridad/usuarios/editar/:id", Name: "Editar Usuario", Parent: "/seguridad/usuarios", Icon: "bi-pencil",
			GetNameFunc: bm.getUserName},
		{Pattern: "/seguridad/usuarios/detalle/:id", Name: "Detalle Usuario", Parent: "/seguridad/usuarios", Icon: "bi-eye",
			GetNameFunc: bm.getUserName},

		// Módulos principales
		{Pattern: "/principal/clientes", Name: "Principal 1.1 - Clientes", Parent: "/dashboard", Icon: "bi-building"},
		{Pattern: "/principal/productos", Name: "Principal 1.2 - Productos", Parent: "/dashboard", Icon: "bi-box"},
		{Pattern: "/principal/facturas", Name: "Principal 2.1 - Facturas", Parent: "/dashboard", Icon: "bi-receipt"},
		{Pattern: "/principal/proveedores", Name: "Principal 2.2 - Proveedores", Parent: "/dashboard", Icon: "bi-truck"},
	}
}

// ============================================
// FUNCIONES PARA OBTENER NOMBRES DINÁMICOS
// ============================================

func (bm *BreadcrumbMiddleware) getPerfilName(c echo.Context, db *sql.DB) string {
	id := bm.extractID(c.Param("id"))
	if id == 0 {
		return "Editar Perfil"
	}

	var nombre string
	query := `SELECT nombre FROM perfiles WHERE id = $1`
	err := db.QueryRow(query, id).Scan(&nombre)
	if err != nil {
		log.Printf("[BREADCRUMB] Error obteniendo nombre de perfil %d: %v", id, err)
		return "Editar Perfil"
	}
	return "Editar: " + nombre
}

func (bm *BreadcrumbMiddleware) getModuloName(c echo.Context, db *sql.DB) string {
	id := bm.extractID(c.Param("id"))
	if id == 0 {
		return "Editar Módulo"
	}

	var nombre string
	query := `SELECT nombre_mostrar FROM modulos WHERE id = $1`
	err := db.QueryRow(query, id).Scan(&nombre)
	if err != nil {
		log.Printf("[BREADCRUMB] Error obteniendo nombre de módulo %d: %v", id, err)
		return "Editar Módulo"
	}
	return "Editar: " + nombre
}

func (bm *BreadcrumbMiddleware) getUserName(c echo.Context, db *sql.DB) string {
	id := bm.extractID(c.Param("id"))
	if id == 0 {
		return "Usuario"
	}

	var nombre string
	query := `SELECT name FROM users WHERE id = $1`
	err := db.QueryRow(query, id).Scan(&nombre)
	if err != nil {
		log.Printf("[BREADCRUMB] Error obteniendo nombre de usuario %d: %v", id, err)
		return "Usuario"
	}
	return nombre
}

func (bm *BreadcrumbMiddleware) extractID(idParam string) int {
	if idParam == "" {
		return 0
	}
	id, err := strconv.Atoi(idParam)
	if err != nil {
		return 0
	}
	return id
}

// ============================================
// FUNCIÓN PRINCIPAL DEL MIDDLEWARE
// ============================================

// Handler - Middleware para generar migajas de pan
func (bm *BreadcrumbMiddleware) Handler(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		breadcrumbs := []Breadcrumb{
			{Name: "Inicio", URL: "/", Icon: "bi-house-door"},
		}

		path := c.Request().URL.Path

		// Buscar en cache
		bm.cacheMutex.RLock()
		cached, exists := bm.routesCache[path]
		bm.cacheMutex.RUnlock()

		var currentConfig *BreadcrumbConfig
		if exists {
			currentConfig = cached
		} else {
			for i := range bm.routes {
				if bm.matchPath(bm.routes[i].Pattern, path) {
					currentConfig = &bm.routes[i]
					bm.cacheMutex.Lock()
					bm.routesCache[path] = currentConfig
					bm.cacheMutex.Unlock()
					break
				}
			}
		}

		if currentConfig != nil {
			breadcrumbs = bm.buildBreadcrumbs(breadcrumbs, currentConfig, c, path)
		}

		c.Set("breadcrumbs", breadcrumbs)
		return next(c)
	}
}

// matchPath - Verifica si el path coincide con el patrón
func (bm *BreadcrumbMiddleware) matchPath(pattern, path string) bool {
	if pattern == path {
		return true
	}

	// Convertir patrón con :id a regex
	if strings.Contains(pattern, ":id") {
		regexPattern := strings.ReplaceAll(pattern, ":id", "([0-9]+)")
		matched, err := regexp.MatchString("^"+regexPattern+"$", path)
		if err == nil && matched {
			return true
		}
	}

	// Para patrones con wildcard
	if strings.HasSuffix(pattern, "/*") {
		basePattern := strings.TrimSuffix(pattern, "/*")
		return strings.HasPrefix(path, basePattern)
	}

	return false
}

// buildBreadcrumbs - Construye la jerarquía de breadcrumbs (CORREGIDO)
func (bm *BreadcrumbMiddleware) buildBreadcrumbs(breadcrumbs []Breadcrumb, current *BreadcrumbConfig, c echo.Context, actualPath string) []Breadcrumb {
	// Obtener nombre dinámico
	name := current.Name
	if current.GetNameFunc != nil && bm.db != nil {
		dynamicName := current.GetNameFunc(c, bm.db)
		if dynamicName != "" {
			name = dynamicName
		}
	}

	// Construir lista de padres
	type parentInfo struct {
		URL  string
		Name string
		Icon string
	}
	parents := []parentInfo{
		{URL: actualPath, Name: name, Icon: current.Icon},
	}

	currentParent := current.Parent
	maxDepth := 10
	depth := 0

	for currentParent != "/" && currentParent != "" && depth < maxDepth {
		found := false
		for _, route := range bm.routes {
			if route.Pattern == currentParent {
				parentName := route.Name
				if route.GetNameFunc != nil && bm.db != nil {
					dynamicName := route.GetNameFunc(c, bm.db)
					if dynamicName != "" {
						parentName = dynamicName
					}
				}
				parents = append([]parentInfo{{
					URL:  route.Pattern,
					Name: parentName,
					Icon: route.Icon,
				}}, parents...)
				currentParent = route.Parent
				found = true
				break
			}
		}
		if !found {
			break
		}
		depth++
	}

	// Agregar a breadcrumbs (evitando duplicados)
	for _, p := range parents {
		lastIdx := len(breadcrumbs) - 1
		if lastIdx >= 0 && breadcrumbs[lastIdx].URL == p.URL {
			continue
		}
		breadcrumbs = append(breadcrumbs, Breadcrumb{
			Name: p.Name,
			URL:  p.URL,
			Icon: p.Icon,
		})
	}

	return breadcrumbs
}
