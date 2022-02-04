FROM golang:1.17-buster as build

WORKDIR /app

COPY go.mod ./
COPY go.sum ./

RUN go mod download

COPY ./cmd ./cmd
COPY ./internal ./internal

RUN go build -o /mergebot ./cmd/mergebot

######### run stage #########
FROM gcr.io/distroless/base-debian10 as run

WORKDIR /
COPY --from=build /mergebot  /mergebot

USER nonroot:nonroot

ENTRYPOINT ["/mergebot"]
