package config

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"

	_ "github.com/lib/pq"
)

// ConnectDB establece la conexión a PostgreSQL
func ConnectDB() (*sql.DB, error) {
	// Railway proporciona DATABASE_URL automáticamente
	dbURL := os.Getenv("DATABASE_URL")

	// DEBUG: Mostrar info de DATABASE_URL
	fmt.Printf("🔍 DEBUG: DATABASE_URL length: %d\n", len(dbURL))
	if dbURL != "" {
		fmt.Println("✅ DATABASE_URL encontrada - usando Railway")
		fmt.Printf("🔍 DEBUG: Primeros 50 chars: %s...\n", func() string {
			if len(dbURL) > 50 {
				return dbURL[:50]
			}
			return dbURL
		}())
	} else {
		fmt.Println("⚠️  DATABASE_URL NO encontrada - usando local")
	}

	var connStr string
	var environment string

	if dbURL == "" {
		// Desarrollo local - usa tu cadena original
		connStr = "postgres://postgres:27100419kAVZ@localhost:5432/TestDb?sslmode=disable"
		environment = "DESARROLLO LOCAL"
		fmt.Println("📌 [CONFIG] Usando configuración local")
	} else {
		// Producción en Railway
		environment = "RAILWAY (PRODUCCIÓN)"
		fmt.Println("📌 [CONFIG] Usando DATABASE_URL de Railway")

		// Tu URL ya está en formato postgresql://, está CORRECTO
		// NO necesitas convertir postgres:// a postgresql://
		connStr = dbURL

		// Solo convertir si empieza con postgres:// (pero Railway usa postgresql://)
		if strings.HasPrefix(dbURL, "postgres://") {
			fmt.Println("⚠️  Convirtiendo postgres:// a postgresql://")
			connStr = strings.Replace(dbURL, "postgres://", "postgresql://", 1)
		}
	}

	// Log seguro (sin mostrar credenciales completas)
	logSafeConnection(connStr, environment)

	// Conectar a la base de datos
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("❌ Error al abrir conexión: %v", err)
	}

	// Configurar pool de conexiones (IMPORTANTE para producción)
	db.SetMaxOpenConns(10)             // REDUCIDO para Railway (25 era mucho)
	db.SetMaxIdleConns(3)              // REDUCIDO
	db.SetConnMaxLifetime(5 * 60 * 60) // 5 horas máximo por conexión
	db.SetConnMaxIdleTime(30 * 60)     // 30 minutos máximo inactiva

	// Verificar conexión con mejor mensaje de error
	err = db.Ping()
	if err != nil {
		// Mostrar error más detallado
		errorMsg := fmt.Sprintf("❌ Error al conectar con la base de datos: %v\n", err)
		errorMsg += fmt.Sprintf("   Environment: %s\n", environment)
		errorMsg += fmt.Sprintf("   Connection string: %s\n", logSafeConnectionForError(connStr))
		return nil, fmt.Errorf(errorMsg)
	}

	// SIEMPRE crear tabla si no existe (tanto en desarrollo como producción)
	err = createTableIfNotExists(db)
	if err != nil {
		return nil, fmt.Errorf("❌ Error creando/verificando tabla: %v", err)
	}

	log.Printf("✅ Conexión a PostgreSQL establecida exitosamente (%s)", environment)
	return db, nil
}

// logSafeConnection muestra la URL de conexión sin credenciales
func logSafeConnection(connStr string, environment string) {
	safeStr := connStr

	// Ocultar credenciales para logs
	if strings.Contains(connStr, "@") {
		parts := strings.SplitN(connStr, "@", 2)
		if len(parts) == 2 {
			// Reemplazar usuario:contraseña por ****
			credPart := parts[0]
			if strings.Contains(credPart, "://") {
				protocolParts := strings.SplitN(credPart, "://", 2)
				if len(protocolParts) == 2 {
					safeStr = protocolParts[0] + "://****@" + parts[1]
				}
			}
		}
	}

	log.Printf("🔗 [%s] Conectando a: %s", environment, safeStr)
}

// Versión para errores (muestra host y puerto)
func logSafeConnectionForError(connStr string) string {
	if strings.Contains(connStr, "@") {
		parts := strings.SplitN(connStr, "@", 2)
		if len(parts) == 2 {
			return "postgresql://****@" + parts[1]
		}
	}
	return connStr
}

func createTableIfNotExists(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		name VARCHAR(100) NOT NULL,
		email VARCHAR(100) UNIQUE NOT NULL,
		phone VARCHAR(20) NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	
	DO $$
	BEGIN
		IF NOT EXISTS (SELECT 1 FROM pg_indexes WHERE indexname = 'idx_users_email') THEN
			CREATE INDEX idx_users_email ON users(email);
		END IF;
	END $$;
	`

	_, err := db.Exec(query)
	if err != nil {
		return fmt.Errorf("❌ Error al crear tabla: %v", err)
	}

	log.Println("✅ Tabla 'users' verificada/creada exitosamente")
	return nil
}
