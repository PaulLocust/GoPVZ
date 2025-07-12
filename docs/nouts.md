1. я столкнулся с небольшой проблемой с созданием бд, в частности сочетание полей createdAt и dateTime
я сделал так:
    в бд за date_time и любую информацию о создании отвечает created_at, но в api сохраняются обозначения из swagger.yaml
2. Для миграций использовал https://github.com/golang-migrate/migrate/releases