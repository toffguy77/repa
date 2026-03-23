import 'package:freezed_annotation/freezed_annotation.dart';

part 'profile.freezed.dart';
part 'profile.g.dart';

@freezed
class MemberProfile with _$MemberProfile {
  const factory MemberProfile({
    required ProfileUser user,
    required UserStats stats,
    required List<ProfileAchievement> achievements,
    required String legend,
    @JsonKey(name: 'season_history') required List<SeasonCardDto> seasonHistory,
  }) = _MemberProfile;

  factory MemberProfile.fromJson(Map<String, dynamic> json) =>
      _$MemberProfileFromJson(json);
}

@freezed
class ProfileUser with _$ProfileUser {
  const factory ProfileUser({
    required String username,
    @JsonKey(name: 'avatar_emoji') String? avatarEmoji,
    @JsonKey(name: 'avatar_url') String? avatarUrl,
  }) = _ProfileUser;

  factory ProfileUser.fromJson(Map<String, dynamic> json) =>
      _$ProfileUserFromJson(json);
}

@freezed
class UserStats with _$UserStats {
  const factory UserStats({
    @JsonKey(name: 'seasons_played') required int seasonsPlayed,
    @JsonKey(name: 'voting_streak') required int votingStreak,
    @JsonKey(name: 'max_voting_streak') required int maxVotingStreak,
    @JsonKey(name: 'guess_accuracy') required double guessAccuracy,
    @JsonKey(name: 'total_votes_cast') required int totalVotesCast,
    @JsonKey(name: 'total_votes_received') required int totalVotesReceived,
    @JsonKey(name: 'top_attribute_all_time') TopAttributeDto? topAttributeAllTime,
  }) = _UserStats;

  factory UserStats.fromJson(Map<String, dynamic> json) =>
      _$UserStatsFromJson(json);
}

@freezed
class TopAttributeDto with _$TopAttributeDto {
  const factory TopAttributeDto({
    @JsonKey(name: 'question_text') required String questionText,
    required double percentage,
  }) = _TopAttributeDto;

  factory TopAttributeDto.fromJson(Map<String, dynamic> json) =>
      _$TopAttributeDtoFromJson(json);
}

@freezed
class ProfileAchievement with _$ProfileAchievement {
  const factory ProfileAchievement({
    required String type,
    Map<String, dynamic>? metadata,
    @JsonKey(name: 'earned_at') required String earnedAt,
  }) = _ProfileAchievement;

  factory ProfileAchievement.fromJson(Map<String, dynamic> json) =>
      _$ProfileAchievementFromJson(json);
}

@freezed
class SeasonCardDto with _$SeasonCardDto {
  const factory SeasonCardDto({
    @JsonKey(name: 'season_id') required String seasonId,
    @JsonKey(name: 'season_number') required int seasonNumber,
    @JsonKey(name: 'top_attribute') required String topAttribute,
    required String category,
    required double percentage,
  }) = _SeasonCardDto;

  factory SeasonCardDto.fromJson(Map<String, dynamic> json) =>
      _$SeasonCardDtoFromJson(json);
}
