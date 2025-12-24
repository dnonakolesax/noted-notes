# Установка модулей и тесты
FROM golang:1.24.7 AS modules

ADD go.mod go.sum /m/
RUN cd /m && go mod download

# RUN make test

# Сборка приложения
FROM golang:1.24.7 AS builder

COPY --from=modules /go/pkg /go/pkg

# Пользователь без прав
RUN useradd -u 10001 notes-runner

RUN mkdir -p /noted-notes
RUN mkdir -p /notes-runner/.aws
ADD . /noted-notes
WORKDIR /noted-notes

RUN go run cmd/migration/main.go
# Сборка
RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 \
    go build -o ./bin/noted-notes ./cmd/main

# Запуск в пустом контейнере
FROM scratch

# Копируем пользователя без прав с прошлого этапа
COPY --from=builder /etc/passwd /etc/passwd
# Запускаем от имени этого пользователя
USER notes-runner

COPY --from=builder /noted-notes/bin/noted-notes /noted-notes
COPY --from=builder /noted-notes/internal/db/sql/requests /internal/db/sql/requests

CMD ["/noted-notes"]
