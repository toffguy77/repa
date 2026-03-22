// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'group.dart';

// **************************************************************************
// FreezedGenerator
// **************************************************************************

T _$identity<T>(T value) => value;

final _privateConstructorUsedError = UnsupportedError(
    'It seems like you constructed your class using `MyClass._()`. This constructor is only meant to be used by freezed and you are not supposed to need it nor use it.\nPlease check the documentation here for more information: https://github.com/rrousselGit/freezed#adding-getters-and-methods-to-our-models');

Group _$GroupFromJson(Map<String, dynamic> json) {
  return _Group.fromJson(json);
}

/// @nodoc
mixin _$Group {
  String get id => throw _privateConstructorUsedError;
  String get name => throw _privateConstructorUsedError;
  @JsonKey(name: 'admin_id')
  String get adminId => throw _privateConstructorUsedError;
  @JsonKey(name: 'invite_code')
  String get inviteCode => throw _privateConstructorUsedError;
  List<String> get categories => throw _privateConstructorUsedError;
  @JsonKey(name: 'telegram_username')
  String? get telegramUsername => throw _privateConstructorUsedError;
  @JsonKey(name: 'created_at')
  String get createdAt => throw _privateConstructorUsedError;

  Map<String, dynamic> toJson() => throw _privateConstructorUsedError;
  @JsonKey(ignore: true)
  $GroupCopyWith<Group> get copyWith => throw _privateConstructorUsedError;
}

/// @nodoc
abstract class $GroupCopyWith<$Res> {
  factory $GroupCopyWith(Group value, $Res Function(Group) then) =
      _$GroupCopyWithImpl<$Res, Group>;
  @useResult
  $Res call(
      {String id,
      String name,
      @JsonKey(name: 'admin_id') String adminId,
      @JsonKey(name: 'invite_code') String inviteCode,
      List<String> categories,
      @JsonKey(name: 'telegram_username') String? telegramUsername,
      @JsonKey(name: 'created_at') String createdAt});
}

/// @nodoc
class _$GroupCopyWithImpl<$Res, $Val extends Group>
    implements $GroupCopyWith<$Res> {
  _$GroupCopyWithImpl(this._value, this._then);

  // ignore: unused_field
  final $Val _value;
  // ignore: unused_field
  final $Res Function($Val) _then;

  @pragma('vm:prefer-inline')
  @override
  $Res call({
    Object? id = null,
    Object? name = null,
    Object? adminId = null,
    Object? inviteCode = null,
    Object? categories = null,
    Object? telegramUsername = freezed,
    Object? createdAt = null,
  }) {
    return _then(_value.copyWith(
      id: null == id
          ? _value.id
          : id // ignore: cast_nullable_to_non_nullable
              as String,
      name: null == name
          ? _value.name
          : name // ignore: cast_nullable_to_non_nullable
              as String,
      adminId: null == adminId
          ? _value.adminId
          : adminId // ignore: cast_nullable_to_non_nullable
              as String,
      inviteCode: null == inviteCode
          ? _value.inviteCode
          : inviteCode // ignore: cast_nullable_to_non_nullable
              as String,
      categories: null == categories
          ? _value.categories
          : categories // ignore: cast_nullable_to_non_nullable
              as List<String>,
      telegramUsername: freezed == telegramUsername
          ? _value.telegramUsername
          : telegramUsername // ignore: cast_nullable_to_non_nullable
              as String?,
      createdAt: null == createdAt
          ? _value.createdAt
          : createdAt // ignore: cast_nullable_to_non_nullable
              as String,
    ) as $Val);
  }
}

/// @nodoc
abstract class _$$GroupImplCopyWith<$Res> implements $GroupCopyWith<$Res> {
  factory _$$GroupImplCopyWith(
          _$GroupImpl value, $Res Function(_$GroupImpl) then) =
      __$$GroupImplCopyWithImpl<$Res>;
  @override
  @useResult
  $Res call(
      {String id,
      String name,
      @JsonKey(name: 'admin_id') String adminId,
      @JsonKey(name: 'invite_code') String inviteCode,
      List<String> categories,
      @JsonKey(name: 'telegram_username') String? telegramUsername,
      @JsonKey(name: 'created_at') String createdAt});
}

/// @nodoc
class __$$GroupImplCopyWithImpl<$Res>
    extends _$GroupCopyWithImpl<$Res, _$GroupImpl>
    implements _$$GroupImplCopyWith<$Res> {
  __$$GroupImplCopyWithImpl(
      _$GroupImpl _value, $Res Function(_$GroupImpl) _then)
      : super(_value, _then);

  @pragma('vm:prefer-inline')
  @override
  $Res call({
    Object? id = null,
    Object? name = null,
    Object? adminId = null,
    Object? inviteCode = null,
    Object? categories = null,
    Object? telegramUsername = freezed,
    Object? createdAt = null,
  }) {
    return _then(_$GroupImpl(
      id: null == id
          ? _value.id
          : id // ignore: cast_nullable_to_non_nullable
              as String,
      name: null == name
          ? _value.name
          : name // ignore: cast_nullable_to_non_nullable
              as String,
      adminId: null == adminId
          ? _value.adminId
          : adminId // ignore: cast_nullable_to_non_nullable
              as String,
      inviteCode: null == inviteCode
          ? _value.inviteCode
          : inviteCode // ignore: cast_nullable_to_non_nullable
              as String,
      categories: null == categories
          ? _value._categories
          : categories // ignore: cast_nullable_to_non_nullable
              as List<String>,
      telegramUsername: freezed == telegramUsername
          ? _value.telegramUsername
          : telegramUsername // ignore: cast_nullable_to_non_nullable
              as String?,
      createdAt: null == createdAt
          ? _value.createdAt
          : createdAt // ignore: cast_nullable_to_non_nullable
              as String,
    ));
  }
}

/// @nodoc
@JsonSerializable()
class _$GroupImpl implements _Group {
  const _$GroupImpl(
      {required this.id,
      required this.name,
      @JsonKey(name: 'admin_id') required this.adminId,
      @JsonKey(name: 'invite_code') required this.inviteCode,
      required final List<String> categories,
      @JsonKey(name: 'telegram_username') this.telegramUsername,
      @JsonKey(name: 'created_at') required this.createdAt})
      : _categories = categories;

