# be-sign-file

A Clean Architecture Golang Starter for Digital Signature & Document Management

---

## 🚀 Giới thiệu
Dự án **be-sign-file** là nền tảng backend quản lý tài liệu và chữ ký số, xây dựng theo chuẩn Clean Architecture với Gin, Gorm, PostgreSQL. Hỗ trợ phát triển nhanh, mở rộng dễ dàng, tích hợp logging hiện đại và các công cụ dev tiện lợi.

## 🧩 Tính năng nổi bật
- Quản lý tài liệu, người dùng, chữ ký số
- Hệ thống logging UI đẹp, lọc theo tháng, xem chi tiết
- API RESTful chuẩn, tài liệu Postman đầy đủ
- Hỗ trợ migration, seeder, script tự động
- Docker & Docker Compose sẵn sàng cho dev/prod

## 📦 Yêu cầu hệ thống
- Go >= 1.20
- PostgreSQL >= 15.0
- Docker (nếu dùng container)

## ⚡️ Khởi động nhanh
### 1. Clone & cấu hình
```bash
git clone https://github.com/PhanPhuc2609/be-sign-file.git
cd be-sign-file
cp .env.example .env # hoặc tự tạo file .env
```

### 2. Chạy bằng Docker Compose (khuyên dùng cho dev)
```bash
docker compose up --build
```
Sau đó có thể chạy các lệnh:
```bash
make init-uuid
make migrate-seed
```
- Truy cập app tại: http://localhost:8888
- Xem logs UI: http://localhost:8888/logs

### 3. Chạy thủ công (không Docker)
- Cài PostgreSQL, tạo DB, cấu hình `.env`
- Chạy migration, seed:
```bash
go run main.go --migrate --seed --run
```
- Hoặc chỉ chạy app:
```bash
go run main.go
```

## 🛠️ Lệnh hữu ích
- **Migration:** `go run main.go --migrate`
- **Seeder:** `go run main.go --seed`
- **Chạy script:** `go run main.go --script:example_script`
- **Kết hợp:** `go run main.go --migrate --seed --run --script:example_script`

## 📝 Tài liệu API
- [Xem tài liệu Postman](https://documenter.getpostman.com/view/29665461/2s9YJaZQCG)

## 🖥️ Logging UI
- Truy cập: `http://localhost:8888/logs`
- Lọc theo tháng, xem chi tiết, UI hiện đại

## 🧑‍💻 Đóng góp
- Pull Request & Issue template chuẩn hóa
- Chào mừng mọi đóng góp, ý tưởng mới!

## 📂 Cấu trúc dự án (rút gọn)
```
be-sign-file/
├── command/           # Lệnh CLI
├── config/            # Cấu hình DB, email, logger
├── controller/        # Xử lý request
├── dto/               # Định nghĩa dữ liệu truyền
├── entity/            # Model nghiệp vụ
├── helpers/           # Hàm tiện ích
├── middleware/        # Middleware Gin
├── migrations/        # Migration, seeder
├── provider/          # Provider nghiệp vụ
├── repository/        # Truy xuất DB
├── routes/            # Định nghĩa route
├── script/            # Script tiện ích
├── service/           # Xử lý nghiệp vụ
├── tests/             # Unit test
├── uploads/           # File upload
├── utils/             # Hàm tiện ích chung
├── docker/            # Dockerfile, cấu hình
├── main.go            # Entry point
└── ...
```

## 📜 License
MIT

---

> **be-sign-file** - Clean Architecture Golang Starter for Digital Signature & Document Management
