// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'voting.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

_$VotingQuestionImpl _$$VotingQuestionImplFromJson(Map<String, dynamic> json) =>
    _$VotingQuestionImpl(
      questionId: json['question_id'] as String,
      text: json['text'] as String,
      category: json['category'] as String,
      answered: json['answered'] as bool,
    );

Map<String, dynamic> _$$VotingQuestionImplToJson(
        _$VotingQuestionImpl instance) =>
    <String, dynamic>{
      'question_id': instance.questionId,
      'text': instance.text,
      'category': instance.category,
      'answered': instance.answered,
    };

_$VotingTargetImpl _$$VotingTargetImplFromJson(Map<String, dynamic> json) =>
    _$VotingTargetImpl(
      userId: json['user_id'] as String,
      username: json['username'] as String,
      avatarEmoji: json['avatar_emoji'] as String?,
      avatarUrl: json['avatar_url'] as String?,
    );

Map<String, dynamic> _$$VotingTargetImplToJson(_$VotingTargetImpl instance) =>
    <String, dynamic>{
      'user_id': instance.userId,
      'username': instance.username,
      'avatar_emoji': instance.avatarEmoji,
      'avatar_url': instance.avatarUrl,
    };

_$VotingProgressImpl _$$VotingProgressImplFromJson(Map<String, dynamic> json) =>
    _$VotingProgressImpl(
      answered: (json['answered'] as num).toInt(),
      total: (json['total'] as num).toInt(),
    );

Map<String, dynamic> _$$VotingProgressImplToJson(
        _$VotingProgressImpl instance) =>
    <String, dynamic>{
      'answered': instance.answered,
      'total': instance.total,
    };

_$VotingSessionImpl _$$VotingSessionImplFromJson(Map<String, dynamic> json) =>
    _$VotingSessionImpl(
      seasonId: json['season_id'] as String,
      questions: (json['questions'] as List<dynamic>)
          .map((e) => VotingQuestion.fromJson(e as Map<String, dynamic>))
          .toList(),
      targets: (json['targets'] as List<dynamic>)
          .map((e) => VotingTarget.fromJson(e as Map<String, dynamic>))
          .toList(),
      progress:
          VotingProgress.fromJson(json['progress'] as Map<String, dynamic>),
    );

Map<String, dynamic> _$$VotingSessionImplToJson(_$VotingSessionImpl instance) =>
    <String, dynamic>{
      'season_id': instance.seasonId,
      'questions': instance.questions,
      'targets': instance.targets,
      'progress': instance.progress,
    };

_$VoteResultDataImpl _$$VoteResultDataImplFromJson(Map<String, dynamic> json) =>
    _$VoteResultDataImpl(
      vote: VoteInfo.fromJson(json['vote'] as Map<String, dynamic>),
      progress:
          VotingProgress.fromJson(json['progress'] as Map<String, dynamic>),
    );

Map<String, dynamic> _$$VoteResultDataImplToJson(
        _$VoteResultDataImpl instance) =>
    <String, dynamic>{
      'vote': instance.vote,
      'progress': instance.progress,
    };

_$VoteInfoImpl _$$VoteInfoImplFromJson(Map<String, dynamic> json) =>
    _$VoteInfoImpl(
      questionId: json['question_id'] as String,
      targetId: json['target_id'] as String,
    );

Map<String, dynamic> _$$VoteInfoImplToJson(_$VoteInfoImpl instance) =>
    <String, dynamic>{
      'question_id': instance.questionId,
      'target_id': instance.targetId,
    };

_$GroupVotingProgressImpl _$$GroupVotingProgressImplFromJson(
        Map<String, dynamic> json) =>
    _$GroupVotingProgressImpl(
      votedCount: (json['voted_count'] as num).toInt(),
      totalCount: (json['total_count'] as num).toInt(),
      quorumReached: json['quorum_reached'] as bool,
      quorumThreshold: (json['quorum_threshold'] as num).toDouble(),
      userVoted: json['user_voted'] as bool,
    );

Map<String, dynamic> _$$GroupVotingProgressImplToJson(
        _$GroupVotingProgressImpl instance) =>
    <String, dynamic>{
      'voted_count': instance.votedCount,
      'total_count': instance.totalCount,
      'quorum_reached': instance.quorumReached,
      'quorum_threshold': instance.quorumThreshold,
      'user_voted': instance.userVoted,
    };
