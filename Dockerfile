FROM ubuntu

ARG server

ENV server=$server

COPY $server .

ENTRYPOINT /${server} 2
