package helpers

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/bwmarrin/discordgo"
	"github.com/slazurin/maple-culvert-tracker/internal/data"
)

type commandUpdateTransport func(*http.Request) (*http.Response, error)

func (transport commandUpdateTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return transport(req)
}

func TestUpdateCommandsClearsExistingDefaultMemberPermissions(t *testing.T) {
	t.Setenv(data.EnvVarDiscordGuildID, "20")
	session, err := discordgo.New("Bot test")
	if err != nil {
		t.Fatal(err)
	}
	session.State.User = &discordgo.User{ID: "10"}
	listURL := discordgo.EndpointApplicationGuildCommands("10", "20") + "?with_localizations=true"
	editURL := discordgo.EndpointApplicationGuildCommand("10", "20", "30")
	edits := 0
	session.Client.Transport = commandUpdateTransport(func(req *http.Request) (*http.Response, error) {
		var response string
		switch {
		case req.Method == http.MethodGet && req.URL.String() == listURL:
			response = `[{"name":"add-secondary-character","id":"30","default_member_permissions":"32"},{"name":"list-secondary-characters","id":"31","default_member_permissions":null}]`
		case req.Method == http.MethodPatch && req.URL.String() == editURL:
			edits++
			var body map[string]json.RawMessage
			if err := json.NewDecoder(req.Body).Decode(&body); err != nil || len(body) != 1 || string(body["default_member_permissions"]) != "null" {
				t.Fatalf("PATCH body = %v, error = %v; want explicit null permissions", body, err)
			}
			response = `{"name":"add-secondary-character","id":"30","default_member_permissions":null}`
		default:
			t.Fatalf("unexpected request: %s %s", req.Method, req.URL)
		}
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(response)), Header: make(http.Header)}, nil
	})
	commands := []*discordgo.ApplicationCommand{{Name: "add-secondary-character"}, {Name: "list-secondary-characters"}}
	if err := UpdateCommands(session, commands); err != nil {
		t.Fatal(err)
	}
	if edits != 1 {
		t.Fatalf("sent %d permission edits, want 1", edits)
	}
}
