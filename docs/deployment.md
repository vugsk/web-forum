# Развёртывание

## Локальный запуск

### Сборка

```bash
# Windows
go build -o forum.exe .

# Linux/macOS
go build -o forum .
```

### Запуск

```bash
# Windows
.\forum.exe

# Linux/macOS
./forum
```

## Туннели для доступа из интернета

### Cloudflare Tunnel (рекомендуется)

**Установка:**

```bash
# Windows
winget install cloudflare.cloudflared

# macOS
brew install cloudflared

# Linux
# Скачайте с https://developers.cloudflare.com/cloudflare-one/connections/connect-networks/downloads/
```

**Запуск (без регистрации):**

```bash
cloudflared tunnel --url http://localhost:8080
```

Получите ссылку вида: `https://random-name.trycloudflare.com`

### Ngrok

**Установка:**

```bash
# Windows
winget install ngrok

# Или скачайте с https://ngrok.com/download
```

**Настройка:**

```bash
# Регистрация на ngrok.com и добавление токена
ngrok config add-authtoken YOUR_TOKEN
```

**Запуск:**

```bash
ngrok http 8080
```

## Docker

### Dockerfile

```dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o forum .

FROM alpine:latest
WORKDIR /app

COPY --from=builder /app/forum .
COPY --from=builder /app/templates ./templates
COPY --from=builder /app/static ./static

EXPOSE 8080
CMD ["./forum"]
```

### docker-compose.yml

```yaml
version: '3.8'

services:
  forum:
    build: .
    ports:
      - "8080:8080"
    environment:
      - DB_HOST=mysql
      - DB_PORT=3306
      - DB_USER=forum
      - DB_PASSWORD=forum_password
      - DB_NAME=webforum
    depends_on:
      mysql:
        condition: service_healthy
    volumes:
      - ./uploads:/app/uploads

  mysql:
    image: mysql:8.0
    environment:
      - MYSQL_ROOT_PASSWORD=root_password
      - MYSQL_DATABASE=webforum
      - MYSQL_USER=forum
      - MYSQL_PASSWORD=forum_password
    volumes:
      - mysql_data:/var/lib/mysql
    healthcheck:
      test: ["CMD", "mysqladmin", "ping", "-h", "localhost"]
      interval: 5s
      timeout: 5s
      retries: 10

volumes:
  mysql_data:
```

**Запуск:**

```bash
docker-compose up -d
```

## Linux сервер

### Systemd сервис

Создайте `/etc/systemd/system/forum.service`:

```ini
[Unit]
Description=Web Forum
After=network.target mysql.service

[Service]
Type=simple
User=www-data
WorkingDirectory=/var/www/forum
ExecStart=/var/www/forum/forum
Restart=always
RestartSec=5
EnvironmentFile=/var/www/forum/.env

[Install]
WantedBy=multi-user.target
```

**Команды:**

```bash
# Перезагрузить конфигурацию
sudo systemctl daemon-reload

# Включить автозапуск
sudo systemctl enable forum

# Запуск
sudo systemctl start forum

# Статус
sudo systemctl status forum

# Логи
sudo journalctl -u forum -f
```

### Nginx reverse proxy

```nginx
server {
    listen 80;
    server_name forum.example.com;
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name forum.example.com;

    ssl_certificate /etc/letsencrypt/live/forum.example.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/forum.example.com/privkey.pem;

    # Загруженные файлы
    location /uploads/ {
        alias /var/www/forum/uploads/;
        expires 30d;
    }

    # Статические файлы
    location /static/ {
        alias /var/www/forum/static/;
        expires 7d;
    }

    # WebSocket
    location /ws/ {
        proxy_pass http://127.0.0.1:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }

    # Остальные запросы
    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    # Лимит загрузки файлов
    client_max_body_size 100M;
}
```

### SSL сертификат (Let's Encrypt)

```bash
sudo apt install certbot python3-certbot-nginx
sudo certbot --nginx -d forum.example.com
```

## Windows сервер

### Как служба Windows

Используйте [NSSM](https://nssm.cc/):

```powershell
# Установка службы
nssm install Forum C:\forum\forum.exe

# Настройка
nssm set Forum AppDirectory C:\forum
nssm set Forum DisplayName "Web Forum"
nssm set Forum Start SERVICE_AUTO_START

# Запуск
nssm start Forum
```

## Резервное копирование

### База данных

```bash
# Бэкап
mysqldump -u root -p webforum > backup_$(date +%Y%m%d).sql

# Восстановление
mysql -u root -p webforum < backup_20251206.sql
```

### Загруженные файлы

```bash
tar -czf uploads_backup.tar.gz uploads/
```

### Автоматический бэкап (cron)

```bash
# crontab -e
0 3 * * * mysqldump -u forum -p'password' webforum > /backups/db_$(date +\%Y\%m\%d).sql
0 4 * * * tar -czf /backups/uploads_$(date +\%Y\%m\%d).tar.gz /var/www/forum/uploads
```

## Мониторинг

### Логи

```bash
# Systemd
journalctl -u forum -f

# Docker
docker logs -f forum_forum_1
```

### Проверка здоровья

```bash
curl http://localhost:8080/
```

