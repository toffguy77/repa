// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'crystals.dart';

// **************************************************************************
// FreezedGenerator
// **************************************************************************

T _$identity<T>(T value) => value;

final _privateConstructorUsedError = UnsupportedError(
    'It seems like you constructed your class using `MyClass._()`. This constructor is only meant to be used by freezed and you are not supposed to need it nor use it.\nPlease check the documentation here for more information: https://github.com/rrousselGit/freezed#adding-getters-and-methods-to-our-models');

CrystalPackage _$CrystalPackageFromJson(Map<String, dynamic> json) {
  return _CrystalPackage.fromJson(json);
}

/// @nodoc
mixin _$CrystalPackage {
  String get id => throw _privateConstructorUsedError;
  int get crystals => throw _privateConstructorUsedError;
  int get bonus => throw _privateConstructorUsedError;
  @JsonKey(name: 'price_kopecks')
  int get priceKopecks => throw _privateConstructorUsedError;

  Map<String, dynamic> toJson() => throw _privateConstructorUsedError;
  @JsonKey(ignore: true)
  $CrystalPackageCopyWith<CrystalPackage> get copyWith =>
      throw _privateConstructorUsedError;
}

/// @nodoc
abstract class $CrystalPackageCopyWith<$Res> {
  factory $CrystalPackageCopyWith(
          CrystalPackage value, $Res Function(CrystalPackage) then) =
      _$CrystalPackageCopyWithImpl<$Res, CrystalPackage>;
  @useResult
  $Res call(
      {String id,
      int crystals,
      int bonus,
      @JsonKey(name: 'price_kopecks') int priceKopecks});
}

/// @nodoc
class _$CrystalPackageCopyWithImpl<$Res, $Val extends CrystalPackage>
    implements $CrystalPackageCopyWith<$Res> {
  _$CrystalPackageCopyWithImpl(this._value, this._then);

  // ignore: unused_field
  final $Val _value;
  // ignore: unused_field
  final $Res Function($Val) _then;

  @pragma('vm:prefer-inline')
  @override
  $Res call({
    Object? id = null,
    Object? crystals = null,
    Object? bonus = null,
    Object? priceKopecks = null,
  }) {
    return _then(_value.copyWith(
      id: null == id
          ? _value.id
          : id // ignore: cast_nullable_to_non_nullable
              as String,
      crystals: null == crystals
          ? _value.crystals
          : crystals // ignore: cast_nullable_to_non_nullable
              as int,
      bonus: null == bonus
          ? _value.bonus
          : bonus // ignore: cast_nullable_to_non_nullable
              as int,
      priceKopecks: null == priceKopecks
          ? _value.priceKopecks
          : priceKopecks // ignore: cast_nullable_to_non_nullable
              as int,
    ) as $Val);
  }
}

/// @nodoc
abstract class _$$CrystalPackageImplCopyWith<$Res>
    implements $CrystalPackageCopyWith<$Res> {
  factory _$$CrystalPackageImplCopyWith(_$CrystalPackageImpl value,
          $Res Function(_$CrystalPackageImpl) then) =
      __$$CrystalPackageImplCopyWithImpl<$Res>;
  @override
  @useResult
  $Res call(
      {String id,
      int crystals,
      int bonus,
      @JsonKey(name: 'price_kopecks') int priceKopecks});
}

/// @nodoc
class __$$CrystalPackageImplCopyWithImpl<$Res>
    extends _$CrystalPackageCopyWithImpl<$Res, _$CrystalPackageImpl>
    implements _$$CrystalPackageImplCopyWith<$Res> {
  __$$CrystalPackageImplCopyWithImpl(
      _$CrystalPackageImpl _value, $Res Function(_$CrystalPackageImpl) _then)
      : super(_value, _then);

  @pragma('vm:prefer-inline')
  @override
  $Res call({
    Object? id = null,
    Object? crystals = null,
    Object? bonus = null,
    Object? priceKopecks = null,
  }) {
    return _then(_$CrystalPackageImpl(
      id: null == id
          ? _value.id
          : id // ignore: cast_nullable_to_non_nullable
              as String,
      crystals: null == crystals
          ? _value.crystals
          : crystals // ignore: cast_nullable_to_non_nullable
              as int,
      bonus: null == bonus
          ? _value.bonus
          : bonus // ignore: cast_nullable_to_non_nullable
              as int,
      priceKopecks: null == priceKopecks
          ? _value.priceKopecks
          : priceKopecks // ignore: cast_nullable_to_non_nullable
              as int,
    ));
  }
}

