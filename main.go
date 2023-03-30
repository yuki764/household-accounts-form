package main

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"google.golang.org/api/sheets/v4"
)

func inputForm(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	// get time
	t := time.Now()
	// print input form
	tpl, err := template.ParseFiles("input-form.html.tpl")
	if err != nil {
		log.Fatalln(err)
	}
	if err := tpl.Execute(w, map[string]interface{}{
		"date":         t.Format(time.DateOnly),
		"categoryList": []string{"生活費", "娯楽", "嗜好品", "交際費", "その他"},
	}); err != nil {
		log.Fatalln(err)
	}
}

func sendAccount(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseForm(); err != nil {
		fmt.Fprintf(w, "error: while parsing params.")
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
		// print input form
		tpl, err := template.ParseFiles("input-form.html.tpl")
		if err != nil {
			log.Fatalln(err)
		}
		if err := tpl.Execute(w, map[string]interface{}{
			"date":         params["date"],
			"categoryList": []string{"生活費", "娯楽", "嗜好品", "交際費", "その他"},
			"submit": map[string]interface{}{
				"date":     params["date"],
				"category": params["category"],
				"price":    params["price"],
				"item":     params["item"],
			},
		}); err != nil {
			log.Fatalln(err)
		}

		// append account entry to Google Sheet
		if err := appendToSheets(params); err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// post account entry to Discord Channel
		if err := postToDiscordChannel(params); err != nil {
			log.Println(err)
		}
	}
}

func appendToSheets(params map[string]interface{}) error {
	// get time
	t, err := time.Parse("2006-01-02", params["date"].(string))
	if err != nil {
		return err
	}

	ctx := context.Background()
	sheetsService, err := sheets.NewService(ctx)
	if err != nil {
		return err
	}
	sheetId := os.Getenv("SHEET_ID")
	rows := [][]interface{}{{`=TEXT(OFFSET(INDIRECT("RC",FALSE),0,1), "yyyy/mm")`, params["date"], params["category"], params["price"], params["item"]}}
	rb := &sheets.ValueRange{Values: rows}

	// calculate financial year
	fy := t.Year()
	if int(t.Month()) < 4 {
		fy = fy - 1
	}

	resp, err := sheetsService.Spreadsheets.Values.Append(sheetId, strconv.Itoa(fy)+"年度!A18", rb).ValueInputOption("USER_ENTERED").InsertDataOption("INSERT_ROWS").Do()
	if err != nil {
		return err
	}
	log.Printf("%#v\n", resp)

	return err
}

func postToDiscordChannel(account map[string]interface{}) error {
	webhookUrl := os.Getenv("DISCORD_WEBHOOK_URL")
	rb := `{"embeds": [{ "color" : 10475956, "fields": [`
	rb += fmt.Sprintf(`{"name": "date", "value": "%v"}`, account["date"]) + `,`
	rb += fmt.Sprintf(`{"name": "price", "value": "%v"}`, account["price"]) + `,`
	rb += fmt.Sprintf(`{"name": "category", "value": "%v", "inline": true}`, account["category"]) + `,`
	rb += fmt.Sprintf(`{"name": "item", "value": "%v", "inline": true}`, account["item"])
	rb += `]}]}`

	req, err := http.NewRequest(
		"POST",
		webhookUrl,
		bytes.NewBuffer([]byte(rb)),
	)
	if err != nil {
		log.Println(err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
	}
	log.Printf("%#v\n", resp)

	defer resp.Body.Close()

	return err
}

func main() {
	log.Printf("Google Sheet ID: %v", os.Getenv("SHEET_ID"))

	http.HandleFunc("/", inputForm)
	http.HandleFunc("/account", sendAccount)
	http.ListenAndServe(":8080", nil)
}
