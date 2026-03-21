// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'user.dart';

// **************************************************************************
// FreezedGenerator
// **************************************************************************

T _$identity<T>(T value) => value;

final _privateConstructorUsedError = UnsupportedError(
    'It seems like you constructed your class using `MyClass._()`. This constructor is only meant to be used by freezed and you are not supposed to need it nor use it.\nPlease check the documentation here for more information: https://github.com/rrousselGit/freezed#adding-getters-and-methods-to-our-models');

User _$UserFromJson(Map<String, dynamic> json) {
  return _User.fromJson(json);
}

/// @nodoc
mixin _$User {
  String get id => throw _privateConstructorUsedError;
  String get username => throw _privateConstructorUsedError;
  @JsonKey(name: 'avatar_url')
  String? get avatarUrl => throw _privateConstructorUsedError;
  @JsonKey(name: 'avatar_emoji')
  String? get avatarEmoji => throw _privateConstructorUsedError;
  @JsonKey(name: 'birth_year')
  int? get birthYear => throw _privateConstructorUsedError;
  @JsonKey(name: 'created_at')
  String get createdAt => throw _privateConstructorUsedError;

  Map<String, dynamic> toJson() => throw _privateConstructorUsedError;
  @JsonKey(ignore: true)
  $UserCopyWith<User> get copyWith => throw _privateConstructorUsedError;
}

/// @nodoc
abstract class $UserCopyWith<$Res> {
  factory $UserCopyWith(User value, $Res Function(User) then) =
      _$UserCopyWithImpl<$Res, User>;
  @useResult
  $Res call(
      {String id,
      String username,
      @JsonKey(name: 'avatar_url') String? avatarUrl,
      @JsonKey(name: 'avatar_emoji') String? avatarEmoji,
      @JsonKey(name: 'birth_year') int? birthYear,
      @JsonKey(name: 'created_at') String createdAt});
}

/// @nodoc
class _$UserCopyWithImpl<$Res, $Val extends User>
    implements $UserCopyWith<$Res> {
  _$UserCopyWithImpl(this._value, this._then);

  // ignore: unused_field
  final $Val _value;
  // ignore: unused_field
  final $Res Function($Val) _then;

  @pragma('vm:prefer-inline')
  @override
  $Res call({
    Object? id = null,
    Object? username = null,
    Object? avatarUrl = freezed,
    Object? avatarEmoji = freezed,
    Object? birthYear = freezed,
    Object? createdAt = null,
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
      avatarUrl: freezed == avatarUrl
          ? _value.avatarUrl
          : avatarUrl // ignore: cast_nullable_to_non_nullable
              as String?,
      avatarEmoji: freezed == avatarEmoji
          ? _value.avatarEmoji
          : avatarEmoji // ignore: cast_nullable_to_non_nullable
              as String?,
      birthYear: freezed == birthYear
          ? _value.birthYear
          : birthYear // ignore: cast_nullable_to_non_nullable
              as int?,
      createdAt: null == createdAt
          ? _value.createdAt
          : createdAt // ignore: cast_nullable_to_non_nullable
              as String,
    ) as $Val);
  }
}

/// @nodoc
abstract class _$$UserImplCopyWith<$Res> implements $UserCopyWith<$Res> {
  factory _$$UserImplCopyWith(
          _$UserImpl value, $Res Function(_$UserImpl) then) =
      __$$UserImplCopyWithImpl<$Res>;
  @override
  @useResult
  $Res call(
      {String id,
      String username,
      @JsonKey(name: 'avatar_url') String? avatarUrl,
      @JsonKey(name: 'avatar_emoji') String? avatarEmoji,
      @JsonKey(name: 'birth_year') int? birthYear,
      @JsonKey(name: 'created_at') String createdAt});
}

