# Валидация входных данных в Resume Generator

## Валидация на клиенте (JavaScript)

### 1. Телефонный номер
- **Автоформатирование**: `+7 (921) 123-45-67`
- **Поддерживаемые форматы**:
  - `+7 9121234567` → `+7 (912) 123-45-67`
  - `8 9121234567` → `+7 (912) 123-45-67`
  - `79121234567` → `+7 (912) 123-45-67`
  - `+1 2015550171` → `+1 (201) 555-01-71`

- **Визуальная обратная связь**: 
  - Поле красным если меньше 10 цифр
  - Подтверждение при валидном номере

### 2. Email
- Проверка формата: `user@example.com`
- Красная граница при некорректном формате

## Валидация на сервере (Go)

### Телефонный номер
```go
isValidPhone(phone string) bool
```
- Минимум **10 цифр**
- Максимум **15 цифр** (международный стандарт)
- Поддерживает любые разделители (+, -, пробелы, скобки)

### Email
```go
isValidEmail(email string) bool
```
- Проверка формата: `name@domain.extension`
- Регулярное выражение для валидации

### Полная валидация Resume
```go
resume.Validate() error
```

Возвращает ошибку если:
- `fullname` пуст или > 100 символов
- `phone` пуст, < 10 цифр, > 15 цифр или > 50 символов
- `email` пуст, невалидный формат или > 100 символов
- `city` > 100 символов
- `summary` > 1000 символов
- `photo` > 5MB, неправильный формат, < 200×200px

## Примеры тестирования

### Валидные номера телефонов
```bash
curl -X POST http://localhost:8080/generate \
  -F "fullname=Иван Иванов" \
  -F "phone=+79121234567" \
  -F "email=ivan@test.com" \
  -F "city=Москв" \
  -F "template=classic"

curl -X POST http://localhost:8080/generate \
  -F "fullname=Test User" \
  -F "phone=+1 (201) 555-0171" \
  -F "email=user@example.com" \
  -F "city=New York" \
  -F "template=modern"
```

### Невалидные данные (ошибки валидации)
```bash
# Телефон слишком короткий (меньше 10 цифр)
curl -X POST http://localhost:8080/generate \
  -F "fullname=Test" \
  -F "phone=123" \
  -F "email=test@test.com" \
  -F "template=classic"
# Ошибка: "phone number must contain 10-15 digits..."

# Email некорректный
curl -X POST http://localhost:8080/generate \
  -F "fullname=Test" \
  -F "phone=79121234567" \
  -F "email=invalidemail" \
  -F "template=classic"
# Ошибка: "email format is invalid"
```

## Стилизация в форме

Поля с ошибками получают:
- Красная граница: `border-color: #ff6b6b`
- Светло-красный фон: `background-color: #fff5f5`
- Подсказка при наведении (title attribute)

## Интеграция

При запуске сервера:
```bash
go run cmd/main.go cmd/fonts.go
# или
./app
```

Всё будет работать автоматически:
- Шрифты скачаны (если нужны)
- Валидация включена на сервере
- Клиентская валидация готова в форме
