[![Build Status](https://build.vio.sh/buildStatus/icon?job=vapor-ware/synse-modbus-ip-plugin/master)](https://build.vio.sh/blue/organizations/jenkins/vapor-ware%2Fsynse-modbus-ip-plugin/activity)
[![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2Fvapor-ware%2Fsynse-modbus-ip-plugin.svg?type=shield)](https://app.fossa.io/projects/git%2Bgithub.com%2Fvapor-ware%2Fsynse-modbus-ip-plugin?ref=badge_shield)
![GitHub release](https://img.shields.io/github/release/vapor-ware/synse-modbus-ip-plugin.svg)

# Synse Modbus TCP/IP Plugin

A plugin for ModBus over TCP/IP for Synse Server.

This plugin is a general-purpose plugin, meaning that there are no device-specific
implementations for this plugin. Instead, a set of default handlers are provided. Registering
devices with the plugin is then simply a matter of passing in the correct configuration.

> Note: By default, the plugin SDK will search for a device handler based on the name of the
> device type. Here, the device handlers for each device type should be manually overridden
> in order to get the functionality required for those devices/outputs. See the supported
> handlers and example config, below.

## Plugin Support

### Outputs

Outputs are referenced by name. Configured device instances specify an output so its device
handler can generate the correct reading type for the raw value of the device read.

Below is a table detailing the Outputs defined by the plugin. For an accounting of Outputs
built-in to the SDK, see [builtins.go](https://github.com/vapor-ware/synse-sdk/blob/master/sdk/output/builtins.go).

| Name | Type | Description | Unit | Precision |
| ---- | ---- | ----------- | ---- | --------- |
| `gallonsPerMin` | flow | Volumetric flow rate reading, measured in gallons per minute. | gallons per minute (gpm) | 4 |
| `inchesWaterColumn` | pressure | Pressure reading, measured in inches of water column. | inches of water column (inch wc) | 8 |

### Device Handlers

Device Handlers define how registers are read from/written to. For this plugin, each device must
explicitly define the device handler it will use, e.g. `handler: input_register`.
Device Handlers should be referenced by name. For examples, see the [example](#example-config)
section below.

| Name | Description | Read | Write | Bulk Read |
| ---- | ----------- | ---- | ----- | --------- |
| `coil` | A handler that reads from coils. | ✗ | ✓ | ✓ |
| `holding_register` | A handler that reads from holding registers. | ✗ | ✓ | ✓ |
| `input_register` | A handler that reads from input registers. | ✗ | ✗ | ✓ |

## Getting Started

### Getting the Plugin

It is recommended to run the plugin in a Docker container. You can pull the image from
[DockerHub][plugin-dockerhub]:

```
docker pull vaporio/modbus-ip-plugin
```

You can also download a plugin binary from the latest [release][plugin-release].

### Running the Plugin

If you are using the plugin binary:

```bash
# The name of the plugin binary may differ depending on whether it is built
# locally or a pre-built binary is used.
$ ./plugin
```

If you are using the docker image:

```bash
$ docker run vaporio/modbus-ip-plugin
```

In either of the above cases, the plugin will run but should ultimately fail because
they are missing device configurations. It is up to you to provide the configurations for
your deployment. For an example, see [docker-compose.yml](docker-compose.yml) and
[devices.yml](example/device/devices.yml).

Since the plugin is a general-use plugin, the device handlers is provides are not specific
to any device. The configured devices must choose the correct handler (see the
[table above](#device-handlers) for supported handlers), and must provide the correct info
(e.g. register address, read width, etc). See the section below for an example configuration.

## Configuration

Device and plugin configuration are described in the [Synse SDK Documentation][sdk-docs].

There is an additional config scheme specific to this plugin for the contents of a configured
device's `data` field. Device `data` may be specified in two places (the prototype config and
the instance config sections). The data scheme describes the resulting unified config from
both sources.

An example:

```yaml
devices:
- type: temperature
  handler: input_register
  data:
    host: 127.0.0.1
    port: 502
    slave_id: 3
    timeout: 5s
    failOnError: false
  instances:
  - info: Temperature Sensor 1
    output: temperature
    data:
      address: 500
      width: 2
      type: f32
  - info: Temperature Sensor 2
    output: temperature
    data:
      address: 502
      width: 2
      type: f32
```

For `Temperature Sensor 1`, the unified data config is:

```yaml
host: 127.0.0.1
port: 502
slave_id: 3
timeout: 5s
failOnError: false
address: 500
width: 2
type: f32
```

| Field | Required | Type | Description |
| ----- | -------- | ---- | ----------- |
| `host` | yes | string | The hostname/ip of the modbus server to connect to. |
| `port` | yes | int | The port number for the modbus server to connect to. |
| `slaveId` | yes | int | The modbus slave id for the device. |
| `address` | yes | int | The register address which holds the output reading. |
| `width` | yes | int | The number of registers to read, starting from the `address`. |
| `type` | yes | string | The type of the data held in the registers (see below). |
| `timeout` | no (default: 5s) | string | The duration to wait for a modbus request to resolve. |
| `failOnError` | no (default: false) | bool | Fail the entire device read if a single output read fails. |

By default, `failOnError` is false, so a failure to read a single register will cause that
failure to be logged, but will *not* cause the entire bulk read to fail. If this is set to true,
all registers must be successfully read in order for the read to complete.

The values that are supported in the `type` field are as follows:

| Type | Description |
| ---- | ----------- |
| `u32`, `uint32` | unsigned 32-bit integer |
| `u64`, `uint64` | unsigned 64-bit integer |
| `s32`, `int32` | signed 32-bit integer |
| `s64`, `int64` | signed 64-bit integer |
| `f32`, `float32` | 32-bit floating point number |
| `f64`, `float64` | 64-bit floating point number |

Note that typically, an `x32` type will have width 2 while an `x64` type will have width 4.

### Example Config

This section shows an example configuration for an eGauge 4115 Power Metering device. It exposes
readings for voltage and frequency via this config.

```yaml
# Sample Config
# -------------

# The config scheme version
version: 3

# Define the device prototype(s) and their instance(s).
devices:
  - type: power
    context:
      model: egauge 4115
    handler: input_register
    data:
      host: 127.0.0.1
      port: 502
      slave_id: 3
    instances:
      # RMS Voltage
      - info: Leg 1 to neutral RMS voltage
        output: voltage
        data:
          address: 500
          width: 2
          type: f32
      - info: Leg 2 to neutral RMS voltage
        output: voltage
        data:
          address: 502
          width: 2
          type: f32

      # Line Frequency
      - info: L1 line frequency
        output: frequency
        data:
          address: 600
          width: 2
          type: f32
      - info: L2 line frequency
        output: frequency
        data:
          address: 602
          width: 2
          type: f32
```

## Feedback

Feedback for this plugin, or any component of the Synse ecosystem, is greatly appreciated!
If you experience any issues, find the documentation unclear, have requests for features,
or just have questions about it, we'd love to know. Feel free to open an issue for any
feedback you may have.

## Contributing

We welcome contributions to the project. The project maintainers actively manage the issues
and pull requests. If you choose to contribute, we ask that you either comment on an existing
issue or open a new one.

## License

This plugin, and all other components of the Synse ecosystem, is released under the
[GPL-3.0](LICENSE) license.

[![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2Fvapor-ware%2Fsynse-modbus-ip-plugin.svg?type=large)](https://app.fossa.io/projects/git%2Bgithub.com%2Fvapor-ware%2Fsynse-modbus-ip-plugin?ref=badge_large)

[plugin-dockerhub]: https://hub.docker.com/r/vaporio/modbus-ip-plugin
[plugin-release]: https://github.com/vapor-ware/synse-modbus-ip-plugin/releases
[sdk-docs]: http://synse-sdk.readthedocs.io/en/latest/user/configuration.html
