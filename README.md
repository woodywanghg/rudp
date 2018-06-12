```
  ________  ___  ___  ________  ________
|\   __  \|\  \|\  \|\   ___ \|\   __  \
\ \  \|\  \ \  \\\  \ \  \_|\ \ \  \|\  \
 \ \   _  _\ \  \\\  \ \  \ \\ \ \   ____\
  \ \  \\  \\ \  \\\  \ \  \_\\ \ \  \___|
   \ \__\\ _\\ \_______\ \_______\ \__\
    \|__|\|__|\|_______|\|_______|\|__|
    
```
## Reliable udp proto
![Overview](https://github.com/woodywanghg/gitpicture/blob/master/overview_ss.png)

## Overview
* **Application data**:&emsp;the data that you want to transmint
* **Reliable UDP**:&emsp;&emsp;&ensp;loss packet check, retransmission, and so on
* **Status monitor**:&emsp;&nbsp;http interface. loss rate, retransmission statistics
* **Bin protocol**:&emsp;&emsp;&emsp;bin protocol, packet by protocolbuf
* **UDP socket**:&ensp;&emsp;&ensp;&emsp;golang udp socket
