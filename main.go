package main

import (
	"database/sql"
	"dhenpresence/database"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"math"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

// --- KONFIGURASI ---
const CAFE_LAT = -7.773025241743875
const CAFE_LONG = 110.41145053080596
const MAX_DISTANCE_METER = 50.0

type Response struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Time    string `json:"time"`
}

type AdminData struct {
	UserFullName, UserAvatar string
	LiveWorkers              []WorkerData
	History                  []HistoryData
	TotalSalary              string
	FilterStart, FilterEnd   string
	CurrentRate              string
}

type WorkerData struct {
	Name, ShiftTime, Duration, Avatar string
}

type HistoryData struct {
	ID                                          int
	Date, Name, PhoneNumber, Shift, Avatar      string
	TotalHours, Status, PermitReason, SalaryEst string
	OriginalSalary                              int64
}

// ---------------------------------------------------------
// SEED DUMMY DATA
// ---------------------------------------------------------
func seedDummyData() {
	log.Println("🌱 Checking dummy data...")

	// 4 Dummy employees
	employees := []struct {
		Username   string
		Password   string
		FullName   string
		Phone      string
		HourlyRate int
		AvatarURL  string
	}{
		{"budi", "budi123", "Budi Santoso", "081234567890", 12000, "https://api.dicebear.com/7.x/adventurer/svg?seed=budi&backgroundColor=b6e3f4"},
		{"siti", "siti123", "Siti Nurhaliza", "081234567891", 13000, "https://api.dicebear.com/7.x/adventurer/svg?seed=siti&backgroundColor=c0aede"},
		{"andi", "andi123", "Andi Wijaya", "081234567892", 11000, "https://api.dicebear.com/7.x/adventurer/svg?seed=andi&backgroundColor=d1d4f9"},
		{"rina", "rina123", "Rina Agustin", "081234567893", 12500, "https://api.dicebear.com/7.x/adventurer/svg?seed=rina&backgroundColor=ffd8be"},
	}

	// Insert employees jika belum ada
	var createdEmployees []int64
	for _, emp := range employees {
		var existingID int
		err := database.DB.QueryRow(database.AdaptQuery("SELECT id FROM users WHERE username = ?"), emp.Username).Scan(&existingID)

		if err == sql.ErrNoRows {
			result, err := database.DB.Exec(
				database.AdaptQuery("INSERT INTO users (username, password, full_name, role, hourly_rate, phone_number, avatar_url) VALUES (?, ?, ?, 'employee', ?, ?, ?)"),
				emp.Username, emp.Password, emp.FullName, emp.HourlyRate, emp.Phone, emp.AvatarURL,
			)
			if err != nil {
				log.Printf("⚠️ Skip %s: %v", emp.FullName, err)
				continue
			}
			userID, _ := result.LastInsertId()
			createdEmployees = append(createdEmployees, userID)
			log.Printf("✅ Created: %s (ID: %d)", emp.FullName, userID)
		} else {
			log.Printf("ℹ️ User %s already exists (ID: %d)", emp.FullName, existingID)
			createdEmployees = append(createdEmployees, int64(existingID))
		}
	}

	if len(createdEmployees) == 0 {
		log.Println("⚠️ No employees to seed attendance")
		return
	}

	// Generate attendance data dari 1 Jan 2026
	log.Println("📅 Generating attendance data...")
	startDate := time.Date(2026, 1, 1, 0, 0, 0, 0, time.Local)
	endDate := time.Now()

	// Get all employee IDs
	rows, _ := database.DB.Query(database.AdaptQuery("SELECT id FROM users WHERE role = 'employee'"))
	var employeeIDs []int64
	for rows.Next() {
		var id int64
		rows.Scan(&id)
		employeeIDs = append(employeeIDs, id)
	}
	rows.Close()

	if len(employeeIDs) == 0 {
		log.Println("⚠️ No employee IDs found")
		return
	}

	log.Printf("👥 Seeding attendance for %d employees", len(employeeIDs))

	// Use math/rand for better randomness
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))

	totalInserted := 0
	for d := startDate; d.Before(endDate) || d.Equal(endDate); d = d.AddDate(0, 0, 1) {
		if d.Weekday() == time.Sunday {
			continue // Skip Sundays
		}

		for _, empID := range employeeIDs {
			// Check if attendance already exists
			var existingID int
			err := database.DB.QueryRow(
				database.AdaptQuery("SELECT id FROM attendance WHERE user_id = ? AND shift_date = ?"),
				empID, d.Format("2006-01-02"),
			).Scan(&existingID)

			if err != sql.ErrNoRows {
				continue // Attendance already exists
			}

			// Random attendance: 85% hadir, 10% izin, 5% sakit
			randomChance := rnd.Float64()

			if randomChance < 0.85 {
				// Hadir - generate realistic shift
				var clockInHour, clockOutHour int
				shiftChoice := rnd.Float64()

				if shiftChoice < 0.6 {
					// 60% Shift Pagi (08:00 - 16:00)
					clockInHour = 8
					clockOutHour = 16
				} else if shiftChoice < 0.9 {
					// 30% Shift Sore (13:00 - 21:00)
					clockInHour = 13
					clockOutHour = 21
				} else {
					// 10% Shift Malam (19:00 - 03:00)
					clockInHour = 19
					clockOutHour = 3
				}

				// Add randomness to clock in (-10 to +30 minutes)
				clockInMinute := rnd.Intn(40) - 10
				if clockInMinute < 0 {
					clockInMinute = 0
				}

				// Clock out with slight variation (0-20 minutes)
				clockOutMinute := rnd.Intn(20)

				clockInTime := time.Date(d.Year(), d.Month(), d.Day(), clockInHour, clockInMinute, 0, 0, time.Local)

				var clockOutTime time.Time
				if clockOutHour < clockInHour {
					// Overnight shift
					clockOutTime = time.Date(d.Year(), d.Month(), d.Day()+1, clockOutHour, clockOutMinute, 0, 0, time.Local)
				} else {
					clockOutTime = time.Date(d.Year(), d.Month(), d.Day(), clockOutHour, clockOutMinute, 0, 0, time.Local)
				}

				// Determine if late (> 15 minutes after scheduled time)
				scheduledTime := time.Date(d.Year(), d.Month(), d.Day(), clockInHour, 0, 0, 0, time.Local)
				isLate := clockInTime.After(scheduledTime.Add(15 * time.Minute))

				var penaltyHours int64
				if isLate {
					lateMinutes := int(clockInTime.Sub(scheduledTime).Minutes())
					if lateMinutes > 60 {
						penaltyHours = 2
					} else if lateMinutes > 15 {
						penaltyHours = 1
					}
				}

				// 15% chance of overtime
				var compensationHours int64
				if rnd.Float64() < 0.15 {
					compensationHours = int64(rnd.Intn(3) + 1) // 1-3 hours
				}

				_, err := database.DB.Exec(
					database.AdaptQuery(`INSERT INTO attendance (user_id, shift_date, clock_in_time, clock_out_time, status, is_late, penalty_hours, compensation_hours, is_auto_closed) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`),
					empID, d.Format("2006-01-02"),
					clockInTime.Format("2006-01-02 15:04:05"),
					clockOutTime.Format("2006-01-02 15:04:05"),
					"", isLate, penaltyHours, compensationHours, false,
				)
				if err == nil {
					totalInserted++
				}

			} else if randomChance < 0.95 {
				// Izin
				reasons := []string{"Urusan keluarga", "Keperluan pribadi", "Acara penting", "Izin dokter"}
				reason := reasons[rnd.Intn(len(reasons))]

				_, err := database.DB.Exec(
					database.AdaptQuery(`INSERT INTO attendance (user_id, shift_date, status, permit_reason) VALUES (?, ?, ?, ?)`),
					empID, d.Format("2006-01-02"), "IZIN: "+reason, reason,
				)
				if err == nil {
					totalInserted++
				}

			} else {
				// Sakit
				reasons := []string{"Flu", "Demam", "Sakit kepala", "Masuk angin"}
				reason := reasons[rnd.Intn(len(reasons))]

				_, err := database.DB.Exec(
					database.AdaptQuery(`INSERT INTO attendance (user_id, shift_date, status, permit_reason) VALUES (?, ?, ?, ?)`),
					empID, d.Format("2006-01-02"), "SAKIT: "+reason, reason,
				)
				if err == nil {
					totalInserted++
				}
			}
		}
	}

	log.Printf("✅ Seed complete: %d attendance records inserted\n", totalInserted)
}

