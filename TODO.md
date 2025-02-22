# Resources Used(Under tools(other) section in Readme.md):

fonts:
icons: feather
illustrations: https://illustrationkit.com/illustrations/halo

# TODO:

add github related files (contribute.md, etc)
add golangci
add sitemap
remove direnv and have taskfile use the .env
make sure that air reloads when libs changes
add title property or daisyui tooltip to all links and buttons
remove htmx and alpinejs
add open graph tags with banner image
add Prometheus(with GORM to collect db stats)
GORM(Add indexes, migration, Logger)
logging (use slog.Error, slog.Info)
add toast after creating, deleting and copying content
update scroll wheel theme
add title/tooltips, whereever needed
Error Handling: Implement robust error handling in all services (e.g., handling database connection issues, RabbitMQ connection issues, message processing failures). Consider dead-letter queues (DLQs) in RabbitMQ for messages that fail to be processed.

## Getting Started (Documentation)

You can run the application locally either using Kubernetes or by running commands directly. Follow these steps:

### Prerequisites

1. **Create .env File**: Refer the [.env.example](.env.example) file in the root of the repository.
2. **Load Environment Variables**: Install [direnv](https://direnv.net/) and run `direnv allow` to load the `.envrc` or `.env` file.
3. **Install Devbox**: Install [Devbox](https://www.jetpack.io/devbox/) and run `devbox shell` to install the required packages and tools. (Optionally, install the [Devbox VSCode extension](https://marketplace.visualstudio.com/items?itemName=jetpack-io.devbox) if you use VSCode).
   NEED TO INSTALL BOTH DEVBOX AND DIRENV VSCODE EXTENSIONS to properly work with vscode
   But probably can skip now since devbox.json is referring to .env file
   Delete envrc file, add it to gitignore and provide command to generate that file

Examples needs to be like

```
devbox run task webapp:watch
```
