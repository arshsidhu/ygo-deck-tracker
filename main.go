package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
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
	ID          int    `json:"id"`
	Name        string `json:"name"`
	GamesWon    int
	GamesLost   int
	MatchesWon  int
	MatchesLost int
	MatchesTied int
	Winner      bool
}

type participant struct {
	participantField `json:"participant"`
}

type matchField struct {
	Player1 int    `json:"player1_id"`
	Player2 int    `json:"player2_id"`
	Winner  int    `json:"winner_id"`
	Scores  string `json:"scores_csv"`
}

type match struct {
	matchField `json:"match"`
}

var DB *sql.DB
var PLAYER_COUNT int
var DECK_COUNT int
var API_USER string
var API_KEY string

func main() {

	API_USER = "asidhuuu"
	API_KEY = "**"

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
		"ID", "PlayerName", "DeckName", "GamesWon", "GamesLost", "MatchesWon", "MatchesLost", "MatchesTied", "TournyWins")
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

	participants := []participant{}
	getParticipants(tid, &participants)
	getScores(tid, participants)
	getWinner(tid, participants)
	saveTournament(participants)

}

func getWinner(tid string, participants []participant) {

	url := fmt.Sprintf("https://challonge.com/%s", tid)

	resp, _ := http.Get(url)

	body, _ := ioutil.ReadAll(resp.Body)

	bodyString := string(body)
	index := strings.Index(bodyString, "<td class='text-center'>1</td>")

	winnerDiv := bodyString[index : index+150]

	for i := 0; i < len(participants); i++ {
		if strings.Contains(winnerDiv, participants[i].Name) {
			participants[i].Winner = true
		} else {
			participants[i].Winner = false
		}
	}

}

func saveTournament(participants []participant) {

	for i := 0; i < len(participants); i++ {

		var name, deckName string
		tmp := strings.Split(participants[i].Name, "-")
		name = strings.ToLower(strings.Trim(tmp[0], " "))
		deckName = strings.ToLower(strings.Trim(tmp[1], " "))
		fmt.Println(name)
		fmt.Println(deckName)

		isWinner := 0
		if participants[i].Winner {
			isWinner = 1
		}

		sqlStatement := `SELECT * FROM "Decks" WHERE "PlayerName" = $1 AND "DeckName" = $2`
		row := DB.QueryRow(sqlStatement, name, deckName)

		var data deck
		switch err := row.Scan(&data.ID, &data.PlayerName, &data.DeckName,
			&data.GamesWon, &data.GamesLost, &data.MatchesWon,
			&data.MatchesLost, &data.MatchesTied, &data.TournyWins); err {
		case sql.ErrNoRows:
			sqlStatement := `INSERT INTO "Decks"(
				"ID", "PlayerName", "DeckName", "GamesWon", "GamesLost", "MatchesWon", "MatchesLost", "MatchesTied", "TournyWins")
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9);`

			_, err := DB.Exec(sqlStatement, DECK_COUNT+1, name, deckName, participants[i].GamesWon,
				participants[i].GamesLost, participants[i].MatchesWon, participants[i].MatchesLost, participants[i].MatchesTied, isWinner)
			if err != nil {
				panic(err)
			}

			DECK_COUNT += 1
		case nil:
			data.GamesWon += participants[i].GamesWon
			data.GamesLost += participants[i].GamesLost
			data.MatchesWon += participants[i].MatchesWon
			data.MatchesLost += participants[i].MatchesLost
			data.MatchesTied += participants[i].MatchesTied
			data.TournyWins += isWinner

			sqlStatement := `UPDATE "Decks"
				SET "GamesWon" = $3, "GamesLost" = $4, "MatchesWon" = $5, "MatchesLost" = $6, "MatchesTied" = $7, "TournyWins" = $8
				WHERE "PlayerName" = $1 AND "DeckName" = $2`

			_, err := DB.Exec(sqlStatement, name, deckName, data.GamesWon,
				data.GamesLost, data.MatchesWon, data.MatchesLost, data.MatchesTied, data.TournyWins)
			if err != nil {
				panic(err)
			}
		default:
			panic(err)
		}
	}

}

func getScores(tid string, participants []participant) {

	url := fmt.Sprintf("https://%s:%s@api.challonge.com/v1/tournaments/%s/matches.json", API_USER, API_KEY, tid)

	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()

	matches := []match{}
	json.NewDecoder(resp.Body).Decode(&matches)

	for _, m := range matches {
		p1 := findParticipant(m.Player1, participants)
		p2 := findParticipant(m.Player2, participants)

		var score1, score2 int
		_, err := fmt.Sscanf(m.Scores, "%d-%d", &score1, &score2)
		if err != nil {
			panic(err)
		}

		p1.GamesWon += score1
		p2.GamesWon += score2

		p1.GamesLost += score2
		p2.GamesLost += score1

		if m.Winner == 0 {
			p1.MatchesTied += 1
			p2.MatchesTied += 1
		} else if p1.ID == m.Winner {
			p1.MatchesWon += 1
			p2.MatchesLost += 1
		} else {
			p2.MatchesWon += 1
			p1.MatchesLost += 1
		}
	}
}

func getParticipants(tid string, target interface{}) error {

	url := fmt.Sprintf("https://%s:%s@api.challonge.com/v1/tournaments/%s/participants.json", API_USER, API_KEY, tid)

	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()

	return json.NewDecoder(resp.Body).Decode(target)
}

func findParticipant(pid int, participants []participant) *participant {

	for i := 0; i < len(participants); i++ {
		if participants[i].ID == pid {
			return &participants[i]
		}
	}

	panic("pid not found")
}
