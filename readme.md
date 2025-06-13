# MeshCentral Client

This is a simple client for [MeshCentral](https://github.com/Ylianst/MeshCentral) that allows you to interact with the MeshCentral server via the Websocket API. This project is not affiliated with MeshCentral in any way.

_This project is for personal use and is not intended to be all encompassing. It is a simple client that allows me to interact with MeshCentral in a way that I find useful._

## Functionality

* List / search devices
* Meshrouter replacment (tcp port forward, udp support possible eventually)
* Connect to devices via SSH
* Cross platform (Windows and Linux, macos not tested)

## Usage

Tool is under active development. The best way to see the currently available commands is to run `mcc help`.

That being said, here is a rough outline of the usage:

```bash
# Start a route with a specified nodeid
$ mcc route -L 8080:127.0.0.1:80 -i <nodeid>

# Don't know the nodeid? Search for it interactively
$ mcc search

# Don't want to search and route separately? Just exclude the nodeid and it will prompt you to search
$ mcc route -L 8080:127.0.0.1:80

# Want to see all the devices?
$ mcc ls

# SSH directly to a device (supports interactive mode as well)
$ mcc ssh -i <nodeid>

# SSH as a proxy, useful for VSCode remote development
$ mcc ssh -i <nodeid> --proxy

# SSH to a device that the mesh node can see but doesn't have a nodeid (useful for network devices)
$ mcc ssh user@192.168.1.1 -i <nodeid>
```

### Explaining the Port Forward

Usage is very similar to the `ssh` command (thats why the flag is `-L`). The format is as follows:

```
8080:127.0.0.1:80
^    ^         ^
|    |         |--- Destination port
|    |------------- Destination IP (optional, can be excluded)
|------------------ Local port (also optional, random port will be assigned)
```

## Contribute / Build

This project leverages devbox. To start a development shell:

```bash
$ devbox shell
```

To build the project, run:

```bash
$ devbox run build-linux
```

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
