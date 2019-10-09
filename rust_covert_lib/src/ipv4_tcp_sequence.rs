extern crate cancellable_timer;
extern crate crossbeam;
extern crate crossbeam_channel;
extern crate pnet;
extern crate rand;
extern crate rayon;
extern crate socket2;

use pnet::transport::transport_channel;
use pnet::transport::TransportChannelType::Layer3;

use pnet::packet::ip::IpNextHeaderProtocols;
use pnet::packet::ipv4::{self, Ipv4Packet, MutableIpv4Packet};
use pnet::packet::tcp::{self, MutableTcpPacket, TcpFlags, TcpPacket};
use pnet::packet::Packet;

use std::io;
use std::net::{IpAddr, Ipv4Addr};
use std::thread;
use std::time::Duration;

use socket2::{Domain, Protocol, Socket, Type};

use cancellable_timer::Timer;

#[derive(Clone)]
pub enum Delim {
    Protocol,
    None,
}

/// The configuration of an ipv4_tcp_sequence covert channel
/// This structure recognizes two IP-port pairs, the friend and the origin.
/// The friend is the node you are sending messages to.
/// The origin is the source IP-port that the friend will read when it receives packets from you.
/// The origin is either your IP-port pair or, if we are in bounce mode, the IP-port pair
/// of the TCP server used to bounce messages off
#[derive(Clone)]
pub struct Config {
    pub friend_ip: Ipv4Addr,
    pub origin_ip: Ipv4Addr,
    pub friend_port: u16,
    pub origin_port: u16,
    /// In bounce mode, packets are not sent to the friend directly. Instead, they are sent
    /// to a bouncer running a TCP socket on the origin IP-port. The packet SYN has the source IP-port
    /// spoofed as the friend IP-port, so that when the bouncer replies with a SYN-ACK packet it will
    /// be transmitted to your friend.
    pub bounce: bool,
    /// The delimiter to use to deliniate messages. Currently it is either no deliniation (Delim::None)
    /// or delinieating by a TCP packet with a specific flag (Delim::Protocol).
    /// Default is Delim::Protocol.
    pub delimiter: Delim,
    /// A function to retrieve a delay to implement between sent packets. By default this
    /// function returns a delay of 0 ms, but users can set it to a longer time or even to
    /// their favourite distribution.
    pub get_delay: fn() -> Duration,
}

impl Config {
    /// Creates a new configuration
    /// # Arguments
    ///
    /// * `friend_ip` - The IP address of your friend.
    /// * `origin_ip` - The IP address of the message origin (either your IP or the IP of the bouncer TCP server.
    /// * `friend_port` - The port number of your friend.
    /// * `origin_port` - The port number of the message origin (either your IP or the IP of the bouncer TCP server.
    ///
    /// # Remarks
    /// This is a convenience wrapper for the default config
    /// It forces the user to set the ip addresses and port numbers
    pub fn new(
        friend_ip: Ipv4Addr,
        origin_ip: Ipv4Addr,
        friend_port: u16,
        origin_port: u16,
    ) -> Config {
        let mut def = Config::default();
        def.friend_ip = friend_ip;
        def.origin_ip = origin_ip;
        def.friend_port = friend_port;
        def.origin_port = origin_port;
        def
    }
}

impl Default for Config {
    fn default() -> Config {
        Config {
            friend_ip: Ipv4Addr::new(127, 0, 0, 1),
            origin_ip: Ipv4Addr::new(127, 0, 0, 1),
            friend_port: 0,
            origin_port: 0,
            bounce: false,
            delimiter: Delim::Protocol,
            get_delay: || Duration::from_millis(0),
        }
    }
}

/// Creates a bi-directional ipv4_tcp_sequence covert channel
/// # Arguments
///
/// * `conf` - The Config struct.
///
pub fn channel(conf: Config) -> io::Result<(Sender, Receiver)> {
    Ok((Sender { conf: conf.clone() }, Receiver { conf: conf }))
}

/// Structure for sending data
pub struct Sender {
    conf: Config,
}

/// Structure for receiving data
pub struct Receiver {
    conf: Config,
}

