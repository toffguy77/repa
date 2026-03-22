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
    @JsonKey(name: 'telegram_username') String? telegramUsername,
    @JsonKey(name: 'created_at') required String createdAt,
  }) = _Group;

  factory Group.fromJson(Map<String, dynamic> json) => _$GroupFromJson(json);
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
  }) = _ActiveSeason;

  factory ActiveSeason.fromJson(Map<String, dynamic> json) =>
      _$ActiveSeasonFromJson(json);
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