// ---------------------------------------------------------
// MAIN
// ---------------------------------------------------------
func main() {
	database.InitDB()
	log.Println("✅ Database siap!")

	// Seed dummy data untuk testing (hanya jika belum ada)
	seedDummyData()

	fs := http.FileServer(http.Dir("assets"))
	http.Handle("/assets/", http.StripPrefix("/assets/", fs))

	// Routes Page
	http.HandleFunc("/", handleLoginView)
	http.HandleFunc("/login", handleLoginPost)
	http.HandleFunc("/register", handleRegisterPost)
	http.HandleFunc("/logout", handleLogout)
	http.HandleFunc("/dashboard", handleDashboard)
	http.HandleFunc("/admin", handleAdminDashboard)

	// Routes API
	http.HandleFunc("/api/clockin", handleAPIClockIn)
	http.HandleFunc("/api/clockout", handleAPIClockOut)
	http.HandleFunc("/api/permit", handleAPIPermit)
	http.HandleFunc("/api/back_to_work", handleAPIBackToWork)
	http.HandleFunc("/api/change_password", handleAPIChangePassword)
	http.HandleFunc("/api/active_shift", handleAPIActiveShift)

	// Routes Admin API
	http.HandleFunc("/admin/update_rate", handleAdminUpdateRate)
	http.HandleFunc("/admin/update_salary", handleAdminUpdateLogSalary)
	http.HandleFunc("/admin/reset_password", handleAdminResetPassword)
	http.HandleFunc("/admin/delete_log", handleAdminDeleteLog)
	http.HandleFunc("/admin/delete_user", handleAdminDeleteUser)
	http.HandleFunc("/admin/manage_accounts", handleAdminManageAccounts)
	http.HandleFunc("/admin/reports", handleAdminReports)

	// Routes Report API
	http.HandleFunc("/api/employees", handleAPIGetEmployees)
	http.HandleFunc("/api/salary_report", handleAPISalaryReport)
	http.HandleFunc("/api/activity_report", handleAPIActivityReport)

	// Routes PWA
	http.HandleFunc("/manifest.json", func(w http.ResponseWriter, r *http.Request) { http.ServeFile(w, r, "assets/manifest.json") })
	http.HandleFunc("/sw.js", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/javascript")
		http.ServeFile(w, r, "assets/sw.js")
	})

	// Get PORT from environment (untuk Koyeb/Railway) atau default 8080
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Start goroutine untuk auto clock out midnight
	go autoClockOutMidnight()

	log.Printf("☕ DhenPresence ready at http://localhost:%s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

// Auto clock out untuk shift malam di midnight
func autoClockOutMidnight() {
	for {
		now := time.Now()
		// Hitung waktu sampai midnight
		midnight := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
		duration := midnight.Sub(now)

		time.Sleep(duration)

		// Clock out semua yang masih aktif di midnight
		database.DB.Exec(database.AdaptQuery(`UPDATE attendance SET clock_out_time = ?, is_auto_closed = 1 WHERE clock_out_time IS NULL`), midnight)
		log.Println("⏰ Auto clock out midnight executed")
	}
}

// ---------------------------------------------------------
// HANDLERS AUTH
// ---------------------------------------------------------

func handleLoginView(w http.ResponseWriter, r *http.Request) {
	if _, err := r.Cookie("user_session"); err == nil {
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		return
	}
	http.ServeFile(w, r, "views/login.html")
}

func handleLoginPost(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		return
	}
	username := r.FormValue("username")
	password := r.FormValue("password")

	log.Printf("🔐 Login attempt - Username: %s", username)

	var dbPass, role, fullName string
	var userID int
	err := database.DB.QueryRow(database.AdaptQuery("SELECT id, password, role, full_name FROM users WHERE username = ?"), username).Scan(&userID, &dbPass, &role, &fullName)

	if err == sql.ErrNoRows {
		log.Printf("❌ Login Failed for '%s' - User not found in database", username)
		http.Redirect(w, r, "/?error=1", http.StatusSeeOther)
		return
	}

	if err != nil {
		log.Printf("❌ Login Failed for '%s' - Database error: %v", username, err)
		http.Redirect(w, r, "/?error=1", http.StatusSeeOther)
		return
	}

	if dbPass != password {
		log.Printf("❌ Login Failed for '%s' - Password mismatch. Expected: %s, Got: %s", username, dbPass, password)
		http.Redirect(w, r, "/?error=1", http.StatusSeeOther)
		return
	}

	log.Printf("✅ Login Success - User: %s (%s), Role: %s", username, fullName, role)

	http.SetCookie(w, &http.Cookie{Name: "user_session", Value: username, Path: "/", HttpOnly: true})

	if role == "owner" {
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
	} else {
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
	}
}

