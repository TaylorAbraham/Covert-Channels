extern crate clap;

use clap::{App, Arg};
use std::net::Ipv4Addr;
use std::str::FromStr;

pub fn get_addresses(
    def_f_addr: &str,
    def_o_addr: &str,
    def_f_port: &str,
    def_o_port: &str,
) -> (Ipv4Addr, Ipv4Addr, u16, u16) {
    let matches = App::new("Sender Program")
        .author("Michael Dysart <michaelwdysart@gmail.com>")
        .about("Sends bounced packets to a destination")
        .arg(
            Arg::with_name("friend address")
                .long("friend_address")
                .takes_value(true)
                .help("The friend IP address"),
        )
        .arg(
            Arg::with_name("origin address")
                .long("origin_address")
                .takes_value(true)
                .help("The origin IP address"),
        )
        .arg(
            Arg::with_name("friend port")
                .long("friend_port")
                .takes_value(true)
                .help("The friend port"),
        )
        .arg(
            Arg::with_name("origin port")
                .long("origin_port")
                .takes_value(true)
                .help("The origin port"),
        )
        .get_matches();

    let fa_str = matches.value_of("friend address").unwrap_or(def_f_addr);
    let oa_str = matches.value_of("origin address").unwrap_or(def_o_addr);
    let fp_str = matches.value_of("friend port").unwrap_or(def_f_port);
    let op_str = matches.value_of("origin port").unwrap_or(def_o_port);

    let friend_addr = match Ipv4Addr::from_str(fa_str) {
        Ok(a) => a,
        Err(e) => panic!("Friend IP Address Invalid: {}", e),
    };
    let origin_addr = match Ipv4Addr::from_str(oa_str) {
        Ok(a) => a,
        Err(e) => panic!("Origin IP Address Invalid: {}", e),
    };
    let friend_port = match fp_str.parse::<u16>() {
        Ok(a) => a,
        Err(e) => panic!("Friend port Invalid: {}", e),
    };
    let origin_port = match op_str.parse::<u16>() {
        Ok(a) => a,
        Err(e) => panic!("Origin port Invalid: {}", e),
    };

    return (friend_addr, origin_addr, friend_port, origin_port);
}

/// Test for the command line arguments
fn main() {
    let (friend_addr, origin_addr, friend_port, origin_port) =
        get_addresses("192.168.0.112", "192.168.0.111", "8082", "8081");
    println!("Using destination IP address: {}", friend_addr);
    println!("Using bouncer IP address: {}", origin_addr);
    println!("Using destination port: {}", friend_port);
    println!("Using bouncer port: {}", origin_port);
}
