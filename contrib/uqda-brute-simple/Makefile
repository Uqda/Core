.PHONY: all

all: util uqda-brute-multi-curve25519 uqda-brute-multi-ed25519

util: util.c
	gcc -Wall -std=c89 -O3 -c -o util.o util.c

uqda-brute-multi-ed25519: uqda-brute-multi-ed25519.c util.o
	gcc -Wall -std=c89 -O3 -o uqda-brute-multi-ed25519 -lsodium uqda-brute-multi-ed25519.c util.o

uqda-brute-multi-curve25519: uqda-brute-multi-curve25519.c util.o
	gcc -Wall -std=c89 -O3 -o uqda-brute-multi-curve25519 -lsodium uqda-brute-multi-curve25519.c util.o