  factory _$GroupImpl.fromJson(Map<String, dynamic> json) =>
      _$$GroupImplFromJson(json);

  @override
  final String id;
  @override
  final String name;
  @override
  @JsonKey(name: 'admin_id')
  final String adminId;
  @override
  @JsonKey(name: 'invite_code')
  final String inviteCode;
  final List<String> _categories;
  @override
  List<String> get categories {
    if (_categories is EqualUnmodifiableListView) return _categories;
    // ignore: implicit_dynamic_type
    return EqualUnmodifiableListView(_categories);
  }

  @override
  @JsonKey(name: 'telegram_username')
  final String? telegramUsername;
  @override
  @JsonKey(name: 'created_at')
  final String createdAt;

  @override
  String toString() {
    return 'Group(id: $id, name: $name, adminId: $adminId, inviteCode: $inviteCode, categories: $categories, telegramUsername: $telegramUsername, createdAt: $createdAt)';
  }

  @override
  bool operator ==(Object other) {
    return identical(this, other) ||
        (other.runtimeType == runtimeType &&
            other is _$GroupImpl &&
            (identical(other.id, id) || other.id == id) &&
            (identical(other.name, name) || other.name == name) &&
            (identical(other.adminId, adminId) || other.adminId == adminId) &&
            (identical(other.inviteCode, inviteCode) ||
                other.inviteCode == inviteCode) &&
            const DeepCollectionEquality()
                .equals(other._categories, _categories) &&
            (identical(other.telegramUsername, telegramUsername) ||
                other.telegramUsername == telegramUsername) &&
            (identical(other.createdAt, createdAt) ||
                other.createdAt == createdAt));
  }

  @JsonKey(ignore: true)
  @override
  int get hashCode => Object.hash(
      runtimeType,
      id,
      name,
      adminId,
      inviteCode,
      const DeepCollectionEquality().hash(_categories),
      telegramUsername,
      createdAt);

  @JsonKey(ignore: true)
  @override
  @pragma('vm:prefer-inline')
  _$$GroupImplCopyWith<_$GroupImpl> get copyWith =>
      __$$GroupImplCopyWithImpl<_$GroupImpl>(this, _$identity);

  @override
  Map<String, dynamic> toJson() {
    return _$$GroupImplToJson(
      this,
    );
  }
}

abstract class _Group implements Group {
  const factory _Group(
          {required final String id,
          required final String name,
          @JsonKey(name: 'admin_id') required final String adminId,
          @JsonKey(name: 'invite_code') required final String inviteCode,
          required final List<String> categories,
          @JsonKey(name: 'telegram_username') final String? telegramUsername,
          @JsonKey(name: 'created_at') required final String createdAt}) =
      _$GroupImpl;

  factory _Group.fromJson(Map<String, dynamic> json) = _$GroupImpl.fromJson;

  @override
  String get id;
  @override
  String get name;
  @override
  @JsonKey(name: 'admin_id')
  String get adminId;
  @override
  @JsonKey(name: 'invite_code')
  String get inviteCode;
  @override
  List<String> get categories;
  @override
  @JsonKey(name: 'telegram_username')
  String? get telegramUsername;
  @override
  @JsonKey(name: 'created_at')
  String get createdAt;
  @override
  @JsonKey(ignore: true)
  _$$GroupImplCopyWith<_$GroupImpl> get copyWith =>
      throw _privateConstructorUsedError;
}

ActiveSeason _$ActiveSeasonFromJson(Map<String, dynamic> json) {
  return _ActiveSeason.fromJson(json);
}

/// @nodoc
mixin _$ActiveSeason {
  String get id => throw _privateConstructorUsedError;
  String get status => throw _privateConstructorUsedError;
  @JsonKey(name: 'reveal_at')
  String get revealAt => throw _privateConstructorUsedError;
  @JsonKey(name: 'voted_count')
  int get votedCount => throw _privateConstructorUsedError;
  @JsonKey(name: 'total_count')
  int get totalCount => throw _privateConstructorUsedError;
  @JsonKey(name: 'user_voted')
  bool get userVoted => throw _privateConstructorUsedError;

  Map<String, dynamic> toJson() => throw _privateConstructorUsedError;
  @JsonKey(ignore: true)
  $ActiveSeasonCopyWith<ActiveSeason> get copyWith =>
      throw _privateConstructorUsedError;
}

/// @nodoc
abstract class $ActiveSeasonCopyWith<$Res> {
  factory $ActiveSeasonCopyWith(
          ActiveSeason value, $Res Function(ActiveSeason) then) =
      _$ActiveSeasonCopyWithImpl<$Res, ActiveSeason>;
  @useResult
  $Res call(
      {String id,
      String status,
      @JsonKey(name: 'reveal_at') String revealAt,
      @JsonKey(name: 'voted_count') int votedCount,
      @JsonKey(name: 'total_count') int totalCount,
      @JsonKey(name: 'user_voted') bool userVoted});
}

/// @nodoc
class _$ActiveSeasonCopyWithImpl<$Res, $Val extends ActiveSeason>
    implements $ActiveSeasonCopyWith<$Res> {
  _$ActiveSeasonCopyWithImpl(this._value, this._then);

  // ignore: unused_field
  final $Val _value;
  // ignore: unused_field
  final $Res Function($Val) _then;

  @pragma('vm:prefer-inline')
  @override
  $Res call({
    Object? id = null,
    Object? status = null,
    Object? revealAt = null,
    Object? votedCount = null,
    Object? totalCount = null,
    Object? userVoted = null,
  }) {
    return _then(_value.copyWith(
      id: null == id
          ? _value.id
          : id // ignore: cast_nullable_to_non_nullable
              as String,
      status: null == status
          ? _value.status
          : status // ignore: cast_nullable_to_non_nullable
              as String,
      revealAt: null == revealAt
          ? _value.revealAt
          : revealAt // ignore: cast_nullable_to_non_nullable
              as String,
      votedCount: null == votedCount
          ? _value.votedCount
          : votedCount // ignore: cast_nullable_to_non_nullable
              as int,
      totalCount: null == totalCount
          ? _value.totalCount
          : totalCount // ignore: cast_nullable_to_non_nullable
              as int,
      userVoted: null == userVoted
          ? _value.userVoted
          : userVoted // ignore: cast_nullable_to_non_nullable
              as bool,
    ) as $Val);
  }
}