func handleRegisterPost(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		return
	}
	fullname := r.FormValue("fullname")
	phone := r.FormValue("phone")
	username := r.FormValue("reg_username")
	password := r.FormValue("reg_password")
	avatarURL := fmt.Sprintf("https://api.dicebear.com/7.x/adventurer/svg?seed=%s&backgroundColor=b6e3f4,c0aede,d1d4f9", url.QueryEscape(username))

	_, err := database.DB.Exec(database.AdaptQuery("INSERT INTO users (username, password, full_name, role, hourly_rate, phone_number, avatar_url) VALUES (?, ?, ?, 'employee', 10000, ?, ?)"), username, password, fullname, phone, avatarURL)

	if err != nil {
		log.Printf("❌ Register Failed for user '%s'. Error: %v", username, err)
		http.Redirect(w, r, "/?error=register_fail", http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, "/?success=registered", http.StatusSeeOther)
}

func handleLogout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{Name: "user_session", Value: "", Path: "/", MaxAge: -1})
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// ---------------------------------------------------------
// HANDLER DASHBOARD
// ---------------------------------------------------------
func handleDashboard(w http.ResponseWriter, r *http.Request) {
	c, err := r.Cookie("user_session")
	if err != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	var userID int
	var fName, avatar string
	var shiftID sql.NullInt64
	database.DB.QueryRow(database.AdaptQuery(`SELECT u.id, u.full_name, COALESCE(u.avatar_url, ''), a.id FROM users u LEFT JOIN attendance a ON u.id = a.user_id AND a.clock_out_time IS NULL WHERE u.username = ?`), c.Value).Scan(&userID, &fName, &avatar, &shiftID)
	if avatar == "" {
		avatar = "https://api.dicebear.com/7.x/adventurer/svg?seed=default"
	}

	permitStatus := ""
	permitReason := ""
	isSick := false
	var statusDB, reasonDB string
	err = database.DB.QueryRow(database.AdaptQuery(`SELECT status, COALESCE(permit_reason, '-') FROM attendance WHERE user_id = ? AND shift_date = ? AND (status LIKE 'IZIN%' OR status LIKE 'SAKIT%') AND status NOT LIKE 'DIBATALKAN%' ORDER BY id DESC LIMIT 1`), userID, time.Now().Format("2006-01-02")).Scan(&statusDB, &reasonDB)
	if err == nil {
		permitStatus = statusDB
		permitReason = reasonDB
		if strings.Contains(strings.ToUpper(permitStatus), "SAKIT") {
			isSick = true
		}
	}

	data := struct {
		FullName, PermitStatus, PermitReason, AvatarURL string
		IsWorking, IsSick                               bool
	}{fName, permitStatus, permitReason, avatar, shiftID.Valid, isSick}

	tmpl, _ := template.ParseFiles("views/dashboard.html")
	tmpl.Execute(w, data)
}

// ---------------------------------------------------------
// API HANDLERS
// ---------------------------------------------------------
func handleAPIClockIn(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		return
	}
	c, _ := r.Cookie("user_session")
	userLat, _ := strconv.ParseFloat(r.FormValue("lat"), 64)
	userLong, _ := strconv.ParseFloat(r.FormValue("long"), 64)
	shiftClean := strings.Replace(r.FormValue("shift"), ".", ":", -1)
	if calculateDistance(CAFE_LAT, CAFE_LONG, userLat, userLong) > MAX_DISTANCE_METER {
		jsonResponse(w, false, "Kejauhan! Harap mendekat ke Cafe.", "")
		return
	}
	var userID int
	database.DB.QueryRow(database.AdaptQuery("SELECT id FROM users WHERE username = ?"), c.Value).Scan(&userID)

	now := time.Now()
	parsed, _ := time.Parse("15:04", shiftClean)
	expected := time.Date(now.Year(), now.Month(), now.Day(), parsed.Hour(), parsed.Minute(), 0, 0, now.Location())

	// SHIFT WINDOW VALIDATION: ±3 jam dari waktu shift
	windowStart := expected.Add(-3 * time.Hour)
	windowEnd := expected.Add(3 * time.Hour)
	if now.Before(windowStart) || now.After(windowEnd) {
		jsonResponse(w, false, "❌ Belum waktunya! Clock in hanya bisa dilakukan 3 jam sebelum/sesudah shift.", "")
		return
	}

	// Auto clock out user sebelumnya dengan waktu clock in user baru
	database.DB.Exec(database.AdaptQuery(`UPDATE attendance SET clock_out_time = ?, is_auto_closed = 1 WHERE clock_out_time IS NULL AND user_id != ?`), now, userID)

	// HITUNG KETERLAMBATAN
	lateMinutes := now.Sub(expected).Minutes()
	isLate := false
	penalty := 0
	msg := "✅ Berhasil masuk!"

	if lateMinutes > 15 {
		isLate = true
		// Toleransi 15 menit: tidak ada penalty
		// 16-60 menit: potong 1 jam
		// 61+ menit: potong per jam (ceiling)
		if lateMinutes <= 60 {
			penalty = 1
			msg = "⚠️ TELAT " + fmt.Sprintf("%.0f", lateMinutes) + " menit! Potong gaji 1 jam."
		} else {
			// Terlambat 2 jam atau lebih: potong sesuai jam keterlambatan
			hoursLate := int(lateMinutes / 60)
			if int(lateMinutes)%60 > 0 {
				hoursLate++ // Ceiling
			}
			penalty = hoursLate
			msg = "⚠️ TELAT " + fmt.Sprintf("%.0f", lateMinutes) + " menit! Potong gaji " + fmt.Sprintf("%d", penalty) + " jam."
		}

		// TRANSFER OVERTIME KE USER SEBELUMNYA (User 1)
		// Nilai keterlambatan dioper ke user sebelumnya sebagai lembur
		database.DB.Exec(database.AdaptQuery(`
			UPDATE attendance 
			SET compensation_hours = compensation_hours + ? 
			WHERE id = (
				SELECT MAX(id) 
				FROM attendance 
				WHERE user_id != ? AND clock_out_time IS NOT NULL
			)
		`), penalty, userID)
	}

	// Insert record attendance baru
	database.DB.Exec(database.AdaptQuery(`
		INSERT INTO attendance (user_id, shift_date, clock_in_time, is_late, penalty_hours) 
		VALUES (?, ?, ?, ?, ?)
	`), userID, now.Format("2006-01-02"), now, isLate, penalty)

	jsonResponse(w, true, msg, now.Format("15:04"))
}
func handleAPIClockOut(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		return
	}
	c, _ := r.Cookie("user_session")
	var userID int
	database.DB.QueryRow(database.AdaptQuery("SELECT id FROM users WHERE username = ?"), c.Value).Scan(&userID)
	res, _ := database.DB.Exec(database.AdaptQuery(`UPDATE attendance SET clock_out_time = ? WHERE user_id = ? AND clock_out_time IS NULL`), time.Now(), userID)
	if rows, _ := res.RowsAffected(); rows == 0 {
		jsonResponse(w, false, "Belum Clock In!", "")
		return
	}
	jsonResponse(w, true, "Hati-hati di jalan! 👋", time.Now().Format("15:04"))
}
func handleAPIPermit(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		return
	}
	c, _ := r.Cookie("user_session")
	var userID int
	database.DB.QueryRow(database.AdaptQuery("SELECT id FROM users WHERE username = ?"), c.Value).Scan(&userID)
	stat := fmt.Sprintf("%s (%s)", r.FormValue("type"), r.FormValue("shift"))
	if r.FormValue("duration") == "FULL" {
		stat = fmt.Sprintf("%s (FULL DAY)", r.FormValue("type"))
	}
	database.DB.Exec(database.AdaptQuery(`INSERT INTO attendance (user_id, shift_date, clock_in_time, clock_out_time, status, permit_reason, manual_salary) VALUES (?, ?, ?, ?, ?, ?, 0)`), userID, time.Now().Format("2006-01-02"), time.Now(), time.Now(), stat, r.FormValue("reason"))
	jsonResponse(w, true, "Izin Terkirim. Semoga lancar!", "")
}

