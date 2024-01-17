FROM golang:1.21 as go-builder


# Install libsdl2-dev
RUN apt-get update && apt-get install -y --no-install-recommends \
  libsdl2-dev \
  libsdl2-image-dev \
  libsdl2-ttf-dev \
  libsdl2-mixer-dev

RUN rm -rf /var/lib/apt/lists/* && apt-get clean

WORKDIR /app
COPY . .

RUN go mod download
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o main .

ENTRYPOINT ["/app/main"]
