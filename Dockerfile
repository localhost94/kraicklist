FROM golang:alpine

WORKDIR /usr/src/myapp
COPY . .

RUN go get -d -v ./...
RUN go install -v ./...

EXPOSE 3001

CMD ["main"]
