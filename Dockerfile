FROM golang:1.19

WORKDIR /srv
COPY go.mod ./
COPY go.sum ./
RUN go mod download
COPY *.go ./
RUN go build -o /main

EXPOSE 10500

CMD [ "/main" ]