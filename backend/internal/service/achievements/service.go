package achievements

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	db "github.com/repa-app/repa/internal/db/sqlc"
	"github.com/rs/zerolog/log"
	"github.com/sqlc-dev/pqtype"
)

type Service struct {
	queries db.Querier
}

func NewService(queries db.Querier) *Service {
	return &Service{queries: queries}
}

// CalculateAchievements runs all achievement checks for a revealed season.
func (s *Service) CalculateAchievements(ctx context.Context, seasonID string) error {
	season, err := s.queries.GetSeasonByID(ctx, seasonID)
	if err != nil {
		return err
	}

	memberIDs, err := s.queries.GetGroupMemberIDs(ctx, season.GroupID)
	if err != nil {
		return err
	}

	// Get winner per question (needed for accuracy checks)
	winners, err := s.queries.GetWinnerPerQuestion(ctx, seasonID)
	if err != nil {
		return err
	}
	winnerMap := make(map[string]string, len(winners)) // questionID -> winning targetID
	for _, w := range winners {
		winnerMap[w.QuestionID] = w.TargetID
	}

	questionCount := int64(len(winnerMap))

	// Get season question count for first-voter check
	sqCount, err := s.queries.CountSeasonQuestions(ctx, seasonID)
	if err != nil {
		return err
	}

	for _, userID := range memberIDs {
		if err := s.checkUserAchievements(ctx, userID, season, winnerMap, questionCount, sqCount); err != nil {
			log.Error().Err(err).Str("user_id", userID).Str("season_id", seasonID).Msg("failed to check achievements")
		}
	}

	// Update group stats for all members
	if err := s.updateGroupStats(ctx, season, memberIDs); err != nil {
		log.Error().Err(err).Str("season_id", seasonID).Msg("failed to update group stats")
	}

	return nil
}

func (s *Service) checkUserAchievements(
	ctx context.Context,
	userID string,
	season db.Season,
	winnerMap map[string]string,
	questionCount int64,
	seasonQuestionCount int64,
) error {
	groupID := season.GroupID
	seasonID := season.ID

	// Get user's votes in this season
	votes, err := s.queries.GetVotesByVoterInSeason(ctx, db.GetVotesByVoterInSeasonParams{
		SeasonID: seasonID,
		VoterID:  userID,
	})
	if err != nil {
		return err
	}

	didVote := len(votes) > 0

	// --- Accuracy-based achievements ---
	if didVote && questionCount > 0 {
		correctCount := 0
		for _, v := range votes {
			if winner, ok := winnerMap[v.QuestionID]; ok && winner == v.TargetID {
				correctCount++
			}
		}
		accuracy := float64(correctCount) / float64(questionCount)

		// SNIPER: >= 80% accuracy
		if accuracy >= 0.8 {
			s.grantAchievement(ctx, userID, groupID, seasonID, db.AchievementTypeSNIPER, nil)
		}

		// TELEPATH: 100% accuracy
		if correctCount == int(questionCount) {
			s.grantAchievement(ctx, userID, groupID, seasonID, db.AchievementTypeTELEPATH, nil)
		}

		// BLIND: < 20% accuracy
		if accuracy < 0.2 {
			s.grantAchievement(ctx, userID, groupID, seasonID, db.AchievementTypeBLIND, nil)
		}

		// ORACLE: 3 seasons in a row with > 70% accuracy
		s.checkOracle(ctx, userID, groupID, seasonID, accuracy)
	}

	// --- Reputation-based achievements ---
	// MONOPOLIST: >= 70% of votes on a single question
	s.checkMonopolist(ctx, userID, seasonID, groupID)

	// PIONEER: first to receive a top attribute in the group's history
	s.checkPioneer(ctx, userID, seasonID, groupID)

	// LEGEND: same top attribute 5 seasons in a row
	s.checkLegend(ctx, userID, groupID)

	// TODO: EXPERT_OF — угадывал правильно за конкретного участника 5+ раз подряд (cross-season tracking)

	// --- Activity-based achievements ---
	if didVote {
		// FIRST_VOTER: first to complete voting
		s.checkFirstVoter(ctx, userID, seasonID, groupID, seasonQuestionCount)

		// NIGHT_OWL: voted between 23:00-03:00 MSK
		s.checkNightOwl(ctx, userID, seasonID, groupID)
	}

	// RECRUITER: brought 3+ members
	s.checkRecruiter(ctx, userID, groupID)

	return nil
}

