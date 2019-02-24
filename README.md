Test WebSocket Scaling:
=======================

This is just some tests around scaling websocket servers.
By default docker is in use.

### Requirements:
- [docker](https://www.docker.com/)
- docker-compose

### Usage:
1. Build images:
```bash
docker-compose build
```

2. Run containers:
```bash
docker-compose up
```

Open browser on [http://localhost:3000](http://localhost:3000)
and create one more tab on the same url. On the fist tab press
the button "Connect to server 1", on the second tab
press the button "Connect to server 2". It will connect to a
corresponding websocket server. After that just type something
in the text area and submit the form. Your should see that
message was send across the servers.

To manage queue messages and channels use
[http://localhost:4171/](http://localhost:4171/). It's a nsq
admin panel


Tested on docker:
```
 Version:          18.09.1
  API version:      1.39 (minimum version 1.12)
  Go version:       go1.10.6
  Git commit:       4c52b90
  Built:            Wed Jan  9 19:41:49 2019
  OS/Arch:          linux/amd64
  Experimental:     false
```