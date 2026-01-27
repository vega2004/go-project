package handler

import (
	"fmt"
	"net/http"
	"tu-proyecto/model"
	"tu-proyecto/service"

	"github.com/labstack/echo/v4"
)

type UserHandler struct {
	service service.UserService
}

func NewUserHandler(service service.UserService) *UserHandler {
	return &UserHandler{service: service}
}

func (h *UserHandler) Welcome(c echo.Context) error {
	fmt.Println("[HANDLER] Welcome() - Inicio")

	// Obtener breadcrumbs del middleware
	breadcrumbsInterface := c.Get("breadcrumbs")
	var breadcrumbs []map[string]string

	if breadcrumbsInterface != nil {
		breadcrumbs = breadcrumbsInterface.([]map[string]string)
		fmt.Printf("   Breadcrumbs del middleware: %v\n", breadcrumbs)
	} else {
		breadcrumbs = []map[string]string{{"name": "Inicio", "url": "/"}}
		fmt.Println("   Breadcrumbs no encontrados en middleware, usando defaults")
	}

	data := map[string]interface{}{
		"Title":       "Bienvenido al Sistema",
		"Message":     "¡Hola Mundo!",
		"breadcrumbs": breadcrumbs,
	}

	fmt.Printf("   Datos a enviar al template: %+v\n", data)
	fmt.Println("   Renderizando template 'welcome.html'...")

	return c.Render(http.StatusOK, "welcome.html", data)
}

func (h *UserHandler) ShowForm(c echo.Context) error {
	fmt.Println("[HANDLER] ShowForm() - Inicio")

	errorMsg := c.QueryParam("error")
	fmt.Printf("   Error message from query: '%s'\n", errorMsg)

	// Obtener breadcrumbs del middleware
	breadcrumbsInterface := c.Get("breadcrumbs")
	var breadcrumbs []map[string]string

	if breadcrumbsInterface != nil {
		breadcrumbs = breadcrumbsInterface.([]map[string]string)
		fmt.Printf("   Breadcrumbs del middleware: %v\n", breadcrumbs)
	} else {
		breadcrumbs = []map[string]string{
			{"name": "Inicio", "url": "/"},
			{"name": "Registro", "url": "/form"},
		}
		fmt.Println("   Breadcrumbs no encontrados en middleware, usando defaults")
	}

	data := map[string]interface{}{
		"Title":        "Registro de Usuario",
		"RecaptchaKey": "6LdjX1gsAAAAAAmGXFf5nW7EkXqx9xukAwyJR35Z",
		"Error":        errorMsg,
		"breadcrumbs":  breadcrumbs,
	}

	fmt.Printf("   Datos a enviar al template: %+v\n", data)
	fmt.Println("   Renderizando template 'form.html'...")

	return c.Render(http.StatusOK, "form.html", data)
}

func (h *UserHandler) SubmitForm(c echo.Context) error {
	fmt.Println("[HANDLER] SubmitForm() - Inicio")

	var form model.UserForm

	// Obtener datos del formulario
	fmt.Println("   Bind del formulario...")
	if err := c.Bind(&form); err != nil {
		fmt.Printf("   Error en Bind: %v\n", err)
		return c.Redirect(http.StatusSeeOther, "/form?error=Error procesando formulario")
	}

	fmt.Printf("   Datos del formulario recibidos:\n")
	fmt.Printf("     Nombre: %s\n", form.Name)
	fmt.Printf("     Email: %s\n", form.Email)
	fmt.Printf("     Teléfono: %s\n", form.Phone)
	fmt.Printf("     reCAPTCHA token: %s\n", func() string {
		if len(form.RecaptchaToken) > 20 {
			return form.RecaptchaToken[:20] + "..."
		}
		return form.RecaptchaToken
	}())

	// Intentar crear el usuario
	fmt.Println("   Creando usuario en servicio...")
	err := h.service.CreateUser(&form)
	if err != nil {
		fmt.Printf("   Error creando usuario: %v\n", err)
		// Redirigir con mensaje de error
		return c.Redirect(http.StatusSeeOther, "/form?error="+err.Error())
	}

	fmt.Println("   Usuario creado exitosamente")
	fmt.Println("   Redirigiendo a /success")

	// Si todo sale bien, redirigir a éxito
	return c.Redirect(http.StatusSeeOther, "/success")
}

func (h *UserHandler) ShowSuccess(c echo.Context) error {
	fmt.Println("[HANDLER] ShowSuccess() - Inicio")

	// Obtener breadcrumbs del middleware
	breadcrumbsInterface := c.Get("breadcrumbs")
	var breadcrumbs []map[string]string

	if breadcrumbsInterface != nil {
		breadcrumbs = breadcrumbsInterface.([]map[string]string)
		fmt.Printf("   Breadcrumbs del middleware: %v\n", breadcrumbs)
	} else {
		breadcrumbs = []map[string]string{
			{"name": "Inicio", "url": "/"},
			{"name": "Registro", "url": "/form"},
			{"name": "Éxito", "url": "/success"},
		}
		fmt.Println("    Breadcrumbs no encontrados en middleware, usando defaults")
	}

	data := map[string]interface{}{
		"Title":       "Registro Exitoso",
		"breadcrumbs": breadcrumbs,
	}

	fmt.Printf("   Datos a enviar al template: %+v\n", data)
	fmt.Println("   Renderizando template 'success.html'...")

	return c.Render(http.StatusOK, "success.html", data)
}

func (h *UserHandler) Maintenance(c echo.Context) error {
	fmt.Println("[HANDLER] Maintenance() - Inicio")

	// Obtener breadcrumbs del middleware
	breadcrumbsInterface := c.Get("breadcrumbs")
	var breadcrumbs []map[string]string

	if breadcrumbsInterface != nil {
		breadcrumbs = breadcrumbsInterface.([]map[string]string)
		fmt.Printf("   Breadcrumbs del middleware: %v\n", breadcrumbs)
	} else {
		breadcrumbs = []map[string]string{
			{"name": "Inicio", "url": "/"},
			{"name": "Mantenimiento", "url": "/maintenance"},
		}
		fmt.Println("   Breadcrumbs no encontrados en middleware, usando defaults")
	}

	data := map[string]interface{}{
		"Title":       "🔧 Sitio en Mantenimiento",
		"breadcrumbs": breadcrumbs,
	}

	fmt.Printf("   Datos a enviar al template: %+v\n", data)
	fmt.Println("   Renderizando template 'maintenance.html'...")

	return c.Render(http.StatusOK, "maintenance.html", data)
}
