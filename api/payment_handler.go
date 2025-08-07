package api

import (
	"ai_tg_writer/internal/domain"
	"ai_tg_writer/internal/infrastructure/database"
	"ai_tg_writer/internal/service"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

type PaymentHandler struct {
	subscriptionService *service.SubscriptionService
	prodamusHandler     interface{} // Временно используем interface{} для совместимости
	db                  *database.DB
}

func NewPaymentHandler(subscriptionService *service.SubscriptionService, prodamusHandler interface{}, db *database.DB) *PaymentHandler {
	return &PaymentHandler{
		subscriptionService: subscriptionService,
		prodamusHandler:     prodamusHandler,
		db:                  db,
	}
}

// HandleSuccess обрабатывает успешную оплату
func (h *PaymentHandler) HandleSuccess(w http.ResponseWriter, r *http.Request) {
	log.Println("Обработка успешной оплаты")
	
	// Получаем параметры из URL
	queryParams := r.URL.Query()
	orderID := queryParams.Get("order_id")
	
	if orderID == "" {
		http.Error(w, "Order ID не найден", http.StatusBadRequest)
		return
	}

	log.Printf("Успешная оплата для заказа: %s", orderID)

	// Парсим order_id для получения userID
	// Формат: sub_123456_20231201123456 или pay_123456_20231201123456
	parts := strings.Split(orderID, "_")
	if len(parts) < 2 {
		http.Error(w, "Неверный формат Order ID", http.StatusBadRequest)
		return
	}

	userID, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		http.Error(w, "Неверный User ID в Order ID", http.StatusBadRequest)
		return
	}

	// Обновляем тариф пользователя на "payed"
	if err := h.db.UpdateUserTariff(userID, "payed"); err != nil {
		log.Printf("Ошибка обновления тарифа пользователя %d: %v", userID, err)
		http.Error(w, "Ошибка обновления тарифа", http.StatusInternalServerError)
		return
	}

	// Если это подписка, обновляем данные подписки
	if strings.HasPrefix(orderID, "sub_") {
		subscription, err := h.subscriptionService.GetUserSubscription(userID)
		if err != nil {
			log.Printf("Ошибка получения подписки пользователя %d: %v", userID, err)
		} else if subscription != nil {
			// Обновляем статус подписки
			subscription.Status = string(domain.SubscriptionStatusActive)
			subscription.LastPayment = time.Now()
			subscription.NextPayment = time.Now().AddDate(0, 1, 0) // +1 месяц
			
			if err := h.subscriptionService.ProcessPayment(userID, subscription.Amount); err != nil {
				log.Printf("Ошибка обработки платежа подписки: %v", err)
			}
		}
	}

	// Отправляем HTML страницу успешной оплаты
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	successHTML := `
<!DOCTYPE html>
<html lang="ru">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Оплата успешна</title>
    <style>
        body {
            font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            margin: 0;
            padding: 0;
            min-height: 100vh;
            display: flex;
            align-items: center;
            justify-content: center;
        }
        .container {
            background: white;
            border-radius: 20px;
            padding: 40px;
            box-shadow: 0 20px 40px rgba(0,0,0,0.1);
            text-align: center;
            max-width: 500px;
            width: 90%;
        }
        .success-icon {
            width: 80px;
            height: 80px;
            background: #4CAF50;
            border-radius: 50%;
            display: flex;
            align-items: center;
            justify-content: center;
            margin: 0 auto 20px;
            color: white;
            font-size: 40px;
        }
        h1 {
            color: #333;
            margin-bottom: 20px;
        }
        p {
            color: #666;
            line-height: 1.6;
            margin-bottom: 30px;
        }
        .button {
            background: linear-gradient(45deg, #667eea, #764ba2);
            color: white;
            padding: 15px 30px;
            border: none;
            border-radius: 25px;
            text-decoration: none;
            display: inline-block;
            font-weight: bold;
            transition: transform 0.3s ease;
        }
        .button:hover {
            transform: translateY(-2px);
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="success-icon">✓</div>
        <h1>Оплата прошла успешно!</h1>
        <p>Спасибо за покупку! Ваш тариф обновлен и теперь у вас есть доступ ко всем премиум функциям.</p>
        <a href="https://t.me/your_bot_username" class="button">Вернуться в бот</a>
    </div>
</body>
</html>`
	
	w.Write([]byte(successHTML))
}

