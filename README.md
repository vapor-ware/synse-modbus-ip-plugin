[![Build Status](https://build.vio.sh/buildStatus/icon?job=vapor-ware/synse-modbus-ip-plugin/master)](https://build.vio.sh/blue/organizations/jenkins/vapor-ware%2Fsynse-modbus-ip-plugin/activity)
[![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2Fvapor-ware%2Fsynse-modbus-ip-plugin.svg?type=shield)](https://app.fossa.io/projects/git%2Bgithub.com%2Fvapor-ware%2Fsynse-modbus-ip-plugin?ref=badge_shield)
![GitHub release](https://img.shields.io/github/release/vapor-ware/synse-modbus-ip-plugin.svg)

# Synse Modbus TCP/IP Plugin

A plugin for Modbus over TCP/IP for [Synse Server][synse-server].

This plugin is a general-purpose plugin, meaning that there are no device-specific
implementations for this plugin. Instead, a set of default handlers are provided. Registering
devices with the plugin is then simply a matter of passing in the correct configuration.

> Note: By default, the plugin SDK will search for a device handler based on the name of the
> device type. Here, the device handlers for each device type should be manually overridden
> in order to get the functionality required for those devices/outputs. See the supported
> handlers and example config, below.

## Getting Started

### Getting

You can install the modbus plugin via a [release](https://github.com/vapor-ware/synse-modbus-ip-plugin/releases)
binary or via Docker image

```
docker pull vaporio/modbus-ip-plugin
```

If you wish to use a development build, fork and clone the repo and build the plugin
from source.

### Running

The modbus plugin comes with a set of sane [default plugin configurations](config.yml), so
you should be able to run the plugin without much additional configuration. You will, however
need to provide device configurations.

This repo includes a [compose file](docker-compose.yml) which provides a basic example of how
to run the modbus plugin with Synse Server and how to configure devices with it. There is no
emulated modbus backend, so all device reads/writes will fail with this compose file, but it
can serve as a good point of reference.

To run, simply:

```bash
docker-compose up -d
```

You can then use Synse's HTTP API or the [Synse CLI][synse-cli] to query Synse for plugin data.

## Modbus Plugin Configuration

Plugin and device configuration are described in detail in the [SDK Documentation][sdk-docs].

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

| Field         | Required            | Type   | Description                                         |
| ------------- | ------------------- | ------ | --------------------------------------------------- |
| `host`        | yes                 | string | The hostname/ip of the modbus server to connect to. |
| `port`        | yes                 | int    | The port number for the modbus server to connect to. |
| `slaveId`     | yes                 | int    | The modbus slave id for the device. |
| `address`     | yes                 | int    | The register address which holds the output reading. |
| `width`       | yes                 | int    | The number of registers to read, starting from the `address`. |
| `type`        | yes                 | string | The type of the data held in the registers (see below). |
| `timeout`     | no (default: 5s)    | string | The duration to wait for a modbus request to resolve. |
| `failOnError` | no (default: false) | bool   | Fail the entire device read if a single output read fails. |

> By default, `failOnError` is false, so a failure to read a single register will cause that
> failure to be logged, but will *not* cause the entire bulk read to fail. If this is set to true,
> all registers must be successfully read in order for the read to complete.

The values that are supported in the `type` field are as follows:

| Type             | Description             |
| ---------------- | ----------------------- |
| `u32`, `uint32`  | unsigned 32-bit integer |
| `u64`, `uint64`  | unsigned 64-bit integer |
| `s32`, `int32`   | signed 32-bit integer |
| `s64`, `int64`   | signed 64-bit integer |
| `f32`, `float32` | 32-bit floating point number |
| `f64`, `float64` | 64-bit floating point number |

Note that typically, an `*32` type will have width 2 while an `*64` type will have width 4.

### Outputs

Outputs are referenced by name. A single device may have more than one instance
of an output type. A value of `-` in the table below indicates that there is no value
set for that field. The *custom* section describes outputs which this plugin defines
while the *built-in* section describes outputs this plugin uses which are [built-in to
the SDK](https://synse.readthedocs.io/en/latest/sdk/concepts/reading_outputs/#built-ins).

**Custom**

| Name              | Description                                               | Unit     | Type       | Precision |
| ----------------- | --------------------------------------------------------- | :------: | ---------- | :-------: |
| gallonsPerMin     | A measure of volumetric flow rate, in gallons per minute. | gpm      | `flow`     | 4         |
| inchesWaterColumn | A measure of pressure, in inches of water column.         | inch wc  | `pressure` | 8         |

**Built-in**

Since this is a general-purpose plugin, any built-in unit may be specified via device
configuration. See the link above for complete set of built-in outputs.

### Device Handlers

Device Handlers are referenced by name.

| Name             | Description                                  | Outputs | Read  | Write | Bulk Read | Listen |
| ---------------- | -------------------------------------------- | ------- | :---: | :---: | :-------: | :----: |
| coil             | A handler that reads from coils.             | any     | ✗     | ✓     | ✓         | ✗      |
| holding_register | A handler that reads from holding registers. | any     | ✗     | ✓     | ✓         | ✗      |
| input_register   | A handler that reads from input registers.   | any     | ✗     | ✗     | ✓         | ✗      |

### Write Values

This plugin supports the following values when writing to a device via a handler.

| Handler          | Write Action  | Write Data   | Description                                         |
| ---------------- | :-----------: | :----------: | --------------------------------------------------- |
| coil             | `-`           | `0`, `false` | Writing a zero (0x00) value to the register.        |
|                  | `-`           | `1`, `true`  | Writing a one value (0xff00) value to the register. |
| holding_register | `-`           | `uint16`     | The data (uint16) to write to the register.         |

### Example Device Configuration

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

## Compatibility

Below is a table describing the compatibility of plugin versions with Synse platform versions.

|             | Synse v2 | Synse v3 |
| ----------- | -------- | -------- |
| plugin v1.x | ✓        | ✗        |
| plugin v2.x | ✗        | ✓        |

## Troubleshooting

### Debugging

The plugin can be run in debug mode for additional logging. This is done by:

- Setting the `debug` option  to `true` in the plugin configuration YAML ([config.yml](config.yml))

  ```yaml
  debug: true
  ```

- Passing the `--debug` flag when running the binary/image

  ```
  docker run vaporio/modbus-ip-plugin --debug
  ```

- Running the image with the `PLUGIN_DEBUG` environment variable set to `true`

  ```
  docker run -e PLUGIN_DEBUG=true vaporio/modbus-ip-plugin
  ```

### Developing

A [development/debug Dockerfile](Dockerfile.dev) is provided in the project repository to enable
building image which may be useful when developing or debugging a plugin. Unlike the slim `scratch`-based
production image, the development image uses an ubuntu base, bringing with it all the standard command line
tools one would expect. To build a development image:

```
make docker-dev
```

The built image will be tagged using the format `dev-{COMMIT}`, where `COMMIT` is the short commit for
the repository at the time. This image is not published as part of the CI pipeline, but those with access
to the Docker Hub repo may publish manually.

## Contributing / Reporting

If you experience a bug, would like to ask a question, or request a feature, open a
[new issue](https://github.com/vapor-ware/synse-modbus-ip-plugin/issues) and provide as much
context as possible. All contributions, questions, and feedback are welcomed and appreciated.

## License

The Synse Modbus-IP Plugin is licensed under GPLv3. See [LICENSE](LICENSE) for more info.

[![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2Fvapor-ware%2Fsynse-modbus-ip-plugin.svg?type=large)](https://app.fossa.io/projects/git%2Bgithub.com%2Fvapor-ware%2Fsynse-modbus-ip-plugin?ref=badge_large)

[synse-cli]: https://github.com/vapor-ware/synse-cli
[synse-server]: https://github.com/vapor-ware/synse-server
[plugin-dockerhub]: https://hub.docker.com/r/vaporio/modbus-ip-plugin
[plugin-release]: https://github.com/vapor-ware/synse-modbus-ip-plugin/releases
[sdk-docs]: https://synse.readthedocs.io/en/latest/sdk/intro/
