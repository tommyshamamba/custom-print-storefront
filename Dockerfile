FROM golang:1.22 AS build
WORKDIR /src
COPY go.mod ./
COPY backend-go ./backend-go
RUN CGO_ENABLED=0 go build -o /server ./backend-go/cmd/server

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /server /server
EXPOSE 8080
ENTRYPOINT ["/server"]
