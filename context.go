package discordflo

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
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

// SendEmbed is a helper function to easily send embeds.
// em: embed to send
func (ctx *Context) SendEmbed(em *discordgo.MessageEmbed) (*discordgo.Message, error) {
	m, err := ctx.Sess.ChannelMessageSendEmbed(ctx.Mess.ChannelID, em)
	return m, err
}

// QuickSendEmbed is a helper function to easily send strings as an embed.
// s: string to send
func (ctx *Context) QuickSendEmbed(s string) (*discordgo.Message, error) {
	em := ctx.Sess.CreateEmbed(ctx)
	em.Description = s
	return ctx.SendEmbed(em)
}

// SendMessage sends a message to the channel the command came from.
// em: embed to send
func (ctx *Context) SendMessage(s string) (*discordgo.Message, error) {
	return ctx.Sess.ChannelMessageSend(ctx.Mess.ChannelID, s)
}

// ErrUserFinding is the error it gives when more than 1 or no users are found, used to return commands.
var ErrUserFinding = errors.New("Found more than 1 or no users")

func removeDuplicateUsers(list *[]*discordgo.User) {
	found := make(map[string]bool)
	j := 0
	for i, x := range *list {
		if !found[x.ID] {
			found[x.ID] = true
			(*list)[j] = (*list)[i]
			j++
		}
	}
	*list = (*list)[:j]
}

// GetAllUsers is a helper function to return all members
func (ctx *Context) GetAllUsers() (ms []*discordgo.User, err error) {
	servers, err := ctx.Sess.UserGuilds(100, "", "")
	if err != nil {
		return
	}
	for _, server := range servers {
		var g *discordgo.Guild
		g, err = ctx.Sess.State.Guild(server.ID)
		if err != nil {
			continue
		}
		for _, m := range g.Members {
			ms = append(ms, m.User)
		}
	}
	removeDuplicateUsers(&ms)
	return
}

// GetUserByName is a helper function to find a User by string
// query: String to use when finding User
func (ctx *Context) GetUserByName(query string) (members []*discordgo.User, err error) {
	MentionRegex := regexp.MustCompile(`<@!?(\d+)>`)
	var id, discrim string
	if MentionRegex.MatchString(query) {
		id = MentionRegex.FindStringSubmatch(query)[1]
	} else if regexp.MustCompile(`^.*#\d{4}$`).MatchString(query) {
		discrim = query[len(query)-4:]
		query = strings.TrimSpace(query[:len(query)-5])
	}
	var exact, wrongcase, startswith, contains, all []*discordgo.User
	lowerQuery := strings.ToLower(query)
	all, err = ctx.GetAllUsers()
	if err != nil {
		return
	}
	for _, u := range all {
		if id != "" && u.ID == id {
			exact = append(exact, u)
			break
		}
		if discrim != "" && u.Discriminator != discrim {
			continue
		}
		if u.Username == query {
			exact = append(exact, u)
		} else if len(exact) == 0 && strings.ToLower(u.Username) == lowerQuery {
			wrongcase = append(wrongcase, u)
		} else if len(wrongcase) == 0 && strings.HasPrefix(strings.ToLower(u.Username), lowerQuery) {
			startswith = append(startswith, u)
		} else if len(startswith) == 0 && strings.Contains(strings.ToLower(u.Username), lowerQuery) {
			contains = append(contains, u)
		}
	}
	if len(exact) != 0 {
		members = exact
	} else if len(wrongcase) != 0 {
		members = wrongcase
	} else if len(startswith) != 0 {
		members = startswith
	} else {
		members = contains
	}
	return
}

