package main

import (
	"fmt"
	"github.com/alexflint/go-arg"
	"github.com/go-resty/resty/v2"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"log"
	"net/http"
	"strconv"
	"time"
)

func recordMetrics(client *resty.Client, interval time.Duration) {
	go func() {
		logrus.Info("Starting fetch loop")
		for {
			logrus.Debug("Loading users")
			var users struct {
				Users []struct {
					Login string `json:"login"`
				} `json:"users"`
			}
			if r, err := client.R().SetResult(&users).Get("/api/v2/users-management/users"); err != nil {
				logrus.Errorf("Error retrieving users: %v", err)
			} else {
				if r.IsError() {
					logrus.Errorf("Error retrieving users (%d): %s", r.StatusCode(), r.String())
				}
			}
			for _, user := range users.Users {
				logrus.Debugf("Getting tokens for user %s", user.Login)
				var tokens struct {
					UserTokens []struct {
						Name           string `json:"name"`
						CreatedAt      string `json:"createdAt"`
						ExpirationDate string `json:"expirationDate"`
						IsExpired      bool   `json:"isExpired"`
						Type           string `json:"type"`
						Project        struct {
							Key  string `json:"key"`
							Name string `json:"name"`
						} `json:"project"`
					} `json:"userTokens"`
				}
				if r, err := client.R().SetQueryParam("login", user.Login).SetResult(&tokens).Get("api/user_tokens/search"); err != nil {
					logrus.Errorf("Error fetching tokens of user %s: %v", user.Login, err)
				} else {
					if r.IsError() {
						logrus.Errorf("Error fetching tokens of user %s (%d): %s", user.Login, r.StatusCode(), r.String())
					}
				}

				for _, token := range tokens.UserTokens {
					if expirationDate, creationDate, err := getDates(token.ExpirationDate, token.CreatedAt); err == nil {
						expiration.
							WithLabelValues(
								user.Login,
								token.Name,
								token.Type,
								token.Project.Key,
								strconv.FormatBool(token.IsExpired),
							).
							Set(float64(expirationDate.Unix()))
						creation.
							WithLabelValues(
								user.Login,
								token.Name,
								token.Type,
								token.Project.Key,
								strconv.FormatBool(token.IsExpired),
							).
							Set(float64(creationDate.Unix()))
					}

				}
			}
			time.Sleep(interval)
		}
	}()
}

func getDates(expiration string, creation string) (time.Time, time.Time, error) {
	var expirationDate time.Time
	if expiration != "" {
		if d, err := time.Parse("2006-01-02T15:04:05-0700", expiration); err != nil {
			logrus.Errorf("Error converting %s to time.Time: %s", expiration, err)
			return time.Time{}, time.Time{}, err
		} else {
			expirationDate = d
		}
	}
	var creationDate time.Time
	if d, err := time.Parse("2006-01-02T15:04:05-0700", creation); err != nil {
		logrus.Errorf("Error converting %s to time.Time: %s", creation, err)
		return time.Time{}, time.Time{}, err
	} else {
		creationDate = d
	}
	return expirationDate, creationDate, nil
}

var (
	expiration = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "sonarqube_user_tokens_expiration_date_seconds",
		Help: "The expiration of a user token as a unix epoch",
	}, []string{
		"user",
		"token",
		"type",
		"project_key",
		"is_expired",
	})
	creation = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "sonarqube_user_tokens_creation_date_seconds",
		Help: "The creation date of a user token as a unix epoch",
	}, []string{
		"user",
		"token",
		"type",
		"project_key",
		"is_expired",
	})
)

func main() {
	var args struct {
		LogLevel string `arg:"env:EXPORTER_LOGLEVEL" help:"The loglevel to use" default:"INFO"`
		URL      string `arg:"required,env:EXPORTER_URL" help:"The URL to your Sonarqube instance"`
		Token    string `arg:"required,env:EXPORTER_TOKEN" help:"The access token for API requests"`
		Port     int    `arg:"env:EXPORTER_PORT" help:"Port to listen on" default:"8081"`
		Interval int64  `arg:"env:EXPORTER_INTERVAL" help:"Interval to fetch user tokens in minutes" default:"60"`
	}
	arg.MustParse(&args)
	if l, err := logrus.ParseLevel(args.LogLevel); err != nil {
		log.Fatal(err)
	} else {
		logrus.SetLevel(l)
	}

	client := resty.New().SetBaseURL(args.URL).SetAuthToken(args.Token)

	recordMetrics(client, time.Duration(args.Interval)*time.Minute)

	logrus.Infof("Starting Sonarqube user token exporter on port %d", args.Port)
	http.HandleFunc("/health", func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(200)
	})
	http.Handle("/metrics", promhttp.Handler())
	if err := http.ListenAndServe(fmt.Sprintf(":%d", args.Port), nil); err != nil {
		logrus.Errorf("Error starting server: %v", err)
	}
}
