package main

import (
	"fmt"
	"crypto/rsa"
	"crypto/rand"
	"flag"
	"os"
)

func main() {
	fmt.Println("RSA key generator")
	fmt.Println("Copyright James Lawrence\n")

	keysize := flag.Int("keysize", 1024, "the bit-size used for the generated RSA key pair")

	flag.Parse()

	fmt.Printf("Generating RSA key pair with bit-size %v...\n", *keysize)

	key, err := rsa.GenerateKey(rand.Reader, *keysize)

	if err != nil {
		fmt.Printf("Error generating RSA key pair: %v\n", err)
	} else {
		fmt.Printf("Modulus: %v\n", key.N)
		fmt.Printf("Public exponent: %v\n", key.E)
		fmt.Printf("Private exponent: %v\n", key.D)
		fmt.Println("Private primes:")

		for i,p := range key.Primes {
			fmt.Printf("\t%v: %v\n", i, p)
		}
	}

	fmt.Println("\nPress enter to continue...")
	_,_ = os.Stdin.Read([]byte{0})
}