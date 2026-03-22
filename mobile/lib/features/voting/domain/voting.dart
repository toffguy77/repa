import 'package:freezed_annotation/freezed_annotation.dart';

part 'voting.freezed.dart';
part 'voting.g.dart';

@freezed
class VotingQuestion with _$VotingQuestion {
  const factory VotingQuestion({
    @JsonKey(name: 'question_id') required String questionId,
    required String text,
    required String category,
    required bool answered,
  }) = _VotingQuestion;

  factory VotingQuestion.fromJson(Map<String, dynamic> json) =>
      _$VotingQuestionFromJson(json);
}

@freezed
class VotingTarget with _$VotingTarget {
  const factory VotingTarget({
    @JsonKey(name: 'user_id') required String userId,
    required String username,
    @JsonKey(name: 'avatar_emoji') String? avatarEmoji,
    @JsonKey(name: 'avatar_url') String? avatarUrl,
  }) = _VotingTarget;

  factory VotingTarget.fromJson(Map<String, dynamic> json) =>
      _$VotingTargetFromJson(json);
}

@freezed
class VotingProgress with _$VotingProgress {
  const factory VotingProgress({
    required int answered,
    required int total,
  }) = _VotingProgress;

  factory VotingProgress.fromJson(Map<String, dynamic> json) =>
      _$VotingProgressFromJson(json);
}

@freezed
class VotingSession with _$VotingSession {
  const factory VotingSession({
    @JsonKey(name: 'season_id') required String seasonId,
    required List<VotingQuestion> questions,
    required List<VotingTarget> targets,
    required VotingProgress progress,
  }) = _VotingSession;

  factory VotingSession.fromJson(Map<String, dynamic> json) =>
      _$VotingSessionFromJson(json);
}

@freezed
class VoteResultData with _$VoteResultData {
  const factory VoteResultData({
    required VoteInfo vote,
    required VotingProgress progress,
  }) = _VoteResultData;

  factory VoteResultData.fromJson(Map<String, dynamic> json) =>
      _$VoteResultDataFromJson(json);
}

@freezed
class VoteInfo with _$VoteInfo {
  const factory VoteInfo({
    @JsonKey(name: 'question_id') required String questionId,
    @JsonKey(name: 'target_id') required String targetId,
  }) = _VoteInfo;

  factory VoteInfo.fromJson(Map<String, dynamic> json) =>
      _$VoteInfoFromJson(json);
}

@freezed
class GroupVotingProgress with _$GroupVotingProgress {
  const factory GroupVotingProgress({
    @JsonKey(name: 'voted_count') required int votedCount,
    @JsonKey(name: 'total_count') required int totalCount,
    @JsonKey(name: 'quorum_reached') required bool quorumReached,
    @JsonKey(name: 'quorum_threshold') required double quorumThreshold,
    @JsonKey(name: 'user_voted') required bool userVoted,
  }) = _GroupVotingProgress;

  factory GroupVotingProgress.fromJson(Map<String, dynamic> json) =>
      _$GroupVotingProgressFromJson(json);
}
