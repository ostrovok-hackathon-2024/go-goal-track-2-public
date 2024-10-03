FROM python:3.12-slim

WORKDIR /app

RUN pip install poetry==1.4.2

ENV POETRY_NO_INTERACTION=1 \
    POETRY_VIRTUALENVS_IN_PROJECT=1 \
    POETRY_VIRTUALENVS_CREATE=1 \
    POETRY_CACHE_DIR=/tmp/poetry_cache

COPY artifacts ./artifacts

COPY python ./python

WORKDIR /app/python

RUN poetry check
# Install poetry and dependencies
RUN pip install poetry && \
    poetry config virtualenvs.create false && \
    poetry install --with server

ENV PORT=8000

# Set the entrypoint to run the server script using poetry
ENTRYPOINT ["poetry", "run", "server"]
