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
RUN mkdir -p /noted/codes/kernels
ADD . /noted-notes
RUN chmod a=r /noted-notes/internal/db/sql/requests/*
RUN chmod a=r /noted-notes/internal/db/sql/requests
RUN chown -R notes-runner:notes-runner /noted/codes
WORKDIR /noted-notes

# Сборка
RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=1 \
    go build -o ./bin/noted-notes ./cmd/main

# Запуск в пустом контейнере
FROM gcr.io/distroless/cc-debian12

# Копируем пользователя без прав с прошлого этапа
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder --chown=notes-runner:notes-runner /noted-notes/internal/db/sql/requests /requests 
COPY --from=builder --chown=notes-runner:notes-runner /noted/codes /noted/codes
# Запускаем от имени этого пользователя
USER notes-runner

COPY --from=builder /noted-notes/bin/noted-notes /noted-notes

CMD ["/noted-notes"]
