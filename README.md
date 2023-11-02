# `grpc-ssh-socks`
```
!grpcssh contains a better version of this, with client-side DNS resolution
and proper concurrent connection support.
```
```
grpc-ssh-socks is a minimal reverse socks proxy
implementation over gRPC, made to be used as a 
reference in C2 development. A bidirectional 
stream is created over gRPC, and later a SSH 
server is spun up on the client, connected to 
by the server. The server then sets up a socks 
proxy with the forwarded Dial function that 
points to the SSH connection. The approach is 
similar to how Chisel works, with gRPC 
bidirectional streams rather than websockets.
```
```
~/grpc-ssh-socks$ make 
to build this abominable thing. 
```
