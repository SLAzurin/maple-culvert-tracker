package commands

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/lib/pq"
	"github.com/slazurin/maple-culvert-tracker/.gen/mapleculverttrackerdb/public/model"
)

func TestSecondaryCharacterCommandRegistration(t *testing.T) {
	commands := make(map[string]*discordgo.ApplicationCommand)
	for _, command := range Commands {
		commands[command.Name] = command
	}
	for name, options := range map[string][]string{
		"add-secondary-character":    {"character-name", "secondary-character-name", "skip-name-check"},
		"remove-secondary-character": {"secondary-character-name"},
		"list-secondary-characters":  nil,
	} {
		command := commands[name]
		if command == nil || CommandHandlers[name] == nil {
			t.Fatalf("%s or its handler was not registered", name)
		}
		if len(command.Options) != len(options) {
			t.Fatalf("%s: got %d options, want %d", name, len(command.Options), len(options))
		}
		for idx, option := range options {
			got := command.Options[idx]
			if option == "skip-name-check" {
				if got.Name != option || got.Required || got.Type != discordgo.ApplicationCommandOptionBoolean {
					t.Errorf("%s is missing optional boolean %s", name, option)
				}
			} else if got.Name != option || !got.Required {
				t.Errorf("%s is missing required option %s", name, option)
			}
		}
		if command.DefaultMemberPermissions != nil {
			t.Errorf("%s should use standard Discord command permissions", name)
		}
	}
}

func TestSecondaryCharacterDatabase(t *testing.T) {
	dsn := os.Getenv("SECONDARY_CHARACTERS_TEST_DSN")
	if dsn == "" {
		t.Skip("set SECONDARY_CHARACTERS_TEST_DSN to run against a disposable PostgreSQL database")
	}
	database, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { database.Close() })
	database.SetMaxOpenConns(1) // keep search_path on a single connection
	exec := func(query string) {
		t.Helper()
		if _, err := database.Exec(query); err != nil {
			t.Fatal(err)
		}
	}
	schema := pq.QuoteIdentifier(fmt.Sprintf("secondary_characters_test_%d", time.Now().UnixNano()))
	exec("CREATE SCHEMA " + schema)
	t.Cleanup(func() { exec("DROP SCHEMA " + schema + " CASCADE") })
	exec("SET search_path TO " + schema)
	migrate := func(file string) {
		t.Helper()
		contents, err := os.ReadFile("../../db_migrations/" + file)
		if err != nil {
			t.Fatal(err)
		}
		exec(string(contents))
	}
	migrate("1_createdb.up.sql")
	exec(`INSERT INTO characters (id, maple_character_name, discord_user_id) VALUES
		(1, 'MainChar', '10'), (2, 'Inactive', '20'), (3, 'NoSecondaries', '30')`)
	exec(`INSERT INTO character_culvert_scores (character_id, culvert_date, score) VALUES (1, '2026-11-04', 12345)`)
	migrate("4_secondary_characters.up.sql")

	for _, tc := range []struct{ parent, secondary, want string }{
		{"mainchar", "ZetaAlt", `Linked "ZetaAlt" to "MainChar".`},
		{"MainChar", "AlphaAlt", `Linked "AlphaAlt" to "MainChar".`},
		{"MainChar", "alphaalt", "already linked to"},
		{"Inactive", "OtherAlt", `Linked "OtherAlt" to "Inactive".`},
		{"NoSecondaries", "alphaalt", "already linked to another"},
		{"NoSecondaries", "AIphaAlt", "OCR cannot distinguish"},
		{"MainChar", "MAINCHAR", "already tracked"},
		{"Missing", "SpareAlt", "was not found"},
	} {
		msg, err := addSecondaryMapping(database, tc.parent, tc.secondary)
		if err != nil || !strings.Contains(msg, tc.want) {
			t.Fatalf("adding %q to %q got (%q, %v), want %q", tc.secondary, tc.parent, msg, err, tc.want)
		}
	}
	if _, err := database.Exec(`INSERT INTO secondary_characters VALUES (2, 'ALPHAALT')`); err == nil {
		t.Fatal("database allowed a case-insensitive duplicate")
	}
	if msg, err := addRankedSecondaryMapping(database, "Inactive", "niru", "na"); err != nil || msg != `Linked "Niru" to "Inactive".` {
		t.Fatalf("ranked add returned (%q, %v)", msg, err)
	}
	if msg, err := removeSecondaryMapping(database, "niru"); err != nil || msg != `Removed "Niru" from "Inactive".` {
		t.Fatalf("removing ranked character returned (%q, %v)", msg, err)
	}
	if msg, err := addRankedSecondaryMapping(database, "Inactive", "zQ9J8kX7Y6W5", "na"); err != nil || !strings.Contains(msg, "Could not verify") {
		t.Fatalf("adding an unranked name returned (%q, %v)", msg, err)
	}
	if msg, err := removeSecondaryMapping(database, "zQ9J8kX7Y6W5"); err != nil || !strings.Contains(msg, "was not found") {
		t.Fatalf("unranked name was stored: (%q, %v)", msg, err)
	}
	exec(`UPDATE characters SET maple_character_name = 'RenamedMain' WHERE id = 1`)
	active := []model.Characters{{ID: 1, MapleCharacterName: "RenamedMain"}, {ID: 3, MapleCharacterName: "NoSecondaries"}}
	secondaries, err := loadActiveSecondaries(database, active)
	if err != nil {
		t.Fatal(err)
	}
	if got, err := json.Marshal(secondaries); err != nil || string(got) != `{"RenamedMain":["AlphaAlt","ZetaAlt"]}` {
		t.Fatalf("active export = %s, error %v", got, err)
	}
	if empty, err := loadActiveSecondaries(database, nil); err != nil || empty == nil || len(empty) != 0 {
		t.Fatalf("empty roster returned %v, error %v", empty, err)
	}
	if msg, err := removeSecondaryMapping(database, "zetaalt"); err != nil || msg != `Removed "ZetaAlt" from "RenamedMain".` {
		t.Fatalf("remove returned (%q, %v)", msg, err)
	}
	if msg, err := removeSecondaryMapping(database, "zetaalt"); err != nil || !strings.Contains(msg, "was not found") {
		t.Fatalf("second remove returned (%q, %v)", msg, err)
	}
	exec(`DELETE FROM characters WHERE id = 2`)
	var count int
	if err := database.QueryRow(`SELECT count(*) FROM secondary_characters WHERE character_id = 2`).Scan(&count); err != nil || count != 0 {
		t.Fatalf("deleted parent still has %d aliases, error %v", count, err)
	}
	migrate("4_secondary_characters.down.sql")
	var tableExists bool
	if err := database.QueryRow(`SELECT to_regclass('secondary_characters') IS NOT NULL`).Scan(&tableExists); err != nil || tableExists {
		t.Fatalf("secondary table exists after rollback: %v, error %v", tableExists, err)
	}
	var score int
	if err := database.QueryRow(`SELECT score FROM character_culvert_scores WHERE character_id = 1`).Scan(&score); err != nil || score != 12345 {
		t.Fatalf("existing score after rollback = %d, error %v", score, err)
	}
}
