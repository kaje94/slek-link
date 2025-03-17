###############
# Build Stage #
###############

# Using devbox as base build image
FROM jetpackio/devbox:0.14.0 AS build-stage

# Installing your devbox project
WORKDIR /code
RUN mkdir -p /code && chown ${DEVBOX_USER}:${DEVBOX_USER} /code
USER ${DEVBOX_USER}:${DEVBOX_USER}
COPY --chown=${DEVBOX_USER}:${DEVBOX_USER} /libs/devbox-minimal/devbox.json devbox.json
COPY --chown=${DEVBOX_USER}:${DEVBOX_USER} /libs/devbox-minimal/devbox.lock devbox.lock
RUN devbox install
USER root:root

# Copy core files
COPY . .
RUN devbox run -- echo "Installed Packages."
RUN devbox run -- bun install --cwd ./apps/webapp/static
RUN devbox run -- templ generate

# Build go app
ENV CGO_ENABLED 0
ENV GOOS linux
RUN devbox run task webapp:build

###############
## Run stage ##
###############

# Deploy the application binary into a lean image
FROM cgr.dev/chainguard/static:latest AS run-stage
COPY --from=build-stage /code/apps/webapp/bin/main .
COPY --from=build-stage /code/apps/webapp/static/public /static/public
ENTRYPOINT ["./main"]