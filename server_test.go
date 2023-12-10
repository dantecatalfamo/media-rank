package main

import (
	"testing"
)

func TestCalculateNewEloScores(t *testing.T) {
	tests := []struct{
		winnerBefore int
		loserBefore int
		winnerAfter int
		loserAfter int
	}{
		{ 1500, 1500, 1515, 1485 },
		{ 1300, 1500, 1322, 1478 },
		{ 1250, 1670, 1277, 1643 },
		{ 1720, 1321, 1722, 1319 },
	}
	for _, test := range(tests) {
		winnerNewScore, loserNewScore := calculateNewEloScores(test.winnerBefore, test.loserBefore)
		if winnerNewScore != test.winnerAfter {
			t.Errorf("winnerBefore: %d, loserBefore: %d, expected winnerNewScore to be %d, found %d",
				test.winnerBefore, test.loserBefore, test.winnerAfter, winnerNewScore)
		}
		if loserNewScore != test.loserAfter {
			t.Errorf("winnerBefore: %d, loserBefore: %d, expected loserNewScore to be %d, found %d",
				test.winnerBefore, test.loserBefore, test.loserAfter, loserNewScore)
		}
	}
}

func TestServer(t *testing.T) {
	t.Run("MediaCount returns number of rows in media table", func(t *testing.T) {
		s, err := NewServer(":memory:")
		if err != nil {
			t.Fatalf("failed to create new server: %s", err)
		}
		_, err = s.db.Exec("INSERT INTO media(path, sha1sum) VALUES ('a', 'a'), ('b', 'b')")
		if err != nil {
			t.Errorf("failed to insert test data: %s", err)
		}
		rowCount, err := s.MediaCount()
		if err != nil {
			t.Errorf("failed to get row count: %s", err)
		}
		if rowCount != 2 {
			t.Errorf("expected rowCount to be 2, found %d", rowCount)
		}
		_, err = s.db.Exec("INSERT INTO media(path, sha1sum) VALUES ('c', 'c'), ('d', 'd')")
		if err != nil {
			t.Errorf("failed to insert test data: %s", err)
		}
		rowCount, err = s.MediaCount()
		if err != nil {
			t.Errorf("failed to get row count: %s", err)
		}
		if rowCount != 4 {
			t.Errorf("expected rowCount to be 4, found %d", rowCount)
		}
	})

	t.Run("GetMediaInfo returns media info", func(t *testing.T) {
		s, err := NewServer(":memory:")
		if err != nil {
			t.Fatalf("failed to create new server: %s", err)
		}
		_, err = s.db.Exec("INSERT INTO media(path, sha1sum, score, matches) VALUES ('a', 'abc', 123, 0), ('b', 'bcd', 456, 12)")
		if err != nil {
			t.Errorf("failed to insert test data: %s", err)
		}
		media1, err := s.GetMediaInfo(1)
		if err != nil {
			t.Errorf("failed to get media: %s", err)
		}
		if media1.Id != 1 {
			t.Errorf("expected media1.Id to be 1, found %d", media1.Id)
		}
		if media1.Path != "a" {
			t.Errorf("expected media1.Path to be \"a\", found %s", media1.Path)
		}
		if media1.Sha1 != "abc" {
			t.Errorf("expected media1.Sha1 to be \"abc\", found %s", media1.Sha1)
		}
		if media1.Score != 123 {
			t.Errorf("expected media1.Score to be 123, found %d", media1.Score)
		}
		if media1.Matches != 0 {
			t.Errorf("expected media1.Matches to be 0, found %d", media1.Matches)
		}
		media2, err := s.GetMediaInfo(2)
		if err != nil {
			t.Errorf("failed to get media: %s", err)
		}
		if media2.Id != 2 {
			t.Errorf("expected media2.Id to be 2, found %d", media2.Id)
		}
		if media2.Path != "b" {
			t.Errorf("expected media2.Path to be \"b\", found %s", media2.Path)
		}
		if media2.Sha1 != "bcd" {
			t.Errorf("expected media2.Sha1 to be \"bcd\", found %s", media2.Sha1)
		}
		if media2.Score != 456 {
			t.Errorf("expected media2.Score to be 345, found %d", media2.Score)
		}
		if media2.Matches != 12 {
			t.Errorf("expected media2.Matches to be 12, found %d", media2.Matches)
		}
	})

	t.Run("InsertMedia creates new db entries", func(t *testing.T) {
		s, err := NewServer(":memory:")
		if err != nil {
			t.Fatalf("failed to create new server: %s", err)
		}
		row1, err := s.InsertMedia("fakepath", "aaa")
		if err != nil {
			t.Error(err)
		}
		row2, err := s.InsertMedia("anotherfakepath", "bbb")
		if err != nil {
			t.Error(err)
		}
		if row1 == row2 {
			t.Errorf("expected to separate rows, returned %d and %d", row1, row2)
		}
		if row1 != 1 {
			t.Errorf("expected row1 to be 1, found %d", row1)
		}
		if row2 != 2 {
			t.Errorf("expected row2 to be 2, found %d", row2)
		}
		rowCount, err := s.MediaCount()
		if err != nil {
			t.Errorf("failed to get media count: %s", err)
		}
		if rowCount != 2 {
			t.Errorf("expected 2 records, found %d", rowCount)
		}
		media1, err := s.GetMediaInfo(row1)
		if err != nil {
			t.Errorf("failed to get media into: %s", err)
		}
		if media1.Path != "fakepath" {
			t.Errorf("expected media1.Path to be \"fakepath\", found %s", media1.Path)
		}
		if media1.Sha1 != "aaa" {
			t.Errorf("expected media1.Sha1 to be \"aaa\", found %s", media1.Sha1)
		}
		if media1.Score != 1500 {
			t.Errorf("expected media1.Score to be 1500, found %d", media1.Score)
		}
		if media1.Matches != 0 {
			t.Errorf("expected media1.Matches to be 0, found %d", media1.Matches)
		}
		if media1.Id != row1 {
			t.Errorf("expected media1.Id to be %d, found %d", row1, media1.Id)
		}
	})

	t.Run("InsertMedia updates path if sha1 already exists", func(t *testing.T) {
		s, err := NewServer(":memory:")
		if err != nil {
			t.Fatalf("failed to create new server: %s", err)
		}
		row1, err := s.InsertMedia("fakepath", "aaa")
		if err != nil {
			t.Errorf("failed to insert media: %s", err)
		}
		row2, err := s.InsertMedia("differentpath", "aaa")
		if err != nil {
			t.Errorf("failed to update media path: %s", err)
		}
		if row1 != row2 {
			t.Errorf("expected row1 == row2, instead found %d and %d", row1, row2)
		}
		media, err := s.GetMediaInfo(row1)
		if err != nil {
			t.Errorf("failed to get media info: %s", err)
		}
		if media.Path != "differentpath" {
			t.Errorf("expectd media.Path to be \"differentpath\", found: %s", media.Path)
		}
	})

	t.Run("UpdateScores updates db entries correctly", func(t *testing.T) {
		tests := []struct{
			winnerBefore int
			loserBefore int
			winnerAfter int
			loserAfter int
		}{
			{ 1500, 1500, 1515, 1485 },
			{ 1300, 1500, 1322, 1478 },
			{ 1250, 1670, 1277, 1643 },
			{ 1720, 1321, 1722, 1319 },
		}
		for _, test := range(tests) {
			_ = test
			s, err := NewServer(":memory:")
			if err != nil {
				t.Fatalf("failed to create new server: %s", err)
			}
			winnerId, err := s.InsertMedia("fakepath", "aaa")
			if err != nil {
				t.Errorf("failed to insert media: %s", err)
			}
			_, err = s.db.Exec("UPDATE media SET score = ? WHERE id = ?", test.winnerBefore, winnerId)
			if err != nil {
				t.Errorf("failed to set winner score: %s", err)
			}
			loserId, err := s.InsertMedia("alsofakepath", "bbb")
			if err != nil {
				t.Errorf("failed to insert media: %s", err)
			}
			_, err = s.db.Exec("UPDATE media SET score = ? WHERE id = ?", test.loserBefore, loserId)
			if err != nil {
				t.Errorf("failed to set loser score: %s", err)
			}
			if err := s.UpdateScores(winnerId, loserId); err != nil {
				t.Errorf("failed to update scores: %s", err)
			}
			winnerInfo, err := s.GetMediaInfo(winnerId)
			if err != nil {
				t.Errorf("failed to get winner media info: %s", err)
			}
			loserInfo, err := s.GetMediaInfo(loserId)
			if err != nil {
				t.Errorf("failed to get loser media info: %s", err)
			}
			if winnerInfo.Score != test.winnerAfter {
				t.Errorf("expected winnerInfo.Score to be %d, found %d", test.winnerAfter, winnerInfo.Score)
			}
			if winnerInfo.Matches != 1 {
				t.Errorf("expected winnerInfo.Matches to be 1, found %d", winnerInfo.Matches)
			}
			if loserInfo.Score != test.loserAfter {
				t.Errorf("expected loserInfo.Score to be %d, found %d", test.loserAfter, loserInfo.Score)
			}
			if loserInfo.Matches != 1 {
				t.Errorf("expected loserInfo.Matches to be 1, found %d", loserInfo.Matches)
			}
		}
	})
}
