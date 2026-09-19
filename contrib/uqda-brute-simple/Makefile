CC = gcc
CFLAGS = -Wall -std=c89 -O3
LDLIBS = -lsodium

.PHONY: all clean
all: uqda-brute-multi-curve25519 uqda-brute-multi-ed25519

util.o: util.c uqda-brute.h
	$(CC) $(CPPFLAGS) $(CFLAGS) -c -o $@ util.c

uqda-brute-multi-ed25519: uqda-brute-multi-ed25519.c util.o uqda-brute.h
	$(CC) $(CPPFLAGS) $(CFLAGS) $(LDFLAGS) -o $@ uqda-brute-multi-ed25519.c util.o $(LDLIBS)

uqda-brute-multi-curve25519: uqda-brute-multi-curve25519.c util.o uqda-brute.h
	$(CC) $(CPPFLAGS) $(CFLAGS) $(LDFLAGS) -o $@ uqda-brute-multi-curve25519.c util.o $(LDLIBS)

clean:
	rm -f util.o uqda-brute-multi-curve25519 uqda-brute-multi-ed25519
