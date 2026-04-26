# Task Plan: Dockerized PostgreSQL for Local Development

This document outlines the tasks required to implement the Dockerized PostgreSQL for local development feature.

## Phase 1: Foundational Setup

These tasks set up the core Docker and environment configuration.

- [X] T001 Create the `docker-compose.yml` file in the project root.
- [X] T002 [P] Add the `postgres` service to `docker-compose.yml`, including image, environment variables, volume, and port mapping.
- [X] T003 [P] Add the `app` service to `docker-compose.yml`, including build context, dependency on `postgres`, port mapping, and volume for hot-reloading.
- [X] T004 Modify the `env.template` file to replace `DATABASE_URL` with PostgreSQL connection variables (`DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`).
- [X] T005 Update the `.dockerignore` file to ensure `.env` is excluded from the Docker build context.

## Phase 2: User Story 1 - Developer Environment Setup

**Goal**: As a new developer, I want to set up my local environment with a single command.
**Independent Test**: A developer can clone the repo, run `make up`, and have the application running and connected to a persistent PostgreSQL database.

- [X] T006 [US1] Add the `up` command to the `Makefile` to run `docker-compose up -d`.
- [X] T007 [US1] Add the `down` command to the `Makefile` to run `docker-compose down`.
- [X] T008 [US1] Add the `logs` command to the `Makefile` to run `docker-compose logs -f`.
- [X] T009 [US1] Modify the existing `dev` command in the `Makefile` to first execute the `up` command and then start the `air` hot-reload server.
- [X] T010 [US1] Update the main `README.md` with a new "Local Development with Docker" section, explaining the prerequisites and the new `make` commands (`up`, `down`, `dev`, `logs`).

## Phase 3: Polish & Cross-Cutting Concerns

These tasks cover final testing and documentation.

- [ ] T011 Manually test the complete developer setup flow as described in the updated `README.md` and the `Testing Plan` section of `plan.md`.
- [ ] T012 Review and merge all changes.

## Dependencies

- **User Story 1 (US1)** is dependent on the completion of all tasks in **Phase 1**.

## Parallel Execution

- Within Phase 1, tasks **T002** and **T003** can be worked on in parallel after **T001** is complete.

## Implementation Strategy

The implementation will follow an MVP-first approach. The primary goal is to deliver User Story 1, which provides the core value of a one-command developer setup.

- **MVP**: Complete all tasks in Phase 1 and Phase 2. This will provide a fully functional, containerized local development environment.
- **Post-MVP**: The polish phase ensures the new workflow is well-tested and documented.
