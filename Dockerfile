# we are setting up a base image here.
FROM golang:1.25.6-alpine AS builder

# creating a working directory where the code will be present
WORKDIR /app

# Docker caches layers, so your deps only re-download when go.mod/go.sum change, not every time you edit source
COPY go.mod go.sum ./

RUN go mod download

COPY . .

# tells Go to produce a pure Go binary with zero C dependencies
RUN CGO_ENABLED=0 GOOS=linux go build -o market ./cmd/webserver

# now we create a small container that can run the binary we are creating right above
FROM alpine:3.20

# creating a directory for our container
WORKDIR /app

# copy the details from our builder 
COPY --from=builder /app/market .

# just a hint saying "hey, this app binds to 8080" — so developers know what to put on the right side of -p.
EXPOSE 8080

# shouldn't this be entrypoint?
ENTRYPOINT ["./market"] 
# CMD ["./market"]       # can be overridden: docker run image ls
# ENTRYPOINT ["./market"] # container IS the binary; harder to accidentally override