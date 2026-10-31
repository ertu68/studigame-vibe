package handlers

import (
	"database/sql"
	"time"

	"studigame/internal/game"
	"studigame/internal/web"
)

// rowQuerier is what *sql.DB and *sql.Tx share, so effect lookups work both
// inside and outside a transaction.
type rowQuerier interface {
	QueryRow(query string, args ...any) *sql.Row
}

// shopEffect sums one effect type over everything the character owns.
func shopEffect(q rowQuerier, characterID int64, effectType string) (int, error) {
	var sum int
	err := q.QueryRow(
		`SELECT COALESCE(SUM(s.effect_value * ci.quantity), 0)
		 FROM character_items ci JOIN shop_items s ON s.id = ci.item_id
		 WHERE ci.character_id = ? AND s.effect_type = ?`, characterID, effectType,
	).Scan(&sum)
	return sum, err
}

// capacityOf is the container capacity including housing/bike bonuses.
func capacityOf(q rowQuerier, char Character) (int, error) {
	bonus, err := shopEffect(q, char.ID, "container_bonus")
	return game.ContainerCapacity(char.ContainerLevel) + bonus, err
}

// SettleRest books what accrues while the student is not on a shift: mood
// regeneration and the hourly income of owned items. Both anchor on a stored
// timestamp that moves by whole hours only, so partial hours are not lost and
// a reload cannot pay twice. A running shift pauses regeneration. It returns
// a summary when passive income arrived.
func SettleRest(db *sql.DB, char Character, now time.Time) (*ActionSummary, error) {
	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var mood int
	var moodAt, paidAt sql.NullString
	if err := tx.QueryRow(
		`SELECT mood, mood_at, passive_paid_at FROM characters WHERE id = ?`, char.ID,
	).Scan(&mood, &moodAt, &paidAt); err != nil {
		return nil, err
	}
	var busy bool
	if err := tx.QueryRow(
		`SELECT EXISTS(SELECT 1 FROM actions WHERE character_id = ? AND resolved_at IS NULL)`, char.ID,
	).Scan(&busy); err != nil {
		return nil, err
	}

	nowText := now.UTC().Format(time.RFC3339Nano)
	if !busy {
		newAt := nowText // full or unset: no banking of rest
		if since := parseNextAt(moodAt); since != nil && mood < 100 {
			hours := game.WholeHours(*since, now)
			mood = game.RegenMood(mood, hours)
			if mood < 100 {
				newAt = since.Add(time.Duration(hours) * time.Hour).UTC().Format(time.RFC3339Nano)
			}
		}
		if _, err := tx.Exec(`UPDATE characters SET mood = ?, mood_at = ? WHERE id = ?`, mood, newAt, char.ID); err != nil {
			return nil, err
		}
	}

	var summary *ActionSummary
	since := parseNextAt(paidAt)
	if since == nil {
		if _, err := tx.Exec(`UPDATE characters SET passive_paid_at = ? WHERE id = ?`, nowText, char.ID); err != nil {
			return nil, err
		}
	} else if full := game.WholeHours(*since, now); full > 0 {
		rate, err := shopEffect(tx, char.ID, "passive_income")
		if err != nil {
			return nil, err
		}
		next := since.Add(time.Duration(full) * time.Hour).UTC().Format(time.RFC3339Nano)
		if _, err := tx.Exec(`UPDATE characters SET passive_paid_at = ? WHERE id = ?`, next, char.ID); err != nil {
			return nil, err
		}
		if payout := rate * min(full, game.MaxPassiveHours); payout > 0 {
			money, err := moneyOf(tx, char.ID)
			if err != nil {
				return nil, err
			}
			capacity, err := capacityOf(tx, char)
			if err != nil {
				return nil, err
			}
			kept, donated := game.SplitAtCapacity(money, payout, capacity)
			if kept > 0 {
				if _, err := postMoney(tx, char.ID, kept, "passive_income", nil); err != nil {
					return nil, err
				}
			}
			if donated > 0 {
				if _, err := tx.Exec(`UPDATE charity_pool SET cents = cents + ? WHERE edition_id = ?`, donated, char.EditionID); err != nil {
					return nil, err
				}
			}
			summary = &ActionSummary{
				Title:       "Nebeneinkünfte eingegangen",
				Detail:      web.FormatCents(payout) + " aus deinen Anschaffungen",
				Donated:     web.FormatCents(donated),
				HasDonation: donated > 0,
			}
		}
	}
	return summary, tx.Commit()
}

// GetShopItemEffects totals the effects of everything a character owns.
func GetShopItemEffects(db *sql.DB, characterID int64) (passiveIncome, containerBonus, moodBonus int, err error) {
	if passiveIncome, err = shopEffect(db, characterID, "passive_income"); err != nil {
		return
	}
	if containerBonus, err = shopEffect(db, characterID, "container_bonus"); err != nil {
		return
	}
	moodBonus, err = shopEffect(db, characterID, "mood_bonus")
	return
}
