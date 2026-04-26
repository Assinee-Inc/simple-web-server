# Implementation Plan: Dockerized PostgreSQL for Local Development

## Technical Context

The project currently uses SQLite for local development and PostgreSQL for production. The goal is to use Docker to run a PostgreSQL database locally to align the development and production environments. The project does not have a `docker-compose.yml` file, so one will be created. The `Makefile` will be updated to manage the Dockerized environment, and `env.template` will be updated with the new database connection string.

## Constitution Check

This change aligns with the constitution's "Technology Stack" which lists PostgreSQL as the production database. By using PostgreSQL in development, we improve consistency and reduce environment-specific bugs, which indirectly supports the "Testability" and "Modularity" principles.

---

## Phase 0: Research & Design

This phase outlines the necessary research and design decisions for implementing the Dockerized PostgreSQL environment.

### 1. `docker-compose.yml` Structure

A new `docker-compose.yml` file will be created with the following services:

-   **`postgres`**:
    -   Image: `postgres:13-alpine`
    -   Environment variables for the database name, user, and password. These will be sourced from the `.env` file.
    -   A volume to persist the database data between container restarts.
    -   Port mapping to expose the PostgreSQL port to the host machine.
-   **`app`**:
    -   Builds from the local `Dockerfile`.
    -   Depends on the `postgres` service.
    -   Maps the application port.
    -   Loads environment variables from the `.env` file.
    -   Mounts the project directory as a volume to allow for hot-reloading.

### 2. `Makefile` Commands

The `Makefile` will be updated with the following commands:

-   **`up`**: Starts the Docker containers using `docker-compose up -d`.
-   **`down`**: Stops the Docker containers using `docker-compose down`.
-   **`logs`**: Tails the logs of the services using `docker-compose logs -f`.
-   **`dev`**: This command will be updated to first run `up` and then start the `air` hot-reload server.

### 3. `env.template` Changes

The `env.template` file will be updated to include the necessary variables for the PostgreSQL connection:

-   `DATABASE_URL` will be changed to a PostgreSQL connection string.
-   `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME` will be added.

---

## Phase 1: Implementation Plan

This phase details the step-by-step implementation of the feature.

1.  **Create `docker-compose.yml`**:
    -   Create a new file named `docker-compose.yml` in the project root.
    -   Add the `postgres` and `app` services as designed in Phase 0.

2.  **Modify `env.template`**:
    -   Open `env.template`.
    -   Replace `DATABASE_URL=./mydb.db` with the new PostgreSQL variables.

3.  **Modify `Makefile`**:
    -   Open the `Makefile`.
    -   Add the `up`, `down`, and `logs` commands.
    -   Modify the `dev` command.

4.  **Update `README.md`**:
    -   Add a new section to the `README.md` explaining how to set up and run the project using the new Dockerized environment.
    -   Include instructions on installing Docker and Docker Compose.
    -   Explain the new `make` commands.

5.  **Update `.dockerignore`**:
    -   Ensure `.env` and other sensitive files are included in `.dockerignore`.

---

## Phase 2: Testing Plan

This phase describes how to test the implementation to ensure it meets the requirements.

1.  **Environment Setup Test**:
    -   Follow the new instructions in the `README.md`.
    -   Run `make setup-env` and fill in the `.env` file.
    -   Run `make up`.
    -   Verify that the `postgres` and `app` containers start successfully.

2.  **Database Connection Test**:
    -   Access the application and perform an action that interacts with the database (e.g., creating a user).
    -   Verify that the data is saved correctly.

3.  **Data Persistence Test**:
    -   Run `make down` to stop the containers.
    -   Run `make up` again to restart the containers.
    -   Verify that the data created in the previous step is still present.

4.  **Hot-Reload Test**:
    -   Run `make dev`.
    -   Make a change to a Go source file.
    -   Verify that the application automatically rebuilds and restarts.
