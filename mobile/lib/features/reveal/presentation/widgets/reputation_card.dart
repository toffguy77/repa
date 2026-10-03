import 'package:flutter/material.dart';
import '../../domain/reveal.dart';
import 'attribute_bar.dart';
import '../../../../core/theme/app_tokens.dart';

class ReputationCard extends StatelessWidget {
  final MyCard card;
  final VoidCallback? onOpenHidden;
  final bool unlockingHidden;

  const ReputationCard({
    super.key,
    required this.card,
    this.onOpenHidden,
    this.unlockingHidden = false,
  });

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.all(24),
      decoration: BoxDecoration(
        color: context.t.elevation.level1.surface,
        borderRadius: BorderRadius.circular(AppTokens.radius.lg),
        boxShadow: [
          BoxShadow(
            color: context.t.color.accent.withValues(alpha: 0.1),
            blurRadius: 20,
            offset: const Offset(0, 8),
          ),
        ],
      ),
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          // Title
          Text(
            card.reputationTitle,
            style: context.ts.heading1.copyWith(
              color: context.t.color.accent,
              fontSize: 24,
            ),
            textAlign: TextAlign.center,
          ),
          const SizedBox(height: 4),
          if (card.trend != null) _buildTrend(context, card.trend!),
          const SizedBox(height: 24),

          // Top attributes
          ...List.generate(card.topAttributes.length, (i) {
            final attr = card.topAttributes[i];
            return Padding(
              padding: const EdgeInsets.only(bottom: 16),
              child: AttributeBar(
                questionText: attr.questionText,
                percentage: attr.percentage,
                index: i,
              ),
            );
          }),

          // Hidden attributes
          if (card.hiddenAttributes.isNotEmpty) ...[
            const SizedBox(height: 8),
            _buildHiddenSection(context),
          ],
        ],
      ),
    );
  }

  Widget _buildTrend(BuildContext context, TrendDto trend) {
    final icon = trend.change == 'up'
        ? Icons.arrow_upward
        : trend.change == 'down'
            ? Icons.arrow_downward
            : Icons.remove;
    final color = trend.change == 'up'
        ? context.t.color.success
        : trend.change == 'down'
            ? context.t.color.danger
            : context.t.color.textSecondary;

    return Row(
      mainAxisSize: MainAxisSize.min,
      children: [
        Icon(icon, size: 16, color: color),
        const SizedBox(width: 4),
        Text(
          '${trend.delta.abs().toStringAsFixed(1)}%',
          style: context.ts.caption.copyWith(color: color),
        ),
      ],
    );
  }

  Widget _buildHiddenSection(BuildContext context) {
    return Column(
      children: [
        // Blurred placeholder bars
        ...card.hiddenAttributes.take(2).map((_) {
          return Padding(
            padding: const EdgeInsets.only(bottom: 8),
            child: ClipRRect(
              borderRadius: BorderRadius.circular(AppTokens.radius.xs),
              child: Container(
                height: 36,
                decoration: BoxDecoration(
                  color: context.t.color.surface,
                  borderRadius: BorderRadius.circular(AppTokens.radius.xs),
                ),
                child: Center(
                  child: Text(
                    '???',
                    style: context.ts.caption.copyWith(
                      color: context.t.color.textSecondary,
                    ),
                  ),
                ),
              ),
            ),
          );
        }),
        const SizedBox(height: 8),
        SizedBox(
          width: double.infinity,
          child: OutlinedButton.icon(
            onPressed: unlockingHidden ? null : onOpenHidden,
            icon: unlockingHidden
                ? const SizedBox(
                    width: 16,
                    height: 16,
                    child: CircularProgressIndicator(strokeWidth: 2),
                  )
                : const Text('\u{1F48E}'),
            label: const Text('Открыть скрытые (5)'),
            style: OutlinedButton.styleFrom(
              foregroundColor: context.t.color.accent,
              side: BorderSide(color: context.t.color.accent),
              shape: RoundedRectangleBorder(
                borderRadius: BorderRadius.circular(AppTokens.radius.md),
              ),
              padding: const EdgeInsets.symmetric(vertical: 12),
            ),
          ),
        ),
      ],
    );
  }
}
