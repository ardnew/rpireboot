# rpireboot
#### Raspberry Pi service to reboot system on GPIO interrupt

## Installation
Download the source code:
```
go get -u -v github.com/ardnew/rpireboot
```
The `install` make target will compile the Go application and install+start the systemd service:
```
cd "$GOPATH/src/github.com/ardnew/rpireboot"
make install
```
The Go application will automatically configure the GPIO pin 18 as interrupt input, pull-up, and falling edge detection, but you may want to configure the pin in your Raspberry Pi configuration file (`/boot/config.txt`):
```
gpio=18=ip,pu
```

## Configuration
To change the GPIO pin used, or its interrupt parameters, the Go application accepts command line flags (use `-h` for a list of what's available). You will need to modify the systemd service file (and your `/boot/config.txt`) accordingly.
