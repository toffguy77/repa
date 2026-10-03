import 'package:freezed_annotation/freezed_annotation.dart';

part 'crystals.freezed.dart';
part 'crystals.g.dart';

@freezed
class CrystalPackage with _$CrystalPackage {
  const factory CrystalPackage({
    required String id,
    required int crystals,
    required int bonus,
    @JsonKey(name: 'price_kopecks') required int priceKopecks,
  }) = _CrystalPackage;

  factory CrystalPackage.fromJson(Map<String, dynamic> json) =>
      _$CrystalPackageFromJson(json);
}

@freezed
class InitPurchaseResult with _$InitPurchaseResult {
  const factory InitPurchaseResult({
    @JsonKey(name: 'payment_url') required String paymentUrl,
    @JsonKey(name: 'payment_id') required String paymentId,
  }) = _InitPurchaseResult;

  factory InitPurchaseResult.fromJson(Map<String, dynamic> json) =>
      _$InitPurchaseResultFromJson(json);
}

@freezed
class VerifyResult with _$VerifyResult {
  const factory VerifyResult({
    required String status,
    @JsonKey(name: 'new_balance') int? newBalance,
  }) = _VerifyResult;

  factory VerifyResult.fromJson(Map<String, dynamic> json) =>
      _$VerifyResultFromJson(json);
}

/// One movement in the user's crystal balance.
///
/// [isGrant] separates free crystals from purchases, so a balance that grew without a payment is
/// not mysterious — see docs/features/crystals.md.
@freezed
class CrystalHistoryEntry with _$CrystalHistoryEntry {
  const factory CrystalHistoryEntry({
    required int delta,
    required String type,
    required String reason,
    @JsonKey(name: 'created_at') required String createdAt,
    @JsonKey(name: 'is_grant') @Default(false) bool isGrant,
  }) = _CrystalHistoryEntry;

  factory CrystalHistoryEntry.fromJson(Map<String, dynamic> json) =>
      _$CrystalHistoryEntryFromJson(json);
}
