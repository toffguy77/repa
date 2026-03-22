import 'package:flutter/material.dart';
import '../../../../core/theme/app_colors.dart';
import '../../../../core/theme/app_text_styles.dart';
import '../../domain/reveal.dart';
import 'attribute_bar.dart';

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
        color: Colors.white,
        borderRadius: BorderRadius.circular(24),
        boxShadow: [
          BoxShadow(
            color: AppColors.primary.withValues(alpha: 0.1),
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
            style: AppTextStyles.headline1.copyWith(
              color: AppColors.primary,
              fontSize: 24,
            ),
            textAlign: TextAlign.center,
          ),
          const SizedBox(height: 4),
          if (card.trend != null) _buildTrend(card.trend!),
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

  Widget _buildTrend(TrendDto trend) {
    final icon = trend.change == 'up'
        ? Icons.arrow_upward
        : trend.change == 'down'
            ? Icons.arrow_downward
            : Icons.remove;
    final color = trend.change == 'up'
        ? AppColors.success
        : trend.change == 'down'
            ? AppColors.error
            : AppColors.textSecondary;

    return Row(
      mainAxisSize: MainAxisSize.min,
      children: [
        Icon(icon, size: 16, color: color),
        const SizedBox(width: 4),
        Text(
          '${trend.delta.abs().toStringAsFixed(1)}%',
          style: AppTextStyles.caption.copyWith(color: color),
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
              borderRadius: BorderRadius.circular(6),
              child: Container(
                height: 36,
                decoration: BoxDecoration(
                  color: AppColors.surface,
                  borderRadius: BorderRadius.circular(6),
                ),
                child: Center(
                  child: Text(
                    '???',
                    style: AppTextStyles.caption.copyWith(
                      color: AppColors.textSecondary,
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
              foregroundColor: AppColors.primary,
              side: const BorderSide(color: AppColors.primary),
              shape: RoundedRectangleBorder(
                borderRadius: BorderRadius.circular(12),
              ),
              padding: const EdgeInsets.symmetric(vertical: 12),
            ),
          ),
        ),
      ],
    );
  }
}
