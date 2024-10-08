# О! Хакатон — go-goal – Тегирование тарифов

> Описание трека:
>
> - [Тегирование тарифов](https://docs.ostrovok.tech/s/hackathon-track-2)

## Описание решения

> Наше решение использует градиентный бустинг классификаторной модели и TF-IDF для векторизации строк с сублинейным ростом и нормализацией.
>
> Такой подход дарит нам скорость, гибкость, эффективность, масштабируемость. Он наиболее приближен к реальному продуктовому решению, гибкий и позволяющий менять параметры обучения на ходу, эффективный и легковесный.
>
> Масштабируемость достигается благодаря легковесности и гибкости: добавление новой категории в уже существующий набор делается в пару действий добавлением весов и конфига.
>
> Предоставлено множество выходных форматов: `csv`, `tsv`, `json`, `yaml`, `parquet`.
>
> Вместе с CLI, есть API решение с гибкой конфигурацией запроса.
>
> Мы предлагаем высокоскоростные варианты реализации как на Python, так и на Go, с возможностью дальнейшей оптимизации стоимости использования.

```mermaid
sequenceDiagram
    participant CLI
    participant API
    participant Backend
    participant Classifier1
    participant Classifier2
    participant ... as ...
    participant ClassifierN

    CLI->>API: Request
    API->>Backend: Process request
    
    par Parallel classification
        Backend->>Classifier1: Classify
        Backend->>Classifier2: Classify
        Backend-->>...: Classify
        Backend->>ClassifierN: Classify
    end

    Classifier1-->>Backend: Result
    Classifier2-->>Backend: Result
    ...-->>Backend: Results
    ClassifierN-->>Backend: Result
    
    Backend-->>API: Aggregated results
    API-->>CLI: Response
```

### Важно! Используется строгий нейминг для моделей и лейблов.

Должны соответствовать следующему формату: `catboost_model_<category>.cbm`, `labels_<category>.json`, `label_encoder_<category>.npy`.

### Модели

Находятся в [директории](./artifacts/cbm)

### Лейблы

Находятся в [директории](./artifacts/labels)

В зависимости от того, какой язык программирования выбран для решения, необходимо выбрать соответствующую директорию с лейблами. (npy/json).
NPY - для решения на Python, JSON - для решения на Go.

### Tfidf

Находятся в [директории](./artifacts/tfidf)

## Как запускать

### Установка зависимостей

Производится посредством средств языков `go` и `poetry` для решения на Python.

### Установка библиотеки CatBoost для Go

Для корректной работы моделей на Go необходимо иметь библиотеку `catboost` в системе.

> Prebuilt shared library (`*.so` | `*.dylib`) artifacts are available of the [releases](https://github.com/catboost/catboost/releases) page on GitHub CatBoost project.\
> The shared library:
>
> 1. Should be in `/usr/local/lib`
> 2. Or set path in environment `CATBOOST_LIBRARY_PATH`
> 3. Or set path manual in source code `SetSharedLibraryPath` (see example below)
>
> For more information, see <https://catboost.ai/en/docs/concepts/c-plus-plus-api_dynamic-c-pluplus-wrapper>.

### Docker Image

#### Пример сборки Docker Image для решения на Go

```bash
cd go
docker build -t cli-app-go -f cli.Dockerfile .
```

#### Пример сборки Docker Image для решения на Python

```bash
cd python
docker build -t cli-app-python -f cli.Dockerfile .
```

#### Пример запуска Docker Image

```bash
docker run --rm cli-app-go -i "Example Rate Name" -c bedding -c quality
```

#### Пример запуска Docker Image для решения на Python

```bash
docker run --rm cli-app-python -i "Example Rate Name" -c bedding -c quality
```

#### Сборка и запуск Docker Image для серверного решения на Go

```bash
cd go
docker build -t server-app-go -f api.Dockerfile .
docker run --rm -p 8000:8000 server-app-go
```

#### Пример запуска Docker Image для серверного решения на Python

```bash
cd python
docker build -t server-app-python -f server.Dockerfile .
docker run --rm -p 8080:8080 server-app-python
```

### Пример запуска

#### На Go

```bash
cd go
go mod install
go run cmd/cli/main.go -i ../inputs/rates_short.csv -c bedding -c quality
```

#### На Python

```bash
cd python
poetry install
poetry run cli -i ../inputs/rates_short.csv --categories bedding --categories quality
```

## Другие комментарии

### Код обучения моделей

Код обучения моделей находится в [файле](./python/notebooks/data_analysis.ipynb)

### Код использования моделей внутри Jupyter

Код использования моделей внутри Jupyter находится в [файле](./python/notebooks/model_evaluation.ipynb)

### Решение реализованное на Go

Подробное описание можно найти в [описании решения](./go/README.md)

#### Мини-демонстрация работы на Go

![Go](./go/docs/file.gif)
![Go](./go/docs/stdout.gif)
![Go](./go/docs/format.gif)

### Решение реализованное на Python

Подробное описание можно найти в [описании решения](./python/README.md)

#### Мини-демонстрация работы на Python

![Python](./python/docs/stdout.gif)
![Python](./python/docs/file.gif)
![Python](./python/docs/format.gif)

### Бенчмарки

#### Бенчмарки для решения на Go в [директории](./go/bench)

Запуск бенчмарков для решения на Go производится посредством команды:

```bash
cd go
chmod +x ./bench/benchmark.sh
./bench/benchmark.sh
```

Подробные результаты бенчмарков описаны в файле [README.md](./go/bench/README.md).

#### Бенчмарки для решения на Python в [директории](./python/bench)

Запуск бенчмарков для решения на Python производится посредством команды:

```bash
cd python
chmod +x ./bench/benchmark.sh
./bench/benchmark.sh
```

Подробные результаты бенчмарков описаны в файле [README.md](./python/bench/README.md).