func handleAPIBackToWork(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		return
	}
	c, _ := r.Cookie("user_session")
	var userID int
	database.DB.QueryRow(database.AdaptQuery("SELECT id FROM users WHERE username = ?"), c.Value).Scan(&userID)

	// Update status menjadi DIBATALKAN agar tetap tercatat di log
	res, _ := database.DB.Exec(database.AdaptQuery(`UPDATE attendance SET status = 'DIBATALKAN - ' || status WHERE user_id = ? AND shift_date = ? AND (status LIKE 'IZIN%' OR status LIKE 'SAKIT%') AND status NOT LIKE 'DIBATALKAN%'`), userID, time.Now().Format("2006-01-02"))

	if rows, _ := res.RowsAffected(); rows == 0 {
		jsonResponse(w, false, "Tidak ada status izin/sakit hari ini.", "")
		return
	}

	jsonResponse(w, true, "Status izin/sakit dibatalkan. Selamat bekerja! 💪", "")
}

func handleAPIActiveShift(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		return
	}

	// Cek apakah ada shift yang aktif hari ini
	var shiftType string
	err := database.DB.QueryRow(database.AdaptQuery(`
		SELECT 
			CASE 
				WHEN clock_in_time >= ? AND clock_in_time < ? THEN 'long_pagi'
				WHEN clock_in_time >= ? THEN 'long_sore'
				ELSE 'normal'
			END as shift_type
		FROM attendance 
		WHERE shift_date = ? AND clock_out_time IS NULL
		ORDER BY clock_in_time DESC LIMIT 1
	`),
		time.Now().Format("2006-01-02")+" 08:00:00",
		time.Now().Format("2006-01-02")+" 09:00:00",
		time.Now().Format("2006-01-02")+" 16:00:00",
		time.Now().Format("2006-01-02"),
	).Scan(&shiftType)

	if err == sql.ErrNoRows {
		shiftType = "none"
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"active_shift": shiftType,
	})
}

func handleAPIChangePassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		return
	}
	c, _ := r.Cookie("user_session")
	var userID int
	var dbPassword string
	database.DB.QueryRow(database.AdaptQuery("SELECT id, password FROM users WHERE username = ?"), c.Value).Scan(&userID, &dbPassword)

	oldPassword := r.FormValue("old_password")
	newPassword := r.FormValue("new_password")
	confirmPassword := r.FormValue("confirm_password")

	// Validasi password lama
	if dbPassword != oldPassword {
		jsonResponse(w, false, "Password lama salah!", "")
		return
	}

	// Validasi password baru
	if newPassword == "" || len(newPassword) < 6 {
		jsonResponse(w, false, "Password baru minimal 6 karakter!", "")
		return
	}

	// Validasi konfirmasi password
	if newPassword != confirmPassword {
		jsonResponse(w, false, "Konfirmasi password tidak cocok!", "")
		return
	}

	// Update password
	database.DB.Exec(database.AdaptQuery("UPDATE users SET password = ? WHERE id = ?"), newPassword, userID)
	jsonResponse(w, true, "Password berhasil diubah!", "")
}

