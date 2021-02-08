

The following function declares a command struct with some arguments and option groups:
Please see the `examples/commands.go` file for other examples, or refer to the [go-flags Go doc](http://godoc.org/github.com/jessevdk/go-flags)

```go
// MSFInject - Inject an MSF payload into a process.
type MSFInject struct {
	Positional struct {
		PID uint32 `description:"process ID to inject into" required:"1-1"`
	} `positional-args:"yes" required:"yes"`
	MSFOptions `group:"msf options"`
}
```

These are options that we defined in a separate struct, for being reusable by other commands maybe. 
Please refer to the go-flags documentation for a list of available metadata fields.

```go
// MSFOptions - Options applying to all msf-related execution commands.
type MSFOptions struct {
	Payload    string `long:"payload" short:"P" description:"payload type (auto-completed)" default:"meterpreter_reverse_https" value-name:"compatible payloads"`
	LHost      string `long:"lhost" short:"l" description:"listen host" required:"yes"`
	LPort      int    `long:"lport" short:"p" description:"listen port" default:"4444"`
	Encoder    string `long:"encoder" short:"e" description:"MSF encoder" value-name:"msf encoders"`
	Iterations int    `long:"iterations" short:"i" description:"iterations of the encoder" default:"1"`
}
```

The body of the function is pretty self-explanating:

```go
// Execute - Inject an MSF payload into a process.
func (m *MSFInject) Execute(args []string) (err error) {
	session := cctx.Context.Sliver.Session
	if session == nil {
		return
	}

	payloadName := m.Payload
	lhost := m.LHost
	lport := m.LPort
	encoder := m.Encoder
	iterations := m.Iterations

	if lhost == "" {
		fmt.Printf(util.Error+"Invalid lhost '%s', see `help %s`\n", lhost, consts.MsfStr)
		return
	}

	ctrl := make(chan bool)
	msg := fmt.Sprintf("Injecting payload %s %s/%s -> %s:%d ...",
		payloadName, session.OS, session.Arch, lhost, lport)
	go spin.Until(msg, ctrl)
	_, err = transport.RPC.MsfRemote(context.Background(), &clientpb.MSFRemoteReq{
		Payload:    payloadName,
		LHost:      lhost,
		LPort:      uint32(lport),
		Encoder:    encoder,
		Iterations: int32(iterations),
		PID:        m.Positional.PID,
		Request:    ContextRequest(session),
	})
	ctrl <- true
	<-ctrl
	if err != nil {
		fmt.Printf(util.Error+"%s\n", err)
	} else {
		fmt.Printf(util.Info + "Executed payload on target\n")
	}
	return nil
}
```


## Binding the commands

Now we declare the function that registers the command to the console parser defined [here](https://github.com/maxlandon/readline/wiki/Interfacing-With-Go-Flags).

```go
func bindCommands(parser *flags.Parser) (err error) {

	// core console
	// ----------------------------------------------------------------------------------------
	ex, err := parser.AddCommand("exit", // Command string
		"Exit from the client/server console", // Description (completions, help usage)
		"",                                    // Long description
		&Exit{})                               // Command implementation
	ex.Aliases = []string{"quit"}                  // Command aliases

        // You can further set many things from here with ex variable, which is now a *flags.Command !

	return
}

```
