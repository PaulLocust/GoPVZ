# 🚀 GoPVZ - Backend Service for Pickup Points

![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-16+-4169E1?logo=postgresql)
![Docker](https://img.shields.io/badge/Docker-24+-2496ED?logo=docker)
![Swagger](https://img.shields.io/badge/Swagger-3.0-85EA2D?logo=swagger)
![Prometheus](https://img.shields.io/badge/Prometheus-2.47-E6522C?logo=prometheus)

Backend-сервис для сотрудников ПВЗ, который позволяет вносить информацию по заказам в рамках приёмки товаров.

## 🚀 Как запускать

```bash
# 1. Клонируйте репозиторий
git clone https://github.com/PaulLocust/GoPVZ.git

# 2. Настройте окружение
mv .env.example .env

# 3. Запустите сервисы через Docker
docker-compose up -d
```
## 🛠 Технологии
- **Язык**: Go (Gin framework)
- **База данных**: PostgreSQL
- **Инфраструктура**: Docker
- **Документация**: Swagger, OpenAPI
- **Метрики**: Prometheus
- **Архитектура**: Clean Architecture, DDD
- **Тестирование**: Unit-тесты, интеграционные тесты
- **Прочее**: Миграции, автогенерация DTO, JWT аутентификация

## После запуска доступны:
- 📚 http://localhost:8080/swagger API Documentation - основное место где можно поиграться с приложением

- 📊 http://localhost:9000/metrics Prometheus Metrics - сырые метрики prometheus

- 📈 http://localhost:9090 Prometheus UI - удобная визуализация метрик

## 📝 Заметки о ходе выполнения проекта

### 🛠 Технические детали
1. **Миграции**  
   Использован инструмент: [golang-migrate](https://github.com/golang-migrate/migrate/releases)

2. **JWT Авторизация**  
   - Документация по стандартным полям: [IANA JWT](https://www.iana.org/assignments/jwt/jwt.xhtml)  
   - Отладка токенов: [JWT.io](https://jwt.io/)  
   - Как работать с авторизацией в Swagger UI:  
     1. Получить токен через `/login`  
     2. В Swagger UI нажать "Authorize"  
     3. Вставить токен  
     4. Swagger автоматически добавит заголовок Authorization

3. **Swagger документация**  
   Проблема: сгенерированная документация не находилась.  
   Решение: добавлен импорт `_ "GoPVZ/internal/transport/rest/docs"` в `main.go`

### 🔍 Проблемы и решения

5. **Архитектура**  
   В процессе разработки принято решение перейти на Clean Architecture + DDD

6. **Генерация DTO**  
   Проблемы в исходном `swagger.yaml`:  
   - JWT Token был строкой вместо структуры  
   - Автовалидация Email мешала кастомной валидации  
   - Русские названия городов/категорий  
   
   Решения:  
   - Модифицирован `swagger.yaml`  
   - Все названия переведены на английский

### 🧪 Тестирование

7. **Тестовая стратегия**  
   - **Интеграционные тесты**:  
     - Используется Testcontainers для изоляции тестов  
     - На каждый тестовый файл создается отдельная таблица  
     - На каждый тест — уникальное подключение  
   - **Покрытие**:  
     - По 3 теста на каждый домен (auth, pvz)  
       - Unit-тест на usecase  
       - Интеграционные тесты на repo и controller/http

### 📈 Метрики
8. **Сбор метрик**
   - **Технические:**:
     - Количество запросов
     - Время ответа
   - **Бизнесовые:**:
     - Количество созданных ПВЗ
     - Количество созданных приёмок заказов
     - Количество добавленных товаров 