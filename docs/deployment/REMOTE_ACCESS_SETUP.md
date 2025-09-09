# 🌐 Настройка удаленного доступа к мониторингу

## 🎯 Обзор

Инструкции для доступа к системе мониторинга с других ПК (из дома, офиса и т.д.).

## 🔧 Настройка сервера

### 1. Обновляем Docker Compose

Файл `docker-compose.yml` уже обновлен для доступа извне:

```yaml
# Grafana - доступ с любого IP
grafana:
  ports:
    - "0.0.0.0:3000:3000"

# Prometheus - доступ с любого IP  
prometheus:
  ports:
    - "0.0.0.0:9090:9090"
```

### 2. Перезапускаем сервисы

```bash
# Останавливаем все сервисы
docker-compose down

# Запускаем с новыми настройками
docker-compose up -d
```

### 3. Настраиваем приложение

Приложение должно быть доступно на порту 8080:

```bash
# Запускаем приложениеtus.
./start_app.sh
```

## 🌍 Настройка доступа извне

### 1. Определяем IP адрес сервера

```bash
# На сервере
ifconfig | grep "inet " | grep -v 127.0.0.1
# или
ip addr show | grep "inet " | grep -v 127.0.0.1
```

### 2. Настраиваем файрвол (если нужно)

```bash
# Ubuntu/Debian
sudo ufw allow 3000  # Grafana
sudo ufw allow 9090  # Prometheus
sudo ufw allow 8080  # Приложение

# CentOS/RHEL
sudo firewall-cmd --permanent --add-port=3000/tcp
sudo firewall-cmd --permanent --add-port=9090/tcp
sudo firewall-cmd --permanent --add-port=8080/tcp
sudo firewall-cmd --reload
```

### 3. Настраиваем роутер (если нужно)

Если сервер за NAT, настройте проброс портов:
- 3000 → 3000 (Grafana)
- 9090 → 9090 (Prometheus)  
- 8080 → 8080 (Приложение)

## 📱 Доступ с других ПК

### 1. Grafana (Дашборды)

**URL**: `http://YOUR_SERVER_IP:3000`

**Логин**: `admin`  
**Пароль**: `admin`

**Дашборды**:
- "AI TG Writer Monitoring - Fixed" - основной дашборд
- "AI TG Writer - Business Metrics" - бизнес-метрики

### 2. Prometheus (Метрики)

**URL**: `http://YOUR_SERVER_IP:9090`

**Функции**:
- Просмотр метрик
- Выполнение запросов
- Настройка алертов

### 3. Приложение (API)

**URL**: `http://YOUR_SERVER_IP:8080`

**Эндпоинты**:
- `/metrics` - метрики Prometheus
- `/health` - состояние приложения
- `/test-metrics` - генерация тестовых метрик

## 🔒 Безопасность

### 1. Смена паролей по умолчанию

```bash
# В Grafana: Settings → Users → Admin → Change Password
# Новый пароль: ваш_безопасный_пароль
```

### 2. Настройка HTTPS (рекомендуется)

```bash
# Установка nginx с SSL
sudo apt install nginx certbot python3-certbot-nginx

# Получение SSL сертификата
sudo certbot --nginx -d your-domain.com
```

### 3. Ограничение доступа по IP

```bash
# В nginx конфигурации
location / {
    allow 192.168.1.0/24;  # Ваша локальная сеть
    allow YOUR_HOME_IP;     # Ваш домашний IP
    deny all;
}
```

## 📊 Мониторинг с мобильных устройств

### 1. Grafana Mobile App

- **iOS**: Grafana Mobile
- **Android**: Grafana Mobile
- **URL**: `http://YOUR_SERVER_IP:3000`
- **Логин**: `admin` / `admin`

### 2. Веб-браузер

Любой современный браузер поддерживает Grafana.

## 🚨 Алерты и уведомления

### 1. Настройка алертов в Grafana

1. Перейдите в **Alerting** → **Alert Rules**
2. Создайте правило для критических метрик
3. Настройте уведомления (email, Slack, Telegram)

### 2. Примеры алертов

```yaml
# Высокая загрузка CPU
- alert: HighCPUUsage
  expr: 100 - (avg by(instance) (rate(node_cpu_seconds_total{mode="idle"}[5m])) * 100) > 80
  for: 5m
  labels:
    severity: warning
  annotations:
    summary: "High CPU usage detected"

# Много ошибок API
- alert: HighAPIErrors
  expr: rate(http_requests_total{status=~"5.."}[5m]) > 0.1
  for: 2m
  labels:
    severity: critical
  annotations:
    summary: "High API error rate detected"
```

## 🔧 Troubleshooting

### 1. Не могу подключиться к Grafana

```bash
# Проверяем статус контейнера
docker ps | grep grafana

# Проверяем логи
docker logs ai_tg_writer_grafana

# Проверяем порт
netstat -tlnp | grep 3000
```

### 2. Не отображаются метрики

```bash
# Проверяем Prometheus
curl http://YOUR_SERVER_IP:9090/api/v1/targets

# Проверяем приложение
curl http://YOUR_SERVER_IP:8080/metrics
```

### 3. Медленная загрузка

```bash
# Проверяем ресурсы
docker stats

# Увеличиваем память для контейнеров
docker-compose down
docker system prune -a
docker-compose up -d
```

## 📱 Быстрый доступ

### Создайте закладки в браузере:

1. **Grafana**: `http://YOUR_SERVER_IP:3000`
2. **Prometheus**: `http://YOUR_SERVER_IP:9090`
3. **Метрики**: `http://YOUR_SERVER_IP:8080/metrics`
4. **Здоровье**: `http://YOUR_SERVER_IP:8080/health`

### Мобильные виджеты:

- **Grafana Dashboard** - основной дашборд
- **Business Metrics** - бизнес-метрики
- **System Health** - состояние системы

## 🎯 Готово!

Теперь вы можете мониторить систему с любого устройства в любой точке мира! 🌍

**Не забудьте**:
- Сменить пароли по умолчанию
- Настроить HTTPS для безопасности
- Создать алерты для критических метрик
- Регулярно обновлять систему
