# Pixel Tactics - Match Microservice
![Go](https://img.shields.io/badge/go-%2300ADD8.svg?style=for-the-badge&logo=go&logoColor=white)
![RabbitMQ](https://img.shields.io/badge/Rabbitmq-FF6600?style=for-the-badge&logo=rabbitmq&logoColor=white)

This is the match microservice of Pixel Tactics. This service focuses on handling game matches such as match invitation, heroes pickup, turn-based gameplay, etc.

## Installing
First, you need to setup the environment variables. You can create `.env` file in the root directory to do this step. Below are the required envs:
- `USER_MICROSERVICE_URL`: URL for user microservice deployment. Example: `http://localhost:8080` for local development and `https://users.deployment-website.com`.
- `RABBITMQ_CONNECTION_STRING`: Connection string for RabbitMQ in the format of `amqp://USERNAME:PASSWORD@HOST:PORT/`. Example: `amqp://guest:guest@localhost:5672/`.

After setting up the environment variables, you need to install dependencies.
```
go mod tidy
```

Ensure that the User Microservice and RabbitMQ server are active. Then, to run this service, run:
```
go run src/main.go
```

## Developers
- Meervix (Emyr298)
