package main

import (
	"crypto/ed25519"
	"fmt"
	"log"
	"os"

	"github.com/alecthomas/kingpin/v2"

	"github.com/xconnio/wampproto-cli"
	"github.com/xconnio/wampproto-go/auth"
	"github.com/xconnio/wampproto-go/messages"
	"github.com/xconnio/wampproto-go/serializers"
)

const (
	versionString = "0.1.0"
)

type cmd struct {
	parsedCommand string

	output *string

	auth *kingpin.CmdClause
	*CryptoSign

	message    *kingpin.CmdClause
	serializer *string
	*Call
	*Result
	*Register
	*Registered
	*Invocation
	*Yield
	*UnRegister
	*UnRegistered
	*Subscribe
	*Subscribed
	*Publish
}

func parseCmd(args []string) (*cmd, error) {
	app := kingpin.New(args[0], "A tool for testing interoperability between different wampproto implementations.")
	app.Version(versionString).VersionFlag.Short('v')

	authCommand := app.Command("auth", "Authentication commands.")

	cryptoSignCommand := authCommand.Command("cryptosign", "Commands for cryptosign authentication.")
	signChallengeCommand := cryptoSignCommand.Command("sign-challenge", "Sign a cryptosign challenge.")
	verifySignatureCommand := cryptoSignCommand.Command("verify-signature", "Verify a cryptosign challenge.")
	getPubKeyCommand := cryptoSignCommand.Command("get-pubkey",
		"Retrieve the ed25519 public key associated with the provided private key.")

	messageCommand := app.Command("message", "Wampproto messages.")
	callCommand := messageCommand.Command("call", "Call message.")
	resultCommand := messageCommand.Command("result", "Result messages.")
	registerCommand := messageCommand.Command("register", "Register message.")
	registeredCommand := messageCommand.Command("registered", "Registered message.")
	invocationCommand := messageCommand.Command("invocation", "Invocation message.")
	yieldCommand := messageCommand.Command("yield", "Yield message.")
	UnRegisterCommand := messageCommand.Command("unregister", "Unregister message.")
	UnRegisteredCommand := messageCommand.Command("unregistered", "Unregistered message.")
	subscribeCommand := messageCommand.Command("subscribe", "Subscribe message.")
	subscribedCommand := messageCommand.Command("subscribed", "Subscribed message.")
	publishCommand := messageCommand.Command("publish", "Publish message.")
	c := &cmd{
		output: app.Flag("output", "Format of the output.").Default("hex").
			Enum(wampprotocli.HexFormat, wampprotocli.Base64Format),

		auth: authCommand,

		CryptoSign: &CryptoSign{
			cryptosign:        cryptoSignCommand,
			generateChallenge: cryptoSignCommand.Command("generate-challenge", "Generate a cryptosign challenge."),

			signChallenge: signChallengeCommand,
			challenge:     signChallengeCommand.Arg("challenge", "Challenge to sign.").Required().String(),
			privateKey:    signChallengeCommand.Arg("private-key", "Private key to sign challenge.").Required().String(),

			verifySignature: verifySignatureCommand,
			signature:       verifySignatureCommand.Arg("signature", "Signature to verify.").Required().String(),
			publicKey:       verifySignatureCommand.Arg("public-key", "Public key to verify signature.").Required().String(),

			generateKeyPair: cryptoSignCommand.Command("keygen", "Generate a WAMP cryptosign ed25519 keypair."),

			getPublicKey: getPubKeyCommand,
			privateKeyFlag: getPubKeyCommand.Arg("private-key",
				"The ed25519 private key to derive the corresponding public key.").Required().String(),
		},

		message: messageCommand,
		serializer: messageCommand.Flag("serializer", "Serializer to use.").Default(wampprotocli.JsonSerializer).
			Enum(wampprotocli.JsonSerializer, wampprotocli.CborSerializer, wampprotocli.MsgpackSerializer,
				wampprotocli.ProtobufSerializer),

		Call: &Call{
			call:          callCommand,
			callRequestID: callCommand.Arg("request-id", "Call request ID.").Required().Int64(),
			callURI:       callCommand.Arg("procedure", "Procedure to call.").Required().String(),
			callArgs:      callCommand.Arg("args", "Arguments for the call.").Strings(),
			callKwargs:    callCommand.Flag("kwargs", "Keyword argument for the call.").Short('k').StringMap(),
			callOption:    callCommand.Flag("option", "Call options.").Short('o').StringMap(),
		},

		Result: &Result{
			result:          resultCommand,
			resultRequestID: resultCommand.Arg("request-id", "Result request ID.").Required().Int64(),
			resultDetails:   resultCommand.Flag("details", "Result details.").Short('d').StringMap(),
			resultArgs:      resultCommand.Arg("args", "Result Arguments").Strings(),
			resultKwargs:    resultCommand.Flag("kwargs", "Result KW Arguments.").Short('k').StringMap(),
		},

		Register: &Register{
			register:     registerCommand,
			regRequestID: registerCommand.Arg("request-id", "Request request ID.").Required().Int64(),
			regProcedure: registerCommand.Arg("procedure", "Procedure to register.").Required().String(),
			regOptions:   registerCommand.Flag("option", "Register options.").Short('o').StringMap(),
		},

		Registered: &Registered{
			registered:          registeredCommand,
			registeredRequestID: registeredCommand.Arg("request-id", "Registered request ID.").Required().Int64(),
			registrationID:      registeredCommand.Arg("registration-id", "Registration ID.").Required().Int64(),
		},

		Invocation: &Invocation{
			invocation:        invocationCommand,
			invRequestID:      invocationCommand.Arg("request-id", "Invocation request ID.").Required().Int64(),
			invRegistrationID: invocationCommand.Arg("registration-id", "Invocation registration ID.").Required().Int64(),
			invDetails:        invocationCommand.Flag("details", "Invocation details.").Short('d').StringMap(),
			invArgs:           invocationCommand.Arg("args", "Invocation arguments.").Strings(),
			invKwArgs:         invocationCommand.Flag("kwargs", "Invocation KW arguments.").Short('k').StringMap(),
		},

		Yield: &Yield{
			yield:          yieldCommand,
			yieldRequestID: yieldCommand.Arg("request-id", "Yield request ID.").Required().Int64(),
			yieldOptions:   yieldCommand.Flag("option", "Yield options.").Short('o').StringMap(),
			yieldArgs:      yieldCommand.Arg("args", "Yield arguments.").Strings(),
			yieldKwArgs:    yieldCommand.Flag("kwargs", "Yield KW arguments.").Short('k').StringMap(),
		},

		UnRegister: &UnRegister{
			unRegister:          UnRegisterCommand,
			unRegRequestID:      UnRegisterCommand.Arg("request-id", "UnRegister request ID.").Required().Int64(),
			unRegRegistrationID: UnRegisterCommand.Arg("registration-id", "UnRegister registration ID.").Required().Int64(),
		},

		UnRegistered: &UnRegistered{
			unRegistered:          UnRegisteredCommand,
			UnRegisteredRequestID: UnRegisteredCommand.Arg("request-id", "UnRegistered request ID.").Required().Int64(),
		},

		Subscribe: &Subscribe{
			subscribe:          subscribeCommand,
			subscribeRequestID: subscribeCommand.Arg("request-id", "Subscribe request ID.").Required().Int64(),
			subscribeTopic:     subscribeCommand.Arg("topic", "Topic to subscribe.").Required().String(),
			subscribeOptions:   subscribeCommand.Flag("option", "Subscribe options.").Short('o').StringMap(),
		},

		Subscribed: &Subscribed{
			subscribed:          subscribedCommand,
			subscribedRequestID: subscribedCommand.Arg("request-id", "Subscribed request ID.").Required().Int64(),
			subscriptionID:      subscribedCommand.Arg("subscription-id", "Subscription ID.").Required().Int64(),
		},

		Publish: &Publish{
			publish:          publishCommand,
			publishRequestID: publishCommand.Arg("request-id", "Publish request ID.").Required().Int64(),
			publishTopic:     publishCommand.Arg("topic", "Publish topic.").Required().String(),
			publishOptions:   publishCommand.Flag("option", "Publish options.").Short('o').StringMap(),
			publishArgs:      publishCommand.Arg("args", "Publish arguments.").Strings(),
			publishKwArgs:    publishCommand.Flag("kwargs", "Publish Keyword arguments.").Short('k').StringMap(),
		},
	}

	parsedCommand, err := app.Parse(args[1:])
	if err != nil {
		return nil, err
	}
	c.parsedCommand = parsedCommand

	return c, nil
}

