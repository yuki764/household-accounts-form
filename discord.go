package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
)

func postToDiscordChannel(account map[string]interface{}) error {
	webhookUrl := os.Getenv("DISCORD_WEBHOOK_URL")
	/*
		rb := `{"embeds": [{ "color" : 10475956, "fields": [`
		rb += fmt.Sprintf(`{"name": "date", "value": "%v"}`, account["date"]) + `,`
		rb += fmt.Sprintf(`{"name": "price", "value": "%v"}`, account["price"]) + `,`
		rb += fmt.Sprintf(`{"name": "category", "value": "%v", "inline": true}`, account["category"]) + `,`
		rb += fmt.Sprintf(`{"name": "item", "value": "%v", "inline": true}`, account["item"])
		rb += `]}]}`
	*/
	b, err := json.Marshal(&struct {
		Content string `json:"content"`
	}{
		Content: fmt.Sprintf("%s\n%d, %s, %s", account["date"], account["price"], account["category"], account["item"]),
	})
	if err != nil {
		slog.With("error", err).Error("failed to encode JSON")
		return err
	}

	req, err := http.NewRequest(
		"POST",
		webhookUrl,
		bytes.NewBuffer(b),
	)
	if err != nil {
		slog.With("error", err).Error("failed to construct post request")
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		slog.With("error", err).Error("failed to request post webhook")
		return err
	}

	slog.Debug("debug response", "response", resp)
	slog.Info("reported to discord: " + resp.Status)

	return nil
}
