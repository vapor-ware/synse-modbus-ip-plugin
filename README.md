[![CircleCI](https://circleci.com/gh/vapor-ware/synse-modbus-ip-plugin.svg?style=shield)](https://circleci.com/gh/vapor-ware/synse-modbus-ip-plugin)
[![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2Fvapor-ware%2Fsynse-modbus-ip-plugin.svg?type=shield)](https://app.fossa.io/projects/git%2Bgithub.com%2Fvapor-ware%2Fsynse-modbus-ip-plugin?ref=badge_shield)

# Synse Modbus-IP Plugin
A plugin for ModBus over TCP/IP for Synse Server.

This plugin is a general-purpose plugin, meaning that there are no device-specific
implementations for this plugin. Instead, a set of default handlers are provided. Registering
devices with the plugin is then simply a matter of passing in the correct configuration.

> Note: By default, a device kind will search for a handler based on the name of the
> device kind. Here, the device handlers for each device kind should be manually overridden
> in order to get the functionality required for those devices/outputs. See the supported
> handlers and example config, below.

## Plugin Support
### Outputs
Outputs should be referenced by name. A single device can have more than one instance
of a single output type. A value of `-` indicates that there is no value set for that field.

| Name | Description | Unit | Precision | Scaling Factor |
| ---- | ----------- | ---- | --------- | -------------- |
| current | An output type for current (amp) readings. | ampere (A) | 3 | - |
| voltage | An output type for voltage (volt) readings. | volt (V) | 3 | - |
| si-to-kwh.power | An output type for power readings that converts from SI to kWh. | kilowatt hour (kWh) | 5 | 2.77777778e-7 |
| power | An output type for power (W) readings. | watt (W) | 3 | - |
| frequency | An output type for frequency (Hz) readings. | hertz (Hz) | 3 | - |


### Device Handlers
Device Handlers define how registers are read from/written to. Each device should
specify the device handler it will use via the override key, e.g. `handlerName: foobar`.
Device Handlers should be referenced by name.

| Name | Description | Read | Write | Bulk Read |
| ---- | ----------- | ---- | ----- | --------- |
| input_register | A handler that reads from input registers. | ✓ | ✗ | ✗ |


## Getting Started
### Getting the Plugin
You can get the Modbus-IP plugin either by cloning this repo, setting up the project dependencies,
and building the binary or docker image

```bash
# Setup the project
$ make setup

# Build the binary
$ make build

# Build the docker image
$ make docker
```

You can also use a pre-built docker image from [DockerHub][plugin-dockerhub]
```bash
$ docker pull vaporio/modbus-ip-plugin
```

Or a pre-built binary from the latest [release][plugin-release].

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
your deployment.

Since the plugin is a general-use plugin, the device handlers is provides are not specific
to any device. The devices that are specified must choose the correct handler (see the table
above for supported handlers), and must provide the correct info (e.g. register address, read
width, etc). See the section below for an example configuration.

## Configuration
Device and plugin configuration are described in detail in the [Synse SDK Documentation][sdk-docs].

There are two additional plugin-specific config schemes for this plugin: the device data
config (the config in the `data` field of a device instance) and the device output data
config (the config for an output type's `data` field). Examples of these configs are shown
in the next sections.

### Device Data
The device data is specified in the `data` field of an instance, e.g.
```yaml
instances:
  - info: example device
    location: r1b1
    data:
      host: 127.0.0.1
      port: 502
      slaveId: 3
      timeout: 15s
      failOnError: false
```

The supported fields for this config are:

| Field | Required | Type | Description |
| ----- | -------- | ---- | ----------- |
| host  | yes | string | The hostname/ip of the device to connect to. |
| port  | yes | int | The port number for the device to connect to. |
| slaveId | yes | int | The modbus slave id for the device. |
| timeout | no (default: 5s) | string | The duration to wait for a modbus request to resolve. |
| failOnError | no (default: false) | bool | Fail the device entire read if a single output read fails. |


### Device Output Data
The device output data is specified in the `data` field of a device output config, e.g.
```yaml
outputs:
  - type: voltage
    info: Leg 1 to neutral RMS voltage
    data:
      address: 500
      width: 2
      type: f32
```

The supported fields for this config are:

| Field | Required | Type | Description |
| ----- | -------- | ---- | ----------- |
| address  | yes | int | The register address which holds the output reading. |
| width  | yes | int | The number of registers to read, starting from the `address`. |
| type | yes | string | The type fot the data held in the registers. |

The type values that are supported in the `type` field are as follows:

| Type(s) | Description |
| ---- | ----------- |
| `u32`, `uint32` | unsigned 32-bit integer |
| `u64`, `uint64` | unsigned 64-bit integer |
| `s32`, `int32` | signed 32-bit integer |
| `s64`, `int64` | signed 64-bit integer |
| `f32`, `float32` | 32-bit floating point number |
| `f64`, `float64` | 64-bit floating point number |

### Example Config
This section shows an example configuration for an eGauge 4115 Power Metering device. It exposes
readings for voltage and frequency via this config.

```yaml
# Sample Config
# -------------

# The config scheme version
version: 1.0

# Define the location(s) of the device(s)
locations:
  - name: r1b1
    rack:
      name: rack-1
    board:
      name: board-1

# Define the device kinds, instances, and outputs being used.
devices:
  - name: eg4115.power
    instances:
      - info: eGauge 4115 power meter - facility power monitoring
        location: r1b1
        # Specify the name of the handler we will be using
        handlerName: input_register
        data:
          host: 127.0.0.1
          port: 502
          slaveId: 3
          timeout: 15s
          failOnError: false
        outputs:
          # RMS Voltage
          - type: voltage
            info: Leg 1 to neutral RMS voltage
            data:
              address: 500
              width: 2
              type: f32
          - type: voltage
            info: Leg 2 to neutral RMS voltage
            data:
              address: 502
              width: 2
              type: f32
          - type: voltage
            info: Leg 3 to neutral RMS voltage
            data:
              address: 504
              width: 2
              type: f32

          # Line Frequency
          - type: frequency
            info: L1 line frequency
            data:
              address: 1500
              width: 2
              type: f32
          - type: frequency
            info: L2 line frequency
            data:
              address: 1502
              width: 2
              type: f32
          - type: frequency
            info: L3 line frequency
            data:
              address: 1504
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

This plugin, and all other components of the Synse ecosystem, is released under the
[GPL-3.0](LICENSE) license.


[plugin-dockerhub]: https://hub.docker.com/r/vaporio/modbus-ip-plugin
[plugin-release]: https://github.com/vapor-ware/synse-modbus-ip-plugin/releases
[sdk-docs]: http://synse-sdk.readthedocs.io/en/latest/user/configuration.html


## License
[![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2Fvapor-ware%2Fsynse-modbus-ip-plugin.svg?type=large)](https://app.fossa.io/projects/git%2Bgithub.com%2Fvapor-ware%2Fsynse-modbus-ip-plugin?ref=badge_large)