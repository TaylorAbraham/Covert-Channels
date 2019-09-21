Covert Channel
==============

The following contains a simple covert channel. The basic
idea is to send a message to a destination indirectly. 

This program passes messages between a sender and receiver
by coding the message in the sequence number of TCP
SYN packets.

With this program, the sender and receiver can send messages
directly. However, they may also work in bounce mode where the
sender and receiver send messages to an intermediary TCP
server, the bouncer. 

In a normal TCP connection, the client sends a TCP SYN 
packet to the server. The server then replies with a TCP 
SYN/ACK packet. The handshake is complete when the client 
replies with a TCP ACK packet. 

In bouncer mode the sender takes advantage of this handshake to send 
messages to the receiver. The sender first crafts a normal
TCP SYN packet, which means it has the SYN bit set.
However, instead of providing the proper source IP address,
it provides the desired receiver IP address in the source IP
of the IP header. Thus, when the TCP server receives the SYN 
packet, it sends the SYN/ACK to the receiver, instead of the 
sender. This allows messages to be passed indirectly between 
sender and receiver. The sender hides the chars it wants to 
send in the sequence number of the TCP header. The sequence 
number is used in TCP to prevent duplicate packets from 
arriving. In the TCP handshake, the TCP server increments 
the sequence number and places it in the acknowledgement
number of the SYN/ACK packet. Thus to get the message the
receiver just has to read the acknowledgement numbers of 
consecutive SYN/ACK packets that it receives from the 
bouncer.

The library provides several features to allow configuration
of message transmission:

- Message Delimiter: The channel can either read until a buffer is filled
	(i.e. all messages have fixed size) or it can read until a termination
	packet is received.
- Message Bouncing: The channel can be setup to bounce messages off of an
	intermediate TCP socket to help obscure source and destination.
- Transmission Timing: A preliminary system is setup to allow users to control
	the inter packet rate of message transmission. By providing a custom 
	function, the user can control the delay between each packet (with one
	data byte per packet) is sent. By default the time between packets is 0,
	but hypothetically users could set it to a large time or even a random 
	number based on some distribution to better match internet traffic.

# Example

To run this example you will need to install rust. 
Instructions for installing can be found online.

This demo can be run on a single machine. Build
the examples as follows.

```
cd rust\_covert\_lib
cargo test
```
Doing so will build the examples with debug level compilation.

The IP addresses default to the local machine. In one terminal
run:
```
sudo ./target/debug/examples/receiver
```

In a second terminal run:
```
sudo ./target/debug/examples/sender
```

In the sender window you are now able to send messages to the 
receiver. Write you message and hit enter to see the message 
appear in the receiver terminal. The receiver is set to timeout
every 10 seconds if no message is received, at which point
it will print a message to the terminal and wait for another
message.
