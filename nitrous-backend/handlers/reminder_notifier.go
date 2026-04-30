package handlers

import (
	"database/sql"
	"log"
	"nitrous-backend/config"
	"nitrous-backend/database"
	"strings"
	"time"
)

type dueReminder struct {
	ID       string
	UserID   string
	Title    string
	Location string
}

// StartReminderNotifier starts a background worker that turns due reminders into in-app notifications.
func StartReminderNotifier() {
	if database.DB == nil {
		log.Println("Reminder notifier disabled: database is not configured")
		return
	}

	interval := time.Minute
	if parsed, err := time.ParseDuration(strings.TrimSpace(config.AppConfig.ReminderPollInterval)); err == nil && parsed > 0 {
		interval = parsed
	}

	go func() {
		log.Printf("Reminder notifier enabled: polling every %s", interval)
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			generateDueReminderNotifications(50)
			<-ticker.C
		}
	}()
}

func generateDueReminderNotifications(limit int) {
	rows, err := database.DB.Query(`
		SELECT r.id::text, r.user_id::text, e.title, COALESCE(e.location, '')
		FROM reminders r
		JOIN events e ON e.id = r.event_id
		WHERE r.notified_at IS NULL AND r.remind_at <= NOW()
		ORDER BY r.remind_at ASC
		LIMIT $1
	`, limit)
	if err != nil {
		log.Printf("Reminder notifier query failed: %v", err)
		return
	}
	defer rows.Close()

	items := make([]dueReminder, 0)
	for rows.Next() {
		var item dueReminder
		if err := rows.Scan(&item.ID, &item.UserID, &item.Title, &item.Location); err != nil {
			log.Printf("Reminder notifier scan failed: %v", err)
			return
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		log.Printf("Reminder notifier rows failed: %v", err)
		return
	}

	for _, item := range items {
		tx, err := database.DB.Begin()
		if err != nil {
			log.Printf("Reminder notifier tx begin failed: %v", err)
			continue
		}

		body := "Reminder: " + item.Title
		if strings.TrimSpace(item.Location) != "" {
			body = body + " at " + item.Location
		}

		if _, err := tx.Exec(`
			INSERT INTO notifications (id, user_id, title, body, type, created_at)
			VALUES (gen_random_uuid(), $1::uuid, 'Event Reminder', $2, 'reminder', NOW())
		`, item.UserID, body); err != nil {
			_ = tx.Rollback()
			log.Printf("Reminder notifier insert failed for %s: %v", item.ID, err)
			continue
		}

		res, err := tx.Exec(`DELETE FROM reminders WHERE id = $1::uuid AND notified_at IS NULL`, item.ID)
		if err != nil {
			_ = tx.Rollback()
			log.Printf("Reminder notifier delete failed for %s: %v", item.ID, err)
			continue
		}

		affected, _ := res.RowsAffected()
		if affected == 0 {
			_ = tx.Rollback()
			continue
		}

		if err := tx.Commit(); err != nil && err != sql.ErrTxDone {
			log.Printf("Reminder notifier commit failed for %s: %v", item.ID, err)
		}
	}
}
