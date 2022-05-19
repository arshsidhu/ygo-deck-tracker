package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "**"
	dbname   = "YGO"
)

type player struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Wins int    `json:"wins"`
}

type deck struct {
	ID          string `json:"id"`
	PlayerName  string `json:"playerName"`
	DeckName    string `json:"deckName"`
	GamesWon    int    `json:"gamesWon"`
	GamesLost   int    `json:"gamesLost"`
	MatchesWon  int    `json:"matchesWon"`
	MatchesLost int    `json:"matchesLost"`
	MatchesTied int    `json:"matchesTied"`
	TournyWins  int    `json:"tournyWins"`
}

type tourny struct {
	Link string `json:"link"`
}

type participantField struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type participant struct {
	participantField `json:"participant"`
}

var DB *sql.DB
var PLAYER_COUNT int
var DECK_COUNT int

func main() {

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		panic(err)
	}

	DB = db
	PLAYER_COUNT = getCount(`SELECT COUNT(*) FROM "Players"`)
	DECK_COUNT = getCount(`SELECT COUNT(*) FROM "Decks"`)

	fmt.Println("Successfully connected!")

	router := gin.Default()
	router.GET("/players", getPlayers)
	router.GET("/decks", getDecks)
	router.GET("/decks/:name", getDecksByPlayer)

	router.POST("/player", insertPlayer)
	router.POST("/deck", insertDeck)
	router.POST("/tournament", insertTournament)

	router.Run("localhost:8080")
}

func getCount(sqlStatement string) int {
	var count int
	row := DB.QueryRow(sqlStatement)
	err := row.Scan(&count)

	switch err {
	case sql.ErrNoRows:
		fmt.Println("No rows were returned!")
		return -1
	case nil:
		return count
	default:
		panic(err)
	}
}

func getPlayers(c *gin.Context) {
	sqlStatement := `SELECT * FROM "Players"`
	rows, err := DB.Query(sqlStatement)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	var players []player

	for rows.Next() {
		var tmp player
		err = rows.Scan(&tmp.ID, &tmp.Name, &tmp.Wins)
		if err != nil {
			panic(err)
		}
		players = append(players, tmp)
	}

	c.IndentedJSON(http.StatusOK, players)
}

func getDecksByPlayer(c *gin.Context) {

	name := c.Param("name")

	sqlStatement := `SELECT * FROM "Decks" WHERE "PlayerName" = $1`
	rows, err := DB.Query(sqlStatement, name)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	var decks []deck

	for rows.Next() {
		var tmp deck
		err = rows.Scan(&tmp.ID, &tmp.PlayerName, &tmp.DeckName,
			&tmp.GamesWon, &tmp.GamesLost, &tmp.MatchesWon,
			&tmp.MatchesLost, &tmp.MatchesTied, &tmp.TournyWins)
		if err != nil {
			panic(err)
		}
		decks = append(decks, tmp)
	}

	c.IndentedJSON(http.StatusOK, decks)
}

func getDecks(c *gin.Context) {
	sqlStatement := `SELECT * FROM "Decks" ORDER BY "PlayerName"`
	rows, err := DB.Query(sqlStatement)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	var decks []deck

	for rows.Next() {
		var tmp deck
		err = rows.Scan(&tmp.ID, &tmp.PlayerName, &tmp.DeckName,
			&tmp.GamesWon, &tmp.GamesLost, &tmp.MatchesWon,
			&tmp.MatchesLost, &tmp.MatchesTied, &tmp.TournyWins)
		if err != nil {
			panic(err)
		}
		decks = append(decks, tmp)
	}

	c.IndentedJSON(http.StatusOK, decks)
}

func insertPlayer(c *gin.Context) {
	var newPlayer player

	if err := c.BindJSON(&newPlayer); err != nil {
		panic(err)
	}

	sqlStatement := `INSERT INTO 
		"Players"("ID", "Name", "Wins") 
		VALUES ($1, $2, $3)`

	_, err := DB.Exec(sqlStatement, PLAYER_COUNT+1, newPlayer.Name, newPlayer.Wins)
	if err != nil {
		panic(err)
	}

	PLAYER_COUNT += 1

	c.IndentedJSON(http.StatusCreated, newPlayer)
}

func insertDeck(c *gin.Context) {
	var newDeck deck

	if err := c.BindJSON(&newDeck); err != nil {
		panic(err)
	}

	sqlStatement := `INSERT INTO "Decks"(
		"ID", "PlayerName", "DeckName", "GamesWon", "GamesLost", "MatchesWon", "MatchesLost", "MatchesTie", "TournyWins")
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9);`

	_, err := DB.Exec(sqlStatement, DECK_COUNT+1, newDeck.PlayerName, newDeck.DeckName, newDeck.GamesWon,
		newDeck.GamesLost, newDeck.MatchesWon, newDeck.MatchesLost, newDeck.MatchesTied, newDeck.TournyWins)
	if err != nil {
		panic(err)
	}

	DECK_COUNT += 1

	c.IndentedJSON(http.StatusCreated, newDeck)
}

func insertTournament(c *gin.Context) {
	var t tourny
	var tid string

	if err := c.BindJSON(&t); err != nil {
		panic(err)
	}

	tid = strings.Split(t.Link, "/")[1]

	participants := getParticipants(tid, &participant{})

	fmt.Println("**********")
	fmt.Println(participants)

}

func getParticipants(tid string, target interface{}) error {

	apiUser := "asidhuuu"
	apiKey := "**"
	url := fmt.Sprintf("https://%s:%s@api.challonge.com/v1/tournaments/%s/participants.json", apiUser, apiKey, tid)

	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()

	return json.NewDecoder(resp.Body).Decode(target)
}
