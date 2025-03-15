# devbox-services: Local Development Configurations

This workspace provides Devbox configurations for essential services like **PostgreSQL**, **RabbitMQ**, and **Valkey**, enabling you to run the web application locally.

## Troubleshooting Common Issues

Here are solutions to common problems encountered when starting these services with Devbox:

- **PostgreSQL: `.devbox/virtenv/postgresql/data exists but is not empty`**

  If you see this error, it indicates that a previous PostgreSQL data directory exists and is preventing a clean start. To resolve this:

  1.  Delete the `.devbox/virtenv/postgresql/data` directory.
  2.  Restart the PostgreSQL service using Devbox.

- **RabbitMQ: `failed_to_parse_configuration_file`**

  This error often occurs due to a corrupted or incorrect `enabled_plugins` file. To fix it:

  1.  Delete the `devbox.d/jetify-com.devbox-plugins.rabbitmq/conf.d/enabled_plugins` file.
  2.  Restart the RabbitMQ service. Devbox will regenerate the necessary configuration.

## Further Assistance

If you encounter other issues:

- Consult the official [Devbox Services Guide](https://www.jetify.com/docs/devbox/guides/services/) for detailed information on each service.
- If you cannot find a solution in the documentation, please report the issue on the [Devbox GitHub repository](https://github.com/jetify-com/devbox/issues).