/// @nodoc
@JsonSerializable()
class _$CrystalPackageImpl implements _CrystalPackage {
  const _$CrystalPackageImpl(
      {required this.id,
      required this.crystals,
      required this.bonus,
      @JsonKey(name: 'price_kopecks') required this.priceKopecks});

  factory _$CrystalPackageImpl.fromJson(Map<String, dynamic> json) =>
      _$$CrystalPackageImplFromJson(json);

  @override
  final String id;
  @override
  final int crystals;
  @override
  final int bonus;
  @override
  @JsonKey(name: 'price_kopecks')
  final int priceKopecks;

  @override
  String toString() {
    return 'CrystalPackage(id: $id, crystals: $crystals, bonus: $bonus, priceKopecks: $priceKopecks)';
  }

  @override
  bool operator ==(Object other) {
    return identical(this, other) ||
        (other.runtimeType == runtimeType &&
            other is _$CrystalPackageImpl &&
            (identical(other.id, id) || other.id == id) &&
            (identical(other.crystals, crystals) ||
                other.crystals == crystals) &&
            (identical(other.bonus, bonus) || other.bonus == bonus) &&
            (identical(other.priceKopecks, priceKopecks) ||
                other.priceKopecks == priceKopecks));
  }

  @JsonKey(ignore: true)
  @override
  int get hashCode =>
      Object.hash(runtimeType, id, crystals, bonus, priceKopecks);

  @JsonKey(ignore: true)
  @override
  @pragma('vm:prefer-inline')
  _$$CrystalPackageImplCopyWith<_$CrystalPackageImpl> get copyWith =>
      __$$CrystalPackageImplCopyWithImpl<_$CrystalPackageImpl>(
          this, _$identity);

  @override
  Map<String, dynamic> toJson() {
    return _$$CrystalPackageImplToJson(
      this,
    );
  }
}

abstract class _CrystalPackage implements CrystalPackage {
  const factory _CrystalPackage(
          {required final String id,
          required final int crystals,
          required final int bonus,
          @JsonKey(name: 'price_kopecks') required final int priceKopecks}) =
      _$CrystalPackageImpl;

  factory _CrystalPackage.fromJson(Map<String, dynamic> json) =
      _$CrystalPackageImpl.fromJson;

  @override
  String get id;
  @override
  int get crystals;
  @override
  int get bonus;
  @override
  @JsonKey(name: 'price_kopecks')
  int get priceKopecks;
  @override
  @JsonKey(ignore: true)
  _$$CrystalPackageImplCopyWith<_$CrystalPackageImpl> get copyWith =>
      throw _privateConstructorUsedError;
}

InitPurchaseResult _$InitPurchaseResultFromJson(Map<String, dynamic> json) {
  return _InitPurchaseResult.fromJson(json);
}

/// @nodoc
mixin _$InitPurchaseResult {
  @JsonKey(name: 'payment_url')
  String get paymentUrl => throw _privateConstructorUsedError;
  @JsonKey(name: 'payment_id')
  String get paymentId => throw _privateConstructorUsedError;

  Map<String, dynamic> toJson() => throw _privateConstructorUsedError;
  @JsonKey(ignore: true)
  $InitPurchaseResultCopyWith<InitPurchaseResult> get copyWith =>
      throw _privateConstructorUsedError;
}

/// @nodoc
abstract class $InitPurchaseResultCopyWith<$Res> {
  factory $InitPurchaseResultCopyWith(
          InitPurchaseResult value, $Res Function(InitPurchaseResult) then) =
      _$InitPurchaseResultCopyWithImpl<$Res, InitPurchaseResult>;
  @useResult
  $Res call(
      {@JsonKey(name: 'payment_url') String paymentUrl,
      @JsonKey(name: 'payment_id') String paymentId});
}

/// @nodoc
class _$InitPurchaseResultCopyWithImpl<$Res, $Val extends InitPurchaseResult>
    implements $InitPurchaseResultCopyWith<$Res> {
  _$InitPurchaseResultCopyWithImpl(this._value, this._then);

  // ignore: unused_field
  final $Val _value;
  // ignore: unused_field
  final $Res Function($Val) _then;

  @pragma('vm:prefer-inline')
  @override
  $Res call({
    Object? paymentUrl = null,
    Object? paymentId = null,
  }) {
    return _then(_value.copyWith(
      paymentUrl: null == paymentUrl
          ? _value.paymentUrl
          : paymentUrl // ignore: cast_nullable_to_non_nullable
              as String,
      paymentId: null == paymentId
          ? _value.paymentId
          : paymentId // ignore: cast_nullable_to_non_nullable
              as String,
    ) as $Val);
  }
}