/// @nodoc
abstract class _$$ActiveSeasonImplCopyWith<$Res>
    implements $ActiveSeasonCopyWith<$Res> {
  factory _$$ActiveSeasonImplCopyWith(
          _$ActiveSeasonImpl value, $Res Function(_$ActiveSeasonImpl) then) =
      __$$ActiveSeasonImplCopyWithImpl<$Res>;
  @override
  @useResult
  $Res call(
      {String id,
      String status,
      @JsonKey(name: 'reveal_at') String revealAt,
      @JsonKey(name: 'voted_count') int votedCount,
      @JsonKey(name: 'total_count') int totalCount,
      @JsonKey(name: 'user_voted') bool userVoted});
}

/// @nodoc
class __$$ActiveSeasonImplCopyWithImpl<$Res>
    extends _$ActiveSeasonCopyWithImpl<$Res, _$ActiveSeasonImpl>
    implements _$$ActiveSeasonImplCopyWith<$Res> {
  __$$ActiveSeasonImplCopyWithImpl(
      _$ActiveSeasonImpl _value, $Res Function(_$ActiveSeasonImpl) _then)
      : super(_value, _then);

  @pragma('vm:prefer-inline')
  @override
  $Res call({
    Object? id = null,
    Object? status = null,
    Object? revealAt = null,
    Object? votedCount = null,
    Object? totalCount = null,
    Object? userVoted = null,
  }) {
    return _then(_$ActiveSeasonImpl(
      id: null == id
          ? _value.id
          : id // ignore: cast_nullable_to_non_nullable
              as String,
      status: null == status
          ? _value.status
          : status // ignore: cast_nullable_to_non_nullable
              as String,
      revealAt: null == revealAt
          ? _value.revealAt
          : revealAt // ignore: cast_nullable_to_non_nullable
              as String,
      votedCount: null == votedCount
          ? _value.votedCount
          : votedCount // ignore: cast_nullable_to_non_nullable
              as int,
      totalCount: null == totalCount
          ? _value.totalCount
          : totalCount // ignore: cast_nullable_to_non_nullable
              as int,
      userVoted: null == userVoted
          ? _value.userVoted
          : userVoted // ignore: cast_nullable_to_non_nullable
              as bool,
    ));
  }
}

/// @nodoc
@JsonSerializable()
class _$ActiveSeasonImpl implements _ActiveSeason {
  const _$ActiveSeasonImpl(
      {required this.id,
      required this.status,
      @JsonKey(name: 'reveal_at') required this.revealAt,
      @JsonKey(name: 'voted_count') required this.votedCount,
      @JsonKey(name: 'total_count') required this.totalCount,
      @JsonKey(name: 'user_voted') required this.userVoted});

  factory _$ActiveSeasonImpl.fromJson(Map<String, dynamic> json) =>
      _$$ActiveSeasonImplFromJson(json);

  @override
  final String id;
  @override
  final String status;
  @override
  @JsonKey(name: 'reveal_at')
  final String revealAt;
  @override
  @JsonKey(name: 'voted_count')
  final int votedCount;
  @override
  @JsonKey(name: 'total_count')
  final int totalCount;
  @override
  @JsonKey(name: 'user_voted')
  final bool userVoted;

  @override
  String toString() {
    return 'ActiveSeason(id: $id, status: $status, revealAt: $revealAt, votedCount: $votedCount, totalCount: $totalCount, userVoted: $userVoted)';
  }

  @override
  bool operator ==(Object other) {
    return identical(this, other) ||
        (other.runtimeType == runtimeType &&
            other is _$ActiveSeasonImpl &&
            (identical(other.id, id) || other.id == id) &&
            (identical(other.status, status) || other.status == status) &&
            (identical(other.revealAt, revealAt) ||
                other.revealAt == revealAt) &&
            (identical(other.votedCount, votedCount) ||
                other.votedCount == votedCount) &&
            (identical(other.totalCount, totalCount) ||
                other.totalCount == totalCount) &&
            (identical(other.userVoted, userVoted) ||
                other.userVoted == userVoted));
  }

  @JsonKey(ignore: true)
  @override
  int get hashCode => Object.hash(
      runtimeType, id, status, revealAt, votedCount, totalCount, userVoted);

  @JsonKey(ignore: true)
  @override
  @pragma('vm:prefer-inline')
  _$$ActiveSeasonImplCopyWith<_$ActiveSeasonImpl> get copyWith =>
      __$$ActiveSeasonImplCopyWithImpl<_$ActiveSeasonImpl>(this, _$identity);

  @override
  Map<String, dynamic> toJson() {
    return _$$ActiveSeasonImplToJson(
      this,
    );
  }
}

abstract class _ActiveSeason implements ActiveSeason {
  const factory _ActiveSeason(
          {required final String id,
          required final String status,
          @JsonKey(name: 'reveal_at') required final String revealAt,
          @JsonKey(name: 'voted_count') required final int votedCount,
          @JsonKey(name: 'total_count') required final int totalCount,
          @JsonKey(name: 'user_voted') required final bool userVoted}) =
      _$ActiveSeasonImpl;

  factory _ActiveSeason.fromJson(Map<String, dynamic> json) =
      _$ActiveSeasonImpl.fromJson;

  @override
  String get id;
  @override
  String get status;
  @override
  @JsonKey(name: 'reveal_at')
  String get revealAt;
  @override
  @JsonKey(name: 'voted_count')
  int get votedCount;
  @override
  @JsonKey(name: 'total_count')
  int get totalCount;
  @override
  @JsonKey(name: 'user_voted')
  bool get userVoted;
  @override
  @JsonKey(ignore: true)
  _$$ActiveSeasonImplCopyWith<_$ActiveSeasonImpl> get copyWith =>
      throw _privateConstructorUsedError;
}

