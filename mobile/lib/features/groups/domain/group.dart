import 'package:freezed_annotation/freezed_annotation.dart';

part 'group.freezed.dart';
part 'group.g.dart';

@freezed
class Group with _$Group {
  const factory Group({
    required String id,
    required String name,
    @JsonKey(name: 'admin_id') required String adminId,
    @JsonKey(name: 'invite_code') required String inviteCode,
    required List<String> categories,

    /// Whether the group restricts itself to warm and neutral questions. Defaults to false so an
    /// older response without the field is read as an ordinary group rather than crashing.
    @JsonKey(name: 'kind_only') @Default(false) bool kindOnly,
    @JsonKey(name: 'telegram_username') String? telegramUsername,
    @JsonKey(name: 'created_at') required String createdAt,
  }) = _Group;

  factory Group.fromJson(Map<String, dynamic> json) => _$GroupFromJson(json);
}

/// What a pending Reveal is waiting for. Mirrors the backend's `reveal_state`
/// (see internal/eligibility) — unknown values fall back to [scheduled] so a new
/// backend state never crashes an older client.
enum SeasonRevealState {
  @JsonValue('SCHEDULED')
  scheduled,
  @JsonValue('WAITING_FOR_MEMBERS')
  waitingForMembers,
  @JsonValue('WAITING_FOR_VOTERS')
  waitingForVoters,
  @JsonValue('POSTPONED')
  postponed,
  @JsonValue('REVEALED')
  revealed,
}

@freezed
class ActiveSeason with _$ActiveSeason {
  const factory ActiveSeason({
    required String id,
    required String status,
    @JsonKey(name: 'reveal_at') required String revealAt,
    @JsonKey(name: 'voted_count') required int votedCount,
    @JsonKey(name: 'total_count') required int totalCount,
    @JsonKey(name: 'user_voted') required bool userVoted,
    @JsonKey(name: 'reveal_state', unknownEnumValue: SeasonRevealState.scheduled)
    @Default(SeasonRevealState.scheduled)
    SeasonRevealState revealState,
    @JsonKey(name: 'members_needed') @Default(0) int membersNeeded,
    @JsonKey(name: 'voters_needed') @Default(0) int votersNeeded,
  }) = _ActiveSeason;

  factory ActiveSeason.fromJson(Map<String, dynamic> json) =>
      _$ActiveSeasonFromJson(json);
}

/// Russian copy for the pending-Reveal states, kept next to the model so the group
/// screen and the reveal screen cannot drift apart.
extension ActiveSeasonRevealCopy on ActiveSeason {
  /// Whether the Reveal is blocked on people rather than on time.
  bool get isWaitingForPeople =>
      revealState == SeasonRevealState.waitingForMembers ||
      revealState == SeasonRevealState.waitingForVoters;

  /// Short headline for the season card / waiting screen.
  String get revealHeadline {
    switch (revealState) {
      case SeasonRevealState.waitingForMembers:
        return 'Нужно больше людей';
      case SeasonRevealState.waitingForVoters:
        return 'Нужно больше голосов';
      case SeasonRevealState.postponed:
        return 'Репа перенесена';
      case SeasonRevealState.revealed:
        return 'Результаты готовы!';
      case SeasonRevealState.scheduled:
        return 'Ждём пятницы';
    }
  }

  /// One line explaining exactly what is missing.
  String get revealExplanation {
    switch (revealState) {
      case SeasonRevealState.waitingForMembers:
        return 'Позови ещё ${_plural(membersNeeded)} — и репа откроется.';
      case SeasonRevealState.waitingForVoters:
        return 'Ещё ${_plural(votersNeeded)} должны проголосовать.';
      case SeasonRevealState.postponed:
        return 'Голосов не хватило. Голоса сохранены, откроем в следующую пятницу.';
      case SeasonRevealState.revealed:
        return 'Смотри свою карточку.';
      case SeasonRevealState.scheduled:
        return 'Reveal в пятницу в 20:00.';
    }
  }

  static String _plural(int n) {
    final mod100 = n % 100;
    if (mod100 >= 11 && mod100 <= 14) return '$n человек';
    switch (n % 10) {
      case 1:
      case 2:
      case 3:
      case 4:
        return '$n человека';
      default:
        return '$n человек';
    }
  }
}

/// What crossing a group-size threshold changes. Mirrors `eligibility`'s unlock keys — an unknown
/// value falls back to [unknown] so a new server-side threshold does not crash an older client.
enum GrowthUnlock {
  @JsonValue('REVEAL')
  reveal,
  @JsonValue('DETECTOR')
  detector,
  unknown,
}

/// The next group size that changes what the group can do. Absent once every threshold is crossed —
/// the server sends null rather than a further target, and the app must not invent one.
@freezed
class GrowthThreshold with _$GrowthThreshold {
  const factory GrowthThreshold({
    required int size,
    required int needed,
    @JsonKey(unknownEnumValue: GrowthUnlock.unknown)
    @Default(GrowthUnlock.unknown)
    GrowthUnlock unlocks,
  }) = _GrowthThreshold;

  factory GrowthThreshold.fromJson(Map<String, dynamic> json) =>
      _$GrowthThresholdFromJson(json);
}

@freezed
class GroupListItem with _$GroupListItem {
  const factory GroupListItem({
    required String id,
    required String name,
    @JsonKey(name: 'member_count') required int memberCount,
    @JsonKey(name: 'invite_code') required String inviteCode,
    @JsonKey(name: 'telegram_username') String? telegramUsername,
    @JsonKey(name: 'active_season') ActiveSeason? activeSeason,
  }) = _GroupListItem;

  factory GroupListItem.fromJson(Map<String, dynamic> json) =>
      _$GroupListItemFromJson(json);
}

@freezed
class Member with _$Member {
  const factory Member({
    required String id,
    required String username,
    @JsonKey(name: 'avatar_emoji') String? avatarEmoji,
    @JsonKey(name: 'avatar_url') String? avatarUrl,
    @JsonKey(name: 'is_admin') required bool isAdmin,
  }) = _Member;

  factory Member.fromJson(Map<String, dynamic> json) =>
      _$MemberFromJson(json);
}

@freezed
class GroupDetail with _$GroupDetail {
  const factory GroupDetail({
    required Group group,
    required List<Member> members,
    @JsonKey(name: 'active_season') ActiveSeason? activeSeason,

    /// How many members this viewer can actually be rated by — membership minus anyone blocked in
    /// either direction. The anonymity warning is judged against this, because a group of five with
    /// three blocks behaves like a group of two. Falls back to 0, which callers read as "unknown"
    /// and substitute the raw member count for.
    @JsonKey(name: 'effective_member_count') @Default(0) int effectiveMemberCount,

    /// The next group-size threshold, or null once the group has crossed them all. Counted by the
    /// server against [effectiveMemberCount], so it agrees with the anonymity warning.
    @JsonKey(name: 'next_threshold') GrowthThreshold? nextThreshold,
  }) = _GroupDetail;

  factory GroupDetail.fromJson(Map<String, dynamic> json) =>
      _$GroupDetailFromJson(json);
}

@freezed
class JoinPreview with _$JoinPreview {
  const factory JoinPreview({
    required String name,
    @JsonKey(name: 'member_count') required int memberCount,
    @JsonKey(name: 'admin_username') required String adminUsername,
  }) = _JoinPreview;

  factory JoinPreview.fromJson(Map<String, dynamic> json) =>
      _$JoinPreviewFromJson(json);
}
