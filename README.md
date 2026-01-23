# DhenPresence

Sistem Presensi Digital untuk Dhen Coffee.

## Menjalankan Lokal

```bash
go run main.go
```

Buka http://localhost:8080

## Deploy ke Koyeb

1. Setup Supabase dan dapatkan DATABASE_URL
2. Push ke GitHub
3. Deploy ke Koyeb dengan environment variables:
   - `DATABASE_URL` = connection string Supabase
   - `PORT` = 8080

## Login Default

- Username: `owner`
- Password: `kopi123`
