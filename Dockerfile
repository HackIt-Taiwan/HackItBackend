# 使用 Go 基础映像构建应用
FROM golang:1.22 as build
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o server .

# 使用轻量级映像运行应用
FROM alpine:latest
WORKDIR /app
COPY --from=build /app/server /server
RUN chmod +x /server
EXPOSE 5000
CMD ["/server"]