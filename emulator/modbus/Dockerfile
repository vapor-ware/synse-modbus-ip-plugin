#
# Builder Image.
#
FROM docker.io/vaporio/python:3.11 as builder
COPY requirements.txt .

ARG ARCH=amd64

RUN pip install --no-deps --target=/build/lib/python3.11 --platform ${ARCH} -r requirements.txt --no-warn-script-location \
 && rm -rf /root/.cache

# Copy source to builder. Install modbus through setup.py.
COPY . /modbus
RUN pip install --no-deps --prefix=/build --no-warn-script-location /modbus \
 && rm -rf /root/.cache

#
# Modbus Emulator Image.
#

# Use the slim image. Copy in installed modbus from the builder.
FROM docker.io/vaporio/python:3.11-slim
COPY --from=builder /build /usr/local

# Start modbus server / emulator app to test against.
# Port 1502 was chosen because it does not require admin privileges.
# modbus is normally port 502.
EXPOSE 1502
ENTRYPOINT ["modbus"]
