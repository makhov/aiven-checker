aiven-checker
-------------

aiven-checker is a system that monitors website availability over the network, 
produces metrics about this and passes these events through an Aiven Kafka instance 
into an Aiven PostgreSQL database (well, through any Kafka instance into any PostgreSQL database, really)

aiven-checker consists of two parts:
- `checker` performs checks and produces results to Kafka
- `writer` consumes Kafka topic and writes the check result into a PostgreSQL

CI/CD pipeline is on GitHub Actions: https://github.com/makhov/aiven-checker/actions

# Usage

Run `checker`:
```bash
$ docker run --name checker \
    docker.pkg.github.com/makhov/aiven-checker/ac \
    /checker
```
Run `writer`:
```
$ docker run --name writer \
    docker.pkg.github.com/makhov/aiven-checker/ac \
    /writer 
```

# Configuration

Config system is based on environment variables. Dotenv file is supported.
```bash
$ cat .env
DEBUG=true
KAFKA_BROKERS=127.0.0.1:9092
TASKS_FILE_PATH=tasks.e2e.json
DATABASE_DSN="user=postgres host=127.0.0.1"
MIGRATIONS_DIR=./migrations
```

# Tasks file

Tasks file is a simple JSON file with the following format: 

```json
{
  "tasks": [
    {
      "url": "<url to check>",
      "period": "<check period>",
      "regexp": "regexp to check (optional)"
    },
    …
  ]
}
```

See [tasks.e2e.json](./tasks.e2e.json) for more examples.

# Tests

There are several types of tests:
- Unit tests
- Integration tests to test integration with database
- End-to-end tests to test the whole system

You can run tests in docker-compose with all dependencies:
```bash
$ ./scripts/run_test.sh unit
$ ./scripts/run_test.sh int
$ ./scripts/run_test.sh e2e
```

Or locally:
```bash
$ make test-unit
$ make test-int
$ make test-e2e
```

# Migrations

aiven-checker uses [goose](https://github.com/pressly/goose) for migrations. If `MIGRATIONS_DIR` is declared, migrations will be applied when the `writer` starts. 
