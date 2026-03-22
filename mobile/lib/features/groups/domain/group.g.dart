// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'group.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

_$GroupImpl _$$GroupImplFromJson(Map<String, dynamic> json) => _$GroupImpl(
      id: json['id'] as String,
      name: json['name'] as String,
      adminId: json['admin_id'] as String,
      inviteCode: json['invite_code'] as String,
      categories: (json['categories'] as List<dynamic>)
          .map((e) => e as String)
          .toList(),
      telegramUsername: json['telegram_username'] as String?,
      createdAt: json['created_at'] as String,
    );

Map<String, dynamic> _$$GroupImplToJson(_$GroupImpl instance) =>
    <String, dynamic>{
      'id': instance.id,
      'name': instance.name,
      'admin_id': instance.adminId,
      'invite_code': instance.inviteCode,
      'categories': instance.categories,
      'telegram_username': instance.telegramUsername,
      'created_at': instance.createdAt,
    };

_$ActiveSeasonImpl _$$ActiveSeasonImplFromJson(Map<String, dynamic> json) =>
    _$ActiveSeasonImpl(
      id: json['id'] as String,
      status: json['status'] as String,
      revealAt: json['reveal_at'] as String,
      votedCount: (json['voted_count'] as num).toInt(),
      totalCount: (json['total_count'] as num).toInt(),
      userVoted: json['user_voted'] as bool,
    );

Map<String, dynamic> _$$ActiveSeasonImplToJson(_$ActiveSeasonImpl instance) =>
    <String, dynamic>{
      'id': instance.id,
      'status': instance.status,
      'reveal_at': instance.revealAt,
      'voted_count': instance.votedCount,
      'total_count': instance.totalCount,
      'user_voted': instance.userVoted,
    };

_$GroupListItemImpl _$$GroupListItemImplFromJson(Map<String, dynamic> json) =>
    _$GroupListItemImpl(
      id: json['id'] as String,
      name: json['name'] as String,
      memberCount: (json['member_count'] as num).toInt(),
      inviteCode: json['invite_code'] as String,
      telegramUsername: json['telegram_username'] as String?,
      activeSeason: json['active_season'] == null
          ? null
          : ActiveSeason.fromJson(
              json['active_season'] as Map<String, dynamic>),
    );

Map<String, dynamic> _$$GroupListItemImplToJson(_$GroupListItemImpl instance) =>
    <String, dynamic>{
      'id': instance.id,
      'name': instance.name,
      'member_count': instance.memberCount,
      'invite_code': instance.inviteCode,
      'telegram_username': instance.telegramUsername,
      'active_season': instance.activeSeason,
    };

_$MemberImpl _$$MemberImplFromJson(Map<String, dynamic> json) => _$MemberImpl(
      id: json['id'] as String,
      username: json['username'] as String,
      avatarEmoji: json['avatar_emoji'] as String?,
      avatarUrl: json['avatar_url'] as String?,
      isAdmin: json['is_admin'] as bool,
    );

Map<String, dynamic> _$$MemberImplToJson(_$MemberImpl instance) =>
    <String, dynamic>{
      'id': instance.id,
      'username': instance.username,
      'avatar_emoji': instance.avatarEmoji,
      'avatar_url': instance.avatarUrl,
      'is_admin': instance.isAdmin,
    };

_$GroupDetailImpl _$$GroupDetailImplFromJson(Map<String, dynamic> json) =>
    _$GroupDetailImpl(
      group: Group.fromJson(json['group'] as Map<String, dynamic>),
      members: (json['members'] as List<dynamic>)
          .map((e) => Member.fromJson(e as Map<String, dynamic>))
          .toList(),
      activeSeason: json['active_season'] == null
          ? null
          : ActiveSeason.fromJson(
              json['active_season'] as Map<String, dynamic>),
    );

Map<String, dynamic> _$$GroupDetailImplToJson(_$GroupDetailImpl instance) =>
    <String, dynamic>{
      'group': instance.group,
      'members': instance.members,
      'active_season': instance.activeSeason,
    };

_$JoinPreviewImpl _$$JoinPreviewImplFromJson(Map<String, dynamic> json) =>
    _$JoinPreviewImpl(
      name: json['name'] as String,
      memberCount: (json['member_count'] as num).toInt(),
      adminUsername: json['admin_username'] as String,
    );

Map<String, dynamic> _$$JoinPreviewImplToJson(_$JoinPreviewImpl instance) =>
    <String, dynamic>{
      'name': instance.name,
      'member_count': instance.memberCount,
      'admin_username': instance.adminUsername,
    };
