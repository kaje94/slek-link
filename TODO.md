# Resources Used:

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

## Getting Started (Documentation)

You can run the application locally either using Kubernetes or by running commands directly. Follow these steps:

### Prerequisites

1. **Create .env File**: Refer the [.env.example](.env.example) file in the root of the repository.
2. **Load Environment Variables**: Install [direnv](https://direnv.net/) and run `direnv allow` to load the `.envrc` or `.env` file.
3. **Install Devbox**: Install [Devbox](https://www.jetpack.io/devbox/) and run `devbox shell` to install the required packages and tools. (Optionally, install the [Devbox VSCode extension](https://marketplace.visualstudio.com/items?itemName=jetpack-io.devbox) if you use VSCode).