// GuildGetUserByName is a helper function to find a User by string in a guild
// query: String to use when finding User
// GuildID: ID for guild
func (ctx *Context) GuildGetUserByName(query, GuildID string) (members []*discordgo.User, err error) {
	MentionRegex := regexp.MustCompile(`<@!?(\d+)>`)
	var id, discrim string
	if MentionRegex.MatchString(query) {
		id = MentionRegex.FindStringSubmatch(query)[1]
	} else if regexp.MustCompile(`^.*#\d{4}$`).MatchString(query) {
		discrim = query[len(query)-4:]
		query = strings.TrimSpace(query[:len(query)-5])
	}
	var exact, wrongcase, startswith, contains []*discordgo.User
	var all []*discordgo.Member
	lowerQuery := strings.ToLower(query)
	g, err := ctx.Sess.State.Guild(GuildID)
	if err != nil {
		return
	}
	all = g.Members
	if err != nil {
		return
	}
	for _, m := range all {
		nick := m.Nick
		u := m.User
		if id != "" && u.ID == id {
			exact = append(exact, u)
			break
		}
		if discrim != "" && u.Discriminator != discrim {
			continue
		}
		if u.Username == query || (nick != "" && nick == query) {
			exact = append(exact, u)
		} else if len(exact) == 0 && strings.ToLower(u.Username) == lowerQuery || (nick != "" && strings.ToLower(nick) == lowerQuery) {
			wrongcase = append(wrongcase, u)
		} else if len(wrongcase) == 0 && strings.HasPrefix(strings.ToLower(u.Username), lowerQuery) || (nick != "" && strings.HasPrefix(strings.ToLower(nick), lowerQuery)) {
			startswith = append(startswith, u)
		} else if len(startswith) == 0 && strings.Contains(strings.ToLower(u.Username), lowerQuery) || (nick != "" && strings.Contains(strings.ToLower(nick), lowerQuery)) {
			contains = append(contains, u)
		}
	}
	if len(exact) != 0 {
		members = exact
	} else if len(wrongcase) != 0 {
		members = wrongcase
	} else if len(startswith) != 0 {
		members = startswith
	} else {
		members = contains
	}
	return
}

// ParseTooManyUsers is a helper function to create a message for finding too many users
// query: String used when finding User
// users: List of users found
func (ctx *Context) ParseTooManyUsers(query string, users []*discordgo.User) (*discordgo.Message, error) {
	out := fmt.Sprintf("Multiple users found for query **%s**:", query)
	for i := 0; i < 6; i++ {
		if i < len(users) {
			out += "\n - " + users[i].Username + " #" + users[i].Discriminator
		}
	}
	if len(users) > 6 {
		out += "\n**And " + strconv.Itoa(len(users)-6) + " more...**"
	}
	return ctx.SendMessage(out)
}

// GetUser is a helper function to get a user by name
func (ctx *Context) GetUser(args ...string) (user *discordgo.User, err error) {
	var query, gid string
	var users []*discordgo.User

	query = args[0]

	if len(args) > 1 {
		gid = args[1]
	}

	if gid != "" {
		users, err = ctx.GuildGetUserByName(query, gid)
		if err != nil {
			ctx.SendMessage("Error collecting users: " + err.Error())
			return
		}
		if len(users) < 1 {
			users, err = ctx.GetUserByName(query)
			if len(users) < 1 {
				ctx.SendMessage("No user found with name **" + query + "**")
				err = ErrUserFinding
				return
			}
			if err != nil {
				ctx.SendMessage("Error collecting users: " + err.Error())
				return
			}
		}
		if len(users) > 1 {
			ctx.ParseTooManyUsers(query, users)
			err = ErrUserFinding
			return
		}
		user = users[0]
	} else {
		users, err = ctx.GetUserByName(query)
		if err != nil {
			ctx.SendMessage("Error collecting users: " + err.Error())
			return
		}
		if len(users) < 1 {
			ctx.SendMessage("No user found with name **" + query + "**")
			err = ErrUserFinding
			return
		}
		if len(users) > 1 {
			ctx.ParseTooManyUsers(query, users)
			err = ErrUserFinding
			return
		}
		user = users[0]
	}
	if len(users) > 1 {
		ctx.ParseTooManyUsers(query, users)
		err = ErrUserFinding
		return
	}
	return
}