GroupListItem _$GroupListItemFromJson(Map<String, dynamic> json) {
  return _GroupListItem.fromJson(json);
}

/// @nodoc
mixin _$GroupListItem {
  String get id => throw _privateConstructorUsedError;
  String get name => throw _privateConstructorUsedError;
  @JsonKey(name: 'member_count')
  int get memberCount => throw _privateConstructorUsedError;
  @JsonKey(name: 'invite_code')
  String get inviteCode => throw _privateConstructorUsedError;
  @JsonKey(name: 'telegram_username')
  String? get telegramUsername => throw _privateConstructorUsedError;
  @JsonKey(name: 'active_season')
  ActiveSeason? get activeSeason => throw _privateConstructorUsedError;

  Map<String, dynamic> toJson() => throw _privateConstructorUsedError;
  @JsonKey(ignore: true)
  $GroupListItemCopyWith<GroupListItem> get copyWith =>
      throw _privateConstructorUsedError;
}

/// @nodoc
abstract class $GroupListItemCopyWith<$Res> {
  factory $GroupListItemCopyWith(
          GroupListItem value, $Res Function(GroupListItem) then) =
      _$GroupListItemCopyWithImpl<$Res, GroupListItem>;
  @useResult
  $Res call(
      {String id,
      String name,
      @JsonKey(name: 'member_count') int memberCount,
      @JsonKey(name: 'invite_code') String inviteCode,
      @JsonKey(name: 'telegram_username') String? telegramUsername,
      @JsonKey(name: 'active_season') ActiveSeason? activeSeason});

  $ActiveSeasonCopyWith<$Res>? get activeSeason;
}

/// @nodoc
class _$GroupListItemCopyWithImpl<$Res, $Val extends GroupListItem>
    implements $GroupListItemCopyWith<$Res> {
  _$GroupListItemCopyWithImpl(this._value, this._then);

  // ignore: unused_field
  final $Val _value;
  // ignore: unused_field
  final $Res Function($Val) _then;

  @pragma('vm:prefer-inline')
  @override
  $Res call({
    Object? id = null,
    Object? name = null,
    Object? memberCount = null,
    Object? inviteCode = null,
    Object? telegramUsername = freezed,
    Object? activeSeason = freezed,
  }) {
    return _then(_value.copyWith(
      id: null == id
          ? _value.id
          : id // ignore: cast_nullable_to_non_nullable
              as String,
      name: null == name
          ? _value.name
          : name // ignore: cast_nullable_to_non_nullable
              as String,
      memberCount: null == memberCount
          ? _value.memberCount
          : memberCount // ignore: cast_nullable_to_non_nullable
              as int,
      inviteCode: null == inviteCode
          ? _value.inviteCode
          : inviteCode // ignore: cast_nullable_to_non_nullable
              as String,
      telegramUsername: freezed == telegramUsername
          ? _value.telegramUsername
          : telegramUsername // ignore: cast_nullable_to_non_nullable
              as String?,
      activeSeason: freezed == activeSeason
          ? _value.activeSeason
          : activeSeason // ignore: cast_nullable_to_non_nullable
              as ActiveSeason?,
    ) as $Val);
  }

  @override
  @pragma('vm:prefer-inline')
  $ActiveSeasonCopyWith<$Res>? get activeSeason {
    if (_value.activeSeason == null) {
      return null;
    }

    return $ActiveSeasonCopyWith<$Res>(_value.activeSeason!, (value) {
      return _then(_value.copyWith(activeSeason: value) as $Val);
    });
  }
}

/// @nodoc
abstract class _$$GroupListItemImplCopyWith<$Res>
    implements $GroupListItemCopyWith<$Res> {
  factory _$$GroupListItemImplCopyWith(
          _$GroupListItemImpl value, $Res Function(_$GroupListItemImpl) then) =
      __$$GroupListItemImplCopyWithImpl<$Res>;
  @override
  @useResult
  $Res call(
      {String id,
      String name,
      @JsonKey(name: 'member_count') int memberCount,
      @JsonKey(name: 'invite_code') String inviteCode,
      @JsonKey(name: 'telegram_username') String? telegramUsername,
      @JsonKey(name: 'active_season') ActiveSeason? activeSeason});

  @override
  $ActiveSeasonCopyWith<$Res>? get activeSeason;
}

/// @nodoc
class __$$GroupListItemImplCopyWithImpl<$Res>
    extends _$GroupListItemCopyWithImpl<$Res, _$GroupListItemImpl>
    implements _$$GroupListItemImplCopyWith<$Res> {
  __$$GroupListItemImplCopyWithImpl(
      _$GroupListItemImpl _value, $Res Function(_$GroupListItemImpl) _then)
      : super(_value, _then);

  @pragma('vm:prefer-inline')
  @override
  $Res call({
    Object? id = null,
    Object? name = null,
    Object? memberCount = null,
    Object? inviteCode = null,
    Object? telegramUsername = freezed,
    Object? activeSeason = freezed,
  }) {
    return _then(_$GroupListItemImpl(
      id: null == id
          ? _value.id
          : id // ignore: cast_nullable_to_non_nullable
              as String,
      name: null == name
          ? _value.name
          : name // ignore: cast_nullable_to_non_nullable
              as String,
      memberCount: null == memberCount
          ? _value.memberCount
          : memberCount // ignore: cast_nullable_to_non_nullable
              as int,
      inviteCode: null == inviteCode
          ? _value.inviteCode
          : inviteCode // ignore: cast_nullable_to_non_nullable
              as String,
      telegramUsername: freezed == telegramUsername
          ? _value.telegramUsername
          : telegramUsername // ignore: cast_nullable_to_non_nullable
              as String?,
      activeSeason: freezed == activeSeason
          ? _value.activeSeason
          : activeSeason // ignore: cast_nullable_to_non_nullable
              as ActiveSeason?,
    ));
  }
}

