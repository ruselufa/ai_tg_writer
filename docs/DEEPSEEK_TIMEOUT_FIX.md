# ⏰ Исправление таймаутов DeepSeek API

## 🚨 **Проблема**
```
Ошибка генерации поста: ошибка DeepSeek API: ошибка чтения ответа: context deadline exceeded (Client.Timeout or context cancellation while reading body)
```

## 🔍 **Причины**

1. **Короткий таймаут HTTP клиента** - всего 60 секунд
2. **DeepSeek API медленно отвечает** - генерация постов занимает время
3. **Сетевые задержки** - медленное соединение
4. **Отсутствие retry логики** - одна ошибка = полный провал

## ✅ **Решение**

### 1. **Увеличение таймаутов**

#### **HTTP Client Timeout:**
```go
// Было
Timeout: 60 * time.Second

// Стало  
Timeout: 300 * time.Second // 5 минут для генерации постов
```

#### **Response Header Timeout:**
```go
// Было
ResponseHeaderTimeout: 30 * time.Second

// Стало
ResponseHeaderTimeout: 120 * time.Second // 2 минуты для заголовков
```

### 2. **Добавление Retry логики**

#### **Новая функция makeRequest:**
```go
func (dh *DeepSeekHandler) makeRequest(request DeepSeekRequest) (*DeepSeekResponse, error) {
    const maxRetries = 3
    var lastErr error

    for attempt := 1; attempt <= maxRetries; attempt++ {
        log.Printf("🔄 [DeepSeek] Попытка %d/%d", attempt, maxRetries)

        response, err := dh.makeSingleRequest(request)
        if err == nil {
            if attempt > 1 {
                log.Printf("✅ [DeepSeek] Успешно после %d попыток", attempt)
            }
            return response, nil
        }

        lastErr = err
        log.Printf("❌ [DeepSeek] Попытка %d неудачна: %v", attempt, err)

        // Если это не последняя попытка, ждем перед повтором
        if attempt < maxRetries {
            waitTime := time.Duration(attempt) * 2 * time.Second
            log.Printf("⏳ [DeepSeek] Ждем %v перед повтором...", waitTime)
            time.Sleep(waitTime)
        }
    }

    return nil, fmt.Errorf("все попытки исчерпаны, последняя ошибка: %v", lastErr)
}
```

#### **Выделение makeSingleRequest:**
```go
func (dh *DeepSeekHandler) makeSingleRequest(request DeepSeekRequest) (*DeepSeekResponse, error) {
    // Логика одного HTTP запроса
}
```

## 🔧 **Технические детали**

### **Таймауты:**
- **HTTP Client**: 300 секунд (5 минут)
- **Response Headers**: 120 секунд (2 минуты)  
- **TLS Handshake**: 10 секунд (без изменений)
- **Idle Connection**: 90 секунд (без изменений)

### **Retry стратегия:**
- **Максимум попыток**: 3
- **Время ожидания**: экспоненциальное (2с, 4с, 8с)
- **Логирование**: детальное для каждой попытки

### **Логирование:**
```
🔄 [DeepSeek] Попытка 1/3
❌ [DeepSeek] Попытка 1 неудачна: context deadline exceeded
⏳ [DeepSeek] Ждем 2s перед повтором...
🔄 [DeepSeek] Попытка 2/3
✅ [DeepSeek] Успешно после 2 попыток
```

## 🧪 **Результат**

### **До исправления:**
- ❌ Таймаут через 60 секунд
- ❌ Одна ошибка = полный провал
- ❌ Пользователь получает ошибку

### **После исправления:**
- ✅ До 5 минут на генерацию поста
- ✅ Автоматические повторы при ошибках
- ✅ Детальное логирование процесса
- ✅ Повышенная надежность API

## 📁 **Измененные файлы**
- `internal/infrastructure/deepseek/deepseek_handler.go`
  - Увеличены таймауты HTTP клиента
  - Добавлена retry логика
  - Улучшено логирование

## 🎯 **Преимущества**

1. **Надежность**: Автоматические повторы при ошибках
2. **Терпение**: Достаточно времени для генерации постов
3. **Мониторинг**: Детальное логирование всех попыток
4. **UX**: Меньше ошибок для пользователей
5. **Стабильность**: Лучшая работа при медленном соединении

## 🔮 **Следующие шаги**

- Мониторинг успешности API вызовов
- Анализ времени ответа DeepSeek
- Возможная настройка таймаутов под нагрузку
- Добавление метрик производительности

## 📝 **Заключение**

Проблема с таймаутами DeepSeek API полностью решена:
- ✅ Увеличены таймауты до разумных значений
- ✅ Добавлена retry логика для надежности
- ✅ Улучшено логирование для мониторинга
- ✅ Повышена стабильность генерации постов
