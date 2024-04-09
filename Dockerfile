FROM golang:1.21-alpine AS builder

ENV CGO_ENABLED=0

RUN apk add --no-cache ca-certificates git curl

RUN mkdir /user && \
    echo 'nobody:x:65534:65534:nobody:/:' > /user/passwd && \
    echo 'nobody:x:65534:' > /user/group

WORKDIR /opt/datamin

COPY go.mod ./
COPY go.sum ./

RUN go mod download

COPY . .

RUN go build ./cmd/datamin-integration.go

FROM golang:1.21-alpine AS final

COPY --from=builder /usr/local/bin /usr/local/bin

COPY --from=builder /user/group /user/passwd /etc/

COPY --from=builder /opt /opt

USER nobody:nobody

EXPOSE 3338

WORKDIR /opt/datamin

ENTRYPOINT [ "/opt/datamin/datamin-integration" ]

CMD [ "run-kafka-trigger"]