impl Sender {
    /// Sends a message to the friend
    /// # Arguments
    ///
    /// * `data` - The message to send.
    /// * `progress` - An optional channel to report regular updates as to the progress of the sent message.
    ///
    /// # Remarks
    /// If a progress channel is supplied, then the send method will transmit the number of sent bytes when the percentage
    /// of sent bytes increases by at least 1%. This is useful for transmissions where the
    /// get_delay function modulates this method to transmit packets at a slow rate.
    pub fn send(
        &mut self,
        data: &[u8],
        mut progress: Option<crossbeam_channel::Sender<usize>>,
        mut cancel: Option<crossbeam_channel::Receiver<()>>,
    ) -> io::Result<usize> {
        let msg_len = data.len();

        // The receive buffer does not matter, as we are not using it when sending. It can be kept small.
        let mut tx = match transport_channel(128, Layer3(IpNextHeaderProtocols::Tcp)) {
            Ok((tx, _)) => tx,
            Err(e) => return Err(e),
        };

        // In bounce mode, we spoof the source IP-port with the
        // IP-port of our friend. The destination IP-port is set to the machine that
        // we plan to bounce messages off.
        let (src_addr, dst_addr, src_port, dst_port) = match self.conf.bounce {
            true => (
                self.conf.friend_ip,
                self.conf.origin_ip,
                self.conf.friend_port,
                self.conf.origin_port,
            ),
            false => (
                self.conf.origin_ip,
                self.conf.friend_ip,
                self.conf.origin_port,
                self.conf.friend_port,
            ),
        };

        let mut curr_seq = rand::random::<u32>();
        let mut send_count = 0;
        let mut send_percent = 0;

        for c in data {
            loop {
                // Each sequence number must be different from the preceeding number so that
                // the receiver can distinguish duplicate packets being resent by the bouncer
                // when it does not receive a reply in the TCP handshake
                // We loop until we have a different number
                // We send only a single byte each message, but we randomly generate the
                // first 3 bytes of the sequence number to make it harder to distinguish from
                // normal traffic
                let new_seq = (rand::random::<u32>() & 0xFFFFFF00) | ((c & 0xFF) as u32);
                if new_seq != curr_seq {
                    curr_seq = new_seq;
                    break;
                }
            }

            // We send SYN packets because these are meant to simulate the
            // start of a TCP handshake
            let pkt: Ipv4Packet = create_packet(
                curr_seq,
                TcpFlags::SYN,
                src_addr,
                dst_addr,
                src_port,
                dst_port,
            );
            let pkt_len: usize = pkt.packet().len();

            match tx.send_to(pkt, IpAddr::V4(dst_addr)) {
                // We might find out that not sending all of the bytes is a expected occurrence even if no error occurs,
                // in which case we should modify this logic
                Ok(n) if n != pkt_len => {
                    return Err(io::Error::new(
                        io::ErrorKind::Other,
                        "Insufficient bytes size",
                    ))
                }
                Err(e) => return Err(e),
                _ => (),
            }
            send_count += 1;
            match progress {
                Some(s) => {
                    // Only sent another message if the number of sent bytes has increased by at least 1%
                    let current_percent =
                        ((send_count as f64 / msg_len as f64) * 100f64).floor() as u8;
                    if current_percent > send_percent {
                        send_percent = current_percent;
                        match s.send(send_count) {
                            _ => (),
                        }
                    }
                    // In order to avoid using a borrow progress channel,
                    // we re-wrap the channel in a new option
                    progress = Some(s)
                }
                _ => (),
            }
            // In the case of a long running transmission, we
            // must be able to cancel the operation.
            // This is done by cancelling on the sleep
            match cancel {
                Some(rx_cancel) => {
                    match cancelleable_sleep((self.conf.get_delay)(), &rx_cancel) {
                        Ok(_) => (),
                        Err(e) => return Err(e),
                    }
                    cancel = Some(rx_cancel);
                }
                None => thread::sleep((self.conf.get_delay)()),
            }
        }

        /// A function to implement a cancellable sleep
        /// # Arguments
        ///
        /// * `delay` - The time to sleep.
        /// * `rx_cancel` - The channel that the user can use to cancel the wait.
        fn cancelleable_sleep(
            delay: Duration,
            rx_cancel: &crossbeam_channel::Receiver<()>,
        ) -> Result<(), io::Error> {
            let (mut timer, canceller) = match Timer::new2() {
                Ok((timer, canceller)) => (timer, canceller),
                Err(e) => return Err(e),
            };

            return rayon::scope(|scope| {
                let (tx_local, rx_local) = crossbeam_channel::unbounded();
                scope.spawn(move |_| {
                    crossbeam::select! {
                        recv(rx_cancel) -> _ => (),
                        recv(rx_local) -> _ => (),
                    }
                    match canceller.cancel() {
                        _ => (),
                    }
                });
                let out = timer.sleep(delay);
                match tx_local.send(()) {
                    // If the above thread has returned this send will
                    // have a legitimate error, so we can just ignore it
                    _ => (),
                };
                return out;
            });
        }

        // If we are using the protocol to delimit messages, send an extra packet with the ACK flag set
        // If we are not in bounce mode, the ACK will be transmitted directly to the friend.
        // Otherwise, the ACK will be transmitted to the bouncer, which will respond by sending a RST packet
        // to the friend.
        match &self.conf.delimiter {
            Delim::Protocol => {
                // Send packet with ack bit set
                let pkt: Ipv4Packet = create_packet(
                    rand::random::<u32>(),
                    TcpFlags::ACK,
                    src_addr,
                    dst_addr,
                    src_port,
                    dst_port,
                );
                let pkt_len: usize = pkt.packet().len();

                match tx.send_to(pkt, IpAddr::V4(dst_addr)) {
                    Ok(n) if n != pkt_len => {
                        return Err(io::Error::new(
                            io::ErrorKind::Other,
                            "Insufficient bytes size",
                        ))
                    }
                    Err(e) => return Err(e),
                    _ => (),
                }
            }
            _ => (),
        }

        return Ok(msg_len);
    }
}