/// @nodoc
abstract class _$$InitPurchaseResultImplCopyWith<$Res>
    implements $InitPurchaseResultCopyWith<$Res> {
  factory _$$InitPurchaseResultImplCopyWith(_$InitPurchaseResultImpl value,
          $Res Function(_$InitPurchaseResultImpl) then) =
      __$$InitPurchaseResultImplCopyWithImpl<$Res>;
  @override
  @useResult
  $Res call(
      {@JsonKey(name: 'payment_url') String paymentUrl,
      @JsonKey(name: 'payment_id') String paymentId});
}

/// @nodoc
class __$$InitPurchaseResultImplCopyWithImpl<$Res>
    extends _$InitPurchaseResultCopyWithImpl<$Res, _$InitPurchaseResultImpl>
    implements _$$InitPurchaseResultImplCopyWith<$Res> {
  __$$InitPurchaseResultImplCopyWithImpl(_$InitPurchaseResultImpl _value,
      $Res Function(_$InitPurchaseResultImpl) _then)
      : super(_value, _then);

  @pragma('vm:prefer-inline')
  @override
  $Res call({
    Object? paymentUrl = null,
    Object? paymentId = null,
  }) {
    return _then(_$InitPurchaseResultImpl(
      paymentUrl: null == paymentUrl
          ? _value.paymentUrl
          : paymentUrl // ignore: cast_nullable_to_non_nullable
              as String,
      paymentId: null == paymentId
          ? _value.paymentId
          : paymentId // ignore: cast_nullable_to_non_nullable
              as String,
    ));
  }
}

/// @nodoc
@JsonSerializable()
class _$InitPurchaseResultImpl implements _InitPurchaseResult {
  const _$InitPurchaseResultImpl(
      {@JsonKey(name: 'payment_url') required this.paymentUrl,
      @JsonKey(name: 'payment_id') required this.paymentId});

  factory _$InitPurchaseResultImpl.fromJson(Map<String, dynamic> json) =>
      _$$InitPurchaseResultImplFromJson(json);

  @override
  @JsonKey(name: 'payment_url')
  final String paymentUrl;
  @override
  @JsonKey(name: 'payment_id')
  final String paymentId;

  @override
  String toString() {
    return 'InitPurchaseResult(paymentUrl: $paymentUrl, paymentId: $paymentId)';
  }

  @override
  bool operator ==(Object other) {
    return identical(this, other) ||
        (other.runtimeType == runtimeType &&
            other is _$InitPurchaseResultImpl &&
            (identical(other.paymentUrl, paymentUrl) ||
                other.paymentUrl == paymentUrl) &&
            (identical(other.paymentId, paymentId) ||
                other.paymentId == paymentId));
  }

  @JsonKey(ignore: true)
  @override
  int get hashCode => Object.hash(runtimeType, paymentUrl, paymentId);

  @JsonKey(ignore: true)
  @override
  @pragma('vm:prefer-inline')
  _$$InitPurchaseResultImplCopyWith<_$InitPurchaseResultImpl> get copyWith =>
      __$$InitPurchaseResultImplCopyWithImpl<_$InitPurchaseResultImpl>(
          this, _$identity);

  @override
  Map<String, dynamic> toJson() {
    return _$$InitPurchaseResultImplToJson(
      this,
    );
  }
}

abstract class _InitPurchaseResult implements InitPurchaseResult {
  const factory _InitPurchaseResult(
          {@JsonKey(name: 'payment_url') required final String paymentUrl,
          @JsonKey(name: 'payment_id') required final String paymentId}) =
      _$InitPurchaseResultImpl;

  factory _InitPurchaseResult.fromJson(Map<String, dynamic> json) =
      _$InitPurchaseResultImpl.fromJson;

  @override
  @JsonKey(name: 'payment_url')
  String get paymentUrl;
  @override
  @JsonKey(name: 'payment_id')
  String get paymentId;
  @override
  @JsonKey(ignore: true)
  _$$InitPurchaseResultImplCopyWith<_$InitPurchaseResultImpl> get copyWith =>
      throw _privateConstructorUsedError;
}

VerifyResult _$VerifyResultFromJson(Map<String, dynamic> json) {
  return _VerifyResult.fromJson(json);
}

/// @nodoc
mixin _$VerifyResult {
  String get status => throw _privateConstructorUsedError;
  @JsonKey(name: 'new_balance')
  int? get newBalance => throw _privateConstructorUsedError;

  Map<String, dynamic> toJson() => throw _privateConstructorUsedError;
  @JsonKey(ignore: true)
  $VerifyResultCopyWith<VerifyResult> get copyWith =>
      throw _privateConstructorUsedError;
}

