Covert Channel
==============

To create the server:

Build the code.
```
go build main.go
```

Run two instances of the server with different websocket ports (must run in supper user mode).
In one terminal, run:
```
sudo ./main -p 8080
```
In a second terminal, run:
```
sudo ./main -p 8081
```

Open up `test.html` twice in separate browser tabs.

You will be prompted for the port numbers. Choose the two websocket ports chosen 
earlier (in our case, 8080 and 8081).

The config will open for the TCP covert channel. You will see inputs for FriendPort and
OriginPort. These should start at 8123 and 8124. In one of the open tabs, switch these 
ports (to 8124 and 8123 for the FriendPort and OriginPort). This will set up a channel that
is complementary to the other channel.

In both taps, click the `Open` button. If everything works, you should now be able to 
exchange messages by typing into the input and clicking the `Send` button.

Received messages will appear in the `textarea`.

Covert Channel
==============

The following contains a simple covert channel implemented
with the TCP protocol.

The library provides several features to allow configuration
of message transmission:

- Message Delimiter: The channel can either read until a buffer is filled
	(i.e. all messages have fixed size) or it can read until a termination
	packet is received.
- Message Bouncing: The channel can be setup to bounce messages off of an
	intermediate TCP socket to help obscure source and destination.
- Encoding configuration: A preliminary system is setup to allow users
	to control how the bytes are encoded in the TCP header.
	By providing custom encoders, users can select which portion of the 
	header contains the byte. By default the data is held in the sequence
	number.
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
cd go\_covert\_lib
go build sender.go
go build receiver.go
```
Doing so will build the two examples.

The IP addresses default to the local machine. In one terminal
run:
```
sudo ./receiver
```

In a second terminal run:
```
sudo ./sender
```

In the sender window you are now able to send messages to the 
receiver. Write you message and hit enter to see the message 
appear in the receiver terminal. The receiver is set to timeout
every 10 seconds if no message is received, at which point
it will print a message to the terminal and wait for another
message.
