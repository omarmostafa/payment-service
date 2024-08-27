# Payment Gateway Microservice

## Overview

This microservice provides integration with two payment gateways: **Stripe** and **Authorize.Net**. It manages **deposit (cash-in)** and **withdrawal (cash-out)** operations and handles asynchronous callbacks from the payment gateways.

The service is designed to be easily extendable, allowing for the addition of more payment gateways in the future. It routes requests based on the provider specified in the API request and ensures robust error handling and transaction logging.

## Design Document
- https://fixed-dinghy-1ef.notion.site/Payment-Gateway-Microservice-Design-Document-7c2d7c9337ac4612bace2782d5225476
## High-Level Architecture

- **Microservice:** The core of the service, responsible for handling deposit and withdrawal operations.
- **Payment Gateways:**
    - **Stripe:** Communicates with Stripe's REST API for transactions.
    - **Authorize.Net:** Communicates with Authorize.Net using SOAP/XML over HTTP.
- **Routing Logic:** Based on the `provider` parameter in the API request, the service routes the transaction to the appropriate payment gateway.
- **Webhooks:**
    - **Stripe Webhook:** Listens for asynchronous updates from Stripe.
    - **Authorize.Net Webhook:** Listens for asynchronous updates from Authorize.Net.
- **Dockerized Deployment:** The service is containerized for easy deployment and scalability.

## Setup Instructions

### Prerequisites

- Docker
- Docker Compose

### Build and Run the Service

1. Clone the repository:


   ```bash
   git clone git@github.com:omarmostafa/payment-service.git
   cd payment-service
   ```

2. Create a `.env` file in the root directory and add the variables which is exist in .env.example:

2. Build and start the service using Docker Compose:

    ```bash
   docker-compose up --build
    ```

3. The service will be available at `http://localhost:8080/`.

## API Endpoints

- **Deposit Endpoint:**
    - **POST** `/api/deposit`
    - **Description:** Handles deposit (cash-in) requests.
    - **Parameters:** `amount`, `provider`, `currency`, etc.

- **Withdrawal Endpoint:**
    - **POST** `/api/withdraw`
    - **Description:** Handles withdrawal (cash-out) requests.
    - **Parameters:** `amount`, `provider`, `currency`, etc.

## Webhook Endpoints

- **Stripe Webhook:**
    - **POST** `/stripe-webhook`
    - **Description:** Listens for asynchronous events from Stripe.

- **Authorize.Net Webhook:**
    - **POST** `/authorize-webhook`
    - **Description:** Listens for asynchronous events from Authorize.Net.

## Swagger UI

The API documentation is available via Swagger UI. You can access it at:
`http://localhost:8080/swagger/index.html#/`

## Future Enhancements

- **Additional Gateways:** The service is designed to easily integrate with more payment gateways as needed.
- **Adding Unit tests**.