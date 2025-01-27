[![Deploy to GCE](https://github.com/Ulas-Scan/UlaScan_BE/actions/workflows/deploy.yml/badge.svg?branch=main)](https://github.com/Ulas-Scan/UlaScan_BE/actions/workflows/deploy.yml)

# Backend for Review Product Tokopedia BE

Welcome to the backend repository for the Review Product Tokopedia BE! This backend serves the streamlit app for my project. It is a backend service built using Gin, Golang, and GORM.

## API Documentation

The API documentation for this project can be found [here](https://www.postman.com/supply-administrator-61638669/workspace/ulascan/folder/36341500-c7118b3a-2f29-4fcd-9e08-9641cf52d5a8) or [Swagger](http://34.101.79.15/swagger/index.html).

## Backend and Cloud Architecture

![Cloud Architecture](https://github.com/Ulas-Scan/UlaScan_BE/assets/87474722/cbcc7a9a-36c3-4212-9f1a-7d2afe5e0e2e)

## Local Installation for Development

To set up the backend for local development, follow these steps:

1. **Clone the Repository**:
   ```sh
   git clone https://github.com/alifsuryadi/Review_Product_Tokopedia_BE
   ```
2. **Navigate to the Project Folder**:
   ```sh
   cd Review_Product_Tokopedia_BE
   ```
3. **Set Environment Variables**:

   ```sh
   cp .env.example .env
   ```

4. **Run the Application**:
   ```sh
   go run main.go
   ```
