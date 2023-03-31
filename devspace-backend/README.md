# Devspace Backend

This is the backend for Devspace, a platform for managing and deploying projects in a Kubernetes cluster.

## Stack

The backend is built with the following technologies:

- Go 1.16
- Gin web framework
- MongoDB for persistence
- Jenkins for CI/CD
- Kubernetes for orchestration
- Elasticsearch and Kibana for logging

## Getting Started

To run the backend locally, follow these steps:

1. Clone this repository.
2. Make sure you have Go 1.16 installed.
3. Install the required packages: `go get ./...`
4. Create a MongoDB instance and set the `MONGO_URI` environment variable to its connection string.
5. Set the `JENKINS_URL`, `JENKINS_USERNAME`, and `JENKINS_PASSWORD` environment variables to connect to a Jenkins instance.
6. Set the `KUBERNETES_CONFIG` environment variable to the path of a Kubernetes config file.
7. Set the `ELASTICSEARCH_URL` environment variable to connect to an Elasticsearch instance.
8. Run the app: `go run app.go`

## Structure

The project is structured according to hexagonal architecture, with files split by domain. Here is an overview of the directory structure:

- `app/`: Contains the entry point for the application (`app.go`), as well as the handlers for each API endpoint (`handlers/`), repositories for each external service (`repositories/`), and use cases for each business operation (`usecases/`).
- `domain/`: Contains the domain model (`model/`), repository interfaces (`repository/`), and use case interfaces (`usecase/`).
- `interfaces/`: Contains the interfaces for each external service (`persistence/` and `api/handlers/`) and the HTTP server (`server.go`).

## API Endpoints

### Projects

- `POST /projects`: Create a new project.

### Scopes

- `POST /projects/:project_id/scopes`: Create a new scope.
- `POST /projects/:project_id/scopes/:scope_id/deploy`: Deploy a project to a scope.
- `GET /projects/:project_id/scopes/:scope_id`: View a scope.
- `GET /projects/:project_id/scopes/:scope_id/logs`: View logs for a scope.

### Builds

- `POST /projects/:project_id/builds`: Build a project.
- `GET /projects/:project_id/builds/:build_id`: View build status.

### Deploys

- `POST /projects/:project_id/scopes/:scope_id/deploys`: Deploy a project to a scope.
- `GET /projects/:project_id/scopes/:scope_id/deploys/:deploy_id`: View deploy status.

## License

This project is licensed under the MIT License.