/// @nodoc
class __$$UserImplCopyWithImpl<$Res>
    extends _$UserCopyWithImpl<$Res, _$UserImpl>
    implements _$$UserImplCopyWith<$Res> {
  __$$UserImplCopyWithImpl(_$UserImpl _value, $Res Function(_$UserImpl) _then)
      : super(_value, _then);

  @pragma('vm:prefer-inline')
  @override
  $Res call({
    Object? id = null,
    Object? username = null,
    Object? avatarUrl = freezed,
    Object? avatarEmoji = freezed,
    Object? birthYear = freezed,
    Object? createdAt = null,
  }) {
    return _then(_$UserImpl(
      id: null == id
          ? _value.id
          : id // ignore: cast_nullable_to_non_nullable
              as String,
      username: null == username
          ? _value.username
          : username // ignore: cast_nullable_to_non_nullable
              as String,
      avatarUrl: freezed == avatarUrl
          ? _value.avatarUrl
          : avatarUrl // ignore: cast_nullable_to_non_nullable
              as String?,
      avatarEmoji: freezed == avatarEmoji
          ? _value.avatarEmoji
          : avatarEmoji // ignore: cast_nullable_to_non_nullable
              as String?,
      birthYear: freezed == birthYear
          ? _value.birthYear
          : birthYear // ignore: cast_nullable_to_non_nullable
              as int?,
      createdAt: null == createdAt
          ? _value.createdAt
          : createdAt // ignore: cast_nullable_to_non_nullable
              as String,
    ));
  }
}

/// @nodoc
@JsonSerializable()
class _$UserImpl implements _User {
  const _$UserImpl(
      {required this.id,
      required this.username,
      @JsonKey(name: 'avatar_url') this.avatarUrl,
      @JsonKey(name: 'avatar_emoji') this.avatarEmoji,
      @JsonKey(name: 'birth_year') this.birthYear,
      @JsonKey(name: 'created_at') required this.createdAt});

  factory _$UserImpl.fromJson(Map<String, dynamic> json) =>
      _$$UserImplFromJson(json);

  @override
  final String id;
  @override
  final String username;
  @override
  @JsonKey(name: 'avatar_url')
  final String? avatarUrl;
  @override
  @JsonKey(name: 'avatar_emoji')
  final String? avatarEmoji;
  @override
  @JsonKey(name: 'birth_year')
  final int? birthYear;
  @override
  @JsonKey(name: 'created_at')
  final String createdAt;

  @override
  String toString() {
    return 'User(id: $id, username: $username, avatarUrl: $avatarUrl, avatarEmoji: $avatarEmoji, birthYear: $birthYear, createdAt: $createdAt)';
  }

  @override
  bool operator ==(Object other) {
    return identical(this, other) ||
        (other.runtimeType == runtimeType &&
            other is _$UserImpl &&
            (identical(other.id, id) || other.id == id) &&
            (identical(other.username, username) ||
                other.username == username) &&
            (identical(other.avatarUrl, avatarUrl) ||
                other.avatarUrl == avatarUrl) &&
            (identical(other.avatarEmoji, avatarEmoji) ||
                other.avatarEmoji == avatarEmoji) &&
            (identical(other.birthYear, birthYear) ||
                other.birthYear == birthYear) &&
            (identical(other.createdAt, createdAt) ||
                other.createdAt == createdAt));
  }

  @JsonKey(ignore: true)
  @override
  int get hashCode => Object.hash(
      runtimeType, id, username, avatarUrl, avatarEmoji, birthYear, createdAt);

  @JsonKey(ignore: true)
  @override
  @pragma('vm:prefer-inline')
  _$$UserImplCopyWith<_$UserImpl> get copyWith =>
      __$$UserImplCopyWithImpl<_$UserImpl>(this, _$identity);

  @override
  Map<String, dynamic> toJson() {
    return _$$UserImplToJson(
      this,
    );
  }
}

abstract class _User implements User {
  const factory _User(
          {required final String id,
          required final String username,
          @JsonKey(name: 'avatar_url') final String? avatarUrl,
          @JsonKey(name: 'avatar_emoji') final String? avatarEmoji,
          @JsonKey(name: 'birth_year') final int? birthYear,
          @JsonKey(name: 'created_at') required final String createdAt}) =
      _$UserImpl;

  factory _User.fromJson(Map<String, dynamic> json) = _$UserImpl.fromJson;

  @override
  String get id;
  @override
  String get username;
  @override
  @JsonKey(name: 'avatar_url')
  String? get avatarUrl;
  @override
  @JsonKey(name: 'avatar_emoji')
  String? get avatarEmoji;
  @override
  @JsonKey(name: 'birth_year')
  int? get birthYear;
  @override
  @JsonKey(name: 'created_at')
  String get createdAt;
  @override
  @JsonKey(ignore: true)
  _$$UserImplCopyWith<_$UserImpl> get copyWith =>
      throw _privateConstructorUsedError;
}