// ---------------------------------------------------------
// ADMIN HANDLERS
// ---------------------------------------------------------
func handleAdminDashboard(w http.ResponseWriter, r *http.Request) {
	c, err := r.Cookie("user_session")
	if err != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	var userID int
	var role, fName, userAva string
	var rate int64
	if err := database.DB.QueryRow(database.AdaptQuery("SELECT id, role, full_name, COALESCE(avatar_url, ''), hourly_rate FROM users WHERE username = ?"), c.Value).Scan(&userID, &role, &fName, &userAva, &rate); err != nil || role != "owner" {
		http.Error(w, "Access Denied", http.StatusForbidden)
		return
	}
	if userAva == "" {
		userAva = "https://api.dicebear.com/7.x/adventurer/svg?seed=owner"
	}

	rows, _ := database.DB.Query(database.AdaptQuery(`SELECT u.full_name, COALESCE(u.avatar_url, ''), a.clock_in_time FROM attendance a JOIN users u ON a.user_id = u.id WHERE a.clock_out_time IS NULL`))
	var live []WorkerData
	for rows.Next() {
		var n, av string
		var t time.Time
		rows.Scan(&n, &av, &t)
		if av == "" {
			av = "https://api.dicebear.com/7.x/adventurer/svg?seed=" + n
		}
		live = append(live, WorkerData{n, t.Format("15:04"), time.Since(t).Round(time.Minute).String(), av})
	}

	start := r.URL.Query().Get("start")
	end := r.URL.Query().Get("end")
	if start == "" {
		start = time.Now().AddDate(0, 0, -30).Format("2006-01-02")
		end = time.Now().Format("2006-01-02")
	}

	hRows, _ := database.DB.Query(database.AdaptQuery(`SELECT a.id, u.full_name, COALESCE(u.phone_number, '-'), COALESCE(u.avatar_url, ''), u.hourly_rate, a.shift_date, a.clock_in_time, a.clock_out_time, a.is_late, a.penalty_hours, a.compensation_hours, COALESCE(a.manual_salary, 0), COALESCE(a.status, ''), COALESCE(a.permit_reason, '-') FROM attendance a JOIN users u ON a.user_id = u.id WHERE a.clock_out_time IS NOT NULL AND a.shift_date BETWEEN ? AND ? ORDER BY a.clock_in_time DESC`), start, end)
	var hist []HistoryData
	var totSal int64
	for hRows.Next() {
		var id int
		var n, ph, av, d, stat, reas string
		var rt, pen, bon, man int64
		var in, out time.Time
		var late bool
		hRows.Scan(&id, &n, &ph, &av, &rt, &d, &in, &out, &late, &pen, &bon, &man, &stat, &reas)
		if av == "" {
			av = "https://api.dicebear.com/7.x/adventurer/svg?seed=" + n
		}
		var sal int64
		if man > 0 {
			sal = man
		} else if stat == "" || stat == "On Time" {
			dur := out.Sub(in).Hours()
			pay := dur - float64(pen) + float64(bon)
			if pay < 0 {
				pay = 0
			}
			sal = int64(pay * float64(rt))
		} else {
			sal = 0
		}
		totSal += sal
		disp := stat
		if disp == "" {
			disp = "✅ On Time"
			if late {
				disp = "⚠️ TELAT"
			}
			if bon > 0 {
				disp += " (+Lembur)"
			}
		}
		if man > 0 {
			disp += " (✏️)"
		}

		// Pakai formatRupiah biar formatnya Rp 100.000,-
		hist = append(hist, HistoryData{id, in.Format("02 Jan"), n, ph, fmt.Sprintf("%s - %s", in.Format("15:04"), out.Format("15:04")), av, fmt.Sprintf("%.1f Jam", out.Sub(in).Hours()), disp, reas, formatRupiah(sal), sal})
	}
	tmpl, _ := template.ParseFiles("views/admin.html")
	// Pakai formatRupiah juga buat total gaji
	tmpl.Execute(w, AdminData{fName, userAva, live, hist, formatRupiah(totSal), start, end, fmt.Sprintf("%d", rate)})
}

// ADMIN ACTIONS
func handleAdminResetPassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		return
	}
	c, _ := r.Cookie("user_session")
	var role string
	database.DB.QueryRow(database.AdaptQuery("SELECT role FROM users WHERE username = ?"), c.Value).Scan(&role)
	if role != "owner" {
		jsonResponse(w, false, "Denied", "")
		return
	}
	newPassword := r.FormValue("password")
	targetUserID := r.FormValue("id")

	if newPassword == "" || len(newPassword) < 6 {
		jsonResponse(w, false, "Password minimal 6 karakter!", "")
		return
	}

	// Update password dengan query yang benar
	result, err := database.DB.Exec(database.AdaptQuery("UPDATE users SET password = ? WHERE id = ?"), newPassword, targetUserID)
	if err != nil {
		log.Printf("Error updating password: %v", err)
		jsonResponse(w, false, "Error: "+err.Error(), "")
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		jsonResponse(w, false, "User tidak ditemukan!", "")
		return
	}

	jsonResponse(w, true, "Password berhasil direset!", "")
}

func handleAdminDeleteLog(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		return
	}
	c, _ := r.Cookie("user_session")
	var role string
	database.DB.QueryRow(database.AdaptQuery("SELECT role FROM users WHERE username = ?"), c.Value).Scan(&role)
	if role != "owner" {
		jsonResponse(w, false, "Denied", "")
		return
	}

	_, err := database.DB.Exec(database.AdaptQuery("DELETE FROM attendance WHERE id = ?"), r.FormValue("id"))
	if err != nil {
		jsonResponse(w, false, "Error", "")
		return
	}
	jsonResponse(w, true, "Data berhasil dihapus selamanya.", "")
}

func handleAdminDeleteUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		return
	}
	c, _ := r.Cookie("user_session")
	var role string
	database.DB.QueryRow(database.AdaptQuery("SELECT role FROM users WHERE username = ?"), c.Value).Scan(&role)
	if role != "owner" {
		jsonResponse(w, false, "Denied", "")
		return
	}

	userID := r.FormValue("id")

	// Cek apakah user yang akan dihapus adalah owner
	var targetRole string
	database.DB.QueryRow(database.AdaptQuery("SELECT role FROM users WHERE id = ?"), userID).Scan(&targetRole)
	if targetRole == "owner" {
		jsonResponse(w, false, "Tidak dapat menghapus akun owner!", "")
		return
	}

	// Hapus semua attendance records user ini dulu
	database.DB.Exec(database.AdaptQuery("DELETE FROM attendance WHERE user_id = ?"), userID)

	// Hapus user
	_, err := database.DB.Exec(database.AdaptQuery("DELETE FROM users WHERE id = ?"), userID)
	if err != nil {
		jsonResponse(w, false, "Error: "+err.Error(), "")
		return
	}
	jsonResponse(w, true, "Akun berhasil dihapus!", "")
}

func handleAdminManageAccounts(w http.ResponseWriter, r *http.Request) {
	c, err := r.Cookie("user_session")
	if err != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	var role string
	database.DB.QueryRow(database.AdaptQuery("SELECT role FROM users WHERE username = ?"), c.Value).Scan(&role)
	if role != "owner" {
		http.Error(w, "Access Denied", http.StatusForbidden)
		return
	}

	// Ambil semua users
	rows, _ := database.DB.Query(database.AdaptQuery(`SELECT id, username, full_name, role, COALESCE(phone_number, '-'), COALESCE(avatar_url, '') FROM users ORDER BY role DESC, full_name ASC`))
	type UserAccount struct {
		ID                                      int
		Username, FullName, Role, Phone, Avatar string
	}
	var users []UserAccount
	for rows.Next() {
		var u UserAccount
		rows.Scan(&u.ID, &u.Username, &u.FullName, &u.Role, &u.Phone, &u.Avatar)
		if u.Avatar == "" {
			u.Avatar = "https://api.dicebear.com/7.x/adventurer/svg?seed=" + u.Username
		}
		users = append(users, u)
	}

	tmpl, _ := template.ParseFiles("views/manage_accounts.html")
	tmpl.Execute(w, users)
}