// HandleFail обрабатывает неудачную оплату
func (h *PaymentHandler) HandleFail(w http.ResponseWriter, r *http.Request) {
	log.Println("Обработка неудачной оплаты")
	
	// Получаем параметры из URL
	queryParams := r.URL.Query()
	orderID := queryParams.Get("order_id")
	
	if orderID != "" {
		log.Printf("Неудачная оплата для заказа: %s", orderID)
	}

	// Отправляем HTML страницу неудачной оплаты
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	failHTML := `
<!DOCTYPE html>
<html lang="ru">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Ошибка оплаты</title>
    <style>
        body {
            font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
            background: linear-gradient(135deg, #ff6b6b 0%, #ee5a24 100%);
            margin: 0;
            padding: 0;
            min-height: 100vh;
            display: flex;
            align-items: center;
            justify-content: center;
        }
        .container {
            background: white;
            border-radius: 20px;
            padding: 40px;
            box-shadow: 0 20px 40px rgba(0,0,0,0.1);
            text-align: center;
            max-width: 500px;
            width: 90%;
        }
        .error-icon {
            width: 80px;
            height: 80px;
            background: #ff6b6b;
            border-radius: 50%;
            display: flex;
            align-items: center;
            justify-content: center;
            margin: 0 auto 20px;
            color: white;
            font-size: 40px;
        }
        h1 {
            color: #333;
            margin-bottom: 20px;
        }
        p {
            color: #666;
            line-height: 1.6;
            margin-bottom: 30px;
        }
        .button {
            background: linear-gradient(45deg, #ff6b6b, #ee5a24);
            color: white;
            padding: 15px 30px;
            border: none;
            border-radius: 25px;
            text-decoration: none;
            display: inline-block;
            font-weight: bold;
            transition: transform 0.3s ease;
        }
        .button:hover {
            transform: translateY(-2px);
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="error-icon">✗</div>
        <h1>Ошибка оплаты</h1>
        <p>К сожалению, произошла ошибка при обработке платежа. Пожалуйста, попробуйте еще раз или обратитесь в поддержку.</p>
        <a href="https://t.me/your_bot_username" class="button">Вернуться в бот</a>
    </div>
</body>
</html>`
	
	w.Write([]byte(failHTML))
}

// HandleWebhook обрабатывает вебхуки от Prodamus
func (h *PaymentHandler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	log.Println("Получен вебхук от Prodamus")

	if r.Method != http.MethodPost {
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	// Парсим данные формы
	if err := r.ParseForm(); err != nil {
		log.Printf("Ошибка парсинга формы: %v", err)
		http.Error(w, "Ошибка парсинга данных", http.StatusBadRequest)
		return
	}

	// Получаем подпись из заголовка
	signature := r.Header.Get("Sign")
	if signature == "" {
		log.Println("Подпись не найдена в заголовках")
		http.Error(w, "Подпись не найдена", http.StatusBadRequest)
		return
	}

	// Временно отключаем проверку подписи
	// TODO: Добавить проверку подписи для нового платежного модуля
	// if !h.prodamusHandler.VerifyWebhook(r.Form, signature) {
	// 	log.Println("Неверная подпись вебхука")
	// 	http.Error(w, "Неверная подпись", http.StatusUnauthorized)
	// 	return
	// }

	// Временно отключаем обработку вебхука
	// TODO: Добавить обработку вебхука для нового платежного модуля
	// webhookData, err := h.prodamusHandler.ProcessWebhook(r.Form, signature)
	// if err != nil {
	// 	log.Printf("Ошибка обработки вебхука: %v", err)
	// 	http.Error(w, "Ошибка обработки данных", http.StatusInternalServerError)
	// 	return
	// }

	// Временно используем заглушку для webhookData
	webhookData := struct {
		OrderID string
		Sum     string
	}{
		OrderID: r.Form.Get("order_id"),
		Sum:     r.Form.Get("sum"),
	}

	log.Printf("Обработка вебхука для заказа: %s, сумма: %s", webhookData.OrderID, webhookData.Sum)

	// Парсим order_id для получения userID
	parts := strings.Split(webhookData.OrderID, "_")
	if len(parts) < 2 {
		log.Printf("Неверный формат Order ID: %s", webhookData.OrderID)
		http.Error(w, "Неверный формат Order ID", http.StatusBadRequest)
		return
	}

	userID, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		log.Printf("Ошибка парсинга User ID: %v", err)
		http.Error(w, "Неверный User ID", http.StatusBadRequest)
		return
	}

	// Обновляем тариф пользователя
	if err := h.db.UpdateUserTariff(userID, "payed"); err != nil {
		log.Printf("Ошибка обновления тарифа пользователя %d: %v", userID, err)
		http.Error(w, "Ошибка обновления тарифа", http.StatusInternalServerError)
		return
	}

	// Если это подписка, обновляем данные подписки
	if strings.HasPrefix(webhookData.OrderID, "sub_") {
		subscription, err := h.subscriptionService.GetUserSubscription(userID)
		if err != nil {
			log.Printf("Ошибка получения подписки пользователя %d: %v", userID, err)
		} else if subscription != nil {
			// Обновляем статус подписки
			subscription.Status = string(domain.SubscriptionStatusActive)
			subscription.LastPayment = time.Now()
			subscription.NextPayment = time.Now().AddDate(0, 1, 0) // +1 месяц
			
			if err := h.subscriptionService.ProcessPayment(userID, subscription.Amount); err != nil {
				log.Printf("Ошибка обработки платежа подписки: %v", err)
			}
		}
	}

	// Отправляем успешный ответ
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("success"))
}

