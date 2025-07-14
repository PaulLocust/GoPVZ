1. я столкнулся с небольшой проблемой с созданием бд, в частности сочетание полей createdAt и dateTime
я сделал так:
    в бд за date_time и любую информацию о создании отвечает created_at, но в api сохраняются обозначения из swagger.yaml
2. Для миграций использовал https://github.com/golang-migrate/migrate/releases
3. Документация для JWT https://www.iana.org/assignments/jwt/jwt.xhtml, отсюда брать названия полей в Payload, сайт для исследования jwt токенов https://jwt.io/
4. Когда я накатывал swagger, с помощью swaggo, возникла проблема, что мой сгенерированный файл с документацией не мог найтись, решением стало импортировани _ "GoPVZ/internal/transport/rest/docs" в main.go внутри rest директории