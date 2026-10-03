// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'chronicle.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

_$ChronicleEntryImpl _$$ChronicleEntryImplFromJson(Map<String, dynamic> json) =>
    _$ChronicleEntryImpl(
      questionId: json['question_id'] as String,
      questionText: json['question_text'] as String,
      category: json['category'] as String,
      userId: json['user_id'] as String,
      username: json['username'] as String,
      avatarEmoji: json['avatar_emoji'] as String?,
      percentage: (json['percentage'] as num).toDouble(),
      voteCount: (json['vote_count'] as num).toInt(),
      totalVoters: (json['total_voters'] as num).toInt(),
      smallSample: json['small_sample'] as bool? ?? false,
    );

Map<String, dynamic> _$$ChronicleEntryImplToJson(
        _$ChronicleEntryImpl instance) =>
    <String, dynamic>{
      'question_id': instance.questionId,
      'question_text': instance.questionText,
      'category': instance.category,
      'user_id': instance.userId,
      'username': instance.username,
      'avatar_emoji': instance.avatarEmoji,
      'percentage': instance.percentage,
      'vote_count': instance.voteCount,
      'total_voters': instance.totalVoters,
      'small_sample': instance.smallSample,
    };

_$ChronicleSeasonImpl _$$ChronicleSeasonImplFromJson(
        Map<String, dynamic> json) =>
    _$ChronicleSeasonImpl(
      seasonId: json['season_id'] as String,
      number: (json['number'] as num).toInt(),
      revealAt: json['reveal_at'] as String,
      entries: (json['entries'] as List<dynamic>?)
              ?.map((e) => ChronicleEntry.fromJson(e as Map<String, dynamic>))
              .toList() ??
          const [],
    );

Map<String, dynamic> _$$ChronicleSeasonImplToJson(
        _$ChronicleSeasonImpl instance) =>
    <String, dynamic>{
      'season_id': instance.seasonId,
      'number': instance.number,
      'reveal_at': instance.revealAt,
      'entries': instance.entries,
    };

_$StandingEntryImpl _$$StandingEntryImplFromJson(Map<String, dynamic> json) =>
    _$StandingEntryImpl(
      userId: json['user_id'] as String,
      username: json['username'] as String,
      avatarEmoji: json['avatar_emoji'] as String?,
      rank: (json['rank'] as num?)?.toInt() ?? 0,
      ranked: json['ranked'] as bool? ?? false,
      seasonsPlayed: (json['seasons_played'] as num?)?.toInt() ?? 0,
    );

Map<String, dynamic> _$$StandingEntryImplToJson(_$StandingEntryImpl instance) =>
    <String, dynamic>{
      'user_id': instance.userId,
      'username': instance.username,
      'avatar_emoji': instance.avatarEmoji,
      'rank': instance.rank,
      'ranked': instance.ranked,
      'seasons_played': instance.seasonsPlayed,
    };

_$GuessStandingImpl _$$GuessStandingImplFromJson(Map<String, dynamic> json) =>
    _$GuessStandingImpl(
      available: json['available'] as bool? ?? false,
      reason: json['reason'] as String? ?? '',
      members: (json['members'] as List<dynamic>?)
              ?.map((e) => StandingEntry.fromJson(e as Map<String, dynamic>))
              .toList() ??
          const [],
    );

Map<String, dynamic> _$$GuessStandingImplToJson(_$GuessStandingImpl instance) =>
    <String, dynamic>{
      'available': instance.available,
      'reason': instance.reason,
      'members': instance.members,
    };

_$ChronicleImpl _$$ChronicleImplFromJson(Map<String, dynamic> json) =>
    _$ChronicleImpl(
      seasons: (json['seasons'] as List<dynamic>?)
              ?.map((e) => ChronicleSeason.fromJson(e as Map<String, dynamic>))
              .toList() ??
          const [],
      standing: json['standing'] == null
          ? const GuessStanding()
          : GuessStanding.fromJson(json['standing'] as Map<String, dynamic>),
      seasonsTotal: (json['seasons_total'] as num?)?.toInt() ?? 0,
      seasonsShown: (json['seasons_shown'] as num?)?.toInt() ?? 0,
    );

Map<String, dynamic> _$$ChronicleImplToJson(_$ChronicleImpl instance) =>
    <String, dynamic>{
      'seasons': instance.seasons,
      'standing': instance.standing,
      'seasons_total': instance.seasonsTotal,
      'seasons_shown': instance.seasonsShown,
    };
