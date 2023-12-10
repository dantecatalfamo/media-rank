package main

import "testing"

func TestCalculateNewEloScores(t *testing.T) {
	t.Run("verify ELO calculations", func(t *testing.T) {
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
	})
}
