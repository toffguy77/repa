import 'package:flutter/material.dart';
import 'package:flutter/services.dart';

import '../../domain/crystals.dart';
import '../../../../core/theme/app_tokens.dart';

class PackageCard extends StatelessWidget {
  final CrystalPackage package;
  final bool highlighted;
  final bool loading;
  final VoidCallback onTap;

  const PackageCard({
    super.key,
    required this.package,
    this.highlighted = false,
    this.loading = false,
    required this.onTap,
  });

  String get _priceText {
    final rubles = package.priceKopecks ~/ 100;
    return '$rubles \u20BD';
  }

  String get _crystalsText {
    if (package.bonus > 0) {
      return '${package.crystals} + ${package.bonus}';
    }
    return '${package.crystals}';
  }

  @override
  Widget build(BuildContext context) {
    return GestureDetector(
      onTap: loading
          ? null
          : () {
              HapticFeedback.mediumImpact();
              onTap();
            },
      child: Container(
        padding: const EdgeInsets.all(16),
        decoration: BoxDecoration(
          color: context.t.elevation.level1.surface,
          borderRadius: BorderRadius.circular(AppTokens.radius.lg),
          border: Border.all(
            color: highlighted ? context.t.color.accent : context.t.color.surface,
            width: highlighted ? 2 : 1,
          ),
          boxShadow: [
            if (highlighted)
              BoxShadow(
                color: context.t.color.accent.withValues(alpha: 0.15),
                blurRadius: 12,
                offset: const Offset(0, 4),
              ),
          ],
        ),
        child: Column(
          children: [
            if (highlighted)
              Container(
                padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 4),
                margin: const EdgeInsets.only(bottom: 8),
                decoration: BoxDecoration(
                  color: context.t.color.accentFill.withValues(alpha: 0.18),
                  borderRadius: BorderRadius.circular(AppTokens.radius.sm),
                ),
                child: Text(
                  'Популярный',
                  style: context.ts.caption.copyWith(
                    color: context.t.color.accent,
                    fontWeight: FontWeight.w600,
                  ),
                ),
              ),
            Row(
              children: [
                const Text('\u{1F48E}', style: TextStyle(fontSize: 32)),
                const SizedBox(width: 12),
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        _crystalsText,
                        style: context.ts.heading2,
                      ),
                      if (package.bonus > 0)
                        Text(
                          '+${package.bonus} бонус',
                          style: context.ts.caption.copyWith(
                            color: context.t.color.success,
                            fontWeight: FontWeight.w500,
                          ),
                        ),
                    ],
                  ),
                ),
                Container(
                  padding:
                      const EdgeInsets.symmetric(horizontal: 16, vertical: 10),
                  decoration: BoxDecoration(
                    color: highlighted ? context.t.color.accent : context.t.color.surface,
                    borderRadius: BorderRadius.circular(AppTokens.radius.md),
                  ),
                  child: Text(
                    _priceText,
                    style: context.ts.body.copyWith(
                      fontWeight: FontWeight.w600,
                      color: highlighted ? context.t.color.onAccentFill : context.t.color.textPrimary,
                    ),
                  ),
                ),
              ],
            ),
          ],
        ),
      ),
    );
  }
}
