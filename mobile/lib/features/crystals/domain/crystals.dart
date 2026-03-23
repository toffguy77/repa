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
