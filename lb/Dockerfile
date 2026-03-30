FROM golang:1.20-alpine AS build
WORKDIR /src
COPY . .
RUN go build -o /predictive-sentinel ./main.go

FROM alpine:3.18
RUN addgroup -S app && adduser -S -G app app
COPY --from=build /predictive-sentinel /usr/local/bin/predictive-sentinel
EXPOSE 8080
USER app
ENTRYPOINT ["/usr/local/bin/predictive-sentinel"]
