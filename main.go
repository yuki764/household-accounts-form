package main

import (
	"bytes"
	"context"
	"fmt"
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
	fmt.Fprintf(w, `<html>
<head>
	<style>
		input, select {font-size: 2em; width: 80vw;}
	</style>
</head>
<body>
	<form method="post" action="account">
		<input required id="date" name="date" type="date" value="%d-%02d-%02d">
		<select required id="category" name="category">
			<option value="生活費">生活費</option>
			<option value="娯楽">娯楽</option>
			<option value="嗜好品">嗜好品</option>
			<option value="交際費">交際費</option>
			<option value="その他">その他</option>
		</select>
		<input required id="price" name="price" type="number">
		<input required id="item" name="item">
		<input type="submit">
	</form>
</body>
</html>`, t.Year(), int(t.Month()), t.Day())
}

func sendAccount(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
	} else {
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
			fmt.Fprintf(w, `<html>
<head>
	<style>
		input, select {font-size: 2em; width: 80vw;}
	</style>
</head>
<body>
	<form method="post" action="account">
		<input required id="date" name="date" type="date" value="%s">
		<select required id="category" name="category">`, params["date"])

			for _, c := range []string{"生活費", "娯楽", "嗜好品", "交際費", "その他"} {
				s := ""
				if c == params["category"] {
					s = "selected"
				}
				fmt.Fprintf(w, "\t\t\t"+`<option %s value="%s">%s</option>`+"\n", s, c, c)
			}

			fmt.Fprintf(w, "\t\t"+`</select>
		<input required id="price" name="price" type="number">
		<input required id="item" name="item">
		<input type="submit">
	</form>`)
			// print submitted result
			fmt.Fprintf(w, `<ul>`)
			for _, k := range []string{"date", "category", "price", "item"} {
				fmt.Fprintf(w, "<li>%v: %v</li>\n", k, params[k])
			}
			fmt.Fprintf(w, `</ul></body></html>`)

			// append account entry to Google Sheet
			appendToSheets(params)

			// post account entry to Discord Channel
			err := postToDiscordChannel(params)
			if err != nil {
				log.Println(err)
			}
		}
	}
}

func appendToSheets(params map[string]interface{}) {
	// get time
	t := time.Now()
	ctx := context.Background()
	sheetsService, err := sheets.NewService(ctx)
	if err != nil {
		log.Println(err)
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
		log.Println(err)
	}
	log.Printf("%#v\n", resp)
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
