FROM golang:latest

COPY . /go/gelf-go

RUN cd gelf-go && go build -o bin/gelf-go

CMD ["gelf-go/bin/gelf-go"]