/// @nodoc
abstract class $VerifyResultCopyWith<$Res> {
  factory $VerifyResultCopyWith(
          VerifyResult value, $Res Function(VerifyResult) then) =
      _$VerifyResultCopyWithImpl<$Res, VerifyResult>;
  @useResult
  $Res call({String status, @JsonKey(name: 'new_balance') int? newBalance});
}

/// @nodoc
class _$VerifyResultCopyWithImpl<$Res, $Val extends VerifyResult>
    implements $VerifyResultCopyWith<$Res> {
  _$VerifyResultCopyWithImpl(this._value, this._then);

  // ignore: unused_field
  final $Val _value;
  // ignore: unused_field
  final $Res Function($Val) _then;

  @pragma('vm:prefer-inline')
  @override
  $Res call({
    Object? status = null,
    Object? newBalance = freezed,
  }) {
    return _then(_value.copyWith(
      status: null == status
          ? _value.status
          : status // ignore: cast_nullable_to_non_nullable
              as String,
      newBalance: freezed == newBalance
          ? _value.newBalance
          : newBalance // ignore: cast_nullable_to_non_nullable
              as int?,
    ) as $Val);
  }
}

/// @nodoc
abstract class _$$VerifyResultImplCopyWith<$Res>
    implements $VerifyResultCopyWith<$Res> {
  factory _$$VerifyResultImplCopyWith(
          _$VerifyResultImpl value, $Res Function(_$VerifyResultImpl) then) =
      __$$VerifyResultImplCopyWithImpl<$Res>;
  @override
  @useResult
  $Res call({String status, @JsonKey(name: 'new_balance') int? newBalance});
}

/// @nodoc
class __$$VerifyResultImplCopyWithImpl<$Res>
    extends _$VerifyResultCopyWithImpl<$Res, _$VerifyResultImpl>
    implements _$$VerifyResultImplCopyWith<$Res> {
  __$$VerifyResultImplCopyWithImpl(
      _$VerifyResultImpl _value, $Res Function(_$VerifyResultImpl) _then)
      : super(_value, _then);

  @pragma('vm:prefer-inline')
  @override
  $Res call({
    Object? status = null,
    Object? newBalance = freezed,
  }) {
    return _then(_$VerifyResultImpl(
      status: null == status
          ? _value.status
          : status // ignore: cast_nullable_to_non_nullable
              as String,
      newBalance: freezed == newBalance
          ? _value.newBalance
          : newBalance // ignore: cast_nullable_to_non_nullable
              as int?,
    ));
  }
}

/// @nodoc
@JsonSerializable()
class _$VerifyResultImpl implements _VerifyResult {
  const _$VerifyResultImpl(
      {required this.status, @JsonKey(name: 'new_balance') this.newBalance});

  factory _$VerifyResultImpl.fromJson(Map<String, dynamic> json) =>
      _$$VerifyResultImplFromJson(json);

  @override
  final String status;
  @override
  @JsonKey(name: 'new_balance')
  final int? newBalance;

  @override
  String toString() {
    return 'VerifyResult(status: $status, newBalance: $newBalance)';
  }

  @override
  bool operator ==(Object other) {
    return identical(this, other) ||
        (other.runtimeType == runtimeType &&
            other is _$VerifyResultImpl &&
            (identical(other.status, status) || other.status == status) &&
            (identical(other.newBalance, newBalance) ||
                other.newBalance == newBalance));
  }

  @JsonKey(ignore: true)
  @override
  int get hashCode => Object.hash(runtimeType, status, newBalance);

  @JsonKey(ignore: true)
  @override
  @pragma('vm:prefer-inline')
  _$$VerifyResultImplCopyWith<_$VerifyResultImpl> get copyWith =>
      __$$VerifyResultImplCopyWithImpl<_$VerifyResultImpl>(this, _$identity);

  @override
  Map<String, dynamic> toJson() {
    return _$$VerifyResultImplToJson(
      this,
    );
  }
}

abstract class _VerifyResult implements VerifyResult {
  const factory _VerifyResult(
          {required final String status,
          @JsonKey(name: 'new_balance') final int? newBalance}) =
      _$VerifyResultImpl;

  factory _VerifyResult.fromJson(Map<String, dynamic> json) =
      _$VerifyResultImpl.fromJson;

  @override
  String get status;
  @override
  @JsonKey(name: 'new_balance')
  int? get newBalance;
  @override
  @JsonKey(ignore: true)
  _$$VerifyResultImplCopyWith<_$VerifyResultImpl> get copyWith =>
      throw _privateConstructorUsedError;
}
