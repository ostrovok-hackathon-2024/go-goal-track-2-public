FROM python:3.12-slim

WORKDIR /app

# Copy the poetry files
COPY pyproject.toml poetry.lock ./

# Install poetry and dependencies
RUN pip install poetry && \
    poetry config virtualenvs.create false && \
    poetry install --with cli

# Copy the application code
COPY src ./src
COPY config.yaml ./
COPY artifacts ./artifacts

# Set the entrypoint to run the CLI script using poetry
ENTRYPOINT ["poetry", "run", "cli"]