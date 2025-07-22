-- Инициализация базы данных для AI Voice Writer Bot

-- Создание расширений (если нужно)
-- CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Таблица пользователей
CREATE TABLE IF NOT EXISTS users (
    id BIGINT PRIMARY KEY,
    username VARCHAR(255),
    first_name VARCHAR(255),
    last_name VARCHAR(255),
    tariff VARCHAR(50) DEFAULT 'free',
    usage_count INTEGER DEFAULT 0,
    last_usage TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    referral_code VARCHAR(20) UNIQUE,
    referred_by BIGINT REFERENCES users(id)
);

-- Таблица голосовых сообщений
CREATE TABLE IF NOT EXISTS voice_messages (
    id SERIAL PRIMARY KEY,
    user_id BIGINT REFERENCES users(id),
    file_id VARCHAR(255),
    duration INTEGER,
    file_size INTEGER,
    text TEXT,
    rewritten TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Таблица статистики использования
CREATE TABLE IF NOT EXISTS usage_stats (
    id SERIAL PRIMARY KEY,
    user_id BIGINT REFERENCES users(id),
    date DATE DEFAULT CURRENT_DATE,
    usage_count INTEGER DEFAULT 0,
    UNIQUE(user_id, date)
);

-- Таблица подписок (для будущего использования)
CREATE TABLE IF NOT EXISTS subscriptions (
    id SERIAL PRIMARY KEY,
    user_id BIGINT REFERENCES users(id),
    tariff VARCHAR(50) NOT NULL,
    status VARCHAR(20) DEFAULT 'active',
    start_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    end_date TIMESTAMP,
    payment_method VARCHAR(50),
    amount DECIMAL(10,2),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Таблица партнёрской программы (для будущего использования)
CREATE TABLE IF NOT EXISTS referrals (
    id SERIAL PRIMARY KEY,
    referrer_id BIGINT REFERENCES users(id),
    referred_id BIGINT REFERENCES users(id),
    status VARCHAR(20) DEFAULT 'pending',
    commission_earned DECIMAL(10,2) DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(referrer_id, referred_id)
);

-- Индексы для оптимизации
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
CREATE INDEX IF NOT EXISTS idx_users_tariff ON users(tariff);
CREATE INDEX IF NOT EXISTS idx_voice_messages_user_id ON voice_messages(user_id);
CREATE INDEX IF NOT EXISTS idx_voice_messages_created_at ON voice_messages(created_at);
CREATE INDEX IF NOT EXISTS idx_usage_stats_user_date ON usage_stats(user_id, date);
CREATE INDEX IF NOT EXISTS idx_subscriptions_user_id ON subscriptions(user_id);
CREATE INDEX IF NOT EXISTS idx_subscriptions_status ON subscriptions(status);

-- Комментарии к таблицам
COMMENT ON TABLE users IS 'Пользователи бота';
COMMENT ON TABLE voice_messages IS 'Голосовые сообщения пользователей';
COMMENT ON TABLE usage_stats IS 'Статистика использования по дням';
COMMENT ON TABLE subscriptions IS 'Подписки пользователей';
COMMENT ON TABLE referrals IS 'Партнёрская программа';

-- Создание представления для статистики пользователей
CREATE OR REPLACE VIEW user_stats AS
SELECT 
    u.id,
    u.username,
    u.first_name,
    u.last_name,
    u.tariff,
    u.created_at,
    COUNT(vm.id) as total_messages,
    COALESCE(us.usage_count, 0) as today_usage,
    u.last_usage
FROM users u
LEFT JOIN voice_messages vm ON u.id = vm.user_id
LEFT JOIN usage_stats us ON u.id = us.user_id AND us.date = CURRENT_DATE
GROUP BY u.id, u.username, u.first_name, u.last_name, u.tariff, u.created_at, us.usage_count, u.last_usage;

-- Функция для автоматического создания реферального кода
CREATE OR REPLACE FUNCTION generate_referral_code()
RETURNS VARCHAR(20) AS $$
DECLARE
    code VARCHAR(20);
    counter INTEGER := 0;
BEGIN
    LOOP
        -- Генерируем код из 8 символов
        code := upper(substring(md5(random()::text) from 1 for 8));
        
        -- Проверяем уникальность
        IF NOT EXISTS (SELECT 1 FROM users WHERE referral_code = code) THEN
            RETURN code;
        END IF;
        
        counter := counter + 1;
        IF counter > 100 THEN
            RAISE EXCEPTION 'Не удалось сгенерировать уникальный реферальный код';
        END IF;
    END LOOP;
END;
$$ LANGUAGE plpgsql;

-- Триггер для автоматического создания реферального кода при создании пользователя
CREATE OR REPLACE FUNCTION set_referral_code()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.referral_code IS NULL THEN
        NEW.referral_code := generate_referral_code();
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_set_referral_code
    BEFORE INSERT ON users
    FOR EACH ROW
    EXECUTE FUNCTION set_referral_code();

-- Функция для проверки лимитов пользователя
CREATE OR REPLACE FUNCTION check_user_limit(user_id BIGINT)
RETURNS BOOLEAN AS $$
DECLARE
    user_tariff VARCHAR(50);
    today_usage INTEGER;
    daily_limit INTEGER;
BEGIN
    -- Получаем тариф пользователя
    SELECT tariff INTO user_tariff FROM users WHERE id = user_id;
    
    -- Получаем количество использований сегодня
    SELECT COALESCE(usage_count, 0) INTO today_usage 
    FROM usage_stats 
    WHERE user_id = $1 AND date = CURRENT_DATE;
    
    -- Определяем дневной лимит
    CASE user_tariff
        WHEN 'free' THEN daily_limit := 5;
        WHEN 'premium' THEN daily_limit := 999999; -- Практически неограниченно
        ELSE daily_limit := 5;
    END CASE;
    
    RETURN today_usage < daily_limit;
END;
$$ LANGUAGE plpgsql; 