func (s *Service) grantAchievement(
	ctx context.Context,
	userID, groupID, seasonID string,
	achievementType db.AchievementType,
	metadata map[string]any,
) {
	// STREAK_VOTER can be granted multiple times, others only once per group
	if achievementType != db.AchievementTypeSTREAKVOTER {
		count, err := s.queries.HasAchievementInGroup(ctx, db.HasAchievementInGroupParams{
			UserID:          userID,
			GroupID:         groupID,
			AchievementType: achievementType,
		})
		if err != nil {
			log.Error().Err(err).Msg("failed to check existing achievement")
			return
		}
		if count > 0 {
			return
		}
	}

	var meta pqtype.NullRawMessage
	if metadata != nil {
		data, err := json.Marshal(metadata)
		if err == nil {
			meta = pqtype.NullRawMessage{RawMessage: data, Valid: true}
		}
	}

	_, err := s.queries.CreateAchievement(ctx, db.CreateAchievementParams{
		ID:              uuid.New().String(),
		UserID:          userID,
		GroupID:         groupID,
		SeasonID:        sql.NullString{String: seasonID, Valid: seasonID != ""},
		AchievementType: achievementType,
		Metadata:        meta,
	})
	if err != nil {
		log.Error().Err(err).
			Str("user_id", userID).
			Str("type", string(achievementType)).
			Msg("failed to create achievement")
	} else {
		log.Info().
			Str("user_id", userID).
			Str("type", string(achievementType)).
			Msg("achievement granted")
	}
}

// checkOracle checks if user had >70% accuracy for 3 consecutive seasons.
func (s *Service) checkOracle(ctx context.Context, userID, groupID, seasonID string, currentAccuracy float64) {
	if currentAccuracy <= 0.7 {
		return
	}

	// Already has ORACLE?
	count, _ := s.queries.HasAchievementInGroup(ctx, db.HasAchievementInGroupParams{
		UserID:          userID,
		GroupID:         groupID,
		AchievementType: db.AchievementTypeORACLE,
	})
	if count > 0 {
		return
	}

	// Need 2 previous seasons also with > 70% accuracy
	prevSeasons, err := s.queries.GetLastNRevealedSeasons(ctx, db.GetLastNRevealedSeasonsParams{
		GroupID: groupID,
		Limit:   3,
	})
	if err != nil || len(prevSeasons) < 3 {
		return
	}

	// prevSeasons are ordered by number DESC; check the 2 that aren't the current one
	streak := 1 // current season counts
	for _, ps := range prevSeasons {
		if ps.ID == seasonID {
			continue
		}
		acc := s.computeAccuracy(ctx, userID, ps.ID)
		if acc > 0.7 {
			streak++
		} else {
			break
		}
	}

	if streak >= 3 {
		s.grantAchievement(ctx, userID, groupID, seasonID, db.AchievementTypeORACLE, nil)
	}
}

func (s *Service) computeAccuracy(ctx context.Context, userID, seasonID string) float64 {
	winners, err := s.queries.GetWinnerPerQuestion(ctx, seasonID)
	if err != nil || len(winners) == 0 {
		return 0
	}
	winnerMap := make(map[string]string, len(winners))
	for _, w := range winners {
		winnerMap[w.QuestionID] = w.TargetID
	}

	votes, err := s.queries.GetVotesByVoterInSeason(ctx, db.GetVotesByVoterInSeasonParams{
		SeasonID: seasonID,
		VoterID:  userID,
	})
	if err != nil || len(votes) == 0 {
		return 0
	}

	correct := 0
	for _, v := range votes {
		if winner, ok := winnerMap[v.QuestionID]; ok && winner == v.TargetID {
			correct++
		}
	}
	return float64(correct) / float64(len(winnerMap))
}

