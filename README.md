# Invite (Annonyx)

Коммерческий сервис для маркетологов: создание инвайт‑ссылок на домене `invite.annonyx.ru`, сбор статистики и управление подпиской.

## Возможности

- Регистрация и вход (JWT).
- Создание инвайт‑ссылок на целевые URL.
- Статистика кликов по времени (отдельная страница с мини‑графиком по часам).
- Личный кабинет с подпиской: пробный период 7 дней, далее 150 ₽/месяц (ЮKassa).
- Админ‑панель: просмотр пользователей, подписок и ссылок.

## Стек

- Go + Gin
- GORM + SQLite
- HTML/CSS/JS (динамический фронт)

## Быстрый старт

1. Установите Go 1.20+.
2. Создайте `.env` (можно взять готовый пример):

```bash
cp .env .env.local
```

3. Запустите сервер:

```bash
go run ./cmd/server
```

Сервер стартует на `http://localhost:8080`.

## Docker

```bash
docker build -t invite .
docker run --env-file .env -p 8080:8080 -v $(pwd)/invite.db:/app/invite.db invite
```

> Важно: база SQLite хранится в файле `invite.db`. Чтобы данные не терялись при перезапуске контейнера, используйте volume как в примере выше.

## Пример Nginx (proxy)

```nginx
server {
    listen 80;
    server_name invite.annonyx.ru;

    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    }
}
```

## Переменные окружения

- `APP_DOMAIN` — домен, на котором работают инвайт‑ссылки.
- `BASE_URL` — базовый URL сервиса.
- `JWT_SECRET` — секрет для подписи JWT.
- `DATABASE_PATH` — путь к SQLite базе.
- `TRIAL_DAYS` — длительность пробного периода.
- `MONTHLY_PRICE_RUB` — стоимость подписки.
- `YOOKASSA_SHOP_ID`, `YOOKASSA_SECRET` — креды ЮKassa.
- `CHECKOUT_RETURN_URL` — URL возврата после оплаты.
- `ADMIN_EMAIL`, `ADMIN_PASSWORD` — учетная запись администратора (создается автоматически).

## Страницы

- `/login` — вход.
- `/register` — регистрация.
- `/dashboard` — личный кабинет.
- `/stats?id=<ID>` — статистика по ссылке.
- `/admin` — админ‑панель.

## Основные маршруты API

- `POST /api/register` — регистрация.
- `POST /api/login` — вход.
- `GET /api/links` — список ссылок.
- `POST /api/links` — создать ссылку.
- `GET /api/links/:id/stats` — статистика кликов.
- `POST /api/billing/checkout` — создать оплату (ЮKassa).
- `POST /api/webhook/yookassa` — вебхук от ЮKassa.
- `GET /i/:code` — редирект по инвайт‑ссылке (фиксирует клик).
- `GET /api/admin/users` — список пользователей.
- `GET /api/admin/users/:id` — детали пользователя.
- `DELETE /api/admin/users/:id` — удалить пользователя.

## ЮKassa

- Интеграция повторяет официальную схему: создаем платеж через API `v3/payments` с `confirmation.redirect`, используем `Idempotence-Key` и Basic Auth.
- Вебхук ожидает событие `payment.succeeded` и активирует подписку пользователя по metadata (`user_id`, `months`).
- Для продакшена добавьте проверку IP‑адресов и логирование входящих уведомлений.

## Что нужно доработать для продакшена

- Реальная проверка безопасности вебхуков (IP allowlist / подпись).
- Логи и мониторинг.
- Управление ролями и тарифами.
