import 'dart:async';

import 'package:app_links/app_links.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import 'crystals_notifier.dart';
import 'widgets/package_card.dart';
import 'widgets/payment_pending_sheet.dart';
import 'widgets/purchase_success_sheet.dart';
import '../../../core/theme/app_tokens.dart';
import 'widgets/crystal_history_list.dart';

class CrystalsShopScreen extends ConsumerStatefulWidget {
  const CrystalsShopScreen({super.key});

  @override
  ConsumerState<CrystalsShopScreen> createState() =>
      _CrystalsShopScreenState();
}

class _CrystalsShopScreenState extends ConsumerState<CrystalsShopScreen> {
  StreamSubscription<Uri>? _linkSub;

  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((_) {
      ref.read(crystalsProvider.notifier).load();
    });
    _listenDeeplinks();
  }

  void _listenDeeplinks() {
    final appLinks = AppLinks();
    _linkSub = appLinks.uriLinkStream.listen((uri) {
      if (!mounted) return;
      if (uri.path.contains('/payment/return')) {
        final paymentId = ref.read(crystalsProvider).pendingPaymentId;
        if (paymentId != null) {
          Navigator.of(context).popUntil(
              (route) => route.isFirst || route.settings.name == '/shop');
          ref.read(crystalsProvider.notifier).startPolling();
        }
      }
    });
  }

  @override
  void dispose() {
    _linkSub?.cancel();
    super.dispose();
  }

  void _handlePurchase(String packageId) {
    ref.read(crystalsProvider.notifier).initPurchase(packageId);
  }

  void _showPendingSheet() {
    showModalBottomSheet(
      context: context,
      isDismissible: false,
      backgroundColor: Colors.transparent,
      builder: (_) => PaymentPendingSheet(
        onConfirm: () {
          Navigator.of(context).pop();
          ref.read(crystalsProvider.notifier).startPolling();
        },
      ),
    );
  }

  void _showSuccessSheet(int amount, int newBalance) {
    // Update global balance
    ref.read(crystalBalanceProvider.notifier).set(newBalance);

    showModalBottomSheet(
      context: context,
      isDismissible: false,
      backgroundColor: Colors.transparent,
      builder: (_) => PurchaseSuccessSheet(
        amount: amount,
        newBalance: newBalance,
        onDone: () {
          Navigator.of(context).pop();
          ref.read(crystalsProvider.notifier).clearPurchaseSuccess();
          ref.read(crystalsProvider.notifier).refreshBalance();
        },
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    final state = ref.watch(crystalsProvider);

    ref.listen<CrystalsState>(crystalsProvider, (prev, next) {
      // Show pending sheet when payment ID is set
      if (prev?.pendingPaymentId == null && next.pendingPaymentId != null) {
        _showPendingSheet();
      }

      // Show success sheet
      if (prev?.purchasedAmount == null && next.purchasedAmount != null) {
        _showSuccessSheet(next.purchasedAmount!, next.balance);
      }

      // Show error snackbar
      if (prev?.error == null && next.error != null) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text(next.error!),
            backgroundColor: context.t.color.danger,
          ),
        );
        ref.read(crystalsProvider.notifier).clearError();
      }
    });

    return Scaffold(
      appBar: AppBar(
        title: const Text('Магазин'),
      ),
      body: state.loading
          ? const Center(child: CircularProgressIndicator())
          : ListView(
              padding: const EdgeInsets.all(16),
              children: [
                // Balance header
                Container(
                  padding: const EdgeInsets.all(20),
                  decoration: BoxDecoration(
                    gradient: LinearGradient(
                      colors: [
                        context.t.color.accent,
                        context.t.color.accent.withValues(alpha: 0.8),
                      ],
                      begin: Alignment.topLeft,
                      end: Alignment.bottomRight,
                    ),
                    borderRadius: BorderRadius.circular(AppTokens.radius.lg),
                  ),
                  child: Column(
                    children: [
                      const Text(
                        '\u{1F48E}',
                        style: TextStyle(fontSize: 40),
                      ),
                      const SizedBox(height: 8),
                      Text(
                        '${state.balance}',
                        style: AppTokens.text.display1.copyWith(
                          color: context.t.color.onAccentFill,
                          fontFeatures: AppTokens.text.numeric.fontFeatures,
                        ),
                      ),
                      Text(
                        'кристаллов',
                        style: AppTokens.text.label.copyWith(
                          color: context.t.color.onAccentFill.withValues(alpha: 0.85),
                        ),
                      ),
                    ],
                  ),
                ),
                const SizedBox(height: 24),
                Text('Пакеты кристаллов', style: context.ts.heading2),
                const SizedBox(height: 12),
                ...state.packages.map((pkg) {
                  return Padding(
                    padding: const EdgeInsets.only(bottom: 12),
                    child: PackageCard(
                      package: pkg,
                      highlighted: pkg.id == 'popular',
                      loading: state.purchasing,
                      onTap: () => _handlePurchase(pkg.id),
                    ),
                  );
                }),
                const SizedBox(height: 24),
                CrystalHistoryList(entries: state.history),
                const SizedBox(height: 16),
                Text(
                  'Оплата через внешний сервис ЮKassa.\n'
                  'Средства зачисляются автоматически.',
                  style: context.ts.caption,
                  textAlign: TextAlign.center,
                ),
                const SizedBox(height: 32),
              ],
            ),
    );
  }
}
