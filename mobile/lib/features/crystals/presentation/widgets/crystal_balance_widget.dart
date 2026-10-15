import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../crystals_notifier.dart';
import '../../../../core/theme/app_tokens.dart';

class CrystalBalanceWidget extends ConsumerStatefulWidget {
  const CrystalBalanceWidget({super.key});

  @override
  ConsumerState<CrystalBalanceWidget> createState() =>
      _CrystalBalanceWidgetState();
}

class _CrystalBalanceWidgetState extends ConsumerState<CrystalBalanceWidget> {
  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((_) {
      ref.read(crystalBalanceProvider.notifier).load();
    });
  }

  @override
  Widget build(BuildContext context) {
    final balance = ref.watch(crystalBalanceProvider);

    return Semantics(
      label: 'Баланс кристаллов: $balance. Открыть магазин',
      button: true,
      child: GestureDetector(
      onTap: () => context.push('/shop'),
      child: Container(
        padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 6),
        decoration: BoxDecoration(
          color: context.t.color.accentFill.withValues(alpha: 0.18),
          borderRadius: BorderRadius.circular(AppTokens.radius.lg),
        ),
        child: Row(
          mainAxisSize: MainAxisSize.min,
          children: [
            const Text('\u{1F48E}', style: TextStyle(fontSize: 16)),
            const SizedBox(width: 4),
            Text(
              '$balance',
              style: TextStyle(
                fontSize: 14,
                fontWeight: FontWeight.w600,
                color: context.t.color.accent,
              ),
            ),
          ],
        ),
      ),
    ),
    );
  }
}
