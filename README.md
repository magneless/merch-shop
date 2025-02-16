# merch-shop

> Это REST API pet-project, написанный на Go.

## Установка

Инструкция по установке проекта на **Linux** (Ubuntu):

1. Клонируйте репозиторий:
```bash
git clone https://github.com/magneless/merch-shop.git
```
2. Перейдите в директорию проекта:
```bash
cd merch-shop
```
3. Установите зависимости:
```bash
go mod download
```
4. Развертывание БД:
   1. Установите Docker, если он еще не установлен:
   ```bash
   sudo apt update
   sudo apt install docker.io
   sudo systemctl start docker
   sudo systemctl enable docker
   ```
   2. Скачайте образ PostgreSQL:
   ```bash
   docker pull postgres
   ```
   3. Запустите контейнер PostgreSQL:
   ```bash
   docker run --name postgres -e POSTGRES_PASSWORD=qwerty -d -p 5436:5432 postgres
   ```
   4. Запустите утилиту migrate:
   ```bash
   migrate -path ./schema -database 'postgres://postgres:qwerty@localhost:5436/postgres?sslmode=disable' up
   ```
6. Запустите проект:
```bash
DB_PASSWORD=qwerty CONFIG_PATH=cmd/config/local.yaml go run cmd/merch-shop/main.go
```

## Запросы

/api/auth

/api

