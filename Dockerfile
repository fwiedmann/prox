FROM alpine

WORKDIR /code

COPY prox .

RUN apk add libcap && setcap 'cap_net_bind_service=+ep' /code/prox

ENTRYPOINT [ "./prox" ]