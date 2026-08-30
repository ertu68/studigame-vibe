-- BAföG: passive Studierende-Förderung, pays out every 6 h while enabled.
-- bafog_next_at is the timestamp of the next payout; NULL means the student
-- has not (re)applied and receives nothing. The payout itself lands in the
-- container and overflows to charity like any other income.

ALTER TABLE characters ADD COLUMN bafog_next_at TEXT NULL;