// checkMonopolist checks if user got >= 70% of votes on any question.
func (s *Service) checkMonopolist(ctx context.Context, userID, seasonID, groupID string) {
	results, err := s.queries.GetSeasonResultsByUser(ctx, db.GetSeasonResultsByUserParams{
		SeasonID: seasonID,
		TargetID: userID,
	})
	if err != nil {
		return
	}

	for _, r := range results {
		if r.Percentage >= 70 {
			s.grantAchievement(ctx, userID, groupID, seasonID, db.AchievementTypeMONOPOLIST, map[string]any{
				"question_id":   r.QuestionID,
				"question_text": r.QuestionText,
				"percentage":    r.Percentage,
			})
			return
		}
	}
}

// checkPioneer checks if user is the first in this group to get top result for a question.
func (s *Service) checkPioneer(ctx context.Context, userID, seasonID, groupID string) {
	topResults, err := s.queries.GetTopResultPerQuestion(ctx, seasonID)
	if err != nil {
		return
	}

	for _, top := range topResults {
		if top.TargetID != userID {
			continue
		}

		// Check if this question's top result was ever won by anyone in previous seasons
		prevSeasons, err := s.queries.GetRevealedSeasonsForGroup(ctx, groupID)
		if err != nil {
			continue
		}

		isFirstTime := true
		for _, ps := range prevSeasons {
			if ps.ID == seasonID {
				continue
			}
			prevTops, err := s.queries.GetTopResultPerQuestion(ctx, ps.ID)
			if err != nil {
				continue
			}
			for _, pt := range prevTops {
				if pt.QuestionID == top.QuestionID {
					isFirstTime = false
					break
				}
			}
			if !isFirstTime {
				break
			}
		}

		if isFirstTime {
			s.grantAchievement(ctx, userID, groupID, seasonID, db.AchievementTypePIONEER, map[string]any{
				"question_id":   top.QuestionID,
				"question_text": top.QuestionText,
			})
			return
		}
	}
}

// checkLegend checks if user has the same top attribute for 5 seasons in a row.
func (s *Service) checkLegend(ctx context.Context, userID, groupID string) {
	count, _ := s.queries.HasAchievementInGroup(ctx, db.HasAchievementInGroupParams{
		UserID:          userID,
		GroupID:         groupID,
		AchievementType: db.AchievementTypeLEGEND,
	})
	if count > 0 {
		return
	}

	seasons, err := s.queries.GetLastNRevealedSeasons(ctx, db.GetLastNRevealedSeasonsParams{
		GroupID: groupID,
		Limit:   5,
	})
	if err != nil || len(seasons) < 5 {
		return
	}

	// Check if top attribute (question_id) is the same across all 5 seasons
	var topQuestionID string
	for i, season := range seasons {
		top, err := s.queries.GetTopAttributeForUser(ctx, db.GetTopAttributeForUserParams{
			SeasonID: season.ID,
			TargetID: userID,
		})
		if err != nil {
			return
		}
		if i == 0 {
			topQuestionID = top.QuestionID
		} else if top.QuestionID != topQuestionID {
			return
		}
	}

	s.grantAchievement(ctx, userID, groupID, seasons[0].ID, db.AchievementTypeLEGEND, nil)
}

// checkFirstVoter checks if user was the first to complete all votes.
func (s *Service) checkFirstVoter(ctx context.Context, userID, seasonID, groupID string, questionCount int64) {
	if questionCount == 0 {
		return
	}

	firstVoter, err := s.queries.GetFirstCompletedVoter(ctx, db.GetFirstCompletedVoterParams{
		SeasonID: seasonID,
		Column2:  questionCount,
	})
	if err != nil {
		return
	}

	if firstVoter == userID {
		s.grantAchievement(ctx, userID, groupID, seasonID, db.AchievementTypeFIRSTVOTER, nil)
	}
}

// checkNightOwl checks if user's first vote was between 23:00-03:00 MSK.
func (s *Service) checkNightOwl(ctx context.Context, userID, seasonID, groupID string) {
	firstVoteTime, err := s.queries.GetFirstVoteTimeByUser(ctx, db.GetFirstVoteTimeByUserParams{
		SeasonID: seasonID,
		VoterID:  userID,
	})
	if err != nil {
		return
	}

	msk := time.FixedZone("MSK", 3*60*60)
	mskTime := firstVoteTime.In(msk)
	hour := mskTime.Hour()

	if hour >= 23 || hour < 3 {
		s.grantAchievement(ctx, userID, groupID, seasonID, db.AchievementTypeNIGHTOWL, nil)
	}
}

