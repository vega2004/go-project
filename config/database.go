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

		// Asegurar formato correcto para Go
		if strings.HasPrefix(dbURL, "postgres://") {
			connStr = strings.Replace(dbURL, "postgres://", "postgresql://", 1)
		} else {
			connStr = dbURL
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
	db.SetMaxOpenConns(25)             // Máximo de conexiones abiertas
	db.SetMaxIdleConns(5)              // Conexiones inactivas en pool
	db.SetConnMaxLifetime(5 * 60 * 60) // 5 horas máximo por conexión
	db.SetConnMaxIdleTime(30 * 60)     // 30 minutos máximo inactiva

	// Verificar conexión
	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("❌ Error al conectar con la base de datos: %v", err)
	}

	// Crear tabla si no existe (solo en desarrollo o primera vez)
	if environment == "DESARROLLO LOCAL" {
		err = createTableIfNotExists(db)
		if err != nil {
			return nil, err
		}
	} else {
		// En producción, solo verificar que la tabla existe
		err = verifyTableExists(db)
		if err != nil {
			log.Printf("⚠️  Advertencia: %v", err)
			log.Println("📌 Creando tabla en producción...")
			err = createTableIfNotExists(db)
			if err != nil {
				return nil, fmt.Errorf("❌ Error creando tabla en producción: %v", err)
			}
		}
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

func verifyTableExists(db *sql.DB) error {
	query := `
	SELECT EXISTS (
		SELECT FROM information_schema.tables 
		WHERE table_schema = 'public' 
		AND table_name = 'users'
	);
	`

	var exists bool
	err := db.QueryRow(query).Scan(&exists)
	if err != nil {
		return fmt.Errorf("error verificando tabla: %v", err)
	}

	if !exists {
		return fmt.Errorf("la tabla 'users' no existe en la base de datos")
	}

	log.Println("✅ Tabla 'users' existe en la base de datos")
	return nil
}
