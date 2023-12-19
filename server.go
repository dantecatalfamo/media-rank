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
	var dbSpec string
	if dbPath == ":memory:" {
		dbSpec = dbPath
	} else {
		dbSpec = fmt.Sprintf("file:%s?cache=shared&_journal_mode=WAL&_foreign_keys=true", dbPath)
	}
	db, err := sql.Open("sqlite3", dbSpec)
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

func (s *Server) Close() error {
	return s.db.Close()
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

func (s *Server) ComparisonCount() (int64, error) {
	row := s.db.QueryRow("SELECT COUNT(*) FROM comparisons")
	if row.Err() != nil {
		return 0, fmt.Errorf("ComparisonCount failed to query: %w", row.Err())
	}
	var rowCount int64
	if err := row.Scan(&rowCount); err != nil {
		return 0, fmt.Errorf("ComparisonCount failed to scan row: %w", err)
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
	defer rows.Close()

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

func (s *Server) SortedList(descending bool) ([]MediaInfo, error) {
	var order string
	if descending {
		order = "DESC"
	} else {
		order = "ASC"
	}
	query := fmt.Sprintf("SELECT id, path, sha1sum, score, matches FROM media ORDER BY score %s", order)
	count, err := s.MediaCount()
	if err != nil {
		return nil, fmt.Errorf("SortedList failed to get count: %w", err)
	}
	list := make([]MediaInfo, 0, count)

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("SortedList query failed: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var id int64
		var mediaPath string
		var sha1 string
		var score int
		var matches int

		if err := rows.Scan(&id, &mediaPath, &sha1, &score, &matches); err != nil {
			return nil, fmt.Errorf("SortedList failed to scan row: %w", err)
		}

		list = append(list, MediaInfo{Id: id, Path: mediaPath, Sha1: sha1, Score: score, Matches: matches })
	}
	if rows.Err() != nil {
		return nil, fmt.Errorf("ScanList error while iterating: %w", rows.Err())
	}

	return list, nil
}

type Comparison struct {
	Id int64
	Points int
	Winner MediaInfo
	Loser MediaInfo
}

const historyQuery = `
SELECT
  c.id id,
  c.points points,
  w.id winner_id, w.path winner_path, w.sha1sum winner_sha1sum, w.score winner_score, w.matches winner_matches,
  l.id loser_id, l.path loser_path, l.sha1sum loser_sha1sum, l.score loser_score, l.matches loser_matches
FROM comparisons c
JOIN media w ON c.winner_id = w.id
JOIN media l ON c.loser_id = l.id
ORDER BY id DESC
`
func (s *Server) Comparisons() ([]Comparison, error) {
	count, err := s.ComparisonCount()
	if err != nil {
		return nil, fmt.Errorf("Server.Comparisons failed to get count: %w", err)
	}
	list := make([]Comparison, 0, count)

	rows, err := s.db.Query(historyQuery)
	if err != nil {
		return nil, fmt.Errorf("Server.Comparisons query failed")
	}
	defer rows.Close()

	for rows.Next() {
		var id int64
		var points int
		var winnerId int64
		var winnerPath string
		var winnerSha1 string
		var winnerScore int
		var winnerMatches int
		var loserId int64
		var loserPath string
		var loserSha1 string
		var loserScore int
		var loserMatches int

		if err := rows.Scan(
			&id, &points,
			&winnerId, &winnerPath, &winnerSha1, &winnerScore, &winnerMatches,
			&loserId, &loserPath, &loserSha1, &loserScore, &loserMatches,
		); err != nil {
			return nil, fmt.Errorf("Server.Comparisons scan row: %w", err)
		}

		list = append(list, Comparison{
			Id: id,
			Points: points,
			Winner: MediaInfo{
				Id: winnerId,
				Path: winnerPath,
				Sha1: winnerSha1,
				Score: winnerScore,
				Matches: winnerMatches,
			},
			Loser: MediaInfo{
				Id: loserId,
				Path: loserPath,
				Sha1: loserSha1,
				Score: loserScore,
				Matches: loserMatches,
			},
		})
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("Server.Comparisons rows: %w", rows.Err())
	}

	return list, nil
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
