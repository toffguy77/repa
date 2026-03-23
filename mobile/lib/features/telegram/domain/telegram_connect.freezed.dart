// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'telegram_connect.dart';

// **************************************************************************
// FreezedGenerator
// **************************************************************************

T _$identity<T>(T value) => value;

final _privateConstructorUsedError = UnsupportedError(
    'It seems like you constructed your class using `MyClass._()`. This constructor is only meant to be used by freezed and you are not supposed to need it nor use it.\nPlease check the documentation here for more information: https://github.com/rrousselGit/freezed#adding-getters-and-methods-to-our-models');

TelegramConnectCode _$TelegramConnectCodeFromJson(Map<String, dynamic> json) {
  return _TelegramConnectCode.fromJson(json);
}

/// @nodoc
mixin _$TelegramConnectCode {
  @JsonKey(name: 'connect_code')
  String get connectCode => throw _privateConstructorUsedError;
  String get instruction => throw _privateConstructorUsedError;
  @JsonKey(name: 'expires_at')
  String get expiresAt => throw _privateConstructorUsedError;

  Map<String, dynamic> toJson() => throw _privateConstructorUsedError;
  @JsonKey(ignore: true)
  $TelegramConnectCodeCopyWith<TelegramConnectCode> get copyWith =>
      throw _privateConstructorUsedError;
}

/// @nodoc
abstract class $TelegramConnectCodeCopyWith<$Res> {
  factory $TelegramConnectCodeCopyWith(
          TelegramConnectCode value, $Res Function(TelegramConnectCode) then) =
      _$TelegramConnectCodeCopyWithImpl<$Res, TelegramConnectCode>;
  @useResult
  $Res call(
      {@JsonKey(name: 'connect_code') String connectCode,
      String instruction,
      @JsonKey(name: 'expires_at') String expiresAt});
}

/// @nodoc
class _$TelegramConnectCodeCopyWithImpl<$Res, $Val extends TelegramConnectCode>
    implements $TelegramConnectCodeCopyWith<$Res> {
  _$TelegramConnectCodeCopyWithImpl(this._value, this._then);

  // ignore: unused_field
  final $Val _value;
  // ignore: unused_field
  final $Res Function($Val) _then;

  @pragma('vm:prefer-inline')
  @override
  $Res call({
    Object? connectCode = null,
    Object? instruction = null,
    Object? expiresAt = null,
  }) {
    return _then(_value.copyWith(
      connectCode: null == connectCode
          ? _value.connectCode
          : connectCode // ignore: cast_nullable_to_non_nullable
              as String,
      instruction: null == instruction
          ? _value.instruction
          : instruction // ignore: cast_nullable_to_non_nullable
              as String,
      expiresAt: null == expiresAt
          ? _value.expiresAt
          : expiresAt // ignore: cast_nullable_to_non_nullable
              as String,
    ) as $Val);
  }
}

/// @nodoc
abstract class _$$TelegramConnectCodeImplCopyWith<$Res>
    implements $TelegramConnectCodeCopyWith<$Res> {
  factory _$$TelegramConnectCodeImplCopyWith(_$TelegramConnectCodeImpl value,
          $Res Function(_$TelegramConnectCodeImpl) then) =
      __$$TelegramConnectCodeImplCopyWithImpl<$Res>;
  @override
  @useResult
  $Res call(
      {@JsonKey(name: 'connect_code') String connectCode,
      String instruction,
      @JsonKey(name: 'expires_at') String expiresAt});
}

/// @nodoc
class __$$TelegramConnectCodeImplCopyWithImpl<$Res>
    extends _$TelegramConnectCodeCopyWithImpl<$Res, _$TelegramConnectCodeImpl>
    implements _$$TelegramConnectCodeImplCopyWith<$Res> {
  __$$TelegramConnectCodeImplCopyWithImpl(_$TelegramConnectCodeImpl _value,
      $Res Function(_$TelegramConnectCodeImpl) _then)
      : super(_value, _then);

  @pragma('vm:prefer-inline')
  @override
  $Res call({
    Object? connectCode = null,
    Object? instruction = null,
    Object? expiresAt = null,
  }) {
    return _then(_$TelegramConnectCodeImpl(
      connectCode: null == connectCode
          ? _value.connectCode
          : connectCode // ignore: cast_nullable_to_non_nullable
              as String,
      instruction: null == instruction
          ? _value.instruction
          : instruction // ignore: cast_nullable_to_non_nullable
              as String,
      expiresAt: null == expiresAt
          ? _value.expiresAt
          : expiresAt // ignore: cast_nullable_to_non_nullable
              as String,
    ));
  }
}

/// @nodoc
@JsonSerializable()
class _$TelegramConnectCodeImpl implements _TelegramConnectCode {
  const _$TelegramConnectCodeImpl(
      {@JsonKey(name: 'connect_code') required this.connectCode,
      required this.instruction,
      @JsonKey(name: 'expires_at') required this.expiresAt});

  factory _$TelegramConnectCodeImpl.fromJson(Map<String, dynamic> json) =>
      _$$TelegramConnectCodeImplFromJson(json);

  @override
  @JsonKey(name: 'connect_code')
  final String connectCode;
  @override
  final String instruction;
  @override
  @JsonKey(name: 'expires_at')
  final String expiresAt;

  @override
  String toString() {
    return 'TelegramConnectCode(connectCode: $connectCode, instruction: $instruction, expiresAt: $expiresAt)';
  }

  @override
  bool operator ==(Object other) {
    return identical(this, other) ||
        (other.runtimeType == runtimeType &&
            other is _$TelegramConnectCodeImpl &&
            (identical(other.connectCode, connectCode) ||
                other.connectCode == connectCode) &&
            (identical(other.instruction, instruction) ||
                other.instruction == instruction) &&
            (identical(other.expiresAt, expiresAt) ||
                other.expiresAt == expiresAt));
  }

  @JsonKey(ignore: true)
  @override
  int get hashCode =>
      Object.hash(runtimeType, connectCode, instruction, expiresAt);

  @JsonKey(ignore: true)
  @override
  @pragma('vm:prefer-inline')
  _$$TelegramConnectCodeImplCopyWith<_$TelegramConnectCodeImpl> get copyWith =>
      __$$TelegramConnectCodeImplCopyWithImpl<_$TelegramConnectCodeImpl>(
          this, _$identity);

  @override
  Map<String, dynamic> toJson() {
    return _$$TelegramConnectCodeImplToJson(
      this,
    );
  }
}

abstract class _TelegramConnectCode implements TelegramConnectCode {
  const factory _TelegramConnectCode(
          {@JsonKey(name: 'connect_code') required final String connectCode,
          required final String instruction,
          @JsonKey(name: 'expires_at') required final String expiresAt}) =
      _$TelegramConnectCodeImpl;

  factory _TelegramConnectCode.fromJson(Map<String, dynamic> json) =
      _$TelegramConnectCodeImpl.fromJson;

  @override
  @JsonKey(name: 'connect_code')
  String get connectCode;
  @override
  String get instruction;
  @override
  @JsonKey(name: 'expires_at')
  String get expiresAt;
  @override
  @JsonKey(ignore: true)
  _$$TelegramConnectCodeImplCopyWith<_$TelegramConnectCodeImpl> get copyWith =>
      throw _privateConstructorUsedError;
}
