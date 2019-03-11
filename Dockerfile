FROM golang:1.11.5
WORKDIR /go/src/secure
COPY main.go glide.yaml glide.lock ./
COPY app app/
COPY database database/
COPY logger logger/
COPY server server/
RUN curl https://glide.sh/get | sh
RUN glide install
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

FROM alpine:3.8
WORKDIR /root/
RUN apk --no-cache add ca-certificates
COPY --from=0 /go/src/secure/main .
CMD ["./main"]