/// @nodoc
@JsonSerializable()
class _$GroupListItemImpl implements _GroupListItem {
  const _$GroupListItemImpl(
      {required this.id,
      required this.name,
      @JsonKey(name: 'member_count') required this.memberCount,
      @JsonKey(name: 'invite_code') required this.inviteCode,
      @JsonKey(name: 'telegram_username') this.telegramUsername,
      @JsonKey(name: 'active_season') this.activeSeason});

  factory _$GroupListItemImpl.fromJson(Map<String, dynamic> json) =>
      _$$GroupListItemImplFromJson(json);

  @override
  final String id;
  @override
  final String name;
  @override
  @JsonKey(name: 'member_count')
  final int memberCount;
  @override
  @JsonKey(name: 'invite_code')
  final String inviteCode;
  @override
  @JsonKey(name: 'telegram_username')
  final String? telegramUsername;
  @override
  @JsonKey(name: 'active_season')
  final ActiveSeason? activeSeason;

  @override
  String toString() {
    return 'GroupListItem(id: $id, name: $name, memberCount: $memberCount, inviteCode: $inviteCode, telegramUsername: $telegramUsername, activeSeason: $activeSeason)';
  }

  @override
  bool operator ==(Object other) {
    return identical(this, other) ||
        (other.runtimeType == runtimeType &&
            other is _$GroupListItemImpl &&
            (identical(other.id, id) || other.id == id) &&
            (identical(other.name, name) || other.name == name) &&
            (identical(other.memberCount, memberCount) ||
                other.memberCount == memberCount) &&
            (identical(other.inviteCode, inviteCode) ||
                other.inviteCode == inviteCode) &&
            (identical(other.telegramUsername, telegramUsername) ||
                other.telegramUsername == telegramUsername) &&
            (identical(other.activeSeason, activeSeason) ||
                other.activeSeason == activeSeason));
  }

  @JsonKey(ignore: true)
  @override
  int get hashCode => Object.hash(runtimeType, id, name, memberCount,
      inviteCode, telegramUsername, activeSeason);

  @JsonKey(ignore: true)
  @override
  @pragma('vm:prefer-inline')
  _$$GroupListItemImplCopyWith<_$GroupListItemImpl> get copyWith =>
      __$$GroupListItemImplCopyWithImpl<_$GroupListItemImpl>(this, _$identity);

  @override
  Map<String, dynamic> toJson() {
    return _$$GroupListItemImplToJson(
      this,
    );
  }
}

abstract class _GroupListItem implements GroupListItem {
  const factory _GroupListItem(
          {required final String id,
          required final String name,
          @JsonKey(name: 'member_count') required final int memberCount,
          @JsonKey(name: 'invite_code') required final String inviteCode,
          @JsonKey(name: 'telegram_username') final String? telegramUsername,
          @JsonKey(name: 'active_season') final ActiveSeason? activeSeason}) =
      _$GroupListItemImpl;

  factory _GroupListItem.fromJson(Map<String, dynamic> json) =
      _$GroupListItemImpl.fromJson;

  @override
  String get id;
  @override
  String get name;
  @override
  @JsonKey(name: 'member_count')
  int get memberCount;
  @override
  @JsonKey(name: 'invite_code')
  String get inviteCode;
  @override
  @JsonKey(name: 'telegram_username')
  String? get telegramUsername;
  @override
  @JsonKey(name: 'active_season')
  ActiveSeason? get activeSeason;
  @override
  @JsonKey(ignore: true)
  _$$GroupListItemImplCopyWith<_$GroupListItemImpl> get copyWith =>
      throw _privateConstructorUsedError;
}

Member _$MemberFromJson(Map<String, dynamic> json) {
  return _Member.fromJson(json);
}

/// @nodoc
mixin _$Member {
  String get id => throw _privateConstructorUsedError;
  String get username => throw _privateConstructorUsedError;
  @JsonKey(name: 'avatar_emoji')
  String? get avatarEmoji => throw _privateConstructorUsedError;
  @JsonKey(name: 'avatar_url')
  String? get avatarUrl => throw _privateConstructorUsedError;
  @JsonKey(name: 'is_admin')
  bool get isAdmin => throw _privateConstructorUsedError;

  Map<String, dynamic> toJson() => throw _privateConstructorUsedError;
  @JsonKey(ignore: true)
  $MemberCopyWith<Member> get copyWith => throw _privateConstructorUsedError;
}

/// @nodoc
abstract class $MemberCopyWith<$Res> {
  factory $MemberCopyWith(Member value, $Res Function(Member) then) =
      _$MemberCopyWithImpl<$Res, Member>;
  @useResult
  $Res call(
      {String id,
      String username,
      @JsonKey(name: 'avatar_emoji') String? avatarEmoji,
      @JsonKey(name: 'avatar_url') String? avatarUrl,
      @JsonKey(name: 'is_admin') bool isAdmin});
}

/// @nodoc
class _$MemberCopyWithImpl<$Res, $Val extends Member>
    implements $MemberCopyWith<$Res> {
  _$MemberCopyWithImpl(this._value, this._then);

  // ignore: unused_field
  final $Val _value;
  // ignore: unused_field
  final $Res Function($Val) _then;

  @pragma('vm:prefer-inline')
  @override
  $Res call({
    Object? id = null,
    Object? username = null,
    Object? avatarEmoji = freezed,
    Object? avatarUrl = freezed,
    Object? isAdmin = null,
  }) {
    return _then(_value.copyWith(
      id: null == id
          ? _value.id
          : id // ignore: cast_nullable_to_non_nullable
              as String,
      username: null == username
          ? _value.username
          : username // ignore: cast_nullable_to_non_nullable
              as String,
      avatarEmoji: freezed == avatarEmoji
          ? _value.avatarEmoji
          : avatarEmoji // ignore: cast_nullable_to_non_nullable
              as String?,
      avatarUrl: freezed == avatarUrl
          ? _value.avatarUrl
          : avatarUrl // ignore: cast_nullable_to_non_nullable
              as String?,
      isAdmin: null == isAdmin
          ? _value.isAdmin
          : isAdmin // ignore: cast_nullable_to_non_nullable
              as bool,
    ) as $Val);
  }
}

