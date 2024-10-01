FROM python:3.12-slim

RUN pip install poetry==1.4.2

ENV POETRY_NO_INTERACTION=1 \
    POETRY_VIRTUALENVS_IN_PROJECT=1 \
    POETRY_VIRTUALENVS_CREATE=1 \
    POETRY_CACHE_DIR=/tmp/poetry_cache

WORKDIR /app

COPY . .

RUN poetry check
# Install dependencies
RUN poetry config virtualenvs.create false \
    && poetry install --with cli --no-interaction --no-ansi

# Set the entrypoint to run the CLI script using poetry
ENTRYPOINT ["poetry", "run", "cli"]
