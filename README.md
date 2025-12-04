# Shortlink API - Backend

API backend untuk aplikasi web shortlink yang dibangun menggunakan Go (Golang) dengan arsitektur clean code.

## Tech Stack

- **Go (Gin Framework)** - Web framework
- **pgx** - Database ORM dan driver PostgreSQL
- **go-redis** - Redis client untuk caching
- **argon2** - Password hashing
- **golang-jwt** - JWT authentication
- **validator/v10** - Request validation
- **godotenv** - Environment variable management
- **go-argon** - Password hashing dengan Argon2


## Prasyarat

- Go 1.19 atau lebih baru
- PostgreSQL 13 atau lebih baru
- Redis 6 atau lebih baru

## Instalasi

1. Clone repository
```bash
git clone <repository-url>
cd <backend-directory>
```

2. Install dependencies
```bash
go mod download
```

3. Setup environment variables
```bash
cp .env.example .env
```

Edit file `.env` dengan konfigurasi Anda:
```env
# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=your_db_user
DB_PASSWORD=your_db_password
DB_NAME=shortlink_db

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=

# JWT
JWT_SECRET=your-secret-key
JWT_EXPIRATION=24h

# Server
PORT=8080
APP_ENV=development
```

## Cara Menjalankan Backend

### Development Mode
```bash
go run main.go
```

### Production Mode
```bash
go build -o shortlink-api
./shortlink-api
```

### Dengan Air (Hot Reload)
```bash
# Install air
go install github.com/cosmtrek/air@latest

# Run dengan air
air
```

## Migrasi Database

Jalankan migrasi untuk membuat tabel database:

```bash
go run main.go migrate
```

Atau jika Anda memiliki script migrasi terpisah:
```bash
go run cmd/migrate/main.go
```

## Testing Endpoints

Anda dapat menggunakan tools seperti Postman atau curl untuk testing:

### Register User
```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "SecurePassword123",
    "name": "John Doe"
  }'
```

### Login
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "SecurePassword123"
  }'
```

### Create Shortlink
```bash
curl -X POST http://localhost:8080/api/v1/links \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "url": "https://example.com/very-long-url",
    "custom_alias": "mylink"
  }'
```




## Redis Flushing Mechanism

Aplikasi menggunakan Redis untuk caching. Mekanisme flushing:
- Cache akan otomatis di-flush saat ada update data
- TTL default: 1 jam
- Manual flush dapat dilakukan melalui endpoint admin

## Keamanan

- Password di-hash menggunakan Argon2
- JWT untuk autentikasi
- Input validation menggunakan validator/v10
- Rate limiting untuk mencegah abuse
- CORS configuration

## Troubleshooting

### Database Connection Error
- Pastikan PostgreSQL sudah berjalan
- Cek kredensial database di `.env`
- Pastikan database sudah dibuat

### Redis Connection Error
- Pastikan Redis server sudah berjalan
- Cek konfigurasi Redis di `.env`

### Migration Failed
- Cek log error untuk detail
- Pastikan koneksi database tersedia
- Pastikan user memiliki permission yang cukup

## Kontribusi

1. Fork repository
2. Create feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to branch (`git push origin feature/AmazingFeature`)
5. Open Pull Request

## License

[MIT License](LICENSE)

## Contact

Untuk pertanyaan atau dukungan, silakan hubungi [your-email@example.com]