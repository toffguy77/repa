import 'package:freezed_annotation/freezed_annotation.dart';

part 'chronicle.freezed.dart';
part 'chronicle.g.dart';

/// One standout result from a past season: a question, and who led it.
@freezed
class ChronicleEntry with _$ChronicleEntry {
  const factory ChronicleEntry({
    @JsonKey(name: 'question_id') required String questionId,
    @JsonKey(name: 'question_text') required String questionText,
    required String category,
    @JsonKey(name: 'user_id') required String userId,
    required String username,
    @JsonKey(name: 'avatar_emoji') String? avatarEmoji,
    required double percentage,
    @JsonKey(name: 'vote_count') required int voteCount,
    @JsonKey(name: 'total_voters') required int totalVoters,

    /// Whether this result's percentages were computed over too few voters to hide who voted. The
    /// server decides, using the same threshold the detector is gated on — the client must not carry
    /// its own copy of that number.
    @JsonKey(name: 'small_sample') @Default(false) bool smallSample,
  }) = _ChronicleEntry;

  factory ChronicleEntry.fromJson(Map<String, dynamic> json) =>
      _$ChronicleEntryFromJson(json);
}

/// One season's worth of the record.
@freezed
class ChronicleSeason with _$ChronicleSeason {
  const factory ChronicleSeason({
    @JsonKey(name: 'season_id') required String seasonId,
    required int number,
    @JsonKey(name: 'reveal_at') required String revealAt,
    @Default([]) List<ChronicleEntry> entries,
  }) = _ChronicleSeason;

  factory ChronicleSeason.fromJson(Map<String, dynamic> json) =>
      _$ChronicleSeasonFromJson(json);
}

/// One member's place in the guess standing.
@freezed
class StandingEntry with _$StandingEntry {
  const factory StandingEntry({
    @JsonKey(name: 'user_id') required String userId,
    required String username,
    @JsonKey(name: 'avatar_emoji') String? avatarEmoji,

    /// 1-based, or 0 for a member who has not been through a reveal yet.
    @Default(0) int rank,

    /// Whether this member has a place at all. "Has not played here yet" and "placed last" are
    /// different facts, and showing the first as the second says something untrue about a newcomer.
    @Default(false) bool ranked,

    /// How much history the place is based on: a place over two seasons and one over ten are not
    /// comparable, so the basis is shown next to the position.
    @JsonKey(name: 'seasons_played') @Default(0) int seasonsPlayed,

    // The accuracy *figure* is deliberately not part of this model, because the server does not send
    // it: with the chronicle publishing each question's winner, an extreme figure recovers a member's
    // individual votes, and the rolling average can be diffed across weeks into a match count. See
    // openspec/specs/group-chronicle.
  }) = _StandingEntry;

  factory StandingEntry.fromJson(Map<String, dynamic> json) =>
      _$StandingEntryFromJson(json);
}

/// The group's guess standing, or the reason it is not being shown yet.
@freezed
class GuessStanding with _$GuessStanding {
  const factory GuessStanding({
    @Default(false) bool available,
    @Default('') String reason,
    @Default([]) List<StandingEntry> members,
  }) = _GuessStanding;

  factory GuessStanding.fromJson(Map<String, dynamic> json) =>
      _$GuessStandingFromJson(json);
}

@freezed
class Chronicle with _$Chronicle {
  const factory Chronicle({
    @Default([]) List<ChronicleSeason> seasons,
    @Default(GuessStanding()) GuessStanding standing,
    @JsonKey(name: 'seasons_total') @Default(0) int seasonsTotal,
    @JsonKey(name: 'seasons_shown') @Default(0) int seasonsShown,
  }) = _Chronicle;

  factory Chronicle.fromJson(Map<String, dynamic> json) =>
      _$ChronicleFromJson(json);
}

extension ChronicleView on Chronicle {
  bool get isEmpty => seasons.isEmpty;

  /// Whether the chronicle shows less than the group's full history, which is what a "more seasons
  /// exist" line would be about. Not pagination — just honesty about the cap.
  bool get isTruncated => seasonsTotal > seasonsShown;
}

extension StandingEntryView on StandingEntry {
  bool get isRanked => ranked && rank > 0;
}
