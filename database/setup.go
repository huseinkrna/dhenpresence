package database

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/lib/pq"  // PostgreSQL driver
	_ "modernc.org/sqlite" // SQLite driver (fallback lokal)
)

var DB *sql.DB

func InitDB() {
	var err error

	// Cek apakah ada DATABASE_URL (untuk production/Supabase)
	dbURL := os.Getenv("DATABASE_URL")

	if dbURL != "" {
		// MODE PRODUCTION: Pakai PostgreSQL (Supabase)
		log.Println("🌐 Connecting to PostgreSQL (Supabase)...")
		DB, err = sql.Open("postgres", dbURL)
		if err != nil {
			log.Fatal("❌ Gagal connect ke PostgreSQL:", err)
		}

		// Test connection
		if err = DB.Ping(); err != nil {
			log.Fatal("❌ PostgreSQL ping failed:", err)
		}
		log.Println("✅ Connected to PostgreSQL!")

		createTablesPostgres()
	} else {
		// MODE DEVELOPMENT: Pakai SQLite lokal
		log.Println("💾 Using local SQLite database...")
		DB, err = sql.Open("sqlite", "./dhenpresence.db")
		if err != nil {
			log.Fatal(err)
		}
		createTablesSQLite()
	}

	seedUser()
}

// PostgreSQL Tables (untuk Supabase)
func createTablesPostgres() {
	// Tabel Users
	queryUser := `
	CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		username TEXT UNIQUE NOT NULL,
		password TEXT NOT NULL,
		full_name TEXT NOT NULL,
		role TEXT DEFAULT 'staff',
		hourly_rate INTEGER DEFAULT 0,
		phone_number TEXT DEFAULT '-',
		avatar_url TEXT DEFAULT ''
	);`
	if _, err := DB.Exec(queryUser); err != nil {
		log.Println("Tabel users:", err)
	}

	// Tabel Attendance
	queryAbsen := `
	CREATE TABLE IF NOT EXISTS attendance (
		id SERIAL PRIMARY KEY,
		user_id INTEGER,
		shift_date DATE,
		clock_in_time TIMESTAMP,
		clock_out_time TIMESTAMP,
		status TEXT,
		permit_reason TEXT,
		closing_note TEXT,
		is_late BOOLEAN DEFAULT false,
		penalty_hours INTEGER DEFAULT 0,
		compensation_hours INTEGER DEFAULT 0,
		is_auto_closed BOOLEAN DEFAULT false,
		manual_salary INTEGER DEFAULT 0
	);`
	if _, err := DB.Exec(queryAbsen); err != nil {
		log.Println("Tabel attendance:", err)
	}

	// Update default rate
	DB.Exec("UPDATE users SET hourly_rate = 15000 WHERE hourly_rate = 0")
}

// SQLite Tables (untuk development lokal)
func createTablesSQLite() {
	queryUser := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT UNIQUE NOT NULL,
		password TEXT NOT NULL,
		full_name TEXT NOT NULL,
		role TEXT DEFAULT 'staff',
		hourly_rate INTEGER DEFAULT 0
	);`
	if _, err := DB.Exec(queryUser); err != nil {
		log.Fatal("Gagal buat tabel user:", err)
	}

	queryAbsen := `
	CREATE TABLE IF NOT EXISTS attendance (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER,
		shift_date DATE,
		clock_in_time DATETIME,
		clock_out_time DATETIME,
		status TEXT,
		permit_reason TEXT,
		closing_note TEXT,
		is_late BOOLEAN DEFAULT 0,
		penalty_hours INTEGER DEFAULT 0,
		compensation_hours INTEGER DEFAULT 0,
		is_auto_closed BOOLEAN DEFAULT 0
	);`
	if _, err := DB.Exec(queryAbsen); err != nil {
		log.Fatal("Gagal buat tabel absensi:", err)
	}

	// Migrations untuk SQLite
	DB.Exec("UPDATE users SET hourly_rate = 15000 WHERE hourly_rate = 0")
	DB.Exec("ALTER TABLE attendance ADD COLUMN manual_salary INTEGER DEFAULT 0;")
	DB.Exec("ALTER TABLE users ADD COLUMN phone_number TEXT DEFAULT '-';")
	DB.Exec("ALTER TABLE users ADD COLUMN avatar_url TEXT DEFAULT '';")
}

func seedUser() {
	var id int
	err := DB.QueryRow("SELECT id FROM users WHERE username = 'owner'").Scan(&id)

	if err == sql.ErrNoRows {
		_, err := DB.Exec("INSERT INTO users (username, password, full_name, role, hourly_rate) VALUES ($1, $2, $3, $4, $5)",
			"owner", "kopi123", "Owner Dhen Coffee", "owner", 15000)
		if err != nil {
			// Coba dengan syntax SQLite (?)
			DB.Exec("INSERT INTO users (username, password, full_name, role, hourly_rate) VALUES (?, ?, ?, ?, ?)",
				"owner", "kopi123", "Owner Dhen Coffee", "owner", 15000)
		}
		log.Println("✅ User default: owner / kopi123")
	}
}
