# 🔧 Добавления в существующий nginx.conf

## 📋 Что добавить в ваш nginx.conf

### 1. **В основной HTTPS сервер (aiwhisper.ru)**

Добавьте эти блоки в существующий `server` блок для `aiwhisper.ru`:

```nginx
# Grafana (мониторинг)
location /grafana/ {
    proxy_pass http://localhost:3000/;
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;
    
    # WebSocket поддержка
    proxy_http_version 1.1;
    proxy_set_header Upgrade $http_upgrade;
    proxy_set_header Connection "upgrade";
}

# Prometheus (метрики)
location /prometheus/ {
    proxy_pass http://localhost:9090/;
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;
}

# Jaeger (трейсинг)
location /jaeger/ {
    proxy_pass http://localhost:16686/;
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;
}
```

### 2. **Добавить новый сервер для админ-панели мониторинга**

Добавьте этот блок в конец файла (после основного сервера):

```nginx
# Отдельный сервер для мониторинга (только для админов)
server {
    listen 443 ssl http2;
    server_name monitor.aiwhisper.ru;
    
    # SSL сертификаты (используйте те же, что у вас)
    ssl_certificate /etc/letsencrypt/live/aiwhisper.ru/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/aiwhisper.ru/privkey.pem;
    
    # SSL настройки (скопируйте из основного сервера)
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers ECDHE-RSA-AES256-GCM-SHA512:DHE-RSA-AES256-GCM-SHA512:ECDHE-RSA-AES256-GCM-SHA384:DHE-RSA-AES256-GCM-SHA384;
    ssl_prefer_server_ciphers off;
    ssl_session_cache shared:SSL:10m;
    ssl_session_timeout 10m;
    
    # Базовая аутентификация
    auth_basic "Monitoring Access";
    auth_basic_user_file /etc/nginx/.htpasswd;
    
    # Grafana
    location / {
        proxy_pass http://localhost:3000;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        
        # WebSocket поддержка
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
    }
    
    # Prometheus
    location /prometheus/ {
        proxy_pass http://localhost:9090/;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

## 🔧 Настройка после добавления

### 1. **Создать файл паролей для базовой аутентификации**

```bash
# Создайте пароль для доступа к мониторингу
sudo htpasswd -c /etc/nginx/.htpasswd admin
```

### 2. **Добавить поддомен в DNS**

Добавьте A-запись в DNS:
```
monitor.aiwhisper.ru → ваш_IP_адрес
```

### 3. **Проверить конфигурацию**

```bash
# Проверить синтаксис
sudo nginx -t

# Если OK, перезагрузить
sudo systemctl reload nginx
```

## 🌐 Ваши адреса после настройки

### **Основные адреса:**
- **Приложение**: https://aiwhisper.ru
- **Grafana**: https://aiwhisper.ru/grafana/
- **Prometheus**: https://aiwhisper.ru/prometheus/
- **Jaeger**: https://aiwhisper.ru/jaeger/

### **Админ-панель:**
- **URL**: https://monitor.aiwhisper.ru
- **Логин**: admin
- **Пароль**: (который вы создали)

## 🔒 Безопасность

### **Рекомендации:**

1. **Ограничьте доступ по IP** (добавьте в location блоки):
```nginx
location /grafana/ {
    allow 192.168.1.0/24;  # Ваша локальная сеть
    allow YOUR_IP;          # Ваш IP
    deny all;
    
    # остальная конфигурация...
}
```

2. **Используйте сильные пароли** для базовой аутентификации

3. **Регулярно обновляйте** SSL сертификаты

## 🚨 Если что-то не работает

### **Проверка:**
```bash
# Статус nginx
sudo systemctl status nginx

# Логи nginx
sudo tail -f /var/log/nginx/error.log

# Проверка портов
netstat -tlnp | grep -E "(3000|9090|16686)"
```

### **Перезапуск:**
```bash
# Перезапуск nginx
sudo systemctl restart nginx

# Перезапуск Docker сервисов
docker-compose restart grafana prometheus jaeger
```

## 🎯 Готово!

После добавления этих блоков у вас будет:
- ✅ Мониторинг доступен по основному домену
- ✅ Отдельная админ-панель с паролем
- ✅ Все сервисы работают через HTTPS
- ✅ WebSocket поддержка для Grafana
