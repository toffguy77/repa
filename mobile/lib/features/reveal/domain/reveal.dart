import 'package:freezed_annotation/freezed_annotation.dart';

part 'reveal.freezed.dart';
part 'reveal.g.dart';

@freezed
class RevealData with _$RevealData {
  const factory RevealData({
    @JsonKey(name: 'my_card') required MyCard myCard,
    @JsonKey(name: 'group_summary') required GroupSummary groupSummary,
  }) = _RevealData;

  factory RevealData.fromJson(Map<String, dynamic> json) =>
      _$RevealDataFromJson(json);
}

@freezed
class MyCard with _$MyCard {
  const factory MyCard({
    @JsonKey(name: 'top_attributes') required List<AttributeDto> topAttributes,
    @JsonKey(name: 'hidden_attributes')
    required List<AttributeDto> hiddenAttributes,
    @JsonKey(name: 'reputation_title') required String reputationTitle,
    TrendDto? trend,
    @JsonKey(name: 'new_achievements')
    required List<AchievementDto> newAchievements,
    @JsonKey(name: 'card_image_url') required String cardImageUrl,
  }) = _MyCard;

  factory MyCard.fromJson(Map<String, dynamic> json) => _$MyCardFromJson(json);
}

@freezed
class AttributeDto with _$AttributeDto {
  const factory AttributeDto({
    @JsonKey(name: 'question_id') required String questionId,
    @JsonKey(name: 'question_text') required String questionText,
    required String category,
    required double percentage,
    required int rank,
  }) = _AttributeDto;

  factory AttributeDto.fromJson(Map<String, dynamic> json) =>
      _$AttributeDtoFromJson(json);
}

@freezed
class TrendDto with _$TrendDto {
  const factory TrendDto({
    required String attribute,
    required String change,
    required double delta,
  }) = _TrendDto;

  factory TrendDto.fromJson(Map<String, dynamic> json) =>
      _$TrendDtoFromJson(json);
}

@freezed
class AchievementDto with _$AchievementDto {
  const factory AchievementDto({
    required String type,
    Map<String, dynamic>? metadata,
  }) = _AchievementDto;

  factory AchievementDto.fromJson(Map<String, dynamic> json) =>
      _$AchievementDtoFromJson(json);
}

@freezed
class GroupSummary with _$GroupSummary {
  const factory GroupSummary({
    @JsonKey(name: 'top_per_question')
    required List<TopQuestionResult> topPerQuestion,
    @JsonKey(name: 'voter_count') required int voterCount,
  }) = _GroupSummary;

  factory GroupSummary.fromJson(Map<String, dynamic> json) =>
      _$GroupSummaryFromJson(json);
}

@freezed
class TopQuestionResult with _$TopQuestionResult {
  const factory TopQuestionResult({
    @JsonKey(name: 'question_id') required String questionId,
    @JsonKey(name: 'question_text') required String questionText,
    @JsonKey(name: 'user_id') required String userId,
    required String username,
    @JsonKey(name: 'avatar_emoji') String? avatarEmoji,
    required double percentage,
  }) = _TopQuestionResult;

  factory TopQuestionResult.fromJson(Map<String, dynamic> json) =>
      _$TopQuestionResultFromJson(json);
}

@freezed
class MemberCard with _$MemberCard {
  const factory MemberCard({
    @JsonKey(name: 'user_id') required String userId,
    required String username,
    @JsonKey(name: 'avatar_emoji') String? avatarEmoji,
    @JsonKey(name: 'avatar_url') String? avatarUrl,
    @JsonKey(name: 'top_attributes') required List<AttributeDto> topAttributes,
    @JsonKey(name: 'reputation_title') required String reputationTitle,
  }) = _MemberCard;

  factory MemberCard.fromJson(Map<String, dynamic> json) =>
      _$MemberCardFromJson(json);
}

/// One partially revealed voter: enough to guess, not enough to know.
///
/// There is deliberately no username field — mirroring the backend DTO, which has no place to put
/// a full name, so no change to a shared shape can accidentally leak one.
@freezed
class DetectorHint with _$DetectorHint {
  const factory DetectorHint({
    @JsonKey(name: 'first_letter') required String firstLetter,
    @JsonKey(name: 'avatar_emoji') String? avatarEmoji,
    @JsonKey(name: 'avatar_url') String? avatarUrl,
  }) = _DetectorHint;

  factory DetectorHint.fromJson(Map<String, dynamic> json) =>
      _$DetectorHintFromJson(json);
}

@freezed
class DetectorResult with _$DetectorResult {
  const factory DetectorResult({
    required bool purchased,
    required List<VoterProfile> voters,
    @JsonKey(name: 'crystal_balance') required int crystalBalance,
    /// False while the group has fewer than 5 members: a voter list drawn from at most
    /// 4 people is not anonymous, so the backend refuses the purchase (GROUP_TOO_SMALL).
    /// Defaults to true so an older backend keeps working.
    @Default(true) bool available,

    /// The ladder. All default so an older backend keeps working: the free count is 0, no hints
    /// are revealed, and the prices fall back to the shipped values.
    @JsonKey(name: 'voter_count') @Default(0) int voterCount,
    @Default(<DetectorHint>[]) List<DetectorHint> hints,
    @JsonKey(name: 'hint_cost') @Default(3) int hintCost,
    @JsonKey(name: 'full_cost') @Default(10) int fullCost,
    @JsonKey(name: 'hint_available') @Default(false) bool hintAvailable,
  }) = _DetectorResult;

  factory DetectorResult.fromJson(Map<String, dynamic> json) =>
      _$DetectorResultFromJson(json);
}

@freezed
class VoterProfile with _$VoterProfile {
  const factory VoterProfile({
    required String id,
    required String username,
    @JsonKey(name: 'avatar_emoji') String? avatarEmoji,
    @JsonKey(name: 'avatar_url') String? avatarUrl,
  }) = _VoterProfile;

  factory VoterProfile.fromJson(Map<String, dynamic> json) =>
      _$VoterProfileFromJson(json);
}

@freezed
class CardUrlResult with _$CardUrlResult {
  const factory CardUrlResult({
    @JsonKey(name: 'image_url') String? imageUrl,
    required String status,
  }) = _CardUrlResult;

  factory CardUrlResult.fromJson(Map<String, dynamic> json) =>
      _$CardUrlResultFromJson(json);
}

@freezed
class ReactionCounts with _$ReactionCounts {
  const factory ReactionCounts({
    required Map<String, int> counts,
    @JsonKey(name: 'my_emoji') String? myEmoji,
  }) = _ReactionCounts;

  factory ReactionCounts.fromJson(Map<String, dynamic> json) =>
      _$ReactionCountsFromJson(json);
}

/// What a member may know about votes concerning them before the Reveal.
///
/// Deliberately tiny: a count, an optional single emoji, and the reveal time. No identity and no
/// attribute — mirroring the backend shape, which has no field for either.
@freezed
class AnticipationState with _$AnticipationState {
  const factory AnticipationState({
    @JsonKey(name: 'voters_about_me') @Default(0) int votersAboutMe,
    @JsonKey(name: 'teaser_emoji') @Default('') String teaserEmoji,
    @JsonKey(name: 'reveal_at') @Default('') String revealAt,
  }) = _AnticipationState;

  factory AnticipationState.fromJson(Map<String, dynamic> json) =>
      _$AnticipationStateFromJson(json);
}
