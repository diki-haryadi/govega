FROM golang:1.18-alpine
RUN apk update && apk add --no-cache git openssh

# add credentials on build
RUN echo -e "[url \"git@bitbucket.org:\"]\n\tinsteadOf = https://bitbucket.org/" >> /root/.gitconfig

ARG SSH_PRIVATE_KEY
ENV SSH_PRIVATE_KEY=$SSH_PRIVATE_KEY

# Make SSH Dir
RUN mkdir ~/.ssh/

# Create id_rsa from string arg, and set permissions
RUN echo "$SSH_PRIVATE_KEY" > ~/.ssh/id_rsa
RUN chmod 600 ~/.ssh/id_rsa

# make sure your entities is accepted
RUN touch ~/.ssh/known_hosts
RUN apk update && apk add openssh
RUN ssh-keyscan bitbucket.org >> ~/.ssh/known_hosts
WORKDIR /app

# Download Go modules
COPY . .
RUN go mod tidy
RUN go build -o /main main.go
EXPOSE ${PORT}

# Run
CMD [ "/main" ]