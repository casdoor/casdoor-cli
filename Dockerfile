FROM golang:1.22.0

ARG APP_VERSION=0.0.1

WORKDIR /usr/src/casdoor-cli

# pre-copy/cache go.mod for pre-downloading dependencies and only redownloading them in subsequent builds if they change
COPY casdoor-cli/go.mod go.sum ./
RUN go mod download && go mod verify

COPY casdoor-cli/ .
RUN go build -v -o /usr/local/bin/casdoor-cli ./... -ldflags "-X gitlab.com/sdv9972401/casdoor-cli-go/cmd.Version=${APP_VERSION}"

CMD ["app"]