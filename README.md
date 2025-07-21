# Sharepal: Expense Sharing Platform

This project is the backend for my [SharePal application](https://github.com/notnotrachit/sharepal_app), a robust, feature-rich REST API for an expense-sharing application similar to Splitwise. Built with Go (Gin), MongoDB, and Redis, it provides a scalable and production-ready backend solution. It includes advanced features like JWT authentication, detailed expense tracking, group management, debt simplification, and a secure system for handling media uploads

## ✨ Features

*   **Authentication**: Secure user registration and login using JWT with access and refresh tokens.
*   **User & Friendship Management**: Send, accept, or reject friend requests, and manage friend lists.
*   **Group Management**: Create shared expense groups, manage members, and track group-specific balances.
*   **Advanced Expense Tracking**:
    *   Create, update, and delete expenses.
    *   Multiple split types: **equal**, **exact amount**, and **percentage**.
    *   Attach notes and upload receipts.
*   **Balance & Debt Simplification**:
    *   Real-time balance calculation between users.
    *   Smart debt simplification to minimize the number of transactions required for settlement.
*   **Media Handling**:
    *   Upload profile pictures and receipts to an S3-compatible object store (like AWS S3 or Cloudflare R2).
*   **Push Notifications**: Integrated service for sending push notifications.
*   **API Documentation**: Auto-generated interactive API documentation via Swagger.

## 🛠️ Tech Stack

*   **Language**: [Go](https://golang.org/)
*   **Web Framework**: [Gin](https://gin-gonic.com/)
*   **Database**: [MongoDB](https://www.mongodb.com/)
*   **Object Storage**: AWS S3 or any S3-compatible service (e.g., Cloudflare R2)

## 🚀 Getting Started

### Prerequisites

*   Go 1.20+
*   Docker & Docker Compose
*   MongoDB
*   Redis

## 📖 API Documentation

Interactive API documentation is available via Swagger. Once the application is running, access it at:

**http://localhost:8080/swagger/index.html**

## ⚙️ Configuration

All configuration is managed via the `.env` file.

| Variable                        | Description                                                              | Default         |
| ------------------------------- | ------------------------------------------------------------------------ | --------------- |
| `SERVER_ADDR`                   | Server address.                                                          | `localhost`     |
| `SERVER_PORT`                   | Server port.                                                             | `8080`          |
| `MONGO_URI`                     | MongoDB connection string.                                               | `mongodb://...` |
| `MONGO_DATABASE`                | MongoDB database name.                                                   | `exampledb`     |
| `USE_REDIS`                     | Set to `true` to enable Redis for caching.                               | `true`          |
| `REDIS_DEFAULT_ADDR`            | Redis server address.                                                    | `localhost:6379`|
| `JWT_SECRET`                    | Secret key for signing JWTs. **Change this for production.**             | `My.Ultra...`   |
| `JWT_ACCESS_EXPIRATION_MINUTES` | Expiration time for access tokens.                                       | `1440`          |
| `JWT_REFRESH_EXPIRATION_DAYS`   | Expiration time for refresh tokens.                                      | `7`             |
| `MODE`                          | Gin framework mode (`debug` or `release`).                               | `debug`         |
| `AWS_REGION`                    | S3 bucket region. Set to `auto` for Cloudflare R2.                       | `auto`          |
| `AWS_S3_BUCKET`                 | S3 bucket name for media uploads.                                        |                 |
| `AWS_ACCESS_KEY_ID`             | AWS/Cloudflare Access Key ID.                                            |                 |
| `AWS_SECRET_ACCESS_KEY`         | AWS/Cloudflare Secret Access Key.                                        |                 |
| `AWS_S3_ENDPOINT`               | S3 endpoint URL. Required for Cloudflare R2.                             |                 |

## 🙏 Credits

This project was bootstrapped from the excellent [golang-mongodb-rest-api-starter](https://github.com/ebubekiryigit/golang-mongodb-rest-api-starter) by Ebubekir.