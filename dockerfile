FROM golang:alpine3.21

WORKDIR /app

COPY . .

RUN go mod tidy \
    && go build

EXPOSE 80

CMD [ "./test" ]