impl Receiver {
    /// Receives a message from the friend. Returns the number of bytes read.
    /// # Arguments
    ///
    /// * `data` - A buffer to hold the received message.
    /// * `progress` - An optional channel to report regular updates as to the progress of the received message.
    /// * `cancel` - An optional channel to cancel the read operation.
    ///
    /// # Remarks
    /// The receiver progress channel is not yet implemented, and won't be until timeout is build.
    ///
    /// If using the Delim::Protocol, additional bytes are read until the delimiter packet is received
    /// (this packet has special flags set to indicate that the message is done). If the data buffer fills
    /// before this point an error is returned.
    /// Otherwise, the method reads until the data buffer is full.
    ///
    pub fn receive(
        &mut self,
        data: &mut [u8],
        progress: Option<crossbeam_channel::Sender<usize>>,
        cancel: Option<crossbeam_channel::Receiver<()>>,
    ) -> io::Result<usize> {
        // If the buffer is empty we can return immediately
        if data.len() == 0 {
            return Ok(0);
        }

        // We want to be able to cancel the read operation
        // For this reason we can't use pnet, but instead must use Socket2
        // We don't use Socket2 otherwise for two reasons:
        //		First, the interface is somewhat cumbersome (though more similar to c socket functions)
        //		Second, Socket2 does not include IPPROTO_RAW as a Protocol, which is necessary for sending packets
        // 		where we have spoofed source IP addresses. We would have to include the number (255)
        //		directly, which I find less clean
        let socket = match Socket::new(Domain::ipv4(), Type::raw(), Some(Protocol::tcp())) {
            Ok(socket) => socket,
            Err(e) => return Err(e),
        };

        return match cancel {
            Some(rx_cancel) => rayon::scope(|scope| -> io::Result<usize> {
                let (tx_local, rx_local) = crossbeam_channel::unbounded();

				// We copy the socket2 so that we can asynchronously cancel the read operation
				// in the thread below
				// Since the socket2 uses immutable references, it is possible
				// to just use an immutable reference to do this, however this is a slight
				// problem with the library because it means that it is possible to have
				// two immutable references that can perform simultaneous reads or writes
				// Based on my research, concurrent reads (or writes) cannot be performed safely,
				// and you may get illegitimate data. The idea behind rust is that such an operation 
				// is prohibited without mutexes, however in this case using a mutex would be impossible
				// since we want to cancel while a read is occuring. Therefore, the libraries choice
				// to have all operations (even the technically mutable ones) is the simplist and right
				// one. However, since I am using this project to learn more about Rust, I thought it 
				// worthwhile to find out how to do the operation if the library had been implemented 
				// with mutable references to prevent simultaneous reads (or writes). 
				// To do this, we must retrieve the file descriptor for the socket, and use it to create
				// a new socket.
				let sock_copy = unsafe { get_socket_copy(&socket) };

				// We must implement separate functions for Windows and Unix, as Windows has raw sockets
				// but Unix has raw file descriptors
				#[cfg(unix)]
				#[inline]
				unsafe fn get_socket_copy(sock_ref : &Socket) -> Socket {
					use std::os::unix::io::AsRawFd;
					use std::os::unix::io::FromRawFd;
					Socket::from_raw_fd(sock_ref.as_raw_fd())
				}

				#[cfg(windows)]
				#[inline]
				unsafe fn get_socket_copy(sock_ref : &Socket) -> Socket {
					use std::os::windows::io::AsRawSocket;
					use std::os::windows::io::FromRawSocket;
					Socket::from_raw_socket(sock_ref.as_raw_socket())
				}

                scope.spawn(move |_| {
                    crossbeam::select! {
                        recv(rx_cancel) -> _ => (),
                        recv(rx_local) -> _ => (),
                    }
                    match sock_copy.shutdown(std::net::Shutdown::Read) {
                        // For some reason, a legitimate, functional shutdown
                        // will return an error (at least on Linux)
                        // Therefore we can just ignore the error
                        _ => (),
                    };
                });
                let out = self.read(&socket, data, progress);
                match tx_local.send(()) {
                    // If the above thread has returned this send will
                    // have a legitimate error, so we can just ignore it
                    _ => (),
                };
                return out;
            }),
            None => self.read(&socket, data, progress),
        };
    }

