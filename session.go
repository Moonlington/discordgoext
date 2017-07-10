package discordflo

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
)

// A Context struct holds variables for Messages
type Context struct {
	Invoked string
	Argstr  string
	Args    []string
	Channel *discordgo.Channel
	Guild   *discordgo.Guild
	Mess    *discordgo.MessageCreate
	Sess    *FloFloSession
}

// The Command structy stores the command.
type Command struct {
	Name        string
	OnMessage   func(ctx *Context)
	Description string
	Usage       string
	Detailed    string
	Subcommands []*Command
	Category    string
}

// NewCommand handles the creation of Commands.
func NewCommand(name, description, usage, detaileddescription string, onmessage func(ctx *Context)) *Command {
	return &Command{
		Name:        name,
		OnMessage:   onmessage,
		Description: description,
		Subcommands: []*Command{},
	}
}

// FloFloSession handles the bot and it's commands.
type FloFloSession struct {
	*discordgo.Session
	Token                string
	Prefix               string
	Bot                  bool
	Commands             []*Command
	removeMessageHandler func()
}

// New creates a FloFloSession from a token.
func New(token, prefix string, userbot bool) (*FloFloSession, error) {
	s, err := discordgo.New(token)
	if err != nil {
		return nil, err
	}
	flo := &FloFloSession{
		Session:  s,
		Token:    token,
		Prefix:   prefix,
		Bot:      true,
		Commands: []*Command{},
	}
	flo.ChangeMessageHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		// Ignore all messages created by other users
		if flo.Bot {
			if m.Author.ID == s.State.User.ID {
				return
			}
		} else {
			if m.Author.ID != s.State.User.ID {
				return
			}
		}

		if len(m.Content) > 0 && (strings.HasPrefix(strings.ToLower(m.Content), strings.ToLower(flo.Prefix))) {
			// Setting values for the commands
			var ctx *Context
			args := strings.Fields(m.Content[len(flo.Prefix):])
			invoked := args[0]
			args = args[1:]
			argstr := m.Content[len(flo.Prefix)+len(invoked):]
			if argstr != "" {
				argstr = argstr[1:]
			}
			channel, err := s.State.Channel(m.ChannelID)
			if err != nil {
				channel, _ = s.State.PrivateChannel(m.ChannelID)
				ctx = &Context{Invoked: invoked, Argstr: argstr, Args: args, Channel: channel, Guild: nil, Mess: m, Sess: flo}
			} else {
				guild, _ := s.State.Guild(channel.GuildID)
				ctx = &Context{Invoked: invoked, Argstr: argstr, Args: args, Channel: channel, Guild: guild, Mess: m, Sess: flo}
			}

			flo.HandleCommands(ctx)
		}
	})
	return flo, err
}

// ChangeMessageHandler handles the changing of the message handler (Lots of handlers.)
func (f *FloFloSession) ChangeMessageHandler(handler func(s *discordgo.Session, m *discordgo.MessageCreate)) {
	undo := f.AddHandler(handler)
	if f.removeMessageHandler != nil {
		f.removeMessageHandler()
	}
	f.removeMessageHandler = undo
}

// AddCommand handles the adding of Commands to the handler.
func (f *FloFloSession) AddCommand(c *Command, category string) {
	c.Category = category
	f.Commands = append(f.Commands, c)
}

// HandleSubcommands returns the Context and Command that is being called
// ctx: Context used
// called: Command called
func (f *FloFloSession) HandleSubcommands(ctx *Context, called *Command) (*Context, *Command) {
	if len(ctx.Args) != 0 {
		var scalled *Command
		ok := false
		for _, c := range called.Subcommands {
			if strings.ToLower(c.Name) == strings.ToLower(ctx.Args[0]) {
				ok = true
				scalled = c
				break
			}
		}
		if ok {
			ctx.Argstr = ctx.Argstr[len(ctx.Args[0]):]
			if ctx.Argstr != "" {
				ctx.Argstr = ctx.Argstr[1:]
			}
			ctx.Invoked += " " + ctx.Args[0]
			ctx.Args = ctx.Args[1:]
			return f.HandleSubcommands(ctx, scalled)
		}
	}
	return ctx, called
}

// HandleCommands handles the Context and calls Command
// ctx: Context used
func (f *FloFloSession) HandleCommands(ctx *Context) {
	if strings.ToLower(ctx.Invoked) == "help" {
		go f.HelpFunction(ctx)
	} else {
		var called *Command
		ok := false
		for _, c := range f.Commands {
			if strings.ToLower(c.Name) == strings.ToLower(ctx.Invoked) {
				ok = true
				called = c
				break
			}
		}
		if ok {
			rctx, rcalled := f.HandleSubcommands(ctx, called)
			go rcalled.OnMessage(rctx)
		}
	}
}

// CreateEmbed handles the easy creation of Embeds.
func (f *FloFloSession) CreateEmbed(ctx *Context) *discordgo.MessageEmbed {
	color := ctx.Sess.State.UserColor(f.State.User.ID, ctx.Mess.ChannelID)
	return &discordgo.MessageEmbed{Color: color}
}

// HelpFunction handles the Help command for the CommandHandler
// ctx: Context used
func (f *FloFloSession) HelpFunction(ctx *Context) {
	embed := f.CreateEmbed(ctx)
	var desc string
	if len(ctx.Args) != 0 {
		ctx.Invoked = ""
		command := ctx.Args[0]
		var called *Command
		ok := false
		for _, c := range f.Commands {
			if strings.ToLower(c.Name) == strings.ToLower(ctx.Args[0]) {
				ok = true
				called = c
				break
			}
		}
		ctx.Args = ctx.Args[1:]
		if ok {
			sctx, scalled := f.HandleSubcommands(ctx, called)
			if scalled.Detailed == "" {
				scalled.Detailed = scalled.Description
			}
			desc = fmt.Sprintf("`%s%s %s`\n%s", f.Prefix, command+sctx.Invoked, scalled.Usage, scalled.Detailed)
			if len(scalled.Subcommands) != 0 {
				desc += "\n\nSubcommands:"
				desc += fmt.Sprintf(" `%shelp %s [subcommand]` for more info!", f.Prefix, command+sctx.Invoked)
				for _, k := range scalled.Subcommands {
					desc += fmt.Sprintf("\n`%s%s %s` - %s", f.Prefix, command, k.Name, k.Description)
				}
			}
		} else {
			desc = "No command called `" + command + "` found!"
		}
	} else {
		desc = "Commands:"
		desc += fmt.Sprintf(" `%shelp [command]` for more info!", f.Prefix)
		sorted := make(map[string][]*Command)
		for _, c := range f.Commands {
			if c.Category == "" {
				sorted["Uncategorized"] = append(sorted["Uncategorized"], c)
			} else {
				sorted[c.Category] = append(sorted[c.Category], c)
			}
		}
		for k, v := range sorted {
			var fdesc string
			field := &discordgo.MessageEmbedField{Name: k + ":"}
			for _, command := range v {
				fdesc += fmt.Sprintf("\n`%s%s` - %s", f.Prefix, command.Name, command.Description)
			}
			field.Value = fdesc[1:]
			embed.Fields = append(embed.Fields, field)
		}
	}
	embed.Author = &discordgo.MessageEmbedAuthor{Name: f.State.User.Username, IconURL: discordgo.EndpointUserAvatar(f.State.User.ID, f.State.User.Avatar)}
	embed.Description = desc
	f.ChannelMessageSendEmbed(ctx.Mess.ChannelID, embed)
}
