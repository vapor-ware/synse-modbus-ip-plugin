FROM iron/go:dev as builder
WORKDIR /go/src/github.com/vapor-ware/synse-modbus-ip-plugin
COPY . .
RUN make build


FROM iron/go
LABEL maintainer="vapor@vapor.io"

WORKDIR /plugin

COPY --from=builder /go/src/github.com/vapor-ware/synse-modbus-ip-plugin/build/plugin ./plugin
COPY config.yml .

EXPOSE 5001

CMD ["./plugin"]
