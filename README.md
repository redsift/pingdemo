# Ping demo with PF_RING

## Packages

```
$ go get github.com/google/gopacket
```

## Running

As this is a simple demo, there is no IP and MAC resolution. You need to get hold of your gateway MAC and interface MAC/IP.

Note: requires root

```
sudo -E go run pfring.go --device=<eth device> --srcIp=<from> --dstIp=<to> --srcMAC=aa:aa:aa:aa:aa:aa --dstMAC=bb:bb:bb:bb:bb:bb
```