/// @nodoc
abstract class _$$MemberImplCopyWith<$Res> implements $MemberCopyWith<$Res> {
  factory _$$MemberImplCopyWith(
          _$MemberImpl value, $Res Function(_$MemberImpl) then) =
      __$$MemberImplCopyWithImpl<$Res>;
  @override
  @useResult
  $Res call(
      {String id,
      String username,
      @JsonKey(name: 'avatar_emoji') String? avatarEmoji,
      @JsonKey(name: 'avatar_url') String? avatarUrl,
      @JsonKey(name: 'is_admin') bool isAdmin});
}

/// @nodoc
class __$$MemberImplCopyWithImpl<$Res>
    extends _$MemberCopyWithImpl<$Res, _$MemberImpl>
    implements _$$MemberImplCopyWith<$Res> {
  __$$MemberImplCopyWithImpl(
      _$MemberImpl _value, $Res Function(_$MemberImpl) _then)
      : super(_value, _then);

  @pragma('vm:prefer-inline')
  @override
  $Res call({
    Object? id = null,
    Object? username = null,
    Object? avatarEmoji = freezed,
    Object? avatarUrl = freezed,
    Object? isAdmin = null,
  }) {
    return _then(_$MemberImpl(
      id: null == id
          ? _value.id
          : id // ignore: cast_nullable_to_non_nullable
              as String,
      username: null == username
          ? _value.username
          : username // ignore: cast_nullable_to_non_nullable
              as String,
      avatarEmoji: freezed == avatarEmoji
          ? _value.avatarEmoji
          : avatarEmoji // ignore: cast_nullable_to_non_nullable
              as String?,
      avatarUrl: freezed == avatarUrl
          ? _value.avatarUrl
          : avatarUrl // ignore: cast_nullable_to_non_nullable
              as String?,
      isAdmin: null == isAdmin
          ? _value.isAdmin
          : isAdmin // ignore: cast_nullable_to_non_nullable
              as bool,
    ));
  }
}

/// @nodoc
@JsonSerializable()
class _$MemberImpl implements _Member {
  const _$MemberImpl(
      {required this.id,
      required this.username,
      @JsonKey(name: 'avatar_emoji') this.avatarEmoji,
      @JsonKey(name: 'avatar_url') this.avatarUrl,
      @JsonKey(name: 'is_admin') required this.isAdmin});

  factory _$MemberImpl.fromJson(Map<String, dynamic> json) =>
      _$$MemberImplFromJson(json);

  @override
  final String id;
  @override
  final String username;
  @override
  @JsonKey(name: 'avatar_emoji')
  final String? avatarEmoji;
  @override
  @JsonKey(name: 'avatar_url')
  final String? avatarUrl;
  @override
  @JsonKey(name: 'is_admin')
  final bool isAdmin;

  @override
  String toString() {
    return 'Member(id: $id, username: $username, avatarEmoji: $avatarEmoji, avatarUrl: $avatarUrl, isAdmin: $isAdmin)';
  }

  @override
  bool operator ==(Object other) {
    return identical(this, other) ||
        (other.runtimeType == runtimeType &&
            other is _$MemberImpl &&
            (identical(other.id, id) || other.id == id) &&
            (identical(other.username, username) ||
                other.username == username) &&
            (identical(other.avatarEmoji, avatarEmoji) ||
                other.avatarEmoji == avatarEmoji) &&
            (identical(other.avatarUrl, avatarUrl) ||
                other.avatarUrl == avatarUrl) &&
            (identical(other.isAdmin, isAdmin) || other.isAdmin == isAdmin));
  }

  @JsonKey(ignore: true)
  @override
  int get hashCode =>
      Object.hash(runtimeType, id, username, avatarEmoji, avatarUrl, isAdmin);

  @JsonKey(ignore: true)
  @override
  @pragma('vm:prefer-inline')
  _$$MemberImplCopyWith<_$MemberImpl> get copyWith =>
      __$$MemberImplCopyWithImpl<_$MemberImpl>(this, _$identity);

  @override
  Map<String, dynamic> toJson() {
    return _$$MemberImplToJson(
      this,
    );
  }
}

abstract class _Member implements Member {
  const factory _Member(
      {required final String id,
      required final String username,
      @JsonKey(name: 'avatar_emoji') final String? avatarEmoji,
      @JsonKey(name: 'avatar_url') final String? avatarUrl,
      @JsonKey(name: 'is_admin') required final bool isAdmin}) = _$MemberImpl;

  factory _Member.fromJson(Map<String, dynamic> json) = _$MemberImpl.fromJson;

  @override
  String get id;
  @override
  String get username;
  @override
  @JsonKey(name: 'avatar_emoji')
  String? get avatarEmoji;
  @override
  @JsonKey(name: 'avatar_url')
  String? get avatarUrl;
  @override
  @JsonKey(name: 'is_admin')
  bool get isAdmin;
  @override
  @JsonKey(ignore: true)
  _$$MemberImplCopyWith<_$MemberImpl> get copyWith =>
      throw _privateConstructorUsedError;
}

GroupDetail _$GroupDetailFromJson(Map<String, dynamic> json) {
  return _GroupDetail.fromJson(json);
}

/// @nodoc
mixin _$GroupDetail {
  Group get group => throw _privateConstructorUsedError;
  List<Member> get members => throw _privateConstructorUsedError;
  @JsonKey(name: 'active_season')
  ActiveSeason? get activeSeason => throw _privateConstructorUsedError;

  Map<String, dynamic> toJson() => throw _privateConstructorUsedError;
  @JsonKey(ignore: true)
  $GroupDetailCopyWith<GroupDetail> get copyWith =>
      throw _privateConstructorUsedError;
}

/// @nodoc
abstract class $GroupDetailCopyWith<$Res> {
  factory $GroupDetailCopyWith(
          GroupDetail value, $Res Function(GroupDetail) then) =
      _$GroupDetailCopyWithImpl<$Res, GroupDetail>;
  @useResult
  $Res call(
      {Group group,
      List<Member> members,
      @JsonKey(name: 'active_season') ActiveSeason? activeSeason});

