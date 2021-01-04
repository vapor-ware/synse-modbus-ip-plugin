#
# Builder Image
#
FROM vaporio/golang:1.13 as builder

#
# Final Image
#
#FROM vaporio/scratch-ish:1.0.0
FROM ubuntu

LABEL org.label-schema.schema-version="1.0" \
      org.label-schema.name="vaporio/modbus-ip-plugin" \
      org.label-schema.vcs-url="https://github.com/vapor-ware/synse-modbus-ip-plugin" \
      org.label-schema.vendor="Vapor IO"

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

# Add in default plugin configuration.
COPY config.yml /etc/synse/plugin/config/config.yml

# Copy the executable.
COPY synse-modbus-ip-plugin ./plugin

# Put vi on there for now.
RUN apt-get update && apt-get upgrade -y
RUN apt-get install -y vim

ENTRYPOINT ["./plugin"]
