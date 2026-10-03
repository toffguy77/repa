// GENERATED CODE - DO NOT MODIFY BY HAND

part of 'crystals.dart';

// **************************************************************************
// JsonSerializableGenerator
// **************************************************************************

_$CrystalPackageImpl _$$CrystalPackageImplFromJson(Map<String, dynamic> json) =>
    _$CrystalPackageImpl(
      id: json['id'] as String,
      crystals: (json['crystals'] as num).toInt(),
      bonus: (json['bonus'] as num).toInt(),
      priceKopecks: (json['price_kopecks'] as num).toInt(),
    );

Map<String, dynamic> _$$CrystalPackageImplToJson(
        _$CrystalPackageImpl instance) =>
    <String, dynamic>{
      'id': instance.id,
      'crystals': instance.crystals,
      'bonus': instance.bonus,
      'price_kopecks': instance.priceKopecks,
    };

_$InitPurchaseResultImpl _$$InitPurchaseResultImplFromJson(
        Map<String, dynamic> json) =>
    _$InitPurchaseResultImpl(
      paymentUrl: json['payment_url'] as String,
      paymentId: json['payment_id'] as String,
    );

Map<String, dynamic> _$$InitPurchaseResultImplToJson(
        _$InitPurchaseResultImpl instance) =>
    <String, dynamic>{
      'payment_url': instance.paymentUrl,
      'payment_id': instance.paymentId,
    };

_$VerifyResultImpl _$$VerifyResultImplFromJson(Map<String, dynamic> json) =>
    _$VerifyResultImpl(
      status: json['status'] as String,
      newBalance: (json['new_balance'] as num?)?.toInt(),
    );

Map<String, dynamic> _$$VerifyResultImplToJson(_$VerifyResultImpl instance) =>
    <String, dynamic>{
      'status': instance.status,
      'new_balance': instance.newBalance,
    };

_$CrystalHistoryEntryImpl _$$CrystalHistoryEntryImplFromJson(
        Map<String, dynamic> json) =>
    _$CrystalHistoryEntryImpl(
      delta: (json['delta'] as num).toInt(),
      type: json['type'] as String,
      reason: json['reason'] as String,
      createdAt: json['created_at'] as String,
      isGrant: json['is_grant'] as bool? ?? false,
    );

Map<String, dynamic> _$$CrystalHistoryEntryImplToJson(
        _$CrystalHistoryEntryImpl instance) =>
    <String, dynamic>{
      'delta': instance.delta,
      'type': instance.type,
      'reason': instance.reason,
      'created_at': instance.createdAt,
      'is_grant': instance.isGrant,
    };
