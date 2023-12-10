package main

import (
	"database/sql"
	"fmt"
	"math"

	_ "github.com/mattn/go-sqlite3"
)

const schema = `
CREATE TABLE IF NOT EXISTS media (
  id INTEGER PRIMARY KEY,
  path TEXT NOT NULL,
  sha1sum TEXT UNIQUE NOT NULL,
  score INTEGER,
  matches INTEGER
);

CREATE TABLE IF NOT EXISTS comparisons (
  id INTEGER PRIMARY KEY,
  winner_id INTEGER NOT NULL,
  loser_id INTEGER NOT NULL,
  points INTEGER,
  FOREIGN KEY(winner_id) REFERENCES media(id) ON DELETE CASCADE,
  FOREIGN KEY(loser_id) REFERENCES media(id) ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS comparisons_winner_id_idx ON comparisons(winner_id);
CREATE INDEX IF NOT EXISTS comparisons_loser_id_idx ON comparisons(loser_id);

-- Maybe a good idea, maybe not
-- CREATE TRIGGER IF NOT EXISTS update_matches AFTER INSERT ON comparisons
-- BEGIN
--   UPDATE media SET matches = (SELECT COUNT(*) FROM comparisons WHERE winner_id = new.winner_id OR loser_id = new.winner_id) WHERE id = new.winner_id;
--   UPDATE media SET matches = (SELECT COUNT(*) FROM comparisons WHERE winner_id = new.loser_id OR loser_id = new.loser_id) WHERE id = new.loser_id;
-- END;
--
`

func NewServer(dbPath string) (*Server, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("new server open db: %w", err)
	}
	_, err = db.Exec(schema)
	if err != nil {
		return nil, fmt.Errorf("new server migration: %w", err)
	}
	return &Server{ db: db }, nil
}


type Server struct {
	db *sql.DB
}

func (s *Server) updateScores(winnerId int, loserId int) error {
	winnerRow := s.db.QueryRow("SELECT score FROM media WHERE id = ?", winnerId)
	if winnerRow.Err() != nil {
		return fmt.Errorf("update scores fetch winner row: %w", winnerRow.Err())
	}

	var winnerScore int
	if err := winnerRow.Scan(&winnerScore); err != nil {
		return fmt.Errorf("update scores scan winner row: %w", err)
	}

	loserRow := s.db.QueryRow("SELECT score FROM media WHERE id = ?", loserId)
	if loserRow.Err() != nil {
		return fmt.Errorf("update scores fetch loser row: %w", loserRow.Err())
	}

	var loserScore int
	if err := loserRow.Scan(&loserScore); err != nil {
		return fmt.Errorf("update scores scan loser row: %w", err)
	}

	winnerNewScore, loserNewScore := calculateNewEloScores(winnerScore, loserScore)

	pointsDifference := winnerNewScore - winnerScore

	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("update scores create new transaction: %w", err)
	}

	_, err = tx.Exec("INSERT INTO comparisons(winner_id, loser_id, points) VALUES (?, ?, ?)", winnerId, loserId, pointsDifference)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("update scores inserting new comparison: %w ", err)
	}

	if _, err := tx.Exec("UPDATE media SET score = ?, matches = matches + 1 WHERE id = ?", winnerNewScore, winnerId); err != nil {
		tx.Rollback()
		return fmt.Errorf("update scores update winner score: %w", err)
	}

	if _, err := tx.Exec("UPDATE media SET score = ?, matches = matches + 1 WHERE id = ?", loserNewScore, loserId); err != nil {
		tx.Rollback()
		return fmt.Errorf("update scores update loser score: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("update scores commit transaction: %w", err)
	}

	return nil
}

func calculateNewEloScores(winnerScore, loserScore int) (winnerNewScore, loserNewScore int) {
	// Good reference https://www.omnicalculator.com/sports/elo
	developmentCoefficient := 30.0
	influence := 400.0
	winnerExpectation := 1/(1 + math.Pow(10, float64(loserScore - winnerScore) / influence))
	loserExpectation := 1.0 - winnerExpectation
	winnerNewScore = winnerScore + int(developmentCoefficient * (1.0 - winnerExpectation))
	loserNewScore = loserScore + int(developmentCoefficient * (0.0 - loserExpectation))
	return winnerNewScore, loserNewScore
}
