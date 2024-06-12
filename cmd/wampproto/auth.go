package main

import "github.com/alecthomas/kingpin/v2"

type CryptoSign struct {
	cryptosign *kingpin.CmdClause

	generateChallenge *kingpin.CmdClause

	signChallenge *kingpin.CmdClause
	challenge     *string
	privateKey    *string

	verifySignature *kingpin.CmdClause
	signature       *string
	publicKey       *string

	generateKeyPair *kingpin.CmdClause

	getPublicKey   *kingpin.CmdClause
	privateKeyFlag *string
}
