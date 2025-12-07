.PHONY: all

all: util Uqda-brute-multi-curve25519 Uqda-brute-multi-ed25519

util: util.c
	gcc -Wall -std=c89 -O3 -c -o util.o util.c

Uqda-brute-multi-ed25519: Uqda-brute-multi-ed25519.c util.o
	gcc -Wall -std=c89 -O3 -o Uqda-brute-multi-ed25519 -lsodium Uqda-brute-multi-ed25519.c util.o

Uqda-brute-multi-curve25519: Uqda-brute-multi-curve25519.c util.o
	gcc -Wall -std=c89 -O3 -o Uqda-brute-multi-curve25519 -lsodium Uqda-brute-multi-curve25519.c util.o
