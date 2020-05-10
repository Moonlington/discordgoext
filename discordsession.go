package discordgoext

import (
	"fmt"
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"
)

// DiscordSession handles the bot and it's commands.
type DiscordSession struct {
	*discordgo.Session
	Token                string
	Prefix               string
	Bot                  bool
	Commands             []*Command
	removeMessageHandler func()
}

// New creates a DiscordSession from a token.
func New(token, prefix string, userbot bool) (*DiscordSession, error) {
	s, err := discordgo.New(token)
	if err != nil {
		return nil, err
	}
	agent := &DiscordSession{
		Session:  s,
		Token:    token,
		Prefix:   prefix,
		Bot:      true,
		Commands: []*Command{},
	}
	agent.ChangeMessageHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		// Ignore all messages created by other users
		if agent.Bot {
			if m.Author.ID == s.State.User.ID {
				return
			}
		} else {
			if m.Author.ID != s.State.User.ID {
				return
			}
		}

		if len(m.Content) > 0 && (strings.HasPrefix(strings.ToLower(m.Content), strings.ToLower(agent.Prefix))) {
			// Setting values for the commands
			var ctx *Context
			args := strings.Fields(m.Content[len(agent.Prefix):])
			invoked := args[0]
			args = args[1:]
			argstr := m.Content[len(agent.Prefix)+len(invoked):]
			if argstr != "" {
				argstr = argstr[1:]
			}
			channel, err := s.State.Channel(m.ChannelID)
			if err != nil {
				channel, _ = s.State.PrivateChannel(m.ChannelID)
				ctx = &Context{Invoked: invoked, Argstr: argstr, Args: args, Channel: channel, Guild: nil, Mess: m, Sess: agent}
			} else {
				guild, _ := s.State.Guild(channel.GuildID)
				ctx = &Context{Invoked: invoked, Argstr: argstr, Args: args, Channel: channel, Guild: guild, Mess: m, Sess: agent}
			}
			go agent.HandleCommands(ctx)
		}
	})
	return agent, err
}

// ChangeMessageHandler handles the changing of the message handler (Lots of handlers.)
func (s *DiscordSession) ChangeMessageHandler(handler func(s *discordgo.Session, m *discordgo.MessageCreate)) {
	undo := s.AddHandler(handler)
	if s.removeMessageHandler != nil {
		s.removeMessageHandler()
	}
	s.removeMessageHandler = undo
}

// AddCommand handles the adding of Commands to the handler.
func (s *DiscordSession) AddCommand(category string, c *Command) {
	c.Category = category
	s.Commands = append(s.Commands, c)
}

// AddPrivateCommand handles the adding of Private Commands to the handler.
func (s *DiscordSession) AddPrivateCommand(category string, check func(ctx *Context) bool, c *Command) {
	c.Check = check
	s.AddCommand(category, c)
}

// HandleSubcommands returns the Context and Command that is being called
// ctx: Context used
// called: Command called
func (s *DiscordSession) HandleSubcommands(ctx *Context, called *Command) (*Context, *Command) {
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
			return s.HandleSubcommands(ctx, scalled)
		}
	}
	return ctx, called
}

// HandleCommands handles the Context and calls Command
// ctx: Context used
func (s *DiscordSession) HandleCommands(ctx *Context) {
	if strings.ToLower(ctx.Invoked) == "help" {
		go s.HelpFunction(ctx)
	} else {
		var called *Command
		ok := false
		for _, c := range s.Commands {
			if strings.ToLower(c.Name) == strings.ToLower(ctx.Invoked) {
				ok = true
				called = c
				break
			}
		}
		if ok {
			rctx, rcalled := s.HandleSubcommands(ctx, called)
			if rcalled.Check(ctx) {
				defer func() {
					if x := recover(); x != nil {
						log.Printf("Panicked and recovered: %v", x)
					}
				}()
				rcalled.OnMessage(rctx)
			}
		}
	}
}

// CreateEmbed handles the easy creation of Embeds.
func (s *DiscordSession) CreateEmbed(ctx *Context) *discordgo.MessageEmbed {
	color := ctx.Sess.State.UserColor(s.State.User.ID, ctx.Mess.ChannelID)
	return &discordgo.MessageEmbed{Color: color}
}

// HelpFunction handles the Help command for the CommandHandler
// ctx: Context used
func (s *DiscordSession) HelpFunction(ctx *Context) {
	embed := s.CreateEmbed(ctx)
	var desc string
	if len(ctx.Args) != 0 {
		ctx.Invoked = ""
		command := ctx.Args[0]
		var called *Command
		ok := false
		for _, c := range s.Commands {
			if strings.ToLower(c.Name) == strings.ToLower(ctx.Args[0]) {
				ok = true
				called = c
				break
			}
		}
		ctx.Args = ctx.Args[1:]
		if ok {
			sctx, scalled := s.HandleSubcommands(ctx, called)
			if scalled.Detailed == "" {
				scalled.Detailed = scalled.Description
			}
			if scalled.Check(ctx) {
				desc = fmt.Sprintf("`%s%s %s`\n%s", s.Prefix, command+sctx.Invoked, scalled.Usage, scalled.Detailed)
			}
			if len(scalled.Subcommands) != 0 {
				desc += "\n\nSubcommands:"
				desc += fmt.Sprintf(" `%shelp %s [subcommand]` for more info!", s.Prefix, command+sctx.Invoked)
				for _, k := range scalled.Subcommands {
					if k.Check(ctx) {
						desc += fmt.Sprintf("\n`%s%s %s %s` - %s", s.Prefix, command, k.Name, k.Usage, k.Description)
					}
				}
			}
		} else {
			desc = "No command called `" + command + "` found!"
		}
	} else {
		desc = fmt.Sprintf(" `%shelp [command]` for more info!", s.Prefix)
		sorted := make(map[string][]*Command)
		for _, c := range s.Commands {
			if c.Check(ctx) {
				if c.Category == "" {
					sorted["Uncategorized"] = append(sorted["Uncategorized"], c)
				} else {
					sorted[c.Category] = append(sorted[c.Category], c)
				}
			}
		}
		for k, v := range sorted {
			var fdesc string
			field := &discordgo.MessageEmbedField{Name: k + ":"}
			for _, command := range v {
				if command.Check(ctx) {
					usageText := ""
					if len(command.Usage) > 0 {
						usageText += " " + command.Usage
					}
					fdesc += fmt.Sprintf("\n`%s%s%s` - %s", s.Prefix, command.Name, usageText, command.Description)
				}
			}
			field.Value = fdesc[1:]
			embed.Fields = append(embed.Fields, field)
		}
	}
	embed.Author = &discordgo.MessageEmbedAuthor{Name: s.State.User.Username, IconURL: discordgo.EndpointUserAvatar(s.State.User.ID, s.State.User.Avatar)}
	embed.Description = desc
	s.ChannelMessageSendEmbed(ctx.Mess.ChannelID, embed)
}
