package main

import (
	"database/sql"
	"errors"
	"fmt"
	"math"

	_ "github.com/mattn/go-sqlite3"
)

type MediaInfo struct {
	Id int64    `json:"id"`
	Path string `json:"path"`
	Sha1 string `json:"sha1"`
	Score int   `json:"score"`
	Matches int `json:"matches"`
}

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
	// Prevent memory-backed DBs from opening independant DBs on
	// concurrent requests
	db.SetMaxOpenConns(1)
	_, err = db.Exec(schema)
	if err != nil {
		return nil, fmt.Errorf("new server migration: %w", err)
	}
	return &Server{ db: db }, nil
}


type Server struct {
	db *sql.DB
}

const insertMediaQuery = `
INSERT INTO media(path, sha1sum, score, matches) VALUES (?, ?, 1500, 0)
  ON CONFLICT(sha1sum) DO UPDATE SET path = ?
`

func (s *Server) InsertMedia(path string, sha1sum string) (int64, error) {
	result, err := s.db.Exec(insertMediaQuery, path, sha1sum, path)
	if err != nil {
		return 0, fmt.Errorf("failed to insert media into db: %w", err)
	}
	rowId, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get ID of inserted media: %w", err)
	}
	return rowId, nil
}

func (s *Server) GetMediaInfo(mediaId int64) (MediaInfo, error) {
	row := s.db.QueryRow("SELECT id, path, sha1sum, score, matches FROM media WHERE id = ?", mediaId)
	if row.Err() != nil {
		return MediaInfo{}, fmt.Errorf("failed to get media info from db: %w", row.Err())
	}
	var id int64
	var mediaPath string
	var sha1 string
	var score int
	var matches int

	if err := row.Scan(&id, &mediaPath, &sha1, &score, &matches); err != nil {
		return MediaInfo{}, fmt.Errorf("get media failed to scan row: %w", err)
	}

	return MediaInfo{ Id: id, Path: mediaPath, Sha1: sha1, Score: score, Matches: matches }, nil
}

func (s *Server) MediaCount() (int64, error) {
	row := s.db.QueryRow("SELECT COUNT(*) FROM media")
	if row.Err() != nil {
		return 0, fmt.Errorf("media count failed to query: %w", row.Err())
	}
	var rowCount int64
	if err := row.Scan(&rowCount); err != nil {
		return 0, fmt.Errorf("media count failed to scan row: %w", err)
	}

	return rowCount, nil
}

func (s *Server) UpdateScores(winnerId int64, loserId int64) error {
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

var NotEnoughMediaError = errors.New("not enough media in database")

func (s *Server) SelectMediaForComparison() (MediaInfo, MediaInfo, error) {
	// FIXME This can be slow on large tables. It could be preferable
	// to get the number of records, select two random numbers within
	// that range and then query for them manually
	// http://www.titov.net/2005/09/21/do-not-use-order-by-rand-or-how-to-get-random-rows-from-table/
	rows, err := s.db.Query("SELECT id FROM media ORDER BY RANDOM() LIMIT 2")
	if err != nil {
		return MediaInfo{}, MediaInfo{}, fmt.Errorf("select media for comparison query failed: %w", err)
	}

	var id1, id2 int64
	if !rows.Next() {
		return MediaInfo{}, MediaInfo{}, NotEnoughMediaError
	}
	if err := rows.Scan(&id1); err != nil {
		return MediaInfo{}, MediaInfo{}, fmt.Errorf("select comparison failed to scan: %w", err)
	}
	if !rows.Next() {
		return MediaInfo{}, MediaInfo{}, NotEnoughMediaError
	}
	if err := rows.Scan(&id2); err != nil {
		return MediaInfo{}, MediaInfo{}, fmt.Errorf("select comparison failed to scan: %w", err)
	}

	if err := rows.Close(); err != nil {
		return MediaInfo{}, MediaInfo{}, fmt.Errorf("select comparison failed to close rows: %w", err)
	}

	media1, err := s.GetMediaInfo(id1)
	if err != nil {
		return MediaInfo{}, MediaInfo{}, fmt.Errorf("select comparisons failed to get media1 info: %w", err)
	}
	media2, err := s.GetMediaInfo(id2)
	if err != nil {
		return MediaInfo{}, MediaInfo{}, fmt.Errorf("select comparisons failed to get media2 info: %w", err)
	}

	return media1, media2, nil
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
