package main

import (
	"context"
	"log/slog"
	"strconv"
	"time"

	"google.golang.org/api/sheets/v4"
)

func getCategories(ctx context.Context, sheetId string) ([]string, error) {
	srv, err := sheets.NewService(ctx)
	if err != nil {
		return nil, err
	}

	resp, err := srv.Spreadsheets.Values.Get(sheetId, "分類!A:A").Do()
	if err != nil {
		return nil, err
	}
	var categories []string
	for _, v := range resp.Values {
		categories = append(categories, v[0].(string))
	}

	return categories, nil
}

func appendToSheets(sheetId string, params map[string]interface{}) error {
	// get time
	t, err := time.Parse(time.DateOnly, params["date"].(string))
	if err != nil {
		return err
	}

	ctx := context.Background()
	srv, err := sheets.NewService(ctx)
	if err != nil {
		return err
	}
	rows := [][]interface{}{{`=TEXT(OFFSET(INDIRECT("RC",FALSE),0,1), "yyyy/mm")`, params["date"], params["category"], params["price"], params["item"]}}
	rb := &sheets.ValueRange{Values: rows}

	// calculate financial year
	fy := t.Year()
	if int(t.Month()) < 4 {
		fy = fy - 1
	}

	resp, err := srv.Spreadsheets.Values.Append(sheetId, strconv.Itoa(fy)+"年度!A18", rb).ValueInputOption("USER_ENTERED").InsertDataOption("INSERT_ROWS").Do()
	if err != nil {
		return err
	}

	slog.Info("debug", "debug response", resp)
	slog.Info("appended to " + resp.TableRange)

	return err
}
