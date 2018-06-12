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
|Module|Description|
|-|-|
| **Application data**|the data that you want to transmint|
| **Reliable UDP**|loss packet check, retransmission, and so on|
| **Status monitor**|http interface. loss rate, retransmission statistics|
| **Bin protocol**|bin protocol, packet by protocolbuf|
| **UDP socket**|golang udp socket|


## Message sequence
![message sequence](https://github.com/woodywanghg/gitpicture/blob/master/msgsequence.svg)

## Message retransmit sequence
![message sequence](https://github.com/woodywanghg/gitpicture/blob/master/sequenceretransmit_s4.svg)

