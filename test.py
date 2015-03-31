import struct
import os
import socket
import requests
import time

def test():
    address = ('127.0.0.1', 24000)
    #address = ('172.25.1.111', 24001)
    s = socket.socket(socket.AF_INET, socket.SOCK_STREAM)  
    s.connect(address)
    count = 1024*1024/2
    while count:
        s.send("a:test%d\n"%count)
        s.send("b:test%d\n"%count)
        count -= 1



if __name__ == "__main__":
    test()