    /// The private read operation
    /// # Arguments
    ///
    /// * `sock` - The IP socket that is being read from.
    /// * `data` - A buffer to hold the received message.
    /// * `progress` - An optional channel to report regular updates as to the progress of the received message.
    fn read(
        &self,
        sock: &Socket,
        data: &mut [u8],
        _progress: Option<crossbeam_channel::Sender<usize>>,
    ) -> io::Result<usize> {
        let (src_addr, _, src_port, dst_port) = match self.conf.bounce {
            false => (
                self.conf.friend_ip,
                self.conf.origin_ip,
                self.conf.friend_port,
                self.conf.origin_port,
            ),
            true => (
                self.conf.origin_ip,
                self.conf.friend_ip,
                self.conf.origin_port,
                self.conf.friend_port,
            ),
        };

        let mut buf = [0u8; 1024];
        let mut prev_val: Option<u32> = None;
        let mut pos: usize = 0;

        loop {
            match sock.recv(&mut buf) {
                Ok(n) => {
                    if n == 0 {
                        return Err(io::Error::new(io::ErrorKind::Other, "Read cancelled"));
                    } else if n < 8 {
                        continue;
                    // The IP header always has at least 20 bytes When finding IP addresses
                    } else {
                        // The ip header length in bytes (the length in the packet is in 32 bit words)
                        // We don't do a wrapping add because it should be impossible for this operation to overflow
                        let ip_header_length = ((buf[0] & 0x0F) * 4) as usize;
                        // Check if we actully received the correct number of bytes
                        if n < ip_header_length {
                            continue;
                        } else if let Some(ip_packet) = Ipv4Packet::new(&buf[..ip_header_length]) {
                            if let Some(tcp_packet) = TcpPacket::new(&buf[ip_header_length..]) {
                                if ip_packet.get_source() == src_addr {
                                    if tcp_packet.get_source() == src_port
                                        && tcp_packet.get_destination() == dst_port
                                    {
                                        // Check if we have hit the delimiter packet for the message if
                                        // we arre using protocol delimiting
                                        match &self.conf.delimiter {
                                            Delim::Protocol => {
                                                // If we use the protocol to delimit messages,
                                                // then when we stop depends on the bounce status.
                                                // If not in bounce mode, we wait for an ACK directly from the sender (it normally
                                                // only sends SYN packets).
                                                // If in bounce mode, we wait for a packet with the RST flag
                                                // set (which occurs when the sender sends an ACK packet to the TCP server,
                                                // which responds with a RST since the connection was not properly established.
                                                let ended = match self.conf.bounce {
                                                    true => {
                                                        tcp_packet.get_flags() & TcpFlags::RST
                                                            == TcpFlags::RST
                                                    }
                                                    false => {
                                                        tcp_packet.get_flags() & TcpFlags::ACK
                                                            == TcpFlags::ACK
                                                    }
                                                };
                                                if ended {
                                                    return Ok(pos);
                                                }
                                            }
                                            _ => (),
                                        }

                                        // If not in bounce mode, the byte is hidden in the sequence number
                                        // If in bounce mode, the bounce operation of the foreign TCP server causes the original
                                        // sequence number to be shifted into the acknowledgement number and incremented by one.
                                        // We reverse that below.
                                        // In addition, the flags depend on whether or not we are using bounce mode.
                                        let (new_val, expected_flags) = match self.conf.bounce {
                                            true => (
                                                tcp_packet.get_acknowledgement().wrapping_sub(1),
                                                TcpFlags::SYN | TcpFlags::ACK,
                                            ),
                                            false => (tcp_packet.get_sequence(), TcpFlags::SYN),
                                        };

                                        if tcp_packet.get_flags() == expected_flags {
                                            match prev_val {
                                                Some(v) if new_val == v => (),
                                                _ => {
                                                    let new_byte = (new_val & 0xFF) as u8;
                                                    if pos < data.len() {
                                                        data[pos] = new_byte;
                                                    }
                                                    pos += 1;
                                                    // We return if Delim::None and the entire buffer has been filled
                                                    // However, if Delim::Protocol we are waiting for the delim packet
                                                    // We only return an error if one more packet arrives here,
                                                    // since it is possible to fill the entire buffer and then let the delim
                                                    // packet arrive. For that reason we must check that pos < data.len() above
                                                    // each time in case the buffer has filled when Delim::Protocol and at least
                                                    // one addition non delim packet has arrived
                                                    match &self.conf.delimiter {
                                                        Delim::None if pos == data.len() => {
                                                            return Ok(pos)
                                                        }
                                                        Delim::Protocol if pos > data.len() => {
                                                            return Err(io::Error::new(
                                                                io::ErrorKind::Other,
                                                                "Insufficient buffer size",
                                                            ))
                                                        }
                                                        _ => (),
                                                    }
                                                }
                                            }
                                            prev_val = Some(new_val);
                                        }
                                    }
                                }
                            }
                        }
                    }
                }
                Err(e) => return Err(e),
            }
        }
    }
}

