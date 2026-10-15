import '../../../core/api/api_service.dart';
import '../domain/crystals.dart';

class CrystalsRepository {
  final ApiService _api;

  CrystalsRepository(this._api);

  Future<int> getBalance() async {
    final response = await _api.getCrystalBalance();
    return (response['data']['balance'] as num).toInt();
  }

  Future<List<CrystalHistoryEntry>> getHistory() async {
    final response = await _api.getCrystalHistory();
    final list = response['data']['entries'] as List;
    return list
        .map((e) => CrystalHistoryEntry.fromJson(e as Map<String, dynamic>))
        .toList();
  }

  Future<List<CrystalPackage>> getPackages() async {
    final response = await _api.getCrystalPackages();
    final list = response['data']['packages'] as List;
    return list
        .map((e) => CrystalPackage.fromJson(e as Map<String, dynamic>))
        .toList();
  }

  Future<InitPurchaseResult> initPurchase(String packageId) async {
    final response = await _api.initCrystalPurchase({'package_id': packageId});
    return InitPurchaseResult.fromJson(
        response['data'] as Map<String, dynamic>);
  }

  Future<VerifyResult> verifyPurchase(String paymentId) async {
    final response = await _api.verifyCrystalPurchase(paymentId);
    return VerifyResult.fromJson(response['data'] as Map<String, dynamic>);
  }
}
