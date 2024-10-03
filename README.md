# О! Хакатон — go-goal – Тегирование тарифов

> Описание трека:
>
> - [Тегирование тарифов](https://docs.ostrovok.tech/s/hackathon-track-2)

## Описание решения

> Общее описание, на чём основано

## Как запускать

### Установка зависимостей

Производится посредством средств языка `go` и `poetry` для решения на Python.

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

Код обучения моделей находится в директории `python/notebooks/data_analysis.ipynb`.

### Код использования моделей внутри Jupyter

Код использования моделей внутри Jupyter находится в `python/notebooks/model_evaluation.ipynb`.

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
