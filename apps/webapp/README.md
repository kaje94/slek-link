# slek-link: webapp

This workspace contains the web application for slek-link built using [Go Templ](https://templ.guide/)

## Getting Started

1. This project uses [Devbox](https://www.jetpack.io/devbox/) for managing dependencies and task execution.
2. Make sure that RabbitMQ and PosgreSQL depedencies are running locally before running the web app.

### Available Commands

Run the following command to see all available tasks:

```sh
devbox run task
```

### Directory Structure

- **internal/**
  - **components/** – Reusable Templ components
  - **config/** – Configuration files containing environment variables
  - **handlers/** – API, Datastar, AsyncAPI, and authentication handlers
  - **pages/** – Templ components for page routes
- **utils/** – Common utility functions
- **static/** – Static assets (JS, CSS, images, etc.)
