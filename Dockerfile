FROM golang:1.15 AS builder
WORKDIR /app/
COPY go.mod /app/
RUN go mod download
COPY *.go *.html.tpl /app/
RUN go build

FROM debian:buster-slim
LABEL maintainer "Setuu <setuu@neigepluie.net>"
RUN apt-get update \
 && apt-get install -y --no-install-recommends \
    ca-certificates \
 && apt-get clean \
 && rm -rf /var/lib/apt/lists/* \
 && update-ca-certificates
COPY --from=builder /app/household-accounts-form /app/*.html.tpl /app/
WORKDIR /app/
CMD /app/household-accounts-form
EXPOSE 8080
ENV TZ=Asia/Tokyo
