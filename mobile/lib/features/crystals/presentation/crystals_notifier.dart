import 'dart:async';

import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:url_launcher/url_launcher.dart';

import '../../../core/api/api_client.dart';
import '../../../core/providers/api_provider.dart';
import '../data/crystals_repository.dart';
import '../domain/crystals.dart';

class CrystalsState {
  final int balance;
  final List<CrystalPackage> packages;
  final bool loading;
  final bool purchasing;
  final String? pendingPaymentId;
  final String? error;
  final int? purchasedAmount;

  const CrystalsState({
    this.balance = 0,
    this.packages = const [],
    this.loading = true,
    this.purchasing = false,
    this.pendingPaymentId,
    this.error,
    this.purchasedAmount,
  });

  CrystalsState copyWith({
    int? balance,
    List<CrystalPackage>? packages,
    bool? loading,
    bool? purchasing,
    String? pendingPaymentId,
    String? error,
    int? purchasedAmount,
    bool clearPendingPayment = false,
    bool clearError = false,
    bool clearPurchasedAmount = false,
  }) {
    return CrystalsState(
      balance: balance ?? this.balance,
      packages: packages ?? this.packages,
      loading: loading ?? this.loading,
      purchasing: purchasing ?? this.purchasing,
      pendingPaymentId:
          clearPendingPayment ? null : (pendingPaymentId ?? this.pendingPaymentId),
      error: clearError ? null : (error ?? this.error),
      purchasedAmount:
          clearPurchasedAmount ? null : (purchasedAmount ?? this.purchasedAmount),
    );
  }
}

class CrystalsNotifier extends StateNotifier<CrystalsState> {
  final CrystalsRepository _repo;
  Timer? _pollingTimer;
  int _pollAttempts = 0;

  CrystalsNotifier(this._repo) : super(const CrystalsState());

  Future<void> load() async {
    state = state.copyWith(loading: true, clearError: true);
    try {
      final results = await Future.wait([
        _repo.getBalance(),
        _repo.getPackages(),
      ]);
      state = state.copyWith(
        balance: results[0] as int,
        packages: results[1] as List<CrystalPackage>,
        loading: false,
      );
    } on AppException catch (e) {
      state = state.copyWith(loading: false, error: e.message);
    } catch (_) {
      state = state.copyWith(loading: false, error: 'Не удалось загрузить данные');
    }
  }

  Future<void> refreshBalance() async {
    try {
      final balance = await _repo.getBalance();
      state = state.copyWith(balance: balance);
    } catch (_) {}
  }

  Future<void> initPurchase(String packageId) async {
    if (state.purchasing) return;
    state = state.copyWith(
      purchasing: true,
      clearError: true,
      clearPurchasedAmount: true,
    );

    try {
      final result = await _repo.initPurchase(packageId);
      state = state.copyWith(
        pendingPaymentId: result.paymentId,
        purchasing: false,
      );

      final uri = Uri.parse(result.paymentUrl);
      await launchUrl(uri, mode: LaunchMode.externalApplication);
    } on AppException catch (e) {
      state = state.copyWith(purchasing: false, error: e.message);
    } catch (_) {
      state = state.copyWith(purchasing: false, error: 'Не удалось начать покупку');
    }
  }

  void startPolling() {
    final paymentId = state.pendingPaymentId;
    if (paymentId == null) return;

    _pollAttempts = 0;
    _pollingTimer?.cancel();
    _pollingTimer = Timer.periodic(const Duration(seconds: 3), (_) {
      _pollAttempts++;
      if (_pollAttempts > 10) {
        _pollingTimer?.cancel();
        state = state.copyWith(
          error: 'Не удалось подтвердить оплату. Попробуйте позже.',
          clearPendingPayment: true,
        );
        return;
      }
      _verifyPayment(paymentId);
    });
  }

  Future<void> verifyOnReturn(String paymentId) async {
    state = state.copyWith(pendingPaymentId: paymentId);
    startPolling();
  }

  Future<void> _verifyPayment(String paymentId) async {
    try {
      final result = await _repo.verifyPurchase(paymentId);

      if (result.status == 'succeeded') {
        _pollingTimer?.cancel();
        final addedAmount = (result.newBalance ?? state.balance) - state.balance;
        state = state.copyWith(
          balance: result.newBalance ?? state.balance,
          purchasedAmount: addedAmount > 0 ? addedAmount : null,
          clearPendingPayment: true,
        );
      } else if (result.status == 'canceled') {
        _pollingTimer?.cancel();
        state = state.copyWith(
          error: 'Оплата отменена',
          clearPendingPayment: true,
        );
      }
      // pending — continue polling
    } catch (_) {
      // Network error during poll — keep trying
    }
  }

  void clearPurchaseSuccess() {
    state = state.copyWith(clearPurchasedAmount: true);
  }

  void clearError() {
    state = state.copyWith(clearError: true);
  }

  @override
  void dispose() {
    _pollingTimer?.cancel();
    super.dispose();
  }
}

final crystalsProvider =
    StateNotifierProvider.autoDispose<CrystalsNotifier, CrystalsState>((ref) {
  final api = ref.watch(apiServiceProvider);
  final repo = CrystalsRepository(api);
  return CrystalsNotifier(repo);
});

// Global balance provider that can be used across the app
final crystalBalanceProvider = StateNotifierProvider<_BalanceNotifier, int>((ref) {
  final api = ref.watch(apiServiceProvider);
  final repo = CrystalsRepository(api);
  return _BalanceNotifier(repo);
});

class _BalanceNotifier extends StateNotifier<int> {
  final CrystalsRepository _repo;

  _BalanceNotifier(this._repo) : super(0);

  Future<void> load() async {
    try {
      state = await _repo.getBalance();
    } catch (_) {}
  }

  void set(int value) {
    state = value;
  }
}
