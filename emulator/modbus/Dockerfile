#
# Builder
#
FROM vaporio/python:3.7 as builder
COPY requirements.txt .

ARG ARCH=amd64

RUN pip install --no-deps --target=/build/lib/python3.7 --platform ${ARCH} -r requirements.txt --no-warn-script-location \
 && rm -rf /root/.cache

# Copy source to builder. Install modbus through setup.py.
COPY . /modbus
RUN pip install --no-deps --prefix=/build --no-warn-script-location /modbus \
 && rm -rf /root/.cache

#
# Not VSM Image
#

# Use the slim image. Copy in installed modbus from the builder.
FROM vaporio/python:3.7-slim
COPY --from=builder /build /usr/local

# Start app.
EXPOSE 1502:1502
ENTRYPOINT ["modbus"]
