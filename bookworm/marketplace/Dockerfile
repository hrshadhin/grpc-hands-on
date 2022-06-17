ARG PYTHON_VERSION=3.10
FROM python:${PYTHON_VERSION}-slim-bullseye

ARG APP_PORT=5000
LABEL maintainer="dev@hrshadhin.me"


RUN mkdir /service && adduser --disabled-password --no-create-home -q app

COPY marketplace/requirements.txt /service/requirements.txt

RUN python -m venv /py && \
    /py/bin/pip install --upgrade pip && \
    /py/bin/pip install -r /service/requirements.txt

ENV PYTHONUNBUFFERED 1
ENV PATH="/py/bin:$PATH"
ENV FLASK_APP=marketplace.py
ENV FLASK_RUN_PORT=${APP_PORT}

COPY protobufs/ /service/protobufs/
COPY marketplace /service/marketplace
WORKDIR /service/marketplace
RUN python -m grpc_tools.protoc -I ../protobufs --python_out=. --grpc_python_out=. ../protobufs/recommendations.proto


EXPOSE ${APP_PORT}
USER app
ENTRYPOINT [ "flask", "run", "--host=0.0.0.0"]