FROM golang
WORKDIR /src
ADD . .
RUN go mod tidy && \
    go build -o dist/app

FROM alpine
WORKDIR /app
COPY --from=0 /src/config.yaml /app/
COPY --from=0 /src/dist/app    /app/
EXPOSE 8080
ENTRYPOINT ["/app/app"]