func Run(args []string) (string, error) {
	c, err := parseCmd(args)
	if err != nil {
		return "", err
	}

	switch c.parsedCommand {
	case c.generateChallenge.FullCommand():
		challenge, err := auth.GenerateCryptoSignChallenge()
		if err != nil {
			return "", err
		}

		return wampprotocli.FormatOutput(*c.output, challenge)

	case c.signChallenge.FullCommand():
		privateKeyBytes, err := wampprotocli.DecodeHexOrBase64(*c.privateKey)
		if err != nil {
			return "", fmt.Errorf("invalid private-key: %s", err.Error())
		}

		if len(privateKeyBytes) != 32 && len(privateKeyBytes) != 64 {
			return "", fmt.Errorf("invalid private-key: must be of length 32 or 64")
		}

		if len(privateKeyBytes) == 32 {
			privateKeyBytes = ed25519.NewKeyFromSeed(privateKeyBytes)
		}

		signedChallenge, err := auth.SignCryptoSignChallenge(*c.challenge, privateKeyBytes)
		if err != nil {
			return "", err
		}

		return wampprotocli.FormatOutput(*c.output, signedChallenge)

	case c.verifySignature.FullCommand():
		publicKeyBytes, err := wampprotocli.DecodeHexOrBase64(*c.publicKey)
		if err != nil {
			return "", fmt.Errorf("invalid public-key: %s", err.Error())
		}

		if len(publicKeyBytes) != 32 {
			return "", fmt.Errorf("invalid public-key: must be of length 32")
		}

		isVerified, err := auth.VerifyCryptoSignSignature(*c.signature, publicKeyBytes)
		if err != nil {
			return "", err
		}

		if isVerified {
			return "Signature verified successfully", nil
		}

		return "", fmt.Errorf("signature verification failed")

	case c.generateKeyPair.FullCommand():
		publicKey, privateKey, err := auth.GenerateCryptoSignKeyPair()
		if err != nil {
			return "", err
		}

		formatedPubKey, err := wampprotocli.FormatOutput(*c.output, publicKey)
		if err != nil {
			return "", err
		}

		formatedPriKey, err := wampprotocli.FormatOutput(*c.output, privateKey)
		if err != nil {
			return "", err
		}

		return fmt.Sprintf("Public Key: %s\nPrivate Key: %s", formatedPubKey, formatedPriKey), nil

	case c.getPublicKey.FullCommand():
		privateKeyBytes, err := wampprotocli.DecodeHexOrBase64(*c.privateKeyFlag)
		if err != nil {
			return "", fmt.Errorf("invalid private-key: %s", err.Error())
		}

		publicKeyBytes := ed25519.NewKeyFromSeed(privateKeyBytes).Public().(ed25519.PublicKey)

		return wampprotocli.FormatOutputBytes(*c.output, publicKeyBytes)

	case c.call.FullCommand():
		var (
			options   = wampprotocli.StringMapToTypedMap(*c.callOption)
			arguments = wampprotocli.StringsToTypedList(*c.callArgs)
			kwargs    = wampprotocli.StringMapToTypedMap(*c.callKwargs)

			serializer = wampprotocli.SerializerByName(*c.serializer)
		)

		arguments, kwargs = wampprotocli.UpdateArgsKwArgsIfEmpty(arguments, kwargs)

		callMessage := messages.NewCall(*c.callRequestID, options, *c.callURI, arguments, kwargs)

		return serializeMessageAndOutput(serializer, callMessage, *c.output)

	case c.result.FullCommand():
		var (
			details   = wampprotocli.StringMapToTypedMap(*c.resultDetails)
			arguments = wampprotocli.StringsToTypedList(*c.resultArgs)
			kwargs    = wampprotocli.StringMapToTypedMap(*c.resultKwargs)

			serializer = wampprotocli.SerializerByName(*c.serializer)
		)

		arguments, kwargs = wampprotocli.UpdateArgsKwArgsIfEmpty(arguments, kwargs)

		resultMessage := messages.NewResult(*c.resultRequestID, details, arguments, kwargs)

		return serializeMessageAndOutput(serializer, resultMessage, *c.output)

	case c.register.FullCommand():
		var (
			options    = wampprotocli.StringMapToTypedMap(*c.regOptions)
			serializer = wampprotocli.SerializerByName(*c.serializer)
		)

		regMessage := messages.NewRegister(*c.regRequestID, options, *c.regProcedure)

		return serializeMessageAndOutput(serializer, regMessage, *c.output)

	case c.registered.FullCommand():
		var serializer = wampprotocli.SerializerByName(*c.serializer)

		registeredCmd := messages.NewRegistered(*c.registeredRequestID, *c.registrationID)

		return serializeMessageAndOutput(serializer, registeredCmd, *c.output)

	case c.invocation.FullCommand():
		var (
			details   = wampprotocli.StringMapToTypedMap(*c.invDetails)
			arguments = wampprotocli.StringsToTypedList(*c.invArgs)
			kwargs    = wampprotocli.StringMapToTypedMap(*c.invKwArgs)

			serializer = wampprotocli.SerializerByName(*c.serializer)
		)

		arguments, kwargs = wampprotocli.UpdateArgsKwArgsIfEmpty(arguments, kwargs)

		invocationMessage := messages.NewInvocation(*c.invRequestID, *c.invRegistrationID, details, arguments, kwargs)

		return serializeMessageAndOutput(serializer, invocationMessage, *c.output)

	case c.yield.FullCommand():
		var (
			options   = wampprotocli.StringMapToTypedMap(*c.yieldOptions)
			arguments = wampprotocli.StringsToTypedList(*c.yieldArgs)
			kwargs    = wampprotocli.StringMapToTypedMap(*c.yieldKwArgs)

			serializer = wampprotocli.SerializerByName(*c.serializer)
		)

		arguments, kwargs = wampprotocli.UpdateArgsKwArgsIfEmpty(arguments, kwargs)

		yieldMessage := messages.NewYield(*c.yieldRequestID, options, arguments, kwargs)

		return serializeMessageAndOutput(serializer, yieldMessage, *c.output)

	case c.unRegister.FullCommand():
		var serializer = wampprotocli.SerializerByName(*c.serializer)

		unRegisterMessage := messages.NewUnRegister(*c.registeredRequestID, *c.unRegRegistrationID)

		return serializeMessageAndOutput(serializer, unRegisterMessage, *c.output)

	case c.unRegistered.FullCommand():
		var serializer = wampprotocli.SerializerByName(*c.serializer)

		unRegisteredMessage := messages.NewUnRegistered(*c.UnRegisteredRequestID)

		return serializeMessageAndOutput(serializer, unRegisteredMessage, *c.output)

	case c.subscribe.FullCommand():
		var (
			subscribeOptions = wampprotocli.StringMapToTypedMap(*c.subscribeOptions)

			serializer = wampprotocli.SerializerByName(*c.serializer)
		)

		subscribeMessage := messages.NewSubscribe(*c.subscribeRequestID, subscribeOptions, *c.subscribeTopic)

		return serializeMessageAndOutput(serializer, subscribeMessage, *c.output)

	case c.subscribed.FullCommand():
		var serializer = wampprotocli.SerializerByName(*c.serializer)

		subscribedMessage := messages.NewSubscribed(*c.subscribedRequestID, *c.subscriptionID)

		return serializeMessageAndOutput(serializer, subscribedMessage, *c.output)

	case c.publish.FullCommand():
		var (
			publishOptions = wampprotocli.StringMapToTypedMap(*c.publishOptions)
			publishArgs    = wampprotocli.StringsToTypedList(*c.publishArgs)
			publishKwargs  = wampprotocli.StringMapToTypedMap(*c.publishKwArgs)

			serializer = wampprotocli.SerializerByName(*c.serializer)
		)

		publishMessage := messages.NewPublish(*c.publishRequestID, publishOptions, *c.publishTopic, publishArgs,
			publishKwargs)

		return serializeMessageAndOutput(serializer, publishMessage, *c.output)

	}

	return "", nil
}

func serializeMessageAndOutput(serializer serializers.Serializer, message messages.Message,
	outputFormat string) (string, error) {
	serializedMessage, err := serializer.Serialize(message)
	if err != nil {
		return "", err
	}

	return wampprotocli.FormatOutputBytes(outputFormat, serializedMessage)
}

func main() {
	output, err := Run(os.Args)
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println(output)
}