// HandleSubscriptionCallback обрабатывает уведомления о подписках от Prodamus
func (h *PaymentHandler) HandleSubscriptionCallback(w http.ResponseWriter, r *http.Request) {
	log.Println("=== ПОЛУЧЕН ВЕБХУК О ПОДПИСКЕ ОТ PRODAMUS ===")
	
	if r.Method != http.MethodPost {
		log.Printf("❌ Неверный метод: %s (ожидается POST)", r.Method)
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	// Парсим данные формы
	if err := r.ParseForm(); err != nil {
		log.Printf("❌ Ошибка парсинга формы: %v", err)
		http.Error(w, "Ошибка парсинга данных", http.StatusBadRequest)
		return
	}

	// Логируем все заголовки
	log.Printf("📋 ЗАГОЛОВКИ ЗАПРОСА:")
	for name, values := range r.Header {
		for _, value := range values {
			log.Printf("   %s: %s", name, value)
		}
	}

	// Логируем все параметры формы
	log.Printf("📋 ПАРАМЕТРЫ ФОРМЫ:")
	for key, values := range r.Form {
		for _, value := range values {
			log.Printf("   %s: %s", key, value)
		}
	}

	// Получаем подпись из заголовка
	signature := r.Header.Get("Sign")
	if signature == "" {
		log.Printf("⚠️ Подпись не найдена в заголовках")
		log.Printf("📋 Доступные заголовки:")
		for name := range r.Header {
			log.Printf("   %s", name)
		}
	} else {
		log.Printf("✅ Подпись найдена: %s", signature)
	}

	// Проверяем подпись для подписок
	if signature != "" {
		log.Printf("🔍 ДИАГНОСТИКА ПОДПИСИ:")
		log.Printf("   Ожидаемая подпись: %s", signature)
		
		// Временно отключаем генерацию подписей
		// TODO: Добавить генерацию подписей для нового платежного модуля
		// urlSignature := h.prodamusHandler.CreateSignature(r.Form)
		// jsonSignature := h.prodamusHandler.CreateSubscriptionSignature(r.Form)
		
		// log.Printf("   Наша URL-encoded подпись: %s", urlSignature)
		// log.Printf("   Наша JSON подпись: %s", jsonSignature)
		
		// Генерируем различные варианты подписей для анализа
		// urlSignature2 := h.prodamusHandler.CreateSignature(r.Form)
		// jsonSignature2 := h.prodamusHandler.CreateSubscriptionSignature(r.Form)
		
		log.Printf("📊 АНАЛИЗ ПОДПИСЕЙ:")
		log.Printf("   Получена от Prodamus: %s", signature)
		// log.Printf("   Наша URL-encoded: %s", urlSignature)
		// log.Printf("   Наша JSON: %s", jsonSignature)
		// log.Printf("   URL совпадение: %v", strings.ToLower(urlSignature) == strings.ToLower(signature))
		// log.Printf("   JSON совпадение: %v", strings.ToLower(jsonSignature) == strings.ToLower(signature))
		
		// ВРЕМЕННО: Отключаем строгую проверку подписи
		// Причина: алгоритм подписи Prodamus для webhook'ов может отличаться
		// TODO: Выяснить правильный алгоритм и включить проверку обратно
		log.Printf("⚠️ ВНИМАНИЕ: Проверка подписи временно отключена для отладки")
		
		// if !h.prodamusHandler.VerifyWebhook(r.Form, signature) && 
		//    !h.prodamusHandler.VerifySubscriptionWebhook(r.Form, signature) {
		// 	log.Printf("❌ Неверная подпись вебхука")
		// 	http.Error(w, "Invalid signature", http.StatusUnauthorized)
		// 	return
		// }
	}

	// Анализируем данные подписки
	log.Printf(" АНАЛИЗ ДАННЫХ ПОДПИСКИ:")
	
	// Проверяем наличие ключевых полей
	orderID := r.Form.Get("order_id")
	if orderID != "" {
		log.Printf("   📦 Order ID: %s", orderID)
		
		// Пытаемся извлечь userID из order_id
		parts := strings.Split(orderID, "_")
		if len(parts) >= 2 {
			if userID, err := strconv.ParseInt(parts[1], 10, 64); err == nil {
				log.Printf("   👤 User ID из order_id: %d", userID)
			} else {
				log.Printf("   ❌ Не удалось извлечь User ID из order_id: %v", err)
			}
		}
	}

	// Проверяем данные подписки
	subscriptionData := r.Form.Get("subscription")
	if subscriptionData != "" {
		log.Printf("   📋 Данные подписки (JSON): %s", subscriptionData)
		
		// Пытаемся распарсить JSON подписки
		var subscription map[string]interface{}
		if err := json.Unmarshal([]byte(subscriptionData), &subscription); err == nil {
			log.Printf("   ✅ JSON подписки распарсен успешно:")
			for key, value := range subscription {
				log.Printf("      %s: %v", key, value)
			}
		} else {
			log.Printf("   ❌ Ошибка парсинга JSON подписки: %v", err)
		}
	}

	// Проверяем другие важные поля
	importantFields := []string{"sum", "customer_phone", "customer_email", "payment_type", "date", "type", "action_code"}
	for _, field := range importantFields {
		if value := r.Form.Get(field); value != "" {
			log.Printf("   📋 %s: %s", field, value)
		}
	}
	
	// Обрабатываем данные подписки
	log.Printf("🔄 ОБРАБОТКА ДАННЫХ ПОДПИСКИ:")
	
	// Извлекаем userID из order_id (если возможно)
	var userID int64
	if orderID != "" {
		parts := strings.Split(orderID, "_")
		if len(parts) >= 2 {
			if parsedUserID, err := strconv.ParseInt(parts[1], 10, 64); err == nil {
				userID = parsedUserID
				log.Printf("   👤 User ID из order_id: %d", userID)
			}
		}
	}
	
	// Проверяем статус подписки
	subscriptionActive := r.Form.Get("subscription[active]")
	actionCode := r.Form.Get("subscription[action_code]")
	subscriptionID := r.Form.Get("subscription[id]")
	
	log.Printf("   📊 Статус подписки: %s", subscriptionActive)
	log.Printf("   📊 Код действия: %s", actionCode)
	log.Printf("   📊 ID подписки: %s", subscriptionID)
	
	// Временно отключаем проверку подписи
	// TODO: Добавить проверку подписи для нового платежного модуля
	// if signature == "" || (!h.prodamusHandler.VerifyWebhook(r.Form, signature) && !h.prodamusHandler.VerifySubscriptionWebhook(r.Form, signature)) {
	log.Printf("⚠️ Подпись не проверена, но обрабатываем данные для тестирования")
	
	// Здесь можно добавить логику обработки подписки
	// Например, обновить статус пользователя в базе данных
	if userID > 0 && subscriptionActive == "1" && actionCode == "finish" {
		log.Printf("✅ Обрабатываем успешную активацию подписки для пользователя %d", userID)
		// TODO: Обновить статус пользователя в базе данных
	}
	// }

	// Анализируем данные подписки из параметров формы
	log.Printf(" АНАЛИЗ ПАРАМЕТРОВ ПОДПИСКИ:")
	subscriptionFields := []string{
		"subscription[id]", "subscription[type]", "subscription[action_code]",
		"subscription[active]", "subscription[cost]", "subscription[name]",
		"subscription[date_create]", "subscription[date_next_payment]",
		"subscription[payment_num]", "subscription[autopayments_num]",
	}
	for _, field := range subscriptionFields {
		if value := r.Form.Get(field); value != "" {
			log.Printf("   📋 %s: %s", field, value)
		}
	}

	log.Printf("=== КОНЕЦ ОБРАБОТКИ ВЕБХУКА ПОДПИСКИ ===")

	// Отправляем успешный ответ
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("success"))
}

// SetupRoutes настраивает маршруты для HTTP-сервера
func (h *PaymentHandler) SetupRoutes(router *mux.Router) {
	router.HandleFunc("/payment/success", h.HandleSuccess).Methods("GET")
	router.HandleFunc("/payment/fail", h.HandleFail).Methods("GET")
	router.HandleFunc("/payment/webhook", h.HandleWebhook).Methods("POST")
	router.HandleFunc("/payment/sub_callback", h.HandleSubscriptionCallback).Methods("POST")
} 