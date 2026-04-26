# Quickstart: Local Development with Docker

This guide explains how to set up and run the project using the Dockerized local development environment.

## Prerequisites

-   [Docker](https://docs.docker.com/get-docker/)
-   [Docker Compose](https://docs.docker.com/compose/install/)
-   `make`

## Setup

1.  **Clone the repository:**
    ```bash
    git clone https://github.com/anglesson/simple-web-server.git
    cd simple-web-server
    ```

2.  **Set up the environment file:**
    -   Copy the template file:
        ```bash
        make setup-env
        ```
    -   Edit the `.env` file and fill in the required values, especially the `SESSION_AUTH_KEY` and `SESSION_ENC_KEY`.

3.  **Start the environment:**
    ```bash
    make up
    ```
    This command will build the Docker images and start the `app` and `postgres` containers in the background.

## Usage

-   **Start the hot-reload development server:**
    ```bash
    make dev
    ```
    The application will be available at `http://localhost:8080`.

-   **Stop the environment:**
    ```bash
    make down
    ```

-   **View logs:**
    ```bash
    make logs
    ```
