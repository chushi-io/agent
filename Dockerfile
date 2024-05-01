FROM golang:1.22

WORKDIR /app
COPY go.mod go.sum /app/
RUN go mod download
# TODO: Just copy over required files
COPY . ./

RUN CGO_ENABLED=0 GOOS=linux go build -o /agent ./cmd/agent/*

ENTRYPOINT ["/agent"]