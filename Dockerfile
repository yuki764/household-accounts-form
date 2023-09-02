FROM golang:1.21.0-bookworm AS build
WORKDIR /go/src/app/
COPY go.mod go.sum ./
RUN go mod download
COPY *.go ./
RUN CGO_ENABLED=0 go build

FROM gcr.io/distroless/static-debian12:latest
LABEL maintainer "Setuu <setuu@neigepluie.net>"
WORKDIR /app/
COPY --from=build /go/src/app/household-accounts-form .
COPY *.html.tpl .
ENV TZ=Asia/Tokyo
EXPOSE 8080
CMD ["/app/household-accounts-form"]
