package main

import (
	"database/sql"
	"dhenpresence/database"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"math"
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
// MAIN
// ---------------------------------------------------------
func main() {
	database.InitDB()
	log.Println("✅ Database siap!")

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

	// Routes Admin API
	http.HandleFunc("/admin/update_rate", handleAdminUpdateRate)
	http.HandleFunc("/admin/update_salary", handleAdminUpdateLogSalary)
	http.HandleFunc("/admin/reset_password", handleAdminResetPassword)
	http.HandleFunc("/admin/delete_log", handleAdminDeleteLog)
	http.HandleFunc("/admin/delete_user", handleAdminDeleteUser)
	http.HandleFunc("/admin/manage_accounts", handleAdminManageAccounts)

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

	log.Printf("☕ DhenPresence ready at http://localhost:%s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
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

	var dbPass, role string
	err := database.DB.QueryRow(database.AdaptQuery("SELECT password, role FROM users WHERE username = ?"), username).Scan(&dbPass, &role)

	if err == sql.ErrNoRows || dbPass != password {
		log.Printf("❌ Login Failed for user '%s'. DB Error: %v. Password Match: %v", username, err, dbPass == password)
		http.Redirect(w, r, "/?error=1", http.StatusSeeOther)
		return
	}

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
	database.DB.Exec(database.AdaptQuery(`UPDATE attendance SET clock_out_time = ?, is_auto_closed = 1 WHERE clock_out_time IS NULL AND user_id != ?`), time.Now(), userID)
	isLate := false
	penalty := 0
	now := time.Now()
	parsed, _ := time.Parse("15:04", shiftClean)
	expected := time.Date(now.Year(), now.Month(), now.Day(), parsed.Hour(), parsed.Minute(), 0, 0, now.Location())
	if now.Sub(expected).Minutes() > 15 {
		isLate = true
		penalty = 1
		database.DB.Exec(database.AdaptQuery(`UPDATE attendance SET compensation_hours = compensation_hours + 1 WHERE id = (SELECT MAX(id) FROM attendance WHERE user_id != ?)`), userID)
	}
	database.DB.Exec(database.AdaptQuery(`INSERT INTO attendance (user_id, shift_date, clock_in_time, is_late, penalty_hours) VALUES (?, ?, ?, ?, ?)`), userID, now.Format("2006-01-02"), now, isLate, penalty)
	msg := "Berhasil masuk!"
	if isLate {
		msg = "⚠️ TELAT! Potong gaji 1 jam."
	}
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
