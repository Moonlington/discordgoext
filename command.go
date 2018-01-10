package discordflo

// The Command structy stores the command.
type Command struct {
	Name        string
	OnMessage   func(ctx *Context)
	Description string
	Usage       string
	Detailed    string
	Subcommands []*Command
	Category    string
	Check       func(ctx *Context) bool
}

// NewCommand handles the creation of Commands.
func NewCommand(name, description, usage, detaileddescription string, onmessage func(ctx *Context)) *Command {
	return &Command{
		Name:        name,
		OnMessage:   onmessage,
		Usage:       usage,
		Detailed:    detaileddescription,
		Description: description,
		Subcommands: []*Command{},
		Check: func(ctx *Context) bool {
			return true
		},
	}
}

// AddSubCommand handles the addition of extra subcommands to an existing command
func (c *Command) AddSubCommand(sc *Command) *Command {
	c.Subcommands = append(c.Subcommands, sc)
	return c
}
