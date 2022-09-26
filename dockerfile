FROM golang:latest
WORKDIR /app
COPY go.mod ./
COPY go.sum ./
RUN go mod download
COPY *.go ./
RUN go build -o /paddle-game-in-go
CMD ["/paddle-game-in-go"]