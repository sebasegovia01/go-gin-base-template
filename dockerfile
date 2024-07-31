# Etapa de compilación
FROM golang:1.22 AS builder

# Establecer el directorio de trabajo
WORKDIR /app

# Copiar los archivos go.mod y go.sum
COPY go.mod go.sum ./

# Descargar las dependencias
RUN go mod download

# Copiar el código fuente
COPY . .

# Compilar la aplicación
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# Etapa final
FROM alpine:latest  

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copiar el ejecutable compilado desde la etapa de compilación
COPY --from=builder /app/main .

# Exponer el puerto en el que se ejecuta tu aplicación
EXPOSE 8080

# Comando para ejecutar la aplicación
CMD ["./main"]