  $GroupCopyWith<$Res> get group;
  $ActiveSeasonCopyWith<$Res>? get activeSeason;
}

/// @nodoc
class _$GroupDetailCopyWithImpl<$Res, $Val extends GroupDetail>
    implements $GroupDetailCopyWith<$Res> {
  _$GroupDetailCopyWithImpl(this._value, this._then);

  // ignore: unused_field
  final $Val _value;
  // ignore: unused_field
  final $Res Function($Val) _then;

  @pragma('vm:prefer-inline')
  @override
  $Res call({
    Object? group = null,
    Object? members = null,
    Object? activeSeason = freezed,
  }) {
    return _then(_value.copyWith(
      group: null == group
          ? _value.group
          : group // ignore: cast_nullable_to_non_nullable
              as Group,
      members: null == members
          ? _value.members
          : members // ignore: cast_nullable_to_non_nullable
              as List<Member>,
      activeSeason: freezed == activeSeason
          ? _value.activeSeason
          : activeSeason // ignore: cast_nullable_to_non_nullable
              as ActiveSeason?,
    ) as $Val);
  }

  @override
  @pragma('vm:prefer-inline')
  $GroupCopyWith<$Res> get group {
    return $GroupCopyWith<$Res>(_value.group, (value) {
      return _then(_value.copyWith(group: value) as $Val);
    });
  }

  @override
  @pragma('vm:prefer-inline')
  $ActiveSeasonCopyWith<$Res>? get activeSeason {
    if (_value.activeSeason == null) {
      return null;
    }

    return $ActiveSeasonCopyWith<$Res>(_value.activeSeason!, (value) {
      return _then(_value.copyWith(activeSeason: value) as $Val);
    });
  }
}

/// @nodoc
abstract class _$$GroupDetailImplCopyWith<$Res>
    implements $GroupDetailCopyWith<$Res> {
  factory _$$GroupDetailImplCopyWith(
          _$GroupDetailImpl value, $Res Function(_$GroupDetailImpl) then) =
      __$$GroupDetailImplCopyWithImpl<$Res>;
  @override
  @useResult
  $Res call(
      {Group group,
      List<Member> members,
      @JsonKey(name: 'active_season') ActiveSeason? activeSeason});

  @override
  $GroupCopyWith<$Res> get group;
  @override
  $ActiveSeasonCopyWith<$Res>? get activeSeason;
}

/// @nodoc
class __$$GroupDetailImplCopyWithImpl<$Res>
    extends _$GroupDetailCopyWithImpl<$Res, _$GroupDetailImpl>
    implements _$$GroupDetailImplCopyWith<$Res> {
  __$$GroupDetailImplCopyWithImpl(
      _$GroupDetailImpl _value, $Res Function(_$GroupDetailImpl) _then)
      : super(_value, _then);

  @pragma('vm:prefer-inline')
  @override
  $Res call({
    Object? group = null,
    Object? members = null,
    Object? activeSeason = freezed,
  }) {
    return _then(_$GroupDetailImpl(
      group: null == group
          ? _value.group
          : group // ignore: cast_nullable_to_non_nullable
              as Group,
      members: null == members
          ? _value._members
          : members // ignore: cast_nullable_to_non_nullable
              as List<Member>,
      activeSeason: freezed == activeSeason
          ? _value.activeSeason
          : activeSeason // ignore: cast_nullable_to_non_nullable
              as ActiveSeason?,
    ));
  }
}

/// @nodoc
@JsonSerializable()
class _$GroupDetailImpl implements _GroupDetail {
  const _$GroupDetailImpl(
      {required this.group,
      required final List<Member> members,
      @JsonKey(name: 'active_season') this.activeSeason})
      : _members = members;

  factory _$GroupDetailImpl.fromJson(Map<String, dynamic> json) =>
      _$$GroupDetailImplFromJson(json);

  @override
  final Group group;
  final List<Member> _members;
  @override
  List<Member> get members {
    if (_members is EqualUnmodifiableListView) return _members;
    // ignore: implicit_dynamic_type
    return EqualUnmodifiableListView(_members);
  }

  @override
  @JsonKey(name: 'active_season')
  final ActiveSeason? activeSeason;

  @override
  String toString() {
    return 'GroupDetail(group: $group, members: $members, activeSeason: $activeSeason)';
  }

  @override
  bool operator ==(Object other) {
    return identical(this, other) ||
        (other.runtimeType == runtimeType &&
            other is _$GroupDetailImpl &&
            (identical(other.group, group) || other.group == group) &&
            const DeepCollectionEquality().equals(other._members, _members) &&
            (identical(other.activeSeason, activeSeason) ||
                other.activeSeason == activeSeason));
  }

  @JsonKey(ignore: true)
  @override
  int get hashCode => Object.hash(runtimeType, group,
      const DeepCollectionEquality().hash(_members), activeSeason);

  @JsonKey(ignore: true)
  @override
  @pragma('vm:prefer-inline')
  _$$GroupDetailImplCopyWith<_$GroupDetailImpl> get copyWith =>
      __$$GroupDetailImplCopyWithImpl<_$GroupDetailImpl>(this, _$identity);

  @override
  Map<String, dynamic> toJson() {
    return _$$GroupDetailImplToJson(
      this,
    );
  }
}

abstract class _GroupDetail implements GroupDetail {
  const factory _GroupDetail(
          {required final Group group,
          required final List<Member> members,
          @JsonKey(name: 'active_season') final ActiveSeason? activeSeason}) =
      _$GroupDetailImpl;

  factory _GroupDetail.fromJson(Map<String, dynamic> json) =
      _$GroupDetailImpl.fromJson;

  @override
  Group get group;
  @override
  List<Member> get members;
  @override
  @JsonKey(name: 'active_season')
  ActiveSeason? get activeSeason;
  @override
  @JsonKey(ignore: true)
  _$$GroupDetailImplCopyWith<_$GroupDetailImpl> get copyWith =>
      throw _privateConstructorUsedError;
}

JoinPreview _$JoinPreviewFromJson(Map<String, dynamic> json) {
  return _JoinPreview.fromJson(json);
}

