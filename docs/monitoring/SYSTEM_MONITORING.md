# 🖥️ Системный мониторинг AI TG Writer

## 📊 Добавленные системные панели

### 🔥 CPU Usage (Использование процессора)
- **Метрика**: `100 - (avg by (instance) (rate(node_cpu_seconds_total{mode="idle"}[5m])) * 100)`
- **Описание**: Показывает процент использования CPU
- **Диапазон**: 0-100%
- **Алерты**: > 80% - предупреждение, > 95% - критично

### 🧠 Memory Usage (Использование памяти)
- **Метрика**: `(1 - (node_memory_MemAvailable_bytes / node_memory_MemTotal_bytes)) * 100`
- **Описание**: Показывает процент использования RAM
- **Диапазон**: 0-100%
- **Алерты**: > 85% - предупреждение, > 95% - критично

### 💾 Disk Usage (Использование диска)
- **Метрика**: `100 - ((node_filesystem_avail_bytes{mountpoint="/"} * 100) / node_filesystem_size_bytes{mountpoint="/"})`
- **Описание**: Показывает процент использования корневого раздела
- **Диапазон**: 0-100%
- **Алерты**: > 80% - предупреждение, > 90% - критично

### 🌐 Network I/O (Сетевой трафик)
- **Метрики**: 
  - `rate(node_network_receive_bytes_total[5m])` - входящий трафик
  - `rate(node_network_transmit_bytes_total[5m])` - исходящий трафик
- **Описание**: Показывает скорость сетевого трафика по устройствам
- **Единицы**: Bytes/sec

### ⚖️ Load Average (Средняя нагрузка)
- **Метрики**:
  - `node_load1` - нагрузка за 1 минуту
  - `node_load5` - нагрузка за 5 минут
  - `node_load15` - нагрузка за 15 минут
- **Описание**: Показывает среднюю нагрузку системы
- **Алерты**: > количество ядер CPU - предупреждение

### 💿 Disk I/O (Дисковый ввод/вывод)
- **Метрики**:
  - `rate(node_disk_read_bytes_total[5m])` - чтение с диска
  - `rate(node_disk_written_bytes_total[5m])` - запись на диск
- **Описание**: Показывает скорость дисковых операций
- **Единицы**: Bytes/sec

### 🔄 Process Count (Количество процессов)
- **Метрики**:
  - `node_procs_running` - запущенные процессы
  - `node_procs_blocked` - заблокированные процессы
- **Описание**: Показывает количество процессов в системе
- **Алерты**: > 1000 заблокированных процессов - предупреждение

### ⏰ System Uptime (Время работы системы)
- **Метрика**: `node_boot_time_seconds`
- **Описание**: Показывает время последней перезагрузки
- **Формат**: Относительное время (например, "2 days ago")

## 🚀 Запуск системного мониторинга

### 1. Запуск всех сервисов:
```bash
docker-compose up -d
```

### 2. Проверка статуса:
```bash
docker-compose ps
```

### 3. Проверка метрик node_exporter:
```bash
curl http://localhost:9100/metrics | grep node_cpu_seconds_total | head -3
```

### 4. Проверка Prometheus:
```bash
curl http://localhost:9090/api/v1/targets | grep node_exporter
```

## 📈 Просмотр дашборда

1. Откройте Grafana: http://localhost:3000
2. Логин: `admin`, Пароль: `admin`
3. Перейдите в дашборд "AI TG Writer Monitoring - Fixed"
4. Прокрутите вниз до системных панелей

## 🔧 Настройка алертов

### Рекомендуемые пороги:

#### CPU Usage:
- **Warning**: > 80%
- **Critical**: > 95%

#### Memory Usage:
- **Warning**: > 85%
- **Critical**: > 95%

#### Disk Usage:
- **Warning**: > 80%
- **Critical**: > 90%

#### Load Average:
- **Warning**: > количество ядер CPU
- **Critical**: > количество ядер CPU * 2

## 📊 Полезные запросы Prometheus

### Топ процессов по CPU:
```promql
topk(5, rate(node_cpu_seconds_total[5m]))
```

### Топ процессов по памяти:
```promql
topk(5, node_memory_MemAvailable_bytes)
```

### Сетевой трафик по интерфейсам:
```promql
rate(node_network_receive_bytes_total[5m]) + rate(node_network_transmit_bytes_total[5m])
```

### Дисковое пространство по файловым системам:
```promql
100 - ((node_filesystem_avail_bytes * 100) / node_filesystem_size_bytes)
```

## 🎯 Мониторинг производительности

### Ключевые метрики для отслеживания:
1. **CPU Usage** - не должен превышать 80%
2. **Memory Usage** - не должен превышать 85%
3. **Disk Usage** - не должен превышать 80%
4. **Load Average** - не должен превышать количество ядер
5. **Network I/O** - отслеживание аномального трафика
6. **Disk I/O** - отслеживание высокой нагрузки на диск

### Признаки проблем:
- **Высокий CPU** + **Высокий Load** = перегрузка процессора
- **Высокий Memory** + **Много процессов** = нехватка памяти
- **Высокий Disk I/O** + **Медленный отклик** = проблемы с диском
- **Высокий Network I/O** = возможная атака или утечка данных

## 🔍 Отладка

### Если системные метрики не отображаются:

1. **Проверьте node_exporter**:
   ```bash
   curl http://localhost:9100/metrics | grep node_cpu
   ```

2. **Проверьте Prometheus**:
   ```bash
   curl http://localhost:9090/api/v1/targets
   ```

3. **Проверьте логи**:
   ```bash
   docker-compose logs node_exporter
   docker-compose logs prometheus
   ```

4. **Перезапустите сервисы**:
   ```bash
   docker-compose restart node_exporter prometheus grafana
   ```

## 🎉 Результат

Теперь у вас есть полный мониторинг:
- ✅ **Приложение**: HTTP запросы, голосовые сообщения, платежи, ошибки
- ✅ **Система**: CPU, память, диск, сеть, процессы, время работы
- ✅ **Визуализация**: Графики в реальном времени с правильными метриками
- ✅ **Алерты**: Готовые пороги для критических метрик

Система мониторинга полностью готова к использованию! 🚀
