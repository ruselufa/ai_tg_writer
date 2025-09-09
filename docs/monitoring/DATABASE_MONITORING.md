# 🗄️ Мониторинг базы данных AI TG Writer

## 📊 Добавленные панели базы данных

### 🔗 Database Connections (Подключения к БД)
- **Метрики**: 
  - `pg_stat_database_numbackends` - активные подключения
  - `pg_settings_max_connections` - максимальное количество подключений
- **Описание**: Показывает текущее количество подключений к PostgreSQL
- **Алерты**: > 80% от max_connections - предупреждение

### 💾 Database Size (Размер базы данных)
- **Метрика**: `pg_database_size_bytes / 1024 / 1024`
- **Описание**: Показывает размер базы данных в мегабайтах
- **Единицы**: MB
- **Алерты**: > 1GB - предупреждение, > 5GB - критично

### 🔍 Database Queries (Запросы к БД)
- **Метрики**:
  - `rate(pg_stat_database_tup_returned[5m])` - возвращенные кортежи/сек
  - `rate(pg_stat_database_tup_fetched[5m])` - полученные кортежи/сек
- **Описание**: Показывает активность запросов к базе данных
- **Единицы**: Tuples/sec

### 🔒 Database Locks (Блокировки БД)
- **Метрика**: `pg_locks_count`
- **Описание**: Показывает количество блокировок по типам
- **Алерты**: > 100 блокировок - предупреждение

### 💿 Disk Space Usage (Использование диска)
- **Метрики**:
  - `node_filesystem_size_bytes{mountpoint="/"} / 1024 / 1024` - общее пространство (MB)
  - `(node_filesystem_size_bytes - node_filesystem_avail_bytes) / 1024 / 1024` - занято (MB)
  - `node_filesystem_avail_bytes{mountpoint="/"} / 1024 / 1024` - свободно (MB)
- **Описание**: Показывает использование дискового пространства в мегабайтах
- **Единицы**: MB

### 🎯 Database Cache Hit Ratio (Эффективность кэша)
- **Метрика**: `pg_stat_database_blks_hit / (pg_stat_database_blks_hit + pg_stat_database_blks_read) * 100`
- **Описание**: Показывает процент попаданий в кэш базы данных
- **Диапазон**: 0-100%
- **Алерты**: < 95% - предупреждение, < 90% - критично

## 🚀 Запуск мониторинга БД

### 1. Запуск PostgreSQL Exporter:
```bash
docker-compose up -d postgres_exporter
```

### 2. Проверка статуса:
```bash
docker-compose ps | grep postgres
```

### 3. Проверка метрик:
```bash
curl http://localhost:9187/metrics | grep pg_stat_database | head -5
```

## 📈 Просмотр в Grafana

1. Откройте Grafana: http://localhost:3000
2. Перейдите в дашборд "AI TG Writer Monitoring - Fixed"
3. Прокрутите вниз до панелей базы данных (панели 19-24)

## 🔧 Полезные запросы Prometheus

### Топ таблиц по размеру:
```promql
topk(5, pg_stat_user_tables_size_bytes)
```

### Медленные запросы:
```promql
rate(pg_stat_database_tup_returned[5m]) / rate(pg_stat_database_tup_fetched[5m])
```

### Использование индексов:
```promql
rate(pg_stat_user_indexes_idx_tup_read[5m])
```

### Активные транзакции:
```promql
pg_stat_activity_count{state="active"}
```

## 🚨 Рекомендуемые алерты

### Database Connections:
- **Warning**: > 80% от max_connections
- **Critical**: > 95% от max_connections

### Database Size:
- **Warning**: > 1GB
- **Critical**: > 5GB

### Cache Hit Ratio:
- **Warning**: < 95%
- **Critical**: < 90%

### Disk Space:
- **Warning**: > 80% заполнено
- **Critical**: > 90% заполнено

## 🔍 Отладка проблем

### Если метрики БД не отображаются:

1. **Проверьте PostgreSQL Exporter**:
   ```bash
   curl http://localhost:9187/metrics | grep pg_stat_database
   ```

2. **Проверьте подключение к БД**:
   ```bash
   docker-compose logs postgres_exporter
   ```

3. **Проверьте Prometheus**:
   ```bash
   curl http://localhost:9090/api/v1/targets | grep postgres_exporter
   ```

4. **Перезапустите сервисы**:
   ```bash
   docker-compose restart postgres_exporter prometheus grafana
   ```

## 📊 Ключевые метрики для отслеживания

### 1. **Производительность**:
- Cache Hit Ratio > 95%
- Количество запросов в секунду
- Время выполнения запросов

### 2. **Ресурсы**:
- Количество подключений < 80% от максимума
- Размер базы данных
- Использование диска

### 3. **Блокировки**:
- Количество блокировок < 100
- Длительность блокировок
- Deadlocks

## 🎯 Оптимизация БД

### При низком Cache Hit Ratio:
- Увеличить `shared_buffers`
- Настроить `effective_cache_size`
- Добавить индексы

### При высоком количестве подключений:
- Увеличить `max_connections`
- Настроить connection pooling
- Оптимизировать запросы

### При большом размере БД:
- Настроить архивирование старых данных
- Сжать таблицы
- Очистить неиспользуемые данные

## 🎉 Результат

Теперь у вас есть полный мониторинг базы данных:
- ✅ **6 панелей БД** с ключевыми метриками
- ✅ **Дисковое пространство** в мегабайтах
- ✅ **Производительность** запросов и кэша
- ✅ **Ресурсы** подключений и блокировок
- ✅ **Готовые алерты** для критических метрик

**Мониторинг базы данных полностью настроен!** 🚀
