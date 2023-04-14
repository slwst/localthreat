package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/haggen/localthreat/api/web"
	gonanoid "github.com/matoous/go-nanoid"

	"github.com/bwmarrin/discordgo"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

func v1APIHandler(db *pgxpool.Pool, discord *discordgo.Session) web.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !strings.HasPrefix(r.URL.Path, "/v1/") {
				next.ServeHTTP(w, r)
				return
			}

			if err := r.ParseForm(); err != nil {
				log.Print("ParseForm():", err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			w.Header().Set("Content-Type", "application/json; charset=utf-8")

			route := &web.Route{}
			route.Parse(r)

			switch {
			case route.Match("POST", "/v1/reports"):
				src, err := ioutil.ReadAll(r.Body)
				if err != nil {
					panic(err)
				}
				id, err := gonanoid.Nanoid(10)
				if err != nil {
					panic(err)
				}
				report := &Report{
					ID: id,
				}
				report.Parse(string(src))
				data, err := json.Marshal(report)
				if err != nil {
					panic(err)
				}
				if discord != nil {
					channel := os.Getenv("DISCORD_CHANNEL")
					if channel != "" {
						err = discordRelay(data, channel, discord)
						if err != nil {
							panic(err)
						}
					}
				}
				_, err = db.Exec(context.Background(), `INSERT INTO reports VALUES ($1, $2);`, report.ID, report.Data)
				if err != nil {
					panic(err)
				}
				w.Write(data)
				w.WriteHeader(http.StatusCreated)
			case route.Match("GET", "/v1/reports/*"):
				report := &Report{}
				err := db.QueryRow(context.Background(), `SELECT id, time, data FROM reports WHERE id = $1;`, route.Target).Scan(&report.ID, &report.Time, &report.Data)
				if err == pgx.ErrNoRows {
					w.WriteHeader(http.StatusNotFound)
					return
				} else if err != nil {
					panic(err)
				}
				data, err := json.Marshal(report)
				if err != nil {
					panic(err)
				}
				w.Write(data)
				w.WriteHeader(http.StatusOK)
			case route.Match("PATCH", "/v1/reports/*"):
				report := &Report{}
				err := db.QueryRow(context.Background(), `SELECT id, time, data FROM reports WHERE id = $1;`, route.Target).Scan(&report.ID, &report.Time, &report.Data)
				if err == pgx.ErrNoRows {
					w.WriteHeader(http.StatusNotFound)
					return
				} else if err != nil {
					panic(err)
				}
				src, err := ioutil.ReadAll(r.Body)
				if err != nil {
					panic(err)
				}
				report.Parse(string(src))
				_, err = db.Exec(context.Background(), `UPDATE reports SET data = $2 WHERE id = $1;`, report.ID, report.Data)
				if err != nil {
					panic(err)
				}
				data, err := json.Marshal(report)
				if err != nil {
					panic(err)
				}
				w.Write(data)
				w.WriteHeader(http.StatusOK)
			default:
				w.WriteHeader(http.StatusNotFound)
			}
		})
	}
}

func main() {
	database, err := pgxpool.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}
	defer database.Close()

	w := web.New()

	w.Use(web.RecoverHandler())
	w.Use(web.RequestIDHandler())
	w.Use(web.LoggingHandler())
	w.Use(web.RemoteAddrHandler())
	w.Use(web.RateLimiterHandler())
	w.Use(web.CORSHandler())
	//discord init
	token := os.Getenv("DISCORD_TOKEN")
	if token != "" {
		discord, err := discordgo.New("Bot " + token)
		if err != nil {
			log.Fatal(err)
		}
		w.Use(v1APIHandler(database, discord))
	} else {
		w.Use(v1APIHandler(database, nil))
	}

	w.Listen(":" + os.Getenv("PORT"))
}

func discordRelay(report []byte, channel string, discord *discordgo.Session) error {
	fmt.Printf("report is %s", report)
	_, err := discord.ChannelMessageSend(channel, string(report))
	return err
}