func handleAdminUpdateRate(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		n, _ := strconv.Atoi(r.FormValue("rate"))
		database.DB.Exec(database.AdaptQuery("UPDATE users SET hourly_rate = ?"), n)
		jsonResponse(w, true, "Updated", "")
	}
}
func handleAdminUpdateLogSalary(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		n, _ := strconv.Atoi(r.FormValue("salary"))
		database.DB.Exec(database.AdaptQuery("UPDATE attendance SET manual_salary = ? WHERE id = ?"), n, r.FormValue("id"))
		jsonResponse(w, true, "Updated", "")
	}
}
func jsonResponse(w http.ResponseWriter, s bool, m, t string) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(Response{s, m, t})
}
func calculateDistance(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371000
	dLat := (lat2 - lat1) * (math.Pi / 180.0)
	dLon := (lon2 - lon1) * (math.Pi / 180.0)
	a := math.Sin(dLat/2)*math.Sin(dLat/2) + math.Cos(lat1*(math.Pi/180.0))*math.Cos(lat2*(math.Pi/180.0))*math.Sin(dLon/2)*math.Sin(dLon/2)
	return R * 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
}

// --- FUNGSI FORMAT RUPIAH BARU ---
func formatRupiah(amount int64) string {
	str := strconv.FormatInt(amount, 10)
	n := len(str)
	if n <= 3 {
		return "Rp " + str + ",-"
	}
	var res []byte
	rem := n % 3
	if rem > 0 {
		res = append(res, str[:rem]...)
		res = append(res, '.')
	}
	for i := rem; i < n; i += 3 {
		if i > rem {
			res = append(res, '.')
		}
		res = append(res, str[i:i+3]...)
	}
	return "Rp " + string(res) + ",-"
}

// ---------------------------------------------------------
// HANDLER ADMIN REPORTS
// ---------------------------------------------------------
func handleAdminReports(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("user_session")
	if err != nil || cookie.Value == "" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	var role, fullName, avatar string
	err = database.DB.QueryRow(database.AdaptQuery("SELECT role, full_name, COALESCE(avatar_url, '') FROM users WHERE username = ?"), cookie.Value).Scan(&role, &fullName, &avatar)
	if err != nil || role != "owner" {
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		return
	}

	// Set default avatar if empty
	if avatar == "" {
		avatar = "https://api.dicebear.com/7.x/adventurer/svg?seed=" + fullName
	}

	tmpl, err := template.ParseFiles("views/admin_reports.html")
	if err != nil {
		http.Error(w, "Template error", 500)
		return
	}

	data := struct {
		UserFullName string
		UserAvatar   string
	}{
		UserFullName: fullName,
		UserAvatar:   avatar,
	}

	tmpl.Execute(w, data)
}

// ---------------------------------------------------------
// API GET EMPLOYEES
// ---------------------------------------------------------
func handleAPIGetEmployees(w http.ResponseWriter, r *http.Request) {
	rows, err := database.DB.Query(database.AdaptQuery("SELECT id, full_name FROM users WHERE role != 'owner' ORDER BY full_name"))
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": "Database error"})
		return
	}
	defer rows.Close()

	type Employee struct {
		ID       int    `json:"id"`
		FullName string `json:"full_name"`
	}

	var employees []Employee
	for rows.Next() {
		var emp Employee
		rows.Scan(&emp.ID, &emp.FullName)
		employees = append(employees, emp)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":   true,
		"employees": employees,
	})
}

