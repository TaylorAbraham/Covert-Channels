extern crate covert;

mod example_utils;

use example_utils::get_addresses;
use std::io;

use crossbeam_channel::unbounded;

use covert::ipv4_tcp_sequence::{channel, Config};

fn main() {
    println!("Covert Channel Sender!");

    let (friend_addr, origin_addr, friend_port, origin_port) =
        get_addresses("127.0.0.1", "127.0.0.1", "8081", "8082");
    println!("Using destination IP address: {}", friend_addr);
    println!("Using bouncer IP address: {}", origin_addr);
    println!("Using destination port: {}", friend_port);
    println!("Using bouncer port: {}", origin_port);

    let conf = Config::new(friend_addr, origin_addr, friend_port, origin_port);
    let (mut cvt, _) = channel(conf).unwrap();

    loop {
        let (t, r) = unbounded();

        println!("Write your message");
        let mut input = String::new();
        match io::stdin().read_line(&mut input) {
            Ok(mut n) => {
                println!("{} bytes read", n);
                if n > 1024 {
                    n = 1024;
                }
                match cvt.send(&input.as_bytes()[..n], None, Some(r)) {
                    Ok(_) => println!("Msg Sent:"),
                    Err(error) => println!("error: {}", error),
                }
            }
            Err(error) => println!("error: {}", error),
        }
    }
}
