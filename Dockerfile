FROM python:3.6.8-alpine3.7 as dcbuilder
RUN apk add --update --no-cache mariadb-client-libs \
    && apk add --no-cache --virtual .build-deps \
    mariadb-dev \
    gcc \
    musl-dev \
    && pip install mysqlclient==1.4.2.post1 \
    && apk del .build-deps
FROM dcbuilder
COPY . /app
WORKDIR /app
RUN pip install -r requirements.txt
CMD ["python", "main.py"]
