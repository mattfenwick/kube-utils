# syntax=docker/dockerfile:1

FROM python:3.10-slim-buster

WORKDIR /app

RUN pip3 install ruamel.yaml

COPY main.py main.py

ENTRYPOINT [ "python3", "main.py"]
