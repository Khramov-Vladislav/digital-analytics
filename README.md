# Resume Generator Web Application

Это веб‑приложение на **Go** для автоматического создания профессионально оформленного резюме в формате PDF.
Приложение поддерживает русский и английский языки (UTF-8), загрузку фотографии, динамический предпросмотр и несколько стилей оформления в соответствии с ГОСТ 7.0.97-2016.

---

## 🎨 Стили резюме (шаблоны)

Приложение поддерживает два стиля оформления:

### **Классический** (`classic`)
- Чёрно‑белый дизайн, традиционная структура
- Серые блоки для заголовков секций
- Идеален для консервативных работодателей
- Поддержка фотографии (32 мм в правом углу)

### **Современный** (`modern`)
- Синий цветовой акцент (RGB 30, 90, 150)
- Визуальные разделители и цветные элементы
- Навыки отображаются в виде тегов‑подсказок
- Большой заголовок в синем заголовке
- Поддержка фотографии в заголовке

---

## 📋 Особенности

✅ **Русская кодировка**
- Полная поддержка UTF-8 и Cyrillic символов
- Шрифты Liberation Sans (совместимые с DejaVu) встроены

✅ **Фотография**
- Загрузка JPG, JPEG, PNG
- Ограничение: максимум 5 МБ, минимум 200×200 пикселей
- Автоматическое размещение в PDF в зависимости от шаблона

✅ **Динамический предпросмотр**
- Форма слева, предпросмотр справа
- Автоматическое обновление при изменении данных
- Мгновенная визуализация структуры резюме

✅ **Соответствие ГОСТу**
- Структурированная информация (личные данные, опыт, образование, навыки)
- Профессиональное форматирование
- Фиксированные размеры шрифтов и отступы

✅ **Динамические разделы**
- Добавление неограниченного числа позиций опыта и образования
- Без перезагрузки страницы

## Requirements

- Go 1.20+

## Installation

```sh
cd /path/to/digital-analytics
# get dependencies
go get github.com/jung-kurt/gofpdf

# build the server
go build -o bin/resume-generator ./cmd
```

## Running

```sh
./bin/resume-generator -addr=:8080
```

Open `http://localhost:8080` in your browser. The web form appears on the left and the preview pane on the right.
Fill in the fields, add education/experience blocks, choose a template and click "Сгенерировать PDF".

## Project Structure

- `cmd/` – entry point (main.go)
- `internal/handlers` – HTTP handlers and routing
- `internal/models` – data structures and validation
- `internal/pdf` – PDF generation logic
- `templates/` – HTML templates for form and preview
- `static/` – static assets (JS/CSS)

## Notes

The app supports Russian and English UTF-8 input. Uploaded photos are validated (JPG/PNG, ≤5 MB, ≥200×200 px).
Multiple resume templates can be added by editing `internal/pdf/pdf.go` and supplementing CSS/HTML as needed.


