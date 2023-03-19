# Use official golang image as base image
FROM golang:1.19

# Set working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum files into container
COPY go.mod .
COPY go.sum .

# Download dependencies
RUN go mod download

# Copy the rest of the application into the container
COPY . .

# Build the application
RUN go build -o main .

# Expose port to listen on
EXPOSE $PORT

# Set environment variable for Sendinblue API key
ENV API_KEY=${API_KEY}

# Set environment variable for Sendinblue sender email address
ENV SEND_EMAIL=${SEND_EMAIL}

# Set environment variable for recipient email address
ENV RECEIVE_EMAIL=${RECEIVE_EMAIL}

# Start the application
CMD ["./main"]