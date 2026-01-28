package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"

	_ "github.com/lib/pq"  // PostgreSQL driver
	_ "modernc.org/sqlite" // SQLite driver (fallback lokal)
)

var DB *sql.DB
var UsePostgres bool // Flag untuk menandakan apakah menggunakan POSTGRES

// Helper untuk adaptasi query
func AdaptQuery(q string) string {
	if !UsePostgres {
		return q
	}
	// Ganti ? dengan $1, $2, dst.
	n := 1
	for {
		if !strings.Contains(q, "?") {
			break
		}
		q = strings.Replace(q, "?", fmt.Sprintf("$%d", n), 1)
		n++
	}
	return q
}

func InitDB() {
	var err error

	// Cek apakah ada DATABASE_URL (untuk production/Supabase)
	dbURL := os.Getenv("DATABASE_URL")

	if dbURL != "" {
		UsePostgres = true
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
	// Seed Owner
	var id int
	err := DB.QueryRow("SELECT id FROM users WHERE username = 'owner'").Scan(&id)

	if err == sql.ErrNoRows {
		avatarOwner := "https://api.dicebear.com/7.x/adventurer/svg?seed=owner&backgroundColor=b6e3f4"
		_, err := DB.Exec("INSERT INTO users (username, password, full_name, role, hourly_rate, phone_number, avatar_url) VALUES ($1, $2, $3, $4, $5, $6, $7)",
			"owner", "kopi123", "Owner Dhen Coffee", "owner", 15000, "081234567890", avatarOwner)
		if err != nil {
			// Coba dengan syntax SQLite (?)
			DB.Exec("INSERT INTO users (username, password, full_name, role, hourly_rate, phone_number, avatar_url) VALUES (?, ?, ?, ?, ?, ?, ?)",
				"owner", "kopi123", "Owner Dhen Coffee", "owner", 15000, "081234567890", avatarOwner)
		}
		log.Println("✅ User default: owner / kopi123")
	} else {
		// Update owner jika sudah ada tapi belum ada avatar/phone
		avatarOwner := "https://api.dicebear.com/7.x/adventurer/svg?seed=owner&backgroundColor=b6e3f4"
		DB.Exec(AdaptQuery("UPDATE users SET phone_number = ?, avatar_url = ? WHERE username = 'owner' AND (phone_number IS NULL OR phone_number = '' OR phone_number = '-')"),
			"081234567890", avatarOwner)
	}

	// Seed Super Admin (Manajer)
	var managerId int
	err = DB.QueryRow("SELECT id FROM users WHERE username = 'manajer'").Scan(&managerId)

	if err == sql.ErrNoRows {
		avatarManager := "https://api.dicebear.com/7.x/adventurer/svg?seed=manajer&backgroundColor=c0aede"
		_, err := DB.Exec("INSERT INTO users (username, password, full_name, role, hourly_rate, phone_number, avatar_url) VALUES ($1, $2, $3, $4, $5, $6, $7)",
			"manajer", "sidikalang", "Manajer Dhen Coffee", "owner", 15000, "081234567891", avatarManager)
		if err != nil {
			// Coba dengan syntax SQLite (?)
			_, errSqlite := DB.Exec("INSERT INTO users (username, password, full_name, role, hourly_rate, phone_number, avatar_url) VALUES (?, ?, ?, ?, ?, ?, ?)",
				"manajer", "sidikalang", "Manajer Dhen Coffee", "owner", 15000, "081234567891", avatarManager)
			if errSqlite != nil {
				log.Printf("❌ Error creating manajer account: %v", errSqlite)
			}
		}
		log.Println("✅ User default: manajer / sidikalang")
	} else {
		log.Println("✅ User manajer sudah ada di database")
	}
}
