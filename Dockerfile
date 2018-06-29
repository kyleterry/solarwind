FROM golang:1.10-alpine
WORKDIR /go/src/github.com/kyleterry/solarwind
COPY . .
RUN apk --no-cache add bash make git
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin/solarwind .

FROM alpine:3.4
COPY --from=0 /go/src/github.com/kyleterry/solarwind/bin/solarwind /usr/bin/solarwind
ENTRYPOINT ["solarwind"]