/// Builds the TCP packet
///
/// # Arguments
///
/// * `sequence` - The sequence number
/// * `flags`    - Any flags for the TCP header
/// * `src_addr` - The source address
/// * `dst_addr` - The destination address
/// * `src_port` - The source port
/// * `dst_port` - The destination port
fn create_packet<'a>(
    sequence: u32,
    flags: u16,
    src_addr: Ipv4Addr,
    dst_addr: Ipv4Addr,
    src_port: u16,
    dst_port: u16,
) -> Ipv4Packet<'a> {
    const TCP_HEADER_LEN: usize = 20;
    const IPV4_HEADER_LEN: usize = 20;

    let vec_ip: Vec<u8> = vec![0u8; IPV4_HEADER_LEN + TCP_HEADER_LEN];
    let mut ip_header = MutableIpv4Packet::owned(vec_ip).unwrap();

    ip_header.set_version(4);
    // IHL, in number of 32 bit words (min is 5, but this is not checked)
    ip_header.set_header_length(5);
    ip_header.set_total_length(40);
    ip_header.set_ttl(64);
    ip_header.set_next_level_protocol(IpNextHeaderProtocols::Tcp);
    ip_header.set_source(src_addr);
    ip_header.set_destination(dst_addr);

    let vec_tcp = vec![0u8; 20];
    let mut tcp_header = MutableTcpPacket::owned(vec_tcp).unwrap();

    tcp_header.set_source(src_port);
    tcp_header.set_destination(dst_port);
    tcp_header.set_sequence(sequence);

    tcp_header.set_flags(flags);
    // Minimum data offset is 5, but this is not checked
    tcp_header.set_data_offset(5);
    tcp_header.set_window(32768);

    let checksum = tcp::ipv4_checksum(
        &tcp_header.to_immutable(),
        &src_addr.clone(),
        &dst_addr.clone(),
    );
    tcp_header.set_checksum(checksum);

    ip_header.set_payload(tcp_header.packet());

    let checksum = ipv4::checksum(&ip_header.to_immutable());
    ip_header.set_checksum(checksum);

    return ip_header.consume_to_immutable();
}
