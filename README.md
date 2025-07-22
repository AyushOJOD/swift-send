# SwiftSend Backend

Go (Gin) backend for SwiftSend — a file sharing service with large file support using chunked uploads to AWS S3.

---

## 🛠 Setup

### Clone the repo

```bash
git clone https://github.com/AyushOJOD/swift-send-backend.git
cd swift-send-backend
```

### Configure environment

Create a .env file:

```bash
AWS_BUCKET_NAME=your-bucket-name
AWS_REGION=your-region
AWS_ACCESS_KEY_ID=your-access-key
AWS_SECRET_ACCESS_KEY=your-secret-key
```

### Install dependencies

```bash
go mod tidy
```

### Run the server

```bash
go run main.go
```

Server starts on: http://localhost:8080

## 📂 Folder Structure

```bash
├── internal/
│ ├── models/ # File manifest types
│ ├── routes/ # API routes
│ ├── storage/ # S3 client wrapper
│ └── utils/ # Config loader
├── main.go
├── go.mod / go.sum
```

## Notes

- Temporary files are stored in /tmp and cleaned up.
- CORS is enabled for development (all origins allowed).
