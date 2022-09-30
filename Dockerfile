# Build
FROM golang:1.18.3-alpine
RUN go install github.com/dwisiswant0/crlfuzz/cmd/crlfuzz@latest

ENTRYPOINT [ "crlfuzz" ]