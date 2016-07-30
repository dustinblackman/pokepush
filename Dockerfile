FROM alpine:3.4

WORKDIR /app
COPY ./pokepush /app/pokepush

CMD ["/app/pokepush"]