/// @nodoc
mixin _$JoinPreview {
  String get name => throw _privateConstructorUsedError;
  @JsonKey(name: 'member_count')
  int get memberCount => throw _privateConstructorUsedError;
  @JsonKey(name: 'admin_username')
  String get adminUsername => throw _privateConstructorUsedError;

  Map<String, dynamic> toJson() => throw _privateConstructorUsedError;
  @JsonKey(ignore: true)
  $JoinPreviewCopyWith<JoinPreview> get copyWith =>
      throw _privateConstructorUsedError;
}

/// @nodoc
abstract class $JoinPreviewCopyWith<$Res> {
  factory $JoinPreviewCopyWith(
          JoinPreview value, $Res Function(JoinPreview) then) =
      _$JoinPreviewCopyWithImpl<$Res, JoinPreview>;
  @useResult
  $Res call(
      {String name,
      @JsonKey(name: 'member_count') int memberCount,
      @JsonKey(name: 'admin_username') String adminUsername});
}

/// @nodoc
class _$JoinPreviewCopyWithImpl<$Res, $Val extends JoinPreview>
    implements $JoinPreviewCopyWith<$Res> {
  _$JoinPreviewCopyWithImpl(this._value, this._then);

  // ignore: unused_field
  final $Val _value;
  // ignore: unused_field
  final $Res Function($Val) _then;

  @pragma('vm:prefer-inline')
  @override
  $Res call({
    Object? name = null,
    Object? memberCount = null,
    Object? adminUsername = null,
  }) {
    return _then(_value.copyWith(
      name: null == name
          ? _value.name
          : name // ignore: cast_nullable_to_non_nullable
              as String,
      memberCount: null == memberCount
          ? _value.memberCount
          : memberCount // ignore: cast_nullable_to_non_nullable
              as int,
      adminUsername: null == adminUsername
          ? _value.adminUsername
          : adminUsername // ignore: cast_nullable_to_non_nullable
              as String,
    ) as $Val);
  }
}

/// @nodoc
abstract class _$$JoinPreviewImplCopyWith<$Res>
    implements $JoinPreviewCopyWith<$Res> {
  factory _$$JoinPreviewImplCopyWith(
          _$JoinPreviewImpl value, $Res Function(_$JoinPreviewImpl) then) =
      __$$JoinPreviewImplCopyWithImpl<$Res>;
  @override
  @useResult
  $Res call(
      {String name,
      @JsonKey(name: 'member_count') int memberCount,
      @JsonKey(name: 'admin_username') String adminUsername});
}

/// @nodoc
class __$$JoinPreviewImplCopyWithImpl<$Res>
    extends _$JoinPreviewCopyWithImpl<$Res, _$JoinPreviewImpl>
    implements _$$JoinPreviewImplCopyWith<$Res> {
  __$$JoinPreviewImplCopyWithImpl(
      _$JoinPreviewImpl _value, $Res Function(_$JoinPreviewImpl) _then)
      : super(_value, _then);

  @pragma('vm:prefer-inline')
  @override
  $Res call({
    Object? name = null,
    Object? memberCount = null,
    Object? adminUsername = null,
  }) {
    return _then(_$JoinPreviewImpl(
      name: null == name
          ? _value.name
          : name // ignore: cast_nullable_to_non_nullable
              as String,
      memberCount: null == memberCount
          ? _value.memberCount
          : memberCount // ignore: cast_nullable_to_non_nullable
              as int,
      adminUsername: null == adminUsername
          ? _value.adminUsername
          : adminUsername // ignore: cast_nullable_to_non_nullable
              as String,
    ));
  }
}

/// @nodoc
@JsonSerializable()
class _$JoinPreviewImpl implements _JoinPreview {
  const _$JoinPreviewImpl(
      {required this.name,
      @JsonKey(name: 'member_count') required this.memberCount,
      @JsonKey(name: 'admin_username') required this.adminUsername});

  factory _$JoinPreviewImpl.fromJson(Map<String, dynamic> json) =>
      _$$JoinPreviewImplFromJson(json);

  @override
  final String name;
  @override
  @JsonKey(name: 'member_count')
  final int memberCount;
  @override
  @JsonKey(name: 'admin_username')
  final String adminUsername;

  @override
  String toString() {
    return 'JoinPreview(name: $name, memberCount: $memberCount, adminUsername: $adminUsername)';
  }

  @override
  bool operator ==(Object other) {
    return identical(this, other) ||
        (other.runtimeType == runtimeType &&
            other is _$JoinPreviewImpl &&
            (identical(other.name, name) || other.name == name) &&
            (identical(other.memberCount, memberCount) ||
                other.memberCount == memberCount) &&
            (identical(other.adminUsername, adminUsername) ||
                other.adminUsername == adminUsername));
  }

  @JsonKey(ignore: true)
  @override
  int get hashCode =>
      Object.hash(runtimeType, name, memberCount, adminUsername);

  @JsonKey(ignore: true)
  @override
  @pragma('vm:prefer-inline')
  _$$JoinPreviewImplCopyWith<_$JoinPreviewImpl> get copyWith =>
      __$$JoinPreviewImplCopyWithImpl<_$JoinPreviewImpl>(this, _$identity);

  @override
  Map<String, dynamic> toJson() {
    return _$$JoinPreviewImplToJson(
      this,
    );
  }
}

abstract class _JoinPreview implements JoinPreview {
  const factory _JoinPreview(
      {required final String name,
      @JsonKey(name: 'member_count') required final int memberCount,
      @JsonKey(name: 'admin_username')
      required final String adminUsername}) = _$JoinPreviewImpl;

  factory _JoinPreview.fromJson(Map<String, dynamic> json) =
      _$JoinPreviewImpl.fromJson;

  @override
  String get name;
  @override
  @JsonKey(name: 'member_count')
  int get memberCount;
  @override
  @JsonKey(name: 'admin_username')
  String get adminUsername;
  @override
  @JsonKey(ignore: true)
  _$$JoinPreviewImplCopyWith<_$JoinPreviewImpl> get copyWith =>
      throw _privateConstructorUsedError;
}
