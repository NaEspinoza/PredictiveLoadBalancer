FROM golang:1.20-alpine AS build
WORKDIR /src
COPY . .
RUN go build -o /backend ./main.go

FROM alpine:3.18
RUN addgroup -S app && adduser -S -G app app
COPY --from=build /backend /usr/local/bin/backend
EXPOSE 9001
USER app
ENTRYPOINT ["/usr/local/bin/backend"]
