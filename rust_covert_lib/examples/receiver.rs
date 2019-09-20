extern crate covert;

mod example_utils;

use example_utils::get_addresses;

use covert::ipv4_tcp_sequence::{channel, Config};

use crossbeam_channel::unbounded;

use std::{thread, time};

fn main() {
    println!("Covert Channel Receiver!");

    let (friend_addr, origin_addr, friend_port, origin_port) =
        get_addresses("127.0.0.1", "127.0.0.1", "8082", "8081");
    println!("Using destination IP address: {}", friend_addr);
    println!("Using bouncer IP address: {}", origin_addr);
    println!("Using destination port: {}", friend_port);
    println!("Using bouncer port: {}", origin_port);

    let conf = Config::new(friend_addr, origin_addr, friend_port, origin_port);
    let (_, mut cvr) = channel(conf).unwrap();

    loop {
        let (t, r) = unbounded();

        thread::spawn(move || {
            thread::sleep(time::Duration::new(10, 0));
            t.send(());
        });

        println!("Waiting for message");
        let mut buf = [0u8; 1024];
        match cvr.receive(&mut buf[..], None, Some(r)) {
            Ok(n) => {
                let msg_cow = String::from_utf8_lossy(&mut buf[..n]);
                println!("Msg Received: {}", msg_cow.into_owned());
            }
            Err(error) => println!("error: {}", error),
        }
    }
}
