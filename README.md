# Pixel Tactics - Match Microservice
![Go](https://img.shields.io/badge/go-%2300ADD8.svg?style=for-the-badge&logo=go&logoColor=white)
![RabbitMQ](https://img.shields.io/badge/Rabbitmq-FF6600?style=for-the-badge&logo=rabbitmq&logoColor=white)

This is the match microservice of Pixel Tactics. This service focuses on handling game matches such as match invitation, heroes pickup, turn-based gameplay, etc.

## Installing
First, you need to install dependencies.
```
go mod tidy
```

After installing the dependencies, setup the environment variables. You can also use `.env` file in order to do this step. Below are the required envs:
- `USER_MICROSERVICE_URL`: URL for user microservice deployment. Example: `http://localhost:8080` for local development and `https://users.deployment-website.com`.
- `RABBITMQ_CONNECTION_STRING`: Connection string for RabbitMQ in the format of `amqp://USERNAME:PASSWORD@HOST:PORT/`. Example: `amqp://guest:guest@localhost:5672/`.

Before running, ensure that the User Microservice and RabbitMQ server is active. To run this service, run:
```
go run src/main.go
```

## Developers
- Meervix (Emyr298)
