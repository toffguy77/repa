import 'package:flutter/material.dart';

import '../../../../core/theme/app_tokens.dart';
import '../../../../core/widgets/kit/kit.dart';
import '../../domain/crystals.dart';

/// Where the balance came from.
///
/// Exists because crystals can now arrive without a payment (welcome, referral, achievements —
/// see docs/features/crystals.md); a balance that grew on its own should not be mysterious.
class CrystalHistoryList extends StatelessWidget {
  const CrystalHistoryList({super.key, required this.entries});

  final List<CrystalHistoryEntry> entries;

  @override
  Widget build(BuildContext context) {
    if (entries.isEmpty) return const SizedBox.shrink();

    final t = context.t;

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        const AppSectionHeader(title: 'Откуда кристаллы'),
        AppCard(
          padding: EdgeInsets.symmetric(vertical: AppTokens.space.sm),
          child: Column(
            children: [
              for (final entry in entries)
                AppListTile(
                  key: ValueKey('${entry.createdAt}-${entry.delta}'),
                  title: entry.reason.isEmpty
                      ? (entry.isGrant ? 'Подарок' : 'Покупка')
                      : entry.reason,
                  subtitle: entry.isGrant ? 'Бесплатно' : 'Покупка',
                  leading: Icon(
                    entry.isGrant ? Icons.card_giftcard : Icons.shopping_bag_outlined,
                    size: 20,
                    color: entry.isGrant ? t.color.energy : t.color.textSecondary,
                  ),
                  trailing: Text(
                    entry.delta >= 0 ? '+${entry.delta}' : '${entry.delta}',
                    style: AppTokens.text.numeric.copyWith(
                      color: entry.delta >= 0 ? t.color.success : t.color.textSecondary,
                    ),
                  ),
                ),
            ],
          ),
        ),
      ],
    );
  }
}
