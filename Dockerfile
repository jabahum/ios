FROM golang:1.24

WORKDIR /var/www/integrated-outbreak-system/app

RUN go install github.com/air-verse/air@v1.61.7

COPY . .

RUN go mod tidy