# 📱 Panduan Penggunaan DhenPresence
**Sistem Absensi Karyawan Dhen Coffee**

---

## 📋 Daftar Isi
1. [Pengenalan Aplikasi](#pengenalan-aplikasi)
2. [Instalasi di HP (iOS & Android)](#instalasi-di-hp)
3. [Panduan untuk Karyawan](#panduan-untuk-karyawan)
4. [Panduan untuk Super Admin](#panduan-untuk-super-admin)
5. [FAQ & Troubleshooting](#faq--troubleshooting)

---

## 🎯 Pengenalan Aplikasi

**DhenPresence** adalah aplikasi absensi berbasis web (PWA) yang dirancang khusus untuk Dhen Coffee. Aplikasi ini memungkinkan karyawan untuk clock in/out dengan validasi GPS dan admin untuk mengelola data kehadiran serta menghitung gaji otomatis.

### Fitur Utama:
- ✅ Clock In/Out dengan GPS validation
- ✅ 5 Jenis Shift (Pagi, Sore, Malam, Long Pagi, Long Sore)
- ✅ Sistem izin (Permit) dan kembali kerja (Back to Work)
- ✅ Shift Window Policy (±3 jam dari waktu shift)
- ✅ Penalty otomatis untuk keterlambatan
- ✅ Kompensasi lembur otomatis
- ✅ PWA - bisa diinstall di HP seperti aplikasi native
- ✅ Offline support dengan Service Worker
- ✅ Dashboard admin dengan laporan lengkap

### URL Akses:
🌐 **Production**: https://electoral-sybila-polteknuklir-6228667e.koyeb.app/

---

## 📲 Instalasi di HP

### **iOS (iPhone/iPad):**
1. Buka Safari browser
2. Akses URL: https://electoral-sybila-polteknuklir-6228667e.koyeb.app/
3. Tap tombol **Share** (kotak dengan panah ke atas)
4. Scroll ke bawah, pilih **"Add to Home Screen"**
5. Edit nama (opsional), tap **"Add"**
6. Icon DhenPresence akan muncul di Home Screen ✅

**📸 Screenshot iOS - Install:**
```
[Ambil screenshot: Safari > Share button > Add to Home Screen]
[Ambil screenshot: Icon DhenPresence di Home Screen iOS]
```

### **Android:**
1. Buka Chrome browser
2. Akses URL: https://electoral-sybila-polteknuklir-6228667e.koyeb.app/
3. Tap menu titik tiga (⋮) di pojok kanan atas
4. Pilih **"Add to Home screen"** atau **"Install app"**
5. Tap **"Install"**
6. Icon DhenPresence akan muncul di Home Screen ✅

**📸 Screenshot Android - Install:**
```
[Ambil screenshot: Chrome menu > Add to Home screen]
[Ambil screenshot: Icon DhenPresence di Home Screen Android]
```

---

## 👤 Panduan untuk Karyawan

### 1️⃣ **Login**

1. Buka aplikasi DhenPresence
2. Masukkan **Username** dan **Password** yang diberikan admin
3. Klik tombol **"MASUK"**
4. Jika berhasil, akan masuk ke halaman **Dashboard**

**📸 Screenshot - Halaman Login:**
```
[Ambil screenshot: Halaman login dengan form username & password]
```

**Catatan Penting:**
- ⚠️ Username dan password **case-sensitive** (huruf besar/kecil berpengaruh)
- ⚠️ Pastikan tidak ada spasi sebelum/sesudah username/password
- ⚠️ Di HP, matikan autocapitalize keyboard untuk menghindari kesalahan

---

### 2️⃣ **Dashboard Karyawan**

Setelah login, kamu akan melihat dashboard dengan:

#### **A. Header Info:**
- 🕒 **Jam digital** - update real-time setiap detik
- 📅 **Tanggal lengkap** - hari, tanggal, bulan, tahun
- 👤 **Nama & avatar** karyawan
- 🚪 **Tombol Logout** - untuk keluar dari aplikasi

**📸 Screenshot - Dashboard Header:**
```
[Ambil screenshot: Header dashboard dengan jam, tanggal, nama user]
```

#### **B. Pilihan Shift:**

Ada **5 jenis shift** yang bisa dipilih:

| Shift | Waktu Mulai | Keterangan |
|-------|-------------|------------|
| **Pagi** | 08:00 WIB | Shift normal pagi (8-9 jam) |
| **Sore** | 14:40 WIB | Shift normal sore (6-7 jam) |
| **Malam** | 19:40 WIB | Shift normal malam (5-6 jam) |
| **Long Pagi** | 08:00 WIB | Shift panjang pagi (12-13 jam) |
| **Long Sore** | 14:40 WIB | Shift panjang sore (12-13 jam) |

**Cara Pilih Shift:**
1. Klik salah satu tombol shift (contoh: **Pagi**)
2. Tombol akan berubah warna menjadi **hijau gelap** (terpilih)
3. Tombol lain akan menjadi abu-abu (tidak aktif)

**📸 Screenshot - Pilihan Shift:**
```
[Ambil screenshot: 5 tombol shift (Pagi, Sore, Malam, Long Pagi, Long Sore)]
[Ambil screenshot: Tombol Pagi terpilih (hijau gelap)]
```

**⚠️ Aturan Long Shift:**
- Jika ada karyawan yang sedang **Long Pagi**, maka hanya **Long Sore** yang bisa dipilih selanjutnya
- Shift **Pagi, Sore, Malam** akan **disabled** (tidak bisa diklik)
- Tujuan: Memastikan shift panjang tidak tumpang tindih dengan shift pendek

---

### 3️⃣ **Clock In (Masuk Kerja)**

#### **Langkah-langkah Clock In:**

1. **Pilih shift** yang sesuai dengan jadwal kerja kamu
2. Pastikan kamu berada di **lokasi cafe** (dalam radius 50 meter)
3. Klik tombol **"CLOCK IN"** berwarna hijau
4. Browser akan meminta **izin lokasi** → Klik **"Allow"** atau **"Izinkan"**
5. Tunggu proses validasi GPS (1-3 detik)
6. Jika berhasil, akan muncul notifikasi **"✅ Clock In Success!"**

**📸 Screenshot - Clock In:**
```
[Ambil screenshot: Tombol CLOCK IN (hijau) setelah memilih shift]
[Ambil screenshot: Browser meminta izin lokasi]
[Ambil screenshot: Popup sukses "Clock In Success!"]
```

#### **Sistem Shift Window (±3 Jam):**

Clock in hanya bisa dilakukan **±3 jam** dari waktu shift:

**Contoh untuk Shift Pagi (08:00 WIB):**
- ✅ Window dibuka: **05:00 - 11:00 WIB**
- ❌ Di luar window: Tidak bisa clock in

**Contoh Skenario:**

| Waktu Clock In | Shift Pagi (08:00) | Status | Penalty |
|----------------|---------------------|--------|---------|
| 05:00 | ✅ Boleh | Masuk tepat waktu | 0 jam |
| 08:00 | ✅ Boleh | Masuk tepat waktu | 0 jam |
| 08:10 | ✅ Boleh | Telat 10 menit (toleransi) | 0 jam |
| 08:45 | ✅ Boleh | ⚠️ Telat 45 menit | **-1 jam gaji** |
| 10:00 | ✅ Boleh | ⚠️ Telat 120 menit | **-2 jam gaji** |
| 11:30 | ❌ **DITOLAK** | Di luar window | - |
| 04:00 | ❌ **DITOLAK** | Di luar window | - |

**📸 Screenshot - Validasi:**
```
[Ambil screenshot: Pesan error "❌ Belum waktunya! Clock in hanya bisa dilakukan 3 jam sebelum/sesudah shift"]
[Ambil screenshot: Pesan sukses dengan info keterlambatan "⚠️ TELAT 45 menit! Potong gaji 1 jam"]
```

#### **Penalty Keterlambatan:**

Sistem penalty bertingkat:

1. **0-15 menit telat**: Tidak ada penalty (toleransi)
2. **16-60 menit telat**: Potong gaji **1 jam**
3. **61-120 menit telat**: Potong gaji **2 jam**
4. **121+ menit telat**: Potong gaji sesuai jam keterlambatan (ceiling)

**Kompensasi Lembur:**
- Penalty dari user yang telat akan **ditransfer** sebagai **jam lembur** ke user sebelumnya yang clock out
- Contoh: User A telat 45 menit → potong gaji 1 jam → User sebelumnya dapat +1 jam lembur

---

### 4️⃣ **Clock Out (Pulang Kerja)**

#### **Langkah-langkah Clock Out:**

1. Pastikan kamu sudah **clock in** sebelumnya
2. Pastikan masih berada di **lokasi cafe**
3. Klik tombol **"CLOCK OUT"** berwarna merah
4. Browser akan validasi GPS (1-3 detik)
5. Jika berhasil, akan muncul notifikasi **"✅ Clock Out Success!"**
6. Total jam kerja akan dihitung otomatis

**📸 Screenshot - Clock Out:**
```
[Ambil screenshot: Tombol CLOCK OUT (merah) setelah clock in]
[Ambil screenshot: Popup sukses "Clock Out Success!" dengan total jam kerja]
```

#### **Auto Clock Out:**

Sistem memiliki 2 mekanisme auto clock out:

1. **Clock Out Otomatis saat User Lain Clock In**
   - Jika user A sedang bekerja dan user B clock in → User A auto clock out
   - Tujuan: Mencegah overlap shift

2. **Auto Clock Out Tengah Malam (00:00 WIB)**
   - Setiap tengah malam, sistem akan auto clock out semua user yang belum clock out
   - Flag: `is_auto_closed = true`
   - Tujuan: Mencegah shift terbuka selamanya

**📸 Screenshot - Auto Clock Out:**
```
[Ambil screenshot: Notifikasi "User sebelumnya telah di-clock out otomatis"]
```

---

### 5️⃣ **Izin / Permit**

Jika kamu tidak bisa masuk kerja, gunakan fitur **Permit**:

#### **Langkah-langkah Mengajukan Izin:**

1. Di dashboard, klik tombol **"IZIN"** berwarna kuning
2. Modal popup akan muncul dengan 2 pilihan:
   - **SAKIT** - jika kamu sakit
   - **IZIN** - jika ada keperluan lain
3. Klik salah satu pilihan
4. Sistem akan mencatat status izin kamu
5. Akan muncul notifikasi **"✅ Izin dicatat!"**

**📸 Screenshot - Permit:**
```
[Ambil screenshot: Tombol IZIN (kuning)]
[Ambil screenshot: Modal popup dengan pilihan SAKIT dan IZIN]
[Ambil screenshot: Notifikasi sukses "Izin dicatat!"]
```

**Catatan:**
- ⚠️ Setelah izin, kamu **tidak bisa clock in** di hari yang sama
- ⚠️ Status akan menjadi **"Sakit"** atau **"Izin"** di laporan admin
- ⚠️ Tidak ada perhitungan gaji untuk hari izin

---

### 6️⃣ **Kembali Kerja / Back to Work**

Jika kamu sudah mengajukan izin tapi ternyata bisa masuk kerja, gunakan fitur **Back to Work**:

#### **Langkah-langkah Kembali Kerja:**

1. Klik tombol **"KEMBALI KERJA"** berwarna biru
2. Konfirmasi akan muncul
3. Status izin akan **dibatalkan**
4. Kamu bisa clock in normal lagi

**📸 Screenshot - Back to Work:**
```
[Ambil screenshot: Tombol KEMBALI KERJA (biru)]
[Ambil screenshot: Notifikasi sukses "Status izin dibatalkan, silakan clock in"]
```

---

### 7️⃣ **Logout**

Untuk keluar dari aplikasi:

1. Klik tombol **"Logout"** di pojok kanan atas
2. Akan redirect ke halaman login
3. Session akan dihapus (aman)

**📸 Screenshot - Logout:**
```
[Ambil screenshot: Tombol Logout di header]
```

---

## 👨‍💼 Panduan untuk Super Admin

### 1️⃣ **Login Admin**

1. Login dengan akun **role = "owner"**
   - Username: `owner`
   - Password: (sesuai yang di-set)
2. Setelah login, akan otomatis redirect ke **Admin Panel**

**📸 Screenshot - Login Admin:**
```
[Ambil screenshot: Login dengan akun owner]
```

---

### 2️⃣ **Admin Panel - Dashboard**

Admin panel memiliki 4 tab utama:

#### **Tab 1: Live Workers (Karyawan Aktif)**

Menampilkan karyawan yang **sedang bekerja** (sudah clock in, belum clock out):

| Kolom | Keterangan |
|-------|------------|
| **Nama** | Nama lengkap karyawan + avatar |
| **Shift** | Jenis shift (Pagi/Sore/Malam/Long) + waktu |
| **Durasi** | Berapa lama sudah bekerja (real-time update) |

**📸 Screenshot - Live Workers:**
```
[Ambil screenshot: Tab Live Workers dengan list karyawan yang sedang bekerja]
[Ambil screenshot: Durasi yang update real-time (misal: 2 jam 34 menit)]
```

**Contoh:**
```
┌─────────────────────────────────────────────────────┐
│ Live Workers - Sedang Bekerja                       │
├─────────────────────────────────────────────────────┤
│ 👤 John Doe                                         │
│    Shift: Pagi (08:00 WIB)                          │
│    Durasi: 3 jam 45 menit                           │
│                                                      │
│ 👤 Jane Smith                                       │
│    Shift: Long Sore (14:40 WIB)                     │
│    Durasi: 1 jam 12 menit                           │
└─────────────────────────────────────────────────────┘
```

---

#### **Tab 2: Riwayat Absensi (History)**

Menampilkan **semua data absensi** dengan filter tanggal:

| Kolom | Keterangan |
|-------|------------|
| **Tanggal** | Tanggal absensi |
| **Nama** | Nama karyawan + phone |
| **Shift** | Jenis shift + waktu clock in/out |
| **Total Jam** | Durasi kerja dalam jam |
| **Status** | Normal / Telat / Sakit / Izin |
| **Estimasi Gaji** | Perhitungan gaji (jam kerja × tarif per jam) |
| **Aksi** | Tombol EDIT dan HAPUS |

**📸 Screenshot - Riwayat Absensi:**
```
[Ambil screenshot: Tab Riwayat dengan tabel data absensi lengkap]
[Ambil screenshot: Filter tanggal (dari - sampai)]
[Ambil screenshot: Status "TELAT" dengan badge merah]
```

**Filter Tanggal:**
1. Pilih **"Dari Tanggal"** (start date)
2. Pilih **"Sampai Tanggal"** (end date)
3. Klik tombol **"FILTER"**
4. Data akan difilter sesuai range tanggal
5. Klik **"RESET"** untuk menampilkan semua data

**Status Badge:**
- 🟢 **Normal** - Clock in/out tepat waktu
- 🔴 **TELAT** - Clock in terlambat (ada penalty)
- 🟡 **Sakit** - Karyawan izin sakit
- 🔵 **Izin** - Karyawan izin keperluan lain

**Aksi Edit:**
1. Klik tombol **"EDIT"** pada baris yang ingin diubah
2. Modal popup akan muncul dengan form:
   - **Tanggal Shift**
   - **Clock In**
   - **Clock Out**
   - **Manual Salary** (opsional - untuk override gaji)
3. Edit data yang diperlukan
4. Klik **"UPDATE"** untuk menyimpan

**📸 Screenshot - Edit Absensi:**
```
[Ambil screenshot: Modal popup form edit dengan field tanggal, clock in, clock out, manual salary]
```

**Aksi Hapus:**
1. Klik tombol **"HAPUS"** (merah) pada baris yang ingin dihapus
2. Konfirmasi akan muncul
3. Data akan terhapus permanen

**⚠️ Warning:** Data yang sudah dihapus **tidak bisa dikembalikan**!

---

#### **Tab 3: Total Gaji**

Menampilkan **summary perhitungan gaji** berdasarkan filter tanggal:

**Informasi yang ditampilkan:**
- 💰 **Total Gaji Keseluruhan** - Sum dari semua estimasi gaji
- 👥 **Jumlah Karyawan** - Berapa karyawan yang absen
- ⏰ **Total Jam Kerja** - Sum dari semua durasi kerja
- 💵 **Tarif Per Jam** - Rate gaji per jam (configurable)

**📸 Screenshot - Total Gaji:**
```
[Ambil screenshot: Tab Total Gaji dengan card summary]
[Ambil screenshot: Breakdown gaji per karyawan]
```

**Cara Kerja Perhitungan:**

```
Estimasi Gaji = (Total Jam Kerja - Penalty Jam) × Tarif Per Jam
```

**Contoh:**
- Jam Kerja: 8 jam
- Telat: 45 menit → Penalty 1 jam
- Tarif: Rp 10,000/jam
- **Gaji = (8 - 1) × 10,000 = Rp 70,000**

**Manual Salary Override:**
- Admin bisa set gaji manual (tidak pakai formula)
- Contoh use case: Bonus, insentif khusus, dll
- Field: `manual_salary` di database

---

#### **Tab 4: Kelola Akun**

Mengelola **user accounts** (tambah, edit, hapus):

**📸 Screenshot - Kelola Akun:**
```
[Ambil screenshot: Tab Kelola Akun dengan list semua user]
[Ambil screenshot: Form tambah user baru]
```

**Fitur:**

1. **List Semua User**
   - Username
   - Full Name
   - Role (owner/employee)
   - Phone Number
   - Hourly Rate (tarif per jam)
   - Avatar

2. **Tambah User Baru**
   - Klik tombol **"+ TAMBAH USER"**
   - Isi form:
     - Username
     - Password
     - Full Name
     - Role (dropdown: owner/employee)
     - Phone Number
     - Hourly Rate (angka)
   - Avatar akan dibuat otomatis (Dicebear API)
   - Klik **"SIMPAN"**

3. **Edit User**
   - Klik tombol **"EDIT"** pada user
   - Modal popup dengan form yang sama
   - Update data yang diperlukan
   - Klik **"UPDATE"**

4. **Hapus User**
   - Klik tombol **"HAPUS"** (merah)
   - Konfirmasi penghapusan
   - User akan terhapus dari database

**⚠️ Warning:** Jika user dihapus, semua **riwayat absensi** user tersebut akan tetap ada (tidak dihapus)

**📸 Screenshot - Form User:**
```
[Ambil screenshot: Form tambah/edit user dengan semua field]
```

---

### 3️⃣ **Fitur Admin Lainnya**

#### **A. Atur Tarif Per Jam Global**

Di tab **Total Gaji**, ada field untuk set tarif default:

1. Masukkan nilai tarif baru (misal: 15000)
2. Klik **"UPDATE RATE"**
3. Semua karyawan akan pakai tarif ini (kecuali yang punya tarif custom)

**📸 Screenshot - Update Rate:**
```
[Ambil screenshot: Form update tarif per jam]
```

#### **B. Export Data (Future Feature)**

Untuk export data ke Excel/PDF, bisa menggunakan browser print:

1. Buka tab **Riwayat Absensi**
2. Filter sesuai periode yang diinginkan
3. Tekan **Ctrl+P** (Windows) atau **Cmd+P** (Mac)
4. Pilih **"Save as PDF"**
5. Simpan file

**Tips:** Sembunyikan kolom **Aksi** sebelum print untuk hasil lebih rapi.

---

## ❓ FAQ & Troubleshooting

### **1. Kenapa tidak bisa clock in? Muncul error GPS?**

**Kemungkinan penyebab:**
- ❌ Lokasi GPS tidak aktif di HP
- ❌ Browser tidak diberi izin akses lokasi
- ❌ Berada di luar radius 50 meter dari cafe

**Solusi:**
1. **Aktifkan GPS/Location** di HP:
   - Settings → Privacy → Location Services → ON
2. **Izinkan browser akses lokasi:**
   - Browser akan popup minta izin → Klik **"Allow"**
   - Jika terlewat, buka Settings browser → Site settings → Permissions → Location → Allow
3. **Pastikan berada di dalam cafe** (radius 50m dari koordinat cafe)
4. **Coba refresh** halaman (swipe down di PWA)

---

### **2. Kenapa tombol shift tidak bisa diklik (disabled)?**

**Kemungkinan penyebab:**
- ⚠️ Ada karyawan lain yang sedang **Long Shift**

**Penjelasan:**
- Jika ada user yang clock in **Long Pagi** → hanya **Long Sore** yang bisa dipilih
- Shift **Pagi, Sore, Malam** akan disabled
- Tujuan: Mencegah shift pendek overlap dengan shift panjang (12+ jam)

**Solusi:**
- Tunggu user dengan Long Shift clock out dulu
- Atau gunakan Long Shift juga

---

### **3. Aplikasi tidak muncul di Home Screen setelah install?**

**iOS:**
- Cek di **App Library** (swipe ke kiri dari Home Screen)
- Search "DhenPresence"
- Drag icon ke Home Screen

**Android:**
- Cek di **App Drawer** (swipe up dari Home Screen)
- Search "DhenPresence"
- Long press → Add to Home Screen

---

### **4. Kenapa gaji saya terpotong padahal tidak telat?**

**Kemungkinan:**
- Periksa **waktu clock in** di tab Riwayat Admin
- Mungkin clock in sedikit melewati toleransi 15 menit
- Contoh: Shift Pagi 08:00, clock in 08:16 → sudah masuk penalty (1 jam)

**Solusi:**
- Pastikan clock in **paling lambat 15 menit** setelah waktu shift
- Contoh: Shift Pagi 08:00 → clock in sebelum 08:15 (aman)

---

### **5. Bagaimana cara menghapus akun karyawan yang resign?**

**Admin:**
1. Login sebagai **owner**
2. Buka tab **"Kelola Akun"**
3. Cari user yang resign
4. Klik tombol **"HAPUS"** (merah)
5. Konfirmasi penghapusan
6. User akan terhapus (tapi riwayat absensi tetap ada)

---

### **6. Aplikasi lambat atau tidak load?**

**Solusi:**
1. **Periksa koneksi internet** - aplikasi butuh internet untuk clock in/out
2. **Clear cache browser:**
   - Chrome: Settings → Privacy → Clear browsing data
   - Safari: Settings → Safari → Clear History and Website Data
3. **Reinstall PWA:**
   - Hapus icon dari Home Screen
   - Install ulang dari browser
4. **Hubungi admin** jika masalah berlanjut

---

### **7. Lupa password, bagaimana reset?**

**Solusi:**
- Hubungi **admin/owner** untuk reset password
- Admin bisa edit password di tab **"Kelola Akun"**

---

### **8. Clock in berhasil tapi tidak muncul di Live Workers?**

**Kemungkinan:**
- Browser cache belum refresh
- Service Worker masih pakai data lama

**Solusi:**
1. **Refresh halaman** (swipe down di PWA atau Ctrl+R)
2. **Force refresh** (Ctrl+Shift+R atau Cmd+Shift+R)
3. Cek di tab **Riwayat** apakah data sudah masuk

---

### **9. Bagaimana cara update tarif per jam untuk karyawan tertentu?**

**Admin:**
1. Login sebagai **owner**
2. Buka tab **"Kelola Akun"**
3. Klik **"EDIT"** pada user yang ingin diubah
4. Update field **"Hourly Rate"**
5. Klik **"UPDATE"**
6. Tarif baru akan berlaku untuk absensi selanjutnya

---

### **10. Auto clock out tengah malam tidak jalan?**

**Kemungkinan:**
- Server Koyeb sleep/restart
- Goroutine tidak berjalan

**Solusi Admin:**
- Cek logs di Koyeb dashboard
- Restart service jika perlu
- Manual clock out user yang terbuka di tab Riwayat (Edit → Set clock out time)

---

## 📞 Kontak Support

Jika ada pertanyaan atau masalah teknis:

- 📧 **Email**: [admin email]
- 📱 **WhatsApp**: [admin phone]
- 🐛 **Bug Report**: [GitHub Issues URL]

---

## 📝 Changelog

### Version 1.0 (January 2026)
- ✅ Initial release
- ✅ Clock In/Out dengan GPS validation
- ✅ 5 jenis shift (Pagi, Sore, Malam, Long Pagi, Long Sore)
- ✅ Shift Window policy (±3 jam)
- ✅ Escalating penalty untuk keterlambatan
- ✅ Kompensasi lembur otomatis
- ✅ Admin panel lengkap
- ✅ PWA support (iOS & Android)
- ✅ Offline functionality

---

**🎉 Terima kasih telah menggunakan DhenPresence!**

Dibuat dengan ❤️ untuk Dhen Coffee