// ---------------------------------------------------------
// API SALARY REPORT (LAPORAN GAJI INDIVIDU)
// ---------------------------------------------------------
func handleAPISalaryReport(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	startDate := r.URL.Query().Get("start")
	endDate := r.URL.Query().Get("end")

	if userID == "" || startDate == "" || endDate == "" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": "Missing parameters"})
		return
	}

	// Get employee info
	var fullName, phoneNumber, avatarURL string
	var hourlyRate int
	err := database.DB.QueryRow(database.AdaptQuery("SELECT full_name, phone_number, hourly_rate, COALESCE(avatar_url, '') FROM users WHERE id = ?"), userID).Scan(&fullName, &phoneNumber, &hourlyRate, &avatarURL)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": "User not found"})
		return
	}

	// Set default avatar if empty
	if avatarURL == "" {
		avatarURL = "https://api.dicebear.com/7.x/adventurer/svg?seed=" + fullName
	}

	// Get attendance records
	query := database.AdaptQuery(`
		SELECT shift_date, clock_in_time, clock_out_time, status, permit_reason, 
		       is_late, penalty_hours, compensation_hours, manual_salary
		FROM attendance
		WHERE user_id = ? AND shift_date BETWEEN ? AND ?
		ORDER BY shift_date ASC
	`)

	rows, err := database.DB.Query(query, userID, startDate, endDate)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": "Database error"})
		return
	}
	defer rows.Close()

	type DailyDetail struct {
		Date        string  `json:"date"`
		Shift       string  `json:"shift"`
		ClockIn     string  `json:"clock_in"`
		ClockOut    string  `json:"clock_out"`
		Hours       float64 `json:"hours"`
		HoursFormat string  `json:"hours_format"`
		Status      string  `json:"status"`
		Salary      int64   `json:"salary"`
	}

	var details []DailyDetail
	var totalDays, onTimeCount, lateCount, permitCount, sickCount int
	var totalHours, totalPenaltyHours, totalBonusHours float64

	for rows.Next() {
		var shiftDate, status, permitReason string
		var clockInTime, clockOutTime sql.NullString
		var isLate sql.NullBool
		var penaltyHours, compensationHours, manualSalary sql.NullInt64

		rows.Scan(&shiftDate, &clockInTime, &clockOutTime, &status, &permitReason, &isLate, &penaltyHours, &compensationHours, &manualSalary)

		totalDays++

		var hours float64
		var clockInStr, clockOutStr string
		var shiftType string

		if clockInTime.Valid {
			clockInStr = clockInTime.String[11:16] // HH:MM

			// Determine shift type from time
			hour, _ := time.Parse("15:04:05", clockInTime.String[11:])
			if hour.Hour() >= 6 && hour.Hour() < 12 {
				shiftType = "Pagi"
			} else if hour.Hour() >= 12 && hour.Hour() < 18 {
				shiftType = "Sore"
			} else {
				shiftType = "Malam"
			}
		}

		if clockOutTime.Valid {
			clockOutStr = clockOutTime.String[11:16]

			// Calculate hours (dengan menit/detik seperti di dashboard)
			if clockInTime.Valid {
				clockIn, _ := time.Parse("2006-01-02 15:04:05", clockInTime.String)
				clockOut, _ := time.Parse("2006-01-02 15:04:05", clockOutTime.String)

				// Handle overnight shifts
				if clockOut.Before(clockIn) {
					clockOut = clockOut.Add(24 * time.Hour)
				}

				hours = clockOut.Sub(clockIn).Hours()
				totalHours += hours
			}
		}

		// Count status
		if status == "Sakit" {
			sickCount++
		} else if status == "Izin" {
			permitCount++
		} else {
			if isLate.Valid && isLate.Bool {
				lateCount++
			} else {
				onTimeCount++
			}
		}

		// Add penalty and bonus
		if penaltyHours.Valid {
			totalPenaltyHours += float64(penaltyHours.Int64)
		}
		if compensationHours.Valid {
			totalBonusHours += float64(compensationHours.Int64)
		}

		// Determine display status
		displayStatus := "✅ Tepat Waktu"
		if status == "Sakit" {
			displayStatus = "🏥 Sakit"
		} else if status == "Izin" {
			displayStatus = "📝 Izin"
		} else if isLate.Valid && isLate.Bool {
			displayStatus = "⚠️ TELAT"
		}

		// Calculate salary for this day (sama seperti dashboard)
		var daySalary int64
		if status != "Sakit" && status != "Izin" && hours > 0 {
			effectiveHours := hours - float64(penaltyHours.Int64)
			if compensationHours.Valid {
				effectiveHours += float64(compensationHours.Int64)
			}
			if effectiveHours < 0 {
				effectiveHours = 0
			}
			daySalary = int64(effectiveHours * float64(hourlyRate))
		}

		// Format hours seperti dashboard: "8.5 Jam"
		hoursFormat := "-"
		if hours > 0 {
			hoursFormat = fmt.Sprintf("%.1f Jam", hours)
		}

		details = append(details, DailyDetail{
			Date:        shiftDate,
			Shift:       shiftType,
			ClockIn:     clockInStr,
			ClockOut:    clockOutStr,
			Hours:       hours,
			HoursFormat: hoursFormat,
			Status:      displayStatus,
			Salary:      daySalary,
		})
	}

	// Calculate salaries
	grossSalary := int64(totalHours * float64(hourlyRate))
	penaltyAmount := int64(totalPenaltyHours * float64(hourlyRate))
	bonusAmount := int64(totalBonusHours * float64(hourlyRate))
	netSalary := grossSalary - penaltyAmount + bonusAmount

	// Calculate attendance rate
	presentDays := onTimeCount + lateCount
	attendanceRate := 0
	if totalDays > 0 {
		attendanceRate = (presentDays * 100) / totalDays
	}

	response := map[string]interface{}{
		"success": true,
		"employee": map[string]interface{}{
			"full_name":    fullName,
			"phone_number": phoneNumber,
			"hourly_rate":  hourlyRate,
			"avatar_url":   avatarURL,
		},
		"period": startDate + " s/d " + endDate,
		"summary": map[string]interface{}{
			"total_days":      totalDays,
			"total_hours":     totalHours,
			"attendance_rate": attendanceRate,
			"on_time_count":   onTimeCount,
			"late_count":      lateCount,
			"permit_count":    permitCount,
			"sick_count":      sickCount,
			"gross_salary":    grossSalary,
			"penalty_amount":  penaltyAmount,
			"bonus_amount":    bonusAmount,
			"net_salary":      netSalary,
		},
		"details": details,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ---------------------------------------------------------
// API ACTIVITY REPORT (REPORT AKTIVITAS)
// ---------------------------------------------------------
func handleAPIActivityReport(w http.ResponseWriter, r *http.Request) {
	startDate := r.URL.Query().Get("start")
	endDate := r.URL.Query().Get("end")

	if startDate == "" || endDate == "" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": "Missing parameters"})
		return
	}

	// Get daily statistics
	dailyQuery := database.AdaptQuery(`
		SELECT 
			shift_date,
			COUNT(*) as total_attendance,
			COUNT(CASE WHEN status NOT IN ('Izin', 'Sakit') THEN 1 END) as present_count,
			COUNT(CASE WHEN status = 'Izin' THEN 1 END) as permit_count,
			COUNT(CASE WHEN status = 'Sakit' THEN 1 END) as sick_count
		FROM attendance
		WHERE shift_date BETWEEN ? AND ?
		GROUP BY shift_date
		ORDER BY shift_date ASC
	`)

	dailyRows, err := database.DB.Query(dailyQuery, startDate, endDate)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": "Database error"})
		return
	}
	defer dailyRows.Close()

	type DailyStat struct {
		Date                 string  `json:"date"`
		EmployeeCount        int     `json:"employee_count"`
		PresentCount         int     `json:"present_count"`
		PermitCount          int     `json:"permit_count"`
		SickCount            int     `json:"sick_count"`
		TotalHours           float64 `json:"total_hours"`
		AttendancePercentage int     `json:"attendance_percentage"`
		LongShiftCount       int     `json:"long_shift_count"`
	}

	var dailyStats []DailyStat
	var totalAttendance, totalPresent, totalPermit, totalSick int
	var grandTotalHours float64

	for dailyRows.Next() {
		var date string
		var total, present, permit, sick int
		dailyRows.Scan(&date, &total, &present, &permit, &sick)

		// Calculate total hours for this day
		var dayHours float64
		hoursQuery := database.AdaptQuery(`
			SELECT SUM(
				CASE 
					WHEN clock_out_time IS NOT NULL THEN 
						CAST((julianday(clock_out_time) - julianday(clock_in_time)) * 24 AS REAL)
					ELSE 0 
				END
			)
			FROM attendance
			WHERE shift_date = ? AND status NOT IN ('Izin', 'Sakit')
		`)
		database.DB.QueryRow(hoursQuery, date).Scan(&dayHours)

		attendancePercentage := 0
		if total > 0 {
			attendancePercentage = (present * 100) / total
		}

		dailyStats = append(dailyStats, DailyStat{
			Date:                 date,
			EmployeeCount:        total,
			PresentCount:         present,
			PermitCount:          permit,
			SickCount:            sick,
			TotalHours:           dayHours,
			AttendancePercentage: attendancePercentage,
			LongShiftCount:       0, // TODO: implement long shift detection
		})

		totalAttendance += total
		totalPresent += present
		totalPermit += permit
		totalSick += sick
		grandTotalHours += dayHours
	}

	// Calculate average hours per day
	avgHoursPerDay := 0.0
	if len(dailyStats) > 0 {
		avgHoursPerDay = grandTotalHours / float64(len(dailyStats))
	}

	// Get employee rankings
	rankingQuery := database.AdaptQuery(`
		SELECT 
			u.id,
			u.full_name,
			u.hourly_rate,
			COUNT(*) as total_days,
			COUNT(CASE WHEN a.status NOT IN ('Izin', 'Sakit') THEN 1 END) as present_days
		FROM users u
		LEFT JOIN attendance a ON u.id = a.user_id AND a.shift_date BETWEEN ? AND ?
		WHERE u.role != 'owner'
		GROUP BY u.id, u.full_name, u.hourly_rate
		HAVING total_days > 0
	`)

	rankingRows, err := database.DB.Query(rankingQuery, startDate, endDate)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": "Database error"})
		return
	}
	defer rankingRows.Close()

	type Ranking struct {
		FullName       string  `json:"full_name"`
		TotalDays      int     `json:"total_days"`
		PresentDays    int     `json:"present_days"`
		AttendanceRate int     `json:"attendance_rate"`
		TotalHours     float64 `json:"total_hours"`
		TotalSalary    int64   `json:"total_salary"`
	}

	var rankings []Ranking
	var totalSalary int64

	for rankingRows.Next() {
		var userID int
		var fullName string
		var hourlyRate, totalDays, presentDays int
		rankingRows.Scan(&userID, &fullName, &hourlyRate, &totalDays, &presentDays)

		// Calculate total hours for this user
		var userHours float64
		userHoursQuery := database.AdaptQuery(`
			SELECT COALESCE(SUM(
				CASE 
					WHEN clock_out_time IS NOT NULL THEN 
						CAST((julianday(clock_out_time) - julianday(clock_in_time)) * 24 AS REAL)
					ELSE 0 
				END
			), 0)
			FROM attendance
			WHERE user_id = ? AND shift_date BETWEEN ? AND ? AND status NOT IN ('Izin', 'Sakit')
		`)
		database.DB.QueryRow(userHoursQuery, userID, startDate, endDate).Scan(&userHours)

		// Calculate salary (hours * rate)
		userSalary := int64(userHours * float64(hourlyRate))

		attendanceRate := 0
		if totalDays > 0 {
			attendanceRate = (presentDays * 100) / totalDays
		}

		rankings = append(rankings, Ranking{
			FullName:       fullName,
			TotalDays:      totalDays,
			PresentDays:    presentDays,
			AttendanceRate: attendanceRate,
			TotalHours:     userHours,
			TotalSalary:    userSalary,
		})

		totalSalary += userSalary
	}

	// Sort rankings by attendance rate (descending)
	for i := 0; i < len(rankings)-1; i++ {
		for j := i + 1; j < len(rankings); j++ {
			if rankings[j].AttendanceRate > rankings[i].AttendanceRate {
				rankings[i], rankings[j] = rankings[j], rankings[i]
			}
		}
	}

	// Generate insights
	type Insight struct {
		Icon    string `json:"icon"`
		Title   string `json:"title"`
		Message string `json:"message"`
	}

	var insights []Insight

	// Find employees with 5+ late occurrences
	lateQuery := database.AdaptQuery(`
		SELECT u.full_name, COUNT(*) as late_count
		FROM users u
		JOIN attendance a ON u.id = a.user_id
		WHERE a.shift_date BETWEEN ? AND ? AND a.is_late = 1
		GROUP BY u.id, u.full_name
		HAVING late_count >= 5
		ORDER BY late_count DESC
	`)

	lateRows, _ := database.DB.Query(lateQuery, startDate, endDate)
	var lateEmployees []string
	for lateRows.Next() {
		var name string
		var count int
		lateRows.Scan(&name, &count)
		lateEmployees = append(lateEmployees, fmt.Sprintf("%s (%dx)", name, count))
	}
	lateRows.Close()

	if len(lateEmployees) > 0 {
		insights = append(insights, Insight{
			Icon:    "fa-triangle-exclamation",
			Title:   "Keterlambatan Berulang",
			Message: fmt.Sprintf("%d karyawan dengan keterlambatan ≥5x: %s", len(lateEmployees), strings.Join(lateEmployees, ", ")),
		})
	}

	// Find day with lowest attendance
	if len(dailyStats) > 0 {
		minAttendance := dailyStats[0]
		for _, day := range dailyStats {
			if day.AttendancePercentage < minAttendance.AttendancePercentage {
				minAttendance = day
			}
		}
		if minAttendance.AttendancePercentage < 80 {
			insights = append(insights, Insight{
				Icon:    "fa-calendar-xmark",
				Title:   "Kehadiran Rendah",
				Message: fmt.Sprintf("Hari dengan kehadiran terendah: %s (%d%%)", minAttendance.Date, minAttendance.AttendancePercentage),
			})
		}
	}

	// Average attendance rate
	if len(dailyStats) > 0 {
		avgAttendance := 0
		for _, day := range dailyStats {
			avgAttendance += day.AttendancePercentage
		}
		avgAttendance /= len(dailyStats)

		if avgAttendance >= 90 {
			insights = append(insights, Insight{
				Icon:    "fa-trophy",
				Title:   "Performa Excellent!",
				Message: fmt.Sprintf("Rata-rata kehadiran %d%% - Tim sangat disiplin!", avgAttendance),
			})
		} else if avgAttendance < 70 {
			insights = append(insights, Insight{
				Icon:    "fa-chart-line-down",
				Title:   "Perhatian Diperlukan",
				Message: fmt.Sprintf("Rata-rata kehadiran hanya %d%% - Perlu evaluasi dan perbaikan", avgAttendance),
			})
		}
	}

	response := map[string]interface{}{
		"success": true,
		"overview": map[string]interface{}{
			"total_attendance":  totalAttendance,
			"present_count":     totalPresent,
			"permit_count":      totalPermit,
			"sick_count":        totalSick,
			"total_hours":       grandTotalHours,
			"avg_hours_per_day": avgHoursPerDay,
			"total_salary":      totalSalary,
		},
		"daily_stats": dailyStats,
		"rankings":    rankings,
		"insights":    insights,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
