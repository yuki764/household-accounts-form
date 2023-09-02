package main

import (
	"bytes"
	"fmt"
	"log/slog"
	"net/http"
	"os"
)

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
		slog.Default().With("error", err).Error("failed to construct post request")
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		slog.Default().With("error", err).Error("failed to request post webhook")
		return err
	}

	slog.Default().Debug("debug response", "response", resp)
	slog.Default().Info("reported to discord: " + resp.Status)

	return nil
}
