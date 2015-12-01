package player

import (
	"rs2d/player/login"
	"fmt"
	"errors"
	"strings"
	"crypto/md5"
	"bytes"
)

type Player struct {
	Salt string // Password salt for hashing
	PasswordHash []byte
}

var PlayerStore *PlayerStorage

func Authenticate(username string, passwordHash []byte) login.LoginResponseCode {
	handleError := func(err error) login.LoginResponseCode {
		fmt.Printf("Error loading player '%v': %v\n", username, err)
		return login.Fail
	}

	if PlayerStore == nil {
		return handleError(errors.New("player storage system is not initialised"))
	}

	p, err := (*PlayerStore).ReadPlayer(username)

	if err != nil {
		return handleError(err)
	}

	if p == nil {
		return login.InvalidUserOrPass
	}

	if !bytes.Equal(p.PasswordHash, passwordHash) {
		return login.InvalidUserOrPass
	}

	// Todo: any more login checks?

	return login.Success
}

func HashPassword(username, password string) []byte {
	handleError := func(err error) []byte {
		fmt.Printf("Error loading player '%v': %v\n", username, err)
		return nil
	}

	if PlayerStore == nil {
		return handleError(errors.New("player storage system is not initialised"))
	}

	p, err := (*PlayerStore).ReadPlayer(username)

	if err != nil {
		return handleError(err)
	}

	salt := ""

	if p != nil {
		salt = p.Salt
	}

	sum := md5.Sum([]byte(strings.Join([]string{password, salt}, "")))

	return sum[:]
}