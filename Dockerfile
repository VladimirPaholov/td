
FROM golang:1.24 AS builder

WORKDIR /app


COPY go.mod go.sum ./
RUN go mod download

COPY . .


RUN go build -o final_project ./main.go


FROM ubuntu:latest

WORKDIR /app


COPY --from=builder /app/final_project .
COPY --from=builder /app/web ./web
COPY --from=builder /app/scheduler.db ./scheduler.db

EXPOSE 7540

CMD ["./final_project"]

#docker run -p 7540:7540 --name my_app(имя контейнера) final_project