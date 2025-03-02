###############
# Build Stage #
###############

# Using devbox as base build image
FROM jetpackio/devbox:latest AS build-stage

# Installing your devbox project
WORKDIR /code
RUN mkdir -p /code && chown ${DEVBOX_USER}:${DEVBOX_USER} /code
USER ${DEVBOX_USER}:${DEVBOX_USER}
COPY --chown=${DEVBOX_USER}:${DEVBOX_USER} devbox.json devbox.json
COPY --chown=${DEVBOX_USER}:${DEVBOX_USER} devbox.lock devbox.lock
RUN devbox install
USER root:root

# Copy core files
COPY . .

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
COPY --from=build-stage /code/apps/webapp/db-ip-to-country.csv.gz .
ENTRYPOINT ["./main"]