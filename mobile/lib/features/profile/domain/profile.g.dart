// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'profile.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

_$MemberProfileImpl _$$MemberProfileImplFromJson(Map<String, dynamic> json) =>
    _$MemberProfileImpl(
      user: ProfileUser.fromJson(json['user'] as Map<String, dynamic>),
      stats: UserStats.fromJson(json['stats'] as Map<String, dynamic>),
      achievements: (json['achievements'] as List<dynamic>)
          .map((e) => ProfileAchievement.fromJson(e as Map<String, dynamic>))
          .toList(),
      legend: json['legend'] as String,
      seasonHistory: (json['season_history'] as List<dynamic>)
          .map((e) => SeasonCardDto.fromJson(e as Map<String, dynamic>))
          .toList(),
    );

Map<String, dynamic> _$$MemberProfileImplToJson(_$MemberProfileImpl instance) =>
    <String, dynamic>{
      'user': instance.user,
      'stats': instance.stats,
      'achievements': instance.achievements,
      'legend': instance.legend,
      'season_history': instance.seasonHistory,
    };

_$ProfileUserImpl _$$ProfileUserImplFromJson(Map<String, dynamic> json) =>
    _$ProfileUserImpl(
      username: json['username'] as String,
      avatarEmoji: json['avatar_emoji'] as String?,
      avatarUrl: json['avatar_url'] as String?,
    );

Map<String, dynamic> _$$ProfileUserImplToJson(_$ProfileUserImpl instance) =>
    <String, dynamic>{
      'username': instance.username,
      'avatar_emoji': instance.avatarEmoji,
      'avatar_url': instance.avatarUrl,
    };

_$UserStatsImpl _$$UserStatsImplFromJson(Map<String, dynamic> json) =>
    _$UserStatsImpl(
      seasonsPlayed: (json['seasons_played'] as num).toInt(),
      votingStreak: (json['voting_streak'] as num).toInt(),
      maxVotingStreak: (json['max_voting_streak'] as num).toInt(),
      guessAccuracy: (json['guess_accuracy'] as num).toDouble(),
      totalVotesCast: (json['total_votes_cast'] as num).toInt(),
      totalVotesReceived: (json['total_votes_received'] as num).toInt(),
      topAttributeAllTime: json['top_attribute_all_time'] == null
          ? null
          : TopAttributeDto.fromJson(
              json['top_attribute_all_time'] as Map<String, dynamic>),
    );

Map<String, dynamic> _$$UserStatsImplToJson(_$UserStatsImpl instance) =>
    <String, dynamic>{
      'seasons_played': instance.seasonsPlayed,
      'voting_streak': instance.votingStreak,
      'max_voting_streak': instance.maxVotingStreak,
      'guess_accuracy': instance.guessAccuracy,
      'total_votes_cast': instance.totalVotesCast,
      'total_votes_received': instance.totalVotesReceived,
      'top_attribute_all_time': instance.topAttributeAllTime,
    };

_$TopAttributeDtoImpl _$$TopAttributeDtoImplFromJson(
        Map<String, dynamic> json) =>
    _$TopAttributeDtoImpl(
      questionText: json['question_text'] as String,
      percentage: (json['percentage'] as num).toDouble(),
    );

Map<String, dynamic> _$$TopAttributeDtoImplToJson(
        _$TopAttributeDtoImpl instance) =>
    <String, dynamic>{
      'question_text': instance.questionText,
      'percentage': instance.percentage,
    };

_$ProfileAchievementImpl _$$ProfileAchievementImplFromJson(
        Map<String, dynamic> json) =>
    _$ProfileAchievementImpl(
      type: json['type'] as String,
      metadata: json['metadata'] as Map<String, dynamic>?,
      earnedAt: json['earned_at'] as String,
    );

Map<String, dynamic> _$$ProfileAchievementImplToJson(
        _$ProfileAchievementImpl instance) =>
    <String, dynamic>{
      'type': instance.type,
      'metadata': instance.metadata,
      'earned_at': instance.earnedAt,
    };

_$SeasonCardDtoImpl _$$SeasonCardDtoImplFromJson(Map<String, dynamic> json) =>
    _$SeasonCardDtoImpl(
      seasonId: json['season_id'] as String,
      seasonNumber: (json['season_number'] as num).toInt(),
      topAttribute: json['top_attribute'] as String,
      category: json['category'] as String,
      percentage: (json['percentage'] as num).toDouble(),
    );

Map<String, dynamic> _$$SeasonCardDtoImplToJson(_$SeasonCardDtoImpl instance) =>
    <String, dynamic>{
      'season_id': instance.seasonId,
      'season_number': instance.seasonNumber,
      'top_attribute': instance.topAttribute,
      'category': instance.category,
      'percentage': instance.percentage,
    };
