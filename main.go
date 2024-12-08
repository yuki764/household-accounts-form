package main

import (
	"context"
	"fmt"
	"html"
	htpl "html/template"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"
	ttpl "text/template"
	"time"
)

func init() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		// https://cloud.google.com/logging/docs/structured-logging
		ReplaceAttr: func(groups []string, attr slog.Attr) slog.Attr {
			if attr.Key == slog.MessageKey {
				attr.Key = "message"
			}
			if attr.Key == slog.LevelKey {
				attr.Key = "severity"
				level := attr.Value.Any().(slog.Level)
				if level == slog.LevelWarn {
					attr.Value = slog.StringValue("WARNING")
				}
			}
			return attr
		},
	}))
	slog.SetDefault(logger)
}

func main() {
	sheetId := os.Getenv("SPREADSHEET_ID")
	slog.Info("Google Spreadsheet ID: " + sheetId)

	prefix := strings.Replace("/"+os.Getenv("HTTP_PATH_PREFIX")+"/", "//", "/", -1)
	slog.Info("Path Prefix: " + prefix)

	webAppIconSvgUrl := os.Getenv("WEB_APP_ICON_URL_SVG")
	webAppIconPng128Url := os.Getenv("WEB_APP_ICON_URL_PNG_128")
	slog.Info("Web App Icon (SVG): " + webAppIconSvgUrl)
	slog.Info("Web App Icon (PNG, 128px): " + webAppIconPng128Url)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	slog.Info("Port: " + port)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) { http.NotFound(w, r) })

	http.HandleFunc(prefix+"form", renderInputForm(sheetId))
	http.HandleFunc(prefix+"account", sendAccount(sheetId))

	http.HandleFunc(prefix+"manifest.webmanifest", renderWebAppManifest(webAppIconSvgUrl, webAppIconPng128Url))

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		slog.With("error", err).Error("failed to listen port 8080")
		panic(err)
	}
}

func renderInputForm(sheetId string) func(http.ResponseWriter, *http.Request) {
	// get categories from spreadsheet
	ctx := context.Background()
	categoryList, err := getCategories(ctx, sheetId)
	if err != nil {
		slog.With("error", err).Error("failed to retrive category list from spreadsheet")
		panic(err)
	}

	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			fmt.Fprintln(w, "Method Not Allowed")
			return
		}

		// get current date
		date := time.Now().Format(time.DateOnly)

		// get last submit values from GET params if exists
		submit := map[string]interface{}{
			"date":     r.FormValue("submit_date"),
			"category": r.FormValue("submit_category"),
			"price":    r.FormValue("submit_price"),
			"item":     r.FormValue("submit_item"),
		}
		if submit["date"] == "" || submit["category"] == "" || submit["price"] == "" || submit["item"] == "" {
			submit = nil
		} else {
			// overwrite date from last submitted
			date = submit["date"].(string)
		}

		// print input form
		tpl, err := htpl.ParseFiles("templates/input-form.html.tpl")
		if err != nil {
			slog.With("error", err).Error("failed to parse html template")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if err := tpl.Execute(w, map[string]interface{}{
			"date":         date,
			"categoryList": categoryList,
			"submit":       submit,
		}); err != nil {
			slog.With("error", err).Error("failed to render form html from template")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
}

func sendAccount(sheetId string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			fmt.Fprintln(w, "Method Not Allowed")
			return
		}

		if err := r.ParseForm(); err != nil {
			slog.Info("failed to parse form parameters", "error", err)
			w.WriteHeader(http.StatusBadRequest)
		} else {
			params := map[string]interface{}{}
			for k, v := range r.PostForm {
				if k == "price" {
					params[k], _ = strconv.Atoi(v[0])
				} else {
					params[k] = v[0]
				}
			}

			// append account entry to Google Sheet
			if err := appendToSheets(sheetId, params); err != nil {
				slog.Error("failed to append account to spreadsheet", "error", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			// post account entry to Discord Channel
			if err := postToDiscordChannel(params); err != nil {
				slog.Error("failed to post message to discord channel", "error", err)
			}

			get_params := strings.Join([]string{
				"submit_date=" + html.EscapeString(r.PostForm["date"][0]),
				"submit_category=" + html.EscapeString(r.PostForm["category"][0]),
				"submit_price=" + html.EscapeString(r.PostForm["price"][0]),
				"submit_item=" + html.EscapeString(r.PostForm["item"][0]),
			}, "&")

			fmt.Fprintf(w, `<!DOCTYPE html>
<head>
<meta charset="utf-8">
<meta http-equiv="refresh" content="1;URL=form?`+get_params+`">
</head>
<body>
<p>The account has been submitted. Please wait...</p>
</body>
</html>`)
		}
	}
}

func renderWebAppManifest(iconUrlSvg string, iconUrlPng128 string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		tpl, err := ttpl.ParseFiles("templates/manifest.webmanifest.tpl")
		if err != nil {
			slog.With("error", err).Error("failed to parse web app manifest template")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Add("Content-Type", "application/manifest+json")

		if err := tpl.Execute(w, map[string]interface{}{
			"iconUrlSvg":    iconUrlSvg,
			"iconUrlPng128": iconUrlPng128,
		}); err != nil {
			slog.With("error", err).Error("failed to render web app manifest from template")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
}