// checkRecruiter checks if user has 3+ members who joined the group after them.
func (s *Service) checkRecruiter(ctx context.Context, userID, groupID string) {
	count, _ := s.queries.HasAchievementInGroup(ctx, db.HasAchievementInGroupParams{
		UserID:          userID,
		GroupID:         groupID,
		AchievementType: db.AchievementTypeRECRUITER,
	})
	if count > 0 {
		return
	}

	membersAfter, err := s.queries.CountMembersJoinedAfterUser(ctx, db.CountMembersJoinedAfterUserParams{
		GroupID: groupID,
		UserID:  userID,
	})
	if err != nil {
		return
	}

	if membersAfter >= 3 {
		s.grantAchievement(ctx, userID, groupID, "", db.AchievementTypeRECRUITER, nil)
	}
}

// updateGroupStats updates UserGroupStat for all members after a season reveal.
func (s *Service) updateGroupStats(ctx context.Context, season db.Season, memberIDs []string) error {
	for _, userID := range memberIDs {
		// Get existing stats
		stats, err := s.queries.GetUserGroupStats(ctx, db.GetUserGroupStatsParams{
			UserID:  userID,
			GroupID: season.GroupID,
		})
		if err != nil && err != sql.ErrNoRows {
			return err
		}

		// Did user vote this season?
		voteCount, err := s.queries.CountVotesCastByUser(ctx, db.CountVotesCastByUserParams{
			SeasonID: season.ID,
			VoterID:  userID,
		})
		if err != nil {
			return err
		}
		didVote := voteCount > 0

		// Votes received
		votesReceived, err := s.queries.CountVotesReceivedByUser(ctx, db.CountVotesReceivedByUserParams{
			SeasonID: season.ID,
			TargetID: userID,
		})
		if err != nil {
			return err
		}

		newSeasonsPlayed := stats.SeasonsPlayed + 1

		// Voting streak
		newStreak := int32(0)
		if didVote {
			newStreak = stats.VotingStreak + 1
		}
		maxStreak := stats.MaxVotingStreak
		if newStreak > maxStreak {
			maxStreak = newStreak
		}

		// Guess accuracy (current season)
		seasonAccuracy := float64(0)
		if didVote {
			winners, _ := s.queries.GetWinnerPerQuestion(ctx, season.ID)
			if len(winners) > 0 {
				winnerMap := make(map[string]string)
				for _, w := range winners {
					winnerMap[w.QuestionID] = w.TargetID
				}
				votes, _ := s.queries.GetVotesByVoterInSeason(ctx, db.GetVotesByVoterInSeasonParams{
					SeasonID: season.ID,
					VoterID:  userID,
				})
				correct := 0
				for _, v := range votes {
					if winner, ok := winnerMap[v.QuestionID]; ok && winner == v.TargetID {
						correct++
					}
				}
				seasonAccuracy = float64(correct) / float64(len(winnerMap))
			}
		}

		// Rolling average accuracy (weighted with previous)
		newAccuracy := seasonAccuracy
		if stats.SeasonsPlayed > 0 {
			weight := float64(stats.SeasonsPlayed)
			if weight > 4 {
				weight = 4 // cap at 5-season window
			}
			newAccuracy = (stats.GuessAccuracy*weight + seasonAccuracy) / (weight + 1)
		}

		_, err = s.queries.UpsertUserGroupStats(ctx, db.UpsertUserGroupStatsParams{
			ID:                 uuid.New().String(),
			UserID:             userID,
			GroupID:            season.GroupID,
			SeasonsPlayed:      newSeasonsPlayed,
			VotingStreak:       newStreak,
			MaxVotingStreak:    maxStreak,
			GuessAccuracy:      newAccuracy,
			TotalVotesCast:     stats.TotalVotesCast + int32(voteCount),
			TotalVotesReceived: stats.TotalVotesReceived + int32(votesReceived),
		})
		if err != nil {
			return err
		}

		// Grant STREAK_VOTER at milestones
		if didVote && (newStreak == 5 || newStreak == 10 || newStreak == 20) {
			s.grantAchievement(ctx, userID, season.GroupID, season.ID, db.AchievementTypeSTREAKVOTER, map[string]any{
				"streak": newStreak,
			})
		}
	}
	return nil
}
