FROM golang
WORKDIR /etc/apiNetworkDelayMonitor/
ADD . .
RUN go mod tidy && \
    go install
ENTRYPOINT ["apiNetworkDelayMonitor"]


#FROM alpine
#WORKDIR /app
#COPY --from=0 /src/config.yaml /app/
#COPY --from=0 /src/dist/app    /app/
#EXPOSE 8080
#ENTRYPOINT ["/app/app"]