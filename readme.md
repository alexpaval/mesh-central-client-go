# MeshCentral Client

This is a simple client for [MeshCentral](https://github.com/Ylianst/MeshCentral) that allows you to interact with the MeshCentral server via the Websocket API. This project is not affiliated with MeshCentral in any way.

_This project is for personal use and is not intended to be all encompassing. It is a simple client that allows me to interact with MeshCentral in a way that I find useful._

## Functionality

* List / search devices
* Meshrouter replacment (tcp port forward, udp support possible eventually)
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
```

### Explaining the Port Forward

Usage is very similar to the `ssh` command (thats why the flag is `-L`). The format is as follows:

```
8080:127.0.0.1:80
^    ^         ^
|    |         |--- Destination port
|    |------------- Destination IP (optional, can be excluded)
|------------------ Local port
```

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
