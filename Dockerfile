# 1. Chọn base image có sẵn Go
FROM golang:1.25-alpine

# 3. Thiết lập thư mục làm việc bên trong container
WORKDIR /go/src/Travia.backend

# 4. Copy các file quản lý thư viện (go.mod, go.sum) trước để tối ưu cache
COPY go.mod go.sum ./
RUN go mod download

# 5. Copy toàn bộ code từ máy local vào container
COPY . .

# 6. Biên dịch ứng dụng Go
#RUN go build -o main .

# 7. Lệnh chạy ứng dụng khi container khởi động
#CMD ["go", "run", "main.go"]