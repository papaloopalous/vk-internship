FROM tarantool/tarantool:2.10.4

USER root

RUN apk add --no-cache \
    build-base \
    cmake \
    git \
    curl \
    unzip

RUN timeout 60 tarantoolctl rocks install https://raw.githubusercontent.com/tarantool/vshard/master/vshard-scm-1.rockspec || \
    timeout 60 tarantoolctl rocks install https://raw.githubusercontent.com/tarantool/vshard/1.7.0/vshard-1.7.0-1.rockspec

RUN mkdir -p /var/lib/tarantool && \
    chown -R tarantool:tarantool /var/lib/tarantool && \
    mkdir -p /opt/tarantool && \
    chown -R tarantool:tarantool /opt/tarantool

USER tarantool

COPY --chown=tarantool:tarantool *.lua /opt/tarantool/

WORKDIR /opt/tarantool

ENV TARANTOOL_WORKDIR=/opt/tarantool \
    TARANTOOL_TIMEOUT=10 \
    TARANTOOL_INSTANCE_NAME=default

CMD ["tarantool", "/opt/tarantool/router.